package agg_test

import (
	"testing"
	"time"

	"github.com/gocrud/mgo"
	"github.com/gocrud/mgo/agg"
)

// 测试用户结构
type TestUser struct {
	ID        mgo.ObjectID `bson:"_id,omitempty"`
	Name      string       `bson:"name"`
	City      string       `bson:"city"`
	Age       int          `bson:"age"`
	Status    string       `bson:"status"`
	CreatedAt time.Time    `bson:"created_at"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// 统计结果
type CityStats struct {
	City      string    `bson:"_id"`
	UserCount int       `bson:"user_count"`
	AvgAge    float64   `bson:"avg_age"`
	MaxAge    int       `bson:"max_age"`
	CreatedAt time.Time `bson:"created_at"`
}

const testURI = "mongodb://example:example@localhost:27017/test_agg_cursor?authSource=admin&directConnection=true"

func TestAggCursorPage(t *testing.T) {
	// 跳过集成测试
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, err := mgo.Open(testURI)
	if err != nil {
		t.Skip("MongoDB not available:", err)
	}
	defer db.Close()

	users := mgo.Model[TestUser](db)

	// 清理并插入测试数据
	if err := users.Find().Delete(); err != nil {
		t.Logf("清理数据失败: %v", err)
	}
	insertTestUsers(t, users)

	t.Run("基础游标分页", func(t *testing.T) {
		// 第一页
		page1, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			SortDesc("user_count").
			CursorPage("", 2)

		if err != nil {
			t.Fatalf("第一页查询失败: %v", err)
		}

		if len(page1.Items) == 0 {
			t.Fatal("第一页应该有数据")
		}

		if page1.NextCursor == "" && page1.HasMore {
			t.Error("有更多数据时应该有下一页游标")
		}

		if page1.PrevCursor != "" {
			t.Error("第一页不应该有上一页游标")
		}

		t.Logf("第一页: %d 个城市, HasMore: %v", len(page1.Items), page1.HasMore)

		// 第二页
		if page1.HasMore {
			page2, err := agg.Aggregate[CityStats](users).
				Match(mgo.M{"status": "active"}).
				GroupBy("$city").
				Count("user_count").
				Avg("avg_age", "$age").
				Max("max_age", "$age").
				Max("created_at", "$created_at").
				SortDesc("user_count").
				CursorPage(page1.NextCursor, 2)

			if err != nil {
				t.Fatalf("第二页查询失败: %v", err)
			}

			if page2.PrevCursor == "" {
				t.Error("第二页应该有上一页游标")
			}

			t.Logf("第二页: %d 个城市, HasMore: %v", len(page2.Items), page2.HasMore)
		}
	})

	t.Run("多字段排序游标分页", func(t *testing.T) {
		page, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			Sort(mgo.D{
				{Key: "user_count", Value: -1},
				{Key: "avg_age", Value: 1},
			}).
			CursorPage("", 3)

		if err != nil {
			t.Fatalf("多字段排序分页失败: %v", err)
		}

		if len(page.Items) == 0 {
			t.Fatal("应该有结果")
		}

		t.Logf("多字段排序: %d 个城市", len(page.Items))
		for _, item := range page.Items {
			t.Logf("  城市: %s, 用户数: %d, 平均年龄: %.1f", item.City, item.UserCount, item.AvgAge)
		}
	})

	t.Run("GroupStage直接调用CursorPage", func(t *testing.T) {
		page, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			CursorPage("", 5)

		if err != nil {
			t.Fatalf("GroupStage游标分页失败: %v", err)
		}

		if len(page.Items) == 0 {
			t.Fatal("应该有结果")
		}

		t.Logf("GroupStage分页: %d 个城市", len(page.Items))
	})

	t.Run("无效游标应返回第一页", func(t *testing.T) {
		page, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			SortDesc("user_count").
			CursorPage("invalid_cursor", 3)

		if err != nil {
			t.Fatalf("无效游标查询失败: %v", err)
		}

		if len(page.Items) == 0 {
			t.Fatal("即使游标无效，也应该返回第一页数据")
		}

		t.Logf("无效游标返回第一页: %d 个城市", len(page.Items))
	})

	t.Run("双向翻页测试", func(t *testing.T) {
		// 第一页
		page1, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			SortDesc("user_count").
			CursorPage("", 2)

		if err != nil {
			t.Fatalf("第一页失败: %v", err)
		}

		if !page1.HasMore {
			t.Skip("数据不足，跳过双向翻页测试")
		}

		// 第二页
		page2, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			SortDesc("user_count").
			CursorPage(page1.NextCursor, 2)

		if err != nil {
			t.Fatalf("第二页失败: %v", err)
		}

		// 返回第一页
		backToPage1, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			SortDesc("user_count").
			CursorPage(page2.PrevCursor, 2)

		if err != nil {
			t.Fatalf("返回第一页失败: %v", err)
		}

		// 验证数据一致性
		if len(backToPage1.Items) != len(page1.Items) {
			t.Errorf("返回第一页的数据数量不一致: got %d, want %d", len(backToPage1.Items), len(page1.Items))
		}

		t.Logf("双向翻页测试成功: 第一页 %d 项, 第二页 %d 项, 返回第一页 %d 项",
			len(page1.Items), len(page2.Items), len(backToPage1.Items))
	})

	t.Run("无排序默认按_id排序", func(t *testing.T) {
		page, err := agg.Aggregate[CityStats](users).
			Match(mgo.M{"status": "active"}).
			GroupBy("$city").
			Count("user_count").
			Avg("avg_age", "$age").
			Max("max_age", "$age").
			Max("created_at", "$created_at").
			// 不设置排序
			CursorPage("", 3)

		if err != nil {
			t.Fatalf("无排序查询失败: %v", err)
		}

		if len(page.Items) == 0 {
			t.Fatal("应该有结果")
		}

		t.Logf("无排序（默认_id降序）: %d 个城市", len(page.Items))
	})

	t.Run("遍历所有页", func(t *testing.T) {
		cursor := ""
		pageNum := 0
		totalItems := 0

		for {
			pageNum++
			page, err := agg.Aggregate[CityStats](users).
				Match(mgo.M{"status": "active"}).
				GroupBy("$city").
				Count("user_count").
				Avg("avg_age", "$age").
				Max("max_age", "$age").
				Max("created_at", "$created_at").
				SortDesc("user_count").
				CursorPage(cursor, 2)

			if err != nil {
				t.Fatalf("第 %d 页查询失败: %v", pageNum, err)
			}

			itemCount := len(page.Items)
			totalItems += itemCount
			t.Logf("第 %d 页: %d 个城市", pageNum, itemCount)

			if !page.HasMore {
				break
			}

			cursor = page.NextCursor

			// 防止无限循环
			if pageNum > 100 {
				t.Fatal("页数超过 100，可能出现无限循环")
			}
		}

		t.Logf("总共遍历 %d 页, %d 个城市", pageNum, totalItems)
	})
}

func insertTestUsers(t *testing.T, users *mgo.TypedCollection[TestUser]) {
	cities := []string{"北京", "上海", "深圳", "广州", "杭州"}
	now := time.Now()

	var testUsers []*TestUser
	for i, city := range cities {
		// 每个城市不同数量的用户
		count := 5 - i
		for j := 0; j < count; j++ {
			user := &TestUser{
				Name:      city + "用户" + string(rune('A'+j)),
				City:      city,
				Age:       20 + (j * 10),
				Status:    "active",
				CreatedAt: now.Add(-time.Duration(i*j) * time.Hour),
			}
			testUsers = append(testUsers, user)
		}
	}

	// 逐个插入
	for _, user := range testUsers {
		if _, err := users.Insert(user); err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}
}
