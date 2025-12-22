package mgo_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gocrud/mgo"
)

const testURI = "mongodb://example:example@localhost:27017/mgo_test?authSource=admin&directConnection=true"

// TestUser 测试用户模型
type TestUser struct {
	ID        mgo.ObjectID `bson:"_id,omitempty"`
	Name      string       `bson:"name"`
	Email     string       `bson:"email"`
	Age       int          `bson:"age"`
	Status    string       `bson:"status"`
	Balance   float64      `bson:"balance"`
	Tags      []string     `bson:"tags"`
	CreatedAt time.Time    `bson:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at"`
	DeletedAt *time.Time   `bson:"deleted_at,omitempty"`
}

func (TestUser) CollName() string {
	return "test_users"
}

func setupTestDB(t *testing.T) *mgo.Database {
	db, err := mgo.Open(testURI)
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}
	return db
}

func TestConnection(t *testing.T) {
	t.Run("Open", func(t *testing.T) {
		db, err := mgo.Open(testURI)
		if err != nil {
			t.Fatalf("Open 失败: %v", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			t.Fatalf("Ping 失败: %v", err)
		}
	})

	t.Run("MustOpen", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("MustOpen panic: %v", r)
			}
		}()

		db := mgo.MustOpen(testURI)
		defer db.Close()

		if err := db.Ping(); err != nil {
			t.Fatalf("Ping 失败: %v", err)
		}
	})

	t.Run("Connect", func(t *testing.T) {
		db, err := mgo.Connect(
			mgo.URI(testURI),
			mgo.MaxPoolSize(50),
		)
		if err != nil {
			t.Fatalf("Connect 失败: %v", err)
		}
		defer db.Close()
	})
}

func TestModelCreation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("AutoInferCollectionName", func(t *testing.T) {
		users := mgo.Model[TestUser](db)
		if users.Name() != "test_users" {
			t.Errorf("集合名推断错误，期望 'test_users'，得到 '%s'", users.Name())
		}
	})

	t.Run("ExplicitCollectionName", func(t *testing.T) {
		users := mgo.Model[TestUser](db, "custom_users")
		if users.Name() != "custom_users" {
			t.Errorf("显式集合名错误，期望 'custom_users'，得到 '%s'", users.Name())
		}
	})

	t.Run("WithTimestamps", func(t *testing.T) {
		users := mgo.Model[TestUser](db).WithTimestamps()
		if users.Options().Timestamps == nil || !users.Options().Timestamps.Enabled {
			t.Error("时间戳未启用")
		}
	})

	t.Run("WithSoftDelete", func(t *testing.T) {
		users := mgo.Model[TestUser](db).WithSoftDelete()
		if users.Options().SoftDelete == nil || !users.Options().SoftDelete.Enabled {
			t.Error("软删除未启用")
		}
	})
}

func TestInsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db).WithTimestamps()
	defer users.Truncate()

	t.Run("InsertOne", func(t *testing.T) {
		user := &TestUser{
			Name:   "张三",
			Email:  "zhangsan@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := users.Insert(user)
		if err != nil {
			t.Fatalf("插入失败: %v", err)
		}

		if id == mgo.NilObjectID {
			t.Error("ID 为空")
		}

		// 验证时间戳
		if user.CreatedAt.IsZero() {
			t.Error("CreatedAt 未设置")
		}
		if user.UpdatedAt.IsZero() {
			t.Error("UpdatedAt 未设置")
		}
	})

	t.Run("InsertMany", func(t *testing.T) {
		userList := []*TestUser{
			{Name: "李四", Email: "lisi@example.com", Age: 30},
			{Name: "王五", Email: "wangwu@example.com", Age: 22},
		}

		ids, err := users.InsertMany(userList...)
		if err != nil {
			t.Fatalf("批量插入失败: %v", err)
		}

		if len(ids) != 2 {
			t.Errorf("期望插入 2 条，实际 %d 条", len(ids))
		}
	})
}

func TestQuery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db)
	defer users.Truncate()

	// 准备测试数据
	testUsers := []*TestUser{
		{Name: "张三", Email: "zhang@test.com", Age: 25, Status: "active", Balance: 1000},
		{Name: "李四", Email: "li@test.com", Age: 30, Status: "active", Balance: 500},
		{Name: "王五", Email: "wang@test.com", Age: 22, Status: "pending", Balance: 200},
		{Name: "赵六", Email: "zhao@test.com", Age: 35, Status: "inactive", Balance: 800},
	}
	_, _ = users.InsertMany(testUsers...)

	t.Run("FindByID", func(t *testing.T) {
		id := testUsers[0].ID
		user, err := users.FindByID(id)
		if err != nil {
			t.Fatalf("FindByID 失败: %v", err)
		}

		if user.Name != "张三" {
			t.Errorf("期望名称 '张三'，得到 '%s'", user.Name)
		}
	})

	t.Run("WhereConditions", func(t *testing.T) {
		results, err := users.Find().
			Where("status", "active").
			Where("age", ">", 20).
			All()

		if err != nil {
			t.Fatalf("查询失败: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("期望 2 条结果，得到 %d 条", len(results))
		}
	})

	t.Run("Count", func(t *testing.T) {
		count, err := users.Find().Where("status", "active").Count()
		if err != nil {
			t.Fatalf("Count 失败: %v", err)
		}

		if count != 2 {
			t.Errorf("期望计数 2，得到 %d", count)
		}
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := users.Find().Where("email", "zhang@test.com").Exists()
		if err != nil {
			t.Fatalf("Exists 失败: %v", err)
		}

		if !exists {
			t.Error("期望存在，但返回不存在")
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		page, err := users.Find().PageList(1, 2)
		if err != nil {
			t.Fatalf("分页失败: %v", err)
		}

		if page.Total != 4 {
			t.Errorf("期望总数 4，得到 %d", page.Total)
		}

		if len(page.Items) != 2 {
			t.Errorf("期望每页 2 条，得到 %d 条", len(page.Items))
		}

		if page.Pages != 2 {
			t.Errorf("期望 2 页，得到 %d 页", page.Pages)
		}
	})
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db).WithTimestamps()
	defer users.Truncate()

	// 插入测试数据
	user := &TestUser{
		Name:    "测试用户",
		Email:   "test@example.com",
		Age:     25,
		Status:  "active",
		Balance: 1000,
		Tags:    []string{"tag1"},
	}
	id, _ := users.Insert(user)

	t.Run("Set", func(t *testing.T) {
		err := users.Find().ID(id).Set("status", "inactive").Update()
		if err != nil {
			t.Fatalf("Set 失败: %v", err)
		}

		updated, _ := users.FindByID(id)
		if updated.Status != "inactive" {
			t.Errorf("期望状态 'inactive'，得到 '%s'", updated.Status)
		}
	})

	t.Run("Inc", func(t *testing.T) {
		err := users.Find().ID(id).Inc("age", 5).Update()
		if err != nil {
			t.Fatalf("Inc 失败: %v", err)
		}

		updated, _ := users.FindByID(id)
		if updated.Age != 30 {
			t.Errorf("期望年龄 30，得到 %d", updated.Age)
		}
	})

	t.Run("Push", func(t *testing.T) {
		err := users.Find().ID(id).Push("tags", "tag2").Update()
		if err != nil {
			t.Fatalf("Push 失败: %v", err)
		}

		updated, _ := users.FindByID(id)
		if len(updated.Tags) != 2 {
			t.Errorf("期望 2 个标签，得到 %d 个", len(updated.Tags))
		}
	})

	t.Run("UpdateMany", func(t *testing.T) {
		n, err := users.Find().
			Where("status", "inactive").
			Set("status", "active").
			UpdateMany()

		if err != nil {
			t.Fatalf("UpdateMany 失败: %v", err)
		}

		if n < 1 {
			t.Errorf("期望至少更新 1 条，实际 %d 条", n)
		}
	})
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("HardDelete", func(t *testing.T) {
		users := mgo.Model[TestUser](db)
		defer users.Truncate()

		user := &TestUser{Name: "测试", Email: "test@test.com"}
		id, _ := users.Insert(user)

		err := users.Find().ID(id).Delete()
		if err != nil {
			t.Fatalf("Delete 失败: %v", err)
		}

		_, err = users.FindByID(id)
		if !mgo.IsNoDocuments(err) {
			t.Error("记录应该已被删除")
		}
	})

	t.Run("SoftDelete", func(t *testing.T) {
		users := mgo.Model[TestUser](db, "soft_delete_test").WithSoftDelete()
		defer users.Truncate()

		user := &TestUser{Name: "测试", Email: "test@test.com"}
		id, _ := users.Insert(user)

		// 软删除
		err := users.Find().ID(id).Delete()
		if err != nil {
			t.Fatalf("软删除失败: %v", err)
		}

		// 默认查询应该找不到
		_, err = users.FindByID(id)
		if err == nil {
			t.Error("软删除后默认查询不应该找到记录")
		}

		// WithTrashed 应该能找到
		deleted, err := users.Find().ID(id).WithTrashed().One()
		if err != nil {
			t.Fatalf("WithTrashed 查询失败: %v", err)
		}

		if deleted.DeletedAt == nil {
			t.Error("DeletedAt 应该已设置")
		}
	})

	t.Run("Restore", func(t *testing.T) {
		users := mgo.Model[TestUser](db, "restore_test").WithSoftDelete()
		defer users.Truncate()

		user := &TestUser{Name: "测试", Email: "test@test.com"}
		id, _ := users.Insert(user)

		// 删除
		users.Find().ID(id).Delete()

		// 恢复
		err := users.Find().ID(id).WithTrashed().Restore()
		if err != nil {
			t.Fatalf("Restore 失败: %v", err)
		}

		// 恢复后应该能正常查询到
		restored, err := users.FindByID(id)
		if err != nil {
			t.Fatalf("恢复后查询失败: %v", err)
		}

		if restored.DeletedAt != nil {
			t.Error("DeletedAt 应该为 nil")
		}
	})
}

func TestTimeQuery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "time_test").WithTimestamps()
	defer users.Truncate()

	// 插入测试数据
	for i := 0; i < 5; i++ {
		users.Insert(&TestUser{
			Name:  fmt.Sprintf("User%d", i),
			Email: fmt.Sprintf("user%d@test.com", i),
			Age:   20 + i,
		})
		time.Sleep(100 * time.Millisecond)
	}

	t.Run("WhereToday", func(t *testing.T) {
		results, err := users.Find().WhereToday("created_at").All()
		if err != nil {
			t.Fatalf("WhereToday 失败: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("期望 5 条今天的记录，得到 %d 条", len(results))
		}
	})

	t.Run("WhereLastHours", func(t *testing.T) {
		results, err := users.Find().WhereLastHours("created_at", 1).All()
		if err != nil {
			t.Fatalf("WhereLastHours 失败: %v", err)
		}

		if len(results) != 5 {
			t.Errorf("期望 5 条最近的记录，得到 %d 条", len(results))
		}
	})
}

func TestFilters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "filter_test")
	defer users.Truncate()

	// 准备测试数据
	users.InsertMany(
		&TestUser{Name: "张三", Age: 25, Status: "active", Tags: []string{"vip", "verified"}},
		&TestUser{Name: "李四", Age: 30, Status: "active", Tags: []string{"verified"}},
		&TestUser{Name: "王五", Age: 22, Status: "pending", Tags: []string{"new"}},
	)

	t.Run("Eq", func(t *testing.T) {
		results, err := users.Find().Filter(mgo.Eq("status", "active")).All()
		if err != nil {
			t.Fatalf("Eq 失败: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("期望 2 条，得到 %d 条", len(results))
		}
	})

	t.Run("Gt", func(t *testing.T) {
		results, err := users.Find().Filter(mgo.Gt("age", 25)).All()
		if err != nil {
			t.Fatalf("Gt 失败: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("期望 1 条，得到 %d 条", len(results))
		}
	})

	t.Run("In", func(t *testing.T) {
		results, err := users.Find().Filter(mgo.In("status", "active", "pending")).All()
		if err != nil {
			t.Fatalf("In 失败: %v", err)
		}
		if len(results) != 3 {
			t.Errorf("期望 3 条，得到 %d 条", len(results))
		}
	})

	t.Run("And", func(t *testing.T) {
		results, err := users.Find().Filter(
			mgo.And(
				mgo.Eq("status", "active"),
				mgo.Gt("age", 25),
			),
		).All()

		if err != nil {
			t.Fatalf("And 失败: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("期望 1 条，得到 %d 条", len(results))
		}
	})

	t.Run("Or", func(t *testing.T) {
		results, err := users.Find().Filter(
			mgo.Or(
				mgo.Eq("status", "pending"),
				mgo.Gt("age", 28),
			),
		).All()

		if err != nil {
			t.Fatalf("Or 失败: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("期望 2 条，得到 %d 条", len(results))
		}
	})
}

func TestPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "page_test")
	defer users.Truncate()

	// 插入 25 条数据
	testUsers := make([]*TestUser, 25)
	for i := 0; i < 25; i++ {
		testUsers[i] = &TestUser{
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    20 + i,
			Status: "active",
		}
	}
	_, _ = users.InsertMany(testUsers...)

	t.Run("StandardPagination", func(t *testing.T) {
		page, err := users.Find().PageList(1, 10)
		if err != nil {
			t.Fatalf("分页失败: %v", err)
		}

		if page.Total != 25 {
			t.Errorf("期望总数 25，得到 %d", page.Total)
		}

		if page.Pages != 3 {
			t.Errorf("期望 3 页，得到 %d 页", page.Pages)
		}

		if len(page.Items) != 10 {
			t.Errorf("期望每页 10 条，得到 %d 条", len(page.Items))
		}
	})

	t.Run("SimplePagination", func(t *testing.T) {
		page, err := users.Find().SimplePageList(2, 10)
		if err != nil {
			t.Fatalf("简化分页失败: %v", err)
		}

		if len(page.Items) != 10 {
			t.Errorf("期望 10 条，得到 %d 条", len(page.Items))
		}

		if !page.HasMore {
			t.Error("应该有更多页")
		}
	})
}

func TestTransaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "tx_test")
	defer users.Truncate()

	// 插入测试用户
	user := &TestUser{
		Name:    "测试用户",
		Email:   "tx@test.com",
		Balance: 1000,
	}
	userID, _ := users.Insert(user)

	t.Run("TransactionSuccess", func(t *testing.T) {
		err := db.Transaction(func(sess *mgo.Session) error {
			txUsers := mgo.Model[TestUser](sess, "tx_test")

			return txUsers.Find().ID(userID).Inc("balance", -100).Update()
		})

		if err != nil {
			t.Fatalf("事务失败: %v", err)
		}

		// 验证余额
		updated, _ := users.FindByID(userID)
		if updated.Balance != 900 {
			t.Errorf("期望余额 900，得到 %.2f", updated.Balance)
		}
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		initialBalance := user.Balance

		err := db.Transaction(func(sess *mgo.Session) error {
			txUsers := mgo.Model[TestUser](sess, "tx_test")

			if err := txUsers.Find().ID(userID).Inc("balance", -500).Update(); err != nil {
				return err
			}

			// 返回错误触发回滚
			return fmt.Errorf("模拟错误")
		})

		if err == nil {
			t.Fatal("期望事务失败")
		}

		// 验证余额未变化
		current, _ := users.FindByID(userID)
		if current.Balance == initialBalance-500 {
			t.Error("事务应该已回滚，余额不应该变化")
		}
	})
}

func TestHelpers(t *testing.T) {
	t.Run("ToSnakeCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"UserProfile", "user_profile"},
			{"TestUser", "test_user"},
			{"ID", "i_d"},
		}

		for _, tt := range tests {
			result := mgo.ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%s) = %s，期望 %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Pluralize", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"User", "users"},
			{"City", "cities"},
			{"Person", "people"},
		}

		for _, tt := range tests {
			result := mgo.Pluralize(tt.input)
			if result != tt.expected {
				t.Errorf("Pluralize(%s) = %s，期望 %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("ObjectID", func(t *testing.T) {
		id := mgo.NewObjectID()
		if id == mgo.NilObjectID {
			t.Error("NewObjectID 返回零值")
		}

		hex := id.Hex()
		parsed, err := mgo.ObjectIDFromHex(hex)
		if err != nil {
			t.Fatalf("ObjectIDFromHex 失败: %v", err)
		}

		if parsed != id {
			t.Error("ObjectID 转换不一致")
		}

		if !mgo.IsValidObjectID(hex) {
			t.Error("IsValidObjectID 返回 false")
		}
	})

	t.Run("ParseTime", func(t *testing.T) {
		tests := []string{
			"2024-01-01",
			"2024-01-01 15:04:05",
			"2024-01-01T15:04:05Z",
		}

		for _, dateStr := range tests {
			_, err := mgo.ParseTime(dateStr)
			if err != nil {
				t.Errorf("ParseTime(%s) 失败: %v", dateStr, err)
			}
		}
	})
}

func TestErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "error_test")
	defer users.Truncate()

	t.Run("IsNoDocuments", func(t *testing.T) {
		_, err := users.FindByID(mgo.NewObjectID())
		if !mgo.IsNoDocuments(err) {
			t.Error("应该返回 ErrNoDocuments")
		}
	})

	t.Run("OneOrNil", func(t *testing.T) {
		result := users.Find().Where("email", "nonexistent@test.com").OneOrNil()
		if result != nil {
			t.Error("OneOrNil 应该返回 nil")
		}
	})

	t.Run("AllOrEmpty", func(t *testing.T) {
		results := users.Find().Where("status", "nonexistent").AllOrEmpty()
		if results == nil {
			t.Error("AllOrEmpty 不应该返回 nil")
		}
		if len(results) != 0 {
			t.Errorf("AllOrEmpty 应该返回空切片，得到 %d 条", len(results))
		}
	})
}

func TestCursorPagination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := mgo.Model[TestUser](db, "cursor_page_test")
	defer users.Truncate()

	// 插入 50 条测试数据
	testUsers := make([]*TestUser, 50)
	for i := 0; i < 50; i++ {
		testUsers[i] = &TestUser{
			Name:   fmt.Sprintf("User%02d", i),
			Email:  fmt.Sprintf("user%02d@test.com", i),
			Age:    20 + i,
			Status: "active",
		}
	}
	_, _ = users.InsertMany(testUsers...)

	t.Run("FirstPage", func(t *testing.T) {
		page, err := users.Find().Desc("_id").CursorPage("", 10)
		if err != nil {
			t.Fatalf("游标分页失败: %v", err)
		}

		if len(page.Items) != 10 {
			t.Errorf("期望 10 条记录，得到 %d 条", len(page.Items))
		}

		if !page.HasMore {
			t.Error("应该有更多数据")
		}

		if page.NextCursor == "" {
			t.Error("NextCursor 不应为空")
		}

		if page.PrevCursor != "" {
			t.Error("第一页的 PrevCursor 应为空")
		}
	})

	t.Run("NextPage", func(t *testing.T) {
		// 获取第一页
		page1, err := users.Find().Desc("_id").CursorPage("", 10)
		if err != nil {
			t.Fatalf("第一页失败: %v", err)
		}

		// 使用游标获取第二页
		page2, err := users.Find().Desc("_id").CursorPage(page1.NextCursor, 10)
		if err != nil {
			t.Fatalf("第二页失败: %v", err)
		}

		if len(page2.Items) != 10 {
			t.Errorf("期望 10 条记录，得到 %d 条", len(page2.Items))
		}

		// 验证两页数据不重复
		if page1.Items[0].ID == page2.Items[0].ID {
			t.Error("第一页和第二页的数据不应重复")
		}

		if page2.PrevCursor == "" {
			t.Error("第二页应该有 PrevCursor")
		}
	})

	t.Run("PrevPage", func(t *testing.T) {
		// 先到第二页
		page1, _ := users.Find().Desc("_id").CursorPage("", 10)
		page2, _ := users.Find().Desc("_id").CursorPage(page1.NextCursor, 10)

		// 使用 PrevCursor 返回第一页
		pagePrev, err := users.Find().Desc("_id").CursorPage(page2.PrevCursor, 10)
		if err != nil {
			t.Fatalf("返回上一页失败: %v", err)
		}

		if len(pagePrev.Items) != 10 {
			t.Errorf("期望 10 条记录，得到 %d 条", len(pagePrev.Items))
		}

		// 验证返回的是第一页的数据
		if page1.Items[0].ID != pagePrev.Items[0].ID {
			t.Error("返回的数据应该与第一页相同")
		}
	})

	t.Run("LastPage", func(t *testing.T) {
		// 遍历到最后一页
		var cursor string
		var page *mgo.CursorPage[TestUser]
		var err error

		for {
			page, err = users.Find().Desc("_id").CursorPage(cursor, 10)
			if err != nil {
				t.Fatalf("分页失败: %v", err)
			}

			if !page.HasMore {
				break
			}

			cursor = page.NextCursor
		}

		// 最后一页应该没有下一页
		if page.HasMore {
			t.Error("最后一页不应该有更多数据")
		}

		if page.NextCursor != "" {
			t.Error("最后一页的 NextCursor 应为空")
		}
	})

	t.Run("CustomSort", func(t *testing.T) {
		// 按年龄升序
		page, err := users.Find().Asc("age").CursorPage("", 10)
		if err != nil {
			t.Fatalf("自定义排序失败: %v", err)
		}

		if len(page.Items) < 2 {
			t.Fatal("数据不足")
		}

		// 验证排序正确
		if page.Items[0].Age > page.Items[1].Age {
			t.Error("排序不正确，应该按年龄升序")
		}
	})

	t.Run("MultiFieldSort", func(t *testing.T) {
		// 多字段排序（已修复：使用 D 类型保证排序顺序）
		page, err := users.Find().
			Desc("status").
			Asc("age").
			CursorPage("", 10)

		if err != nil {
			t.Fatalf("多字段排序失败: %v", err)
		}

		if len(page.Items) != 10 {
			t.Errorf("期望 10 条记录，得到 %d 条", len(page.Items))
		}
	})

	t.Run("WithFilter", func(t *testing.T) {
		// 带过滤条件的游标分页
		page, err := users.Find().
			Where("age", ">", 30).
			Desc("age").
			CursorPage("", 10)

		if err != nil {
			t.Fatalf("带过滤条件失败: %v", err)
		}

		// 验证所有结果都符合条件
		for _, user := range page.Items {
			if user.Age <= 30 {
				t.Errorf("User %s 的年龄 %d 不符合条件", user.Name, user.Age)
			}
		}
	})

	t.Run("InvalidCursor", func(t *testing.T) {
		// 无效游标应该返回第一页（容错处理）
		page, err := users.Find().Desc("_id").CursorPage("invalid_cursor", 10)
		if err != nil {
			t.Fatalf("无效游标处理失败: %v", err)
		}

		// 应该返回数据，而不是报错
		if len(page.Items) == 0 {
			t.Error("无效游标应该返回第一页数据")
		}
	})

	t.Run("EmptyResult", func(t *testing.T) {
		// 查询不存在的数据
		page, err := users.Find().
			Where("status", "nonexistent").
			CursorPage("", 10)

		if err != nil {
			t.Fatalf("空结果查询失败: %v", err)
		}

		if len(page.Items) != 0 {
			t.Error("应该返回空结果")
		}

		if page.HasMore {
			t.Error("空结果不应该有更多数据")
		}

		if page.NextCursor != "" {
			t.Error("空结果的 NextCursor 应为空")
		}
	})

	t.Run("PageSizeLimit", func(t *testing.T) {
		// 测试每页数量限制
		page, err := users.Find().CursorPage("", 2000) // 超过最大限制
		if err != nil {
			t.Fatalf("分页失败: %v", err)
		}

		// 应该被限制为最大值（1000）
		if len(page.Items) > 1000 {
			t.Errorf("每页数量超过限制，期望最多 1000，得到 %d", len(page.Items))
		}
	})

	t.Run("Traversal", func(t *testing.T) {
		// 完整遍历所有数据
		var allItems []*TestUser
		var cursor string

		for {
			page, err := users.Find().Desc("_id").CursorPage(cursor, 10)
			if err != nil {
				t.Fatalf("遍历失败: %v", err)
			}

			allItems = append(allItems, page.Items...)

			if !page.HasMore {
				break
			}

			cursor = page.NextCursor
		}

		if len(allItems) != 50 {
			t.Errorf("期望遍历 50 条记录，实际 %d 条", len(allItems))
		}

		// 验证没有重复
		idSet := make(map[mgo.ObjectID]bool)
		for _, item := range allItems {
			if idSet[item.ID] {
				t.Errorf("发现重复记录: %s", item.ID.Hex())
			}
			idSet[item.ID] = true
		}
	})
}

func TestCrossDBTransaction(t *testing.T) {
	// 创建 Client 实例
	client, err := mgo.OpenClient(testURI)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer client.Close()

	// 准备测试数据库和集合
	db1 := client.Database("cross_db_test_1")
	db2 := client.Database("cross_db_test_2")
	defer db1.Drop()
	defer db2.Drop()

	users := mgo.Model[TestUser](db1, "users")
	logs := mgo.Model[TestUser](db2, "logs")

	t.Run("CrossDBTransactionSuccess", func(t *testing.T) {
		// 清理数据
		users.Truncate()
		logs.Truncate()

		// 插入测试用户
		user := &TestUser{
			Name:    "测试用户",
			Email:   "crossdb@test.com",
			Balance: 1000,
		}
		userID, _ := users.Insert(user)

		// 跨库事务
		err := client.Transaction(func(sess *mgo.ClientSession) error {
			usersDB := sess.Database("cross_db_test_1")
			logsDB := sess.Database("cross_db_test_2")

			txUsers := mgo.Model[TestUser](usersDB, "users")
			txLogs := mgo.Model[TestUser](logsDB, "logs")

			// 更新用户余额
			if err := txUsers.Find().ID(userID).Inc("balance", -100).Update(); err != nil {
				return err
			}

			// 记录日志
			log := &TestUser{
				Name:   "转账日志",
				Email:  "log@test.com",
				Status: "completed",
			}
			if _, err := txLogs.Insert(log); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			t.Fatalf("跨库事务失败: %v", err)
		}

		// 验证结果
		updated, _ := users.FindByID(userID)
		if updated.Balance != 900 {
			t.Errorf("期望余额 900，得到 %.2f", updated.Balance)
		}

		logCount, _ := logs.Find().Count()
		if logCount != 1 {
			t.Errorf("期望日志数 1，得到 %d", logCount)
		}
	})

	t.Run("CrossDBTransactionRollback", func(t *testing.T) {
		// 清理数据
		users.Truncate()
		logs.Truncate()

		user := &TestUser{
			Name:    "测试用户",
			Email:   "rollback@test.com",
			Balance: 1000,
		}
		userID, _ := users.Insert(user)

		// 跨库事务（应该回滚）
		err := client.Transaction(func(sess *mgo.ClientSession) error {
			usersDB := sess.Database("cross_db_test_1")
			logsDB := sess.Database("cross_db_test_2")

			txUsers := mgo.Model[TestUser](usersDB, "users")
			txLogs := mgo.Model[TestUser](logsDB, "logs")

			// 更新用户余额
			if err := txUsers.Find().ID(userID).Inc("balance", -100).Update(); err != nil {
				return err
			}

			// 记录日志
			log := &TestUser{
				Name:  "转账日志",
				Email: "log@test.com",
			}
			if _, err := txLogs.Insert(log); err != nil {
				return err
			}

			// 返回错误触发回滚
			return fmt.Errorf("模拟错误")
		})

		if err == nil {
			t.Fatal("期望事务失败")
		}

		// 验证回滚：余额应该保持不变
		current, _ := users.FindByID(userID)
		if current.Balance != 1000 {
			t.Errorf("事务应该已回滚，期望余额 1000，得到 %.2f", current.Balance)
		}

		// 验证回滚：日志应该没有被插入
		logCount, _ := logs.Find().Count()
		if logCount != 0 {
			t.Errorf("事务应该已回滚，期望日志数 0，得到 %d", logCount)
		}
	})

	t.Run("CrossDBWithMultipleDatabases", func(t *testing.T) {
		// 清理数据
		users.Truncate()
		logs.Truncate()

		// 使用 Databases 辅助方法
		err := client.Transaction(func(sess *mgo.ClientSession) error {
			dbs := sess.Databases("cross_db_test_1", "cross_db_test_2")
			db1 := dbs[0]
			db2 := dbs[1]

			txUsers := mgo.Model[TestUser](db1, "users")
			txLogs := mgo.Model[TestUser](db2, "logs")

			// 插入数据
			user := &TestUser{
				Name:   "多库用户",
				Email:  "multi@test.com",
				Status: "active",
			}
			if _, err := txUsers.Insert(user); err != nil {
				return err
			}

			log := &TestUser{
				Name:  "多库日志",
				Email: "multilog@test.com",
			}
			if _, err := txLogs.Insert(log); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			t.Fatalf("多库事务失败: %v", err)
		}

		// 验证数据
		userCount, _ := users.Find().Count()
		logCount, _ := logs.Find().Count()

		if userCount != 1 {
			t.Errorf("期望用户数 1，得到 %d", userCount)
		}
		if logCount != 1 {
			t.Errorf("期望日志数 1，得到 %d", logCount)
		}
	})
}
