package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// User 用户模型
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	Name      string        `bson:"name"`
	Email     string        `bson:"email"`
	Age       int           `bson:"age"`
	City      string        `bson:"city"`
	Status    string        `bson:"status"`
	Tags      []string      `bson:"tags,omitempty"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func main() {
	// MongoDB 连接配置
	uri := "mongodb://example:example@localhost:27017/?directConnection=true"
	ctx := context.Background()

	// 使用 mgo 连接 MongoDB
	client, err := mgo.Connect(ctx, uri)
	if err != nil {
		log.Fatal("❌ 连接 MongoDB 失败:", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal("❌ 断开连接失败:", err)
		}
	}()

	fmt.Println("✅ 成功连接到 MongoDB")
	fmt.Println()

	// 获取集合
	coll := client.Database("test_db").Collection("users")

	// 清空集合（为了演示）
	if err := coll.Drop(ctx); err != nil {
		fmt.Println("⚠️  清空集合失败（可能集合不存在）:", err)
	}

	fmt.Println("==================== MGO 库使用示例 ====================\n")

	// 示例1: 插入操作
	example1_Insert(ctx, coll)

	// 示例2: 基础查询
	example2_BasicQuery(ctx, coll)

	// 示例3: 高级查询
	example3_AdvancedQuery(ctx, coll)

	// 示例4: 更新操作
	example4_Update(ctx, coll)

	// 示例5: 删除操作
	example5_Delete(ctx, coll)

	// 示例6: 聚合操作
	example6_Aggregation(ctx, coll)

	// 示例7: 查询并修改
	example7_FindAndModify(ctx, coll)

	fmt.Println("\n==================== 示例完成 ====================")
}

// 示例1: 插入操作
func example1_Insert(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("📝 示例1: 插入操作")
	fmt.Println("----------------------------------------")

	// 1.1 插入单条文档
	user1 := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		City:      "北京",
		Status:    "active",
		Tags:      []string{"vip", "developer"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := coll.InsertOne(ctx, user1)
	if err != nil {
		log.Fatal("插入失败:", err)
	}
	fmt.Printf("✓ 插入单条文档，ID: %v\n", result.InsertedID)

	// 1.2 批量插入多条文档
	users := []any{
		User{
			Name:      "李四",
			Email:     "lisi@example.com",
			Age:       30,
			City:      "上海",
			Status:    "active",
			Tags:      []string{"manager"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		User{
			Name:      "王五",
			Email:     "wangwu@example.com",
			Age:       22,
			City:      "深圳",
			Status:    "active",
			Tags:      []string{"designer"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		User{
			Name:      "赵六",
			Email:     "zhaoliu@example.com",
			Age:       28,
			City:      "北京",
			Status:    "inactive",
			Tags:      []string{"developer"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		User{
			Name:      "孙七",
			Email:     "sunqi@example.com",
			Age:       35,
			City:      "广州",
			Status:    "active",
			Tags:      []string{"vip", "manager"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	manyResult, err := coll.InsertMany(ctx, users)
	if err != nil {
		log.Fatal("批量插入失败:", err)
	}
	fmt.Printf("✓ 批量插入 %d 条文档成功\n\n", len(manyResult.InsertedIDs))
}

// 示例2: 基础查询
func example2_BasicQuery(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("📖 示例2: 基础查询")
	fmt.Println("----------------------------------------")

	// 2.1 查询单条文档
	var user User
	err := coll.Query(ctx).Eq("name", "张三").One(&user)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 查询单条: %s (年龄: %d, 城市: %s)\n", user.Name, user.Age, user.City)

	// 2.2 查询多条文档
	var activeUsers []User
	err = coll.Query(ctx).Eq("status", "active").All(&activeUsers)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 查询所有 active 用户: 共 %d 人\n", len(activeUsers))
	for _, u := range activeUsers {
		fmt.Printf("  - %s (年龄: %d, 城市: %s)\n", u.Name, u.Age, u.City)
	}

	// 2.3 计数
	count, err := coll.Query(ctx).Eq("status", "active").Count()
	if err != nil {
		log.Fatal("计数失败:", err)
	}
	fmt.Printf("✓ Active 用户数量: %d\n", count)

	// 2.4 判断是否存在
	exists, err := coll.Query(ctx).Eq("email", "zhangsan@example.com").Exists()
	if err != nil {
		log.Fatal("判断失败:", err)
	}
	fmt.Printf("✓ 邮箱 zhangsan@example.com 是否存在: %v\n\n", exists)
}

// 示例3: 高级查询
func example3_AdvancedQuery(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("🔍 示例3: 高级查询")
	fmt.Println("----------------------------------------")

	// 3.1 条件查询（年龄大于等于25）
	var users1 []User
	err := coll.Query(ctx).Gte("age", 25).All(&users1)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 年龄 >= 25 的用户: %d 人\n", len(users1))

	// 3.2 多条件查询（AND）
	var users2 []User
	err = coll.Query(ctx).
		Eq("city", "北京").
		Eq("status", "active").
		All(&users2)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 北京的 active 用户: %d 人\n", len(users2))

	// 3.3 IN 查询
	var users3 []User
	err = coll.Query(ctx).
		In("city", "北京", "上海", "深圳").
		All(&users3)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 在北京、上海、深圳的用户: %d 人\n", len(users3))

	// 3.4 范围查询
	var users4 []User
	err = coll.Query(ctx).Between("age", 25, 30).All(&users4)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 年龄在 25-30 之间的用户: %d 人\n", len(users4))

	// 3.5 字符串包含查询
	var users5 []User
	err = coll.Query(ctx).Contains("email", "example.com").All(&users5)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 邮箱包含 'example.com' 的用户: %d 人\n", len(users5))

	// 3.6 带投影和排序的查询
	var users6 []User
	err = coll.Query(ctx).
		Eq("status", "active").
		Select("name", "email", "age").
		Desc("age").
		Limit(3).
		All(&users6)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Println("✓ Active 用户按年龄降序，取前3名:")
	for i, u := range users6 {
		fmt.Printf("  %d. %s (年龄: %d)\n", i+1, u.Name, u.Age)
	}

	// 3.7 分页查询
	var users7 []User
	err = coll.Query(ctx).
		Eq("status", "active").
		Page(1, 2). // 第1页，每页2条
		All(&users7)
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	fmt.Printf("✓ 分页查询 (第1页，每页2条): %d 条记录\n\n", len(users7))
}

// 示例4: 更新操作
func example4_Update(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("✏️  示例4: 更新操作")
	fmt.Println("----------------------------------------")

	// 4.1 更新单条文档
	update1 := mgo.Update().
		Set("age", 26).
		Set("updated_at", time.Now())

	result1, err := coll.Query(ctx).Eq("name", "张三").UpdateOne(update1)
	if err != nil {
		log.Fatal("更新失败:", err)
	}
	fmt.Printf("✓ 更新张三的年龄: 匹配 %d 条，修改 %d 条\n", result1.MatchedCount, result1.ModifiedCount)

	// 4.2 更新多条文档
	update2 := mgo.Update().
		Set("status", "active").
		Set("updated_at", time.Now())

	result2, err := coll.Query(ctx).Eq("city", "北京").UpdateMany(update2)
	if err != nil {
		log.Fatal("更新失败:", err)
	}
	fmt.Printf("✓ 更新所有北京用户状态: 匹配 %d 条，修改 %d 条\n", result2.MatchedCount, result2.ModifiedCount)

	// 4.3 递增操作
	update3 := mgo.Update().
		Inc("age", 1).
		Set("updated_at", time.Now())

	_, err = coll.Query(ctx).Eq("name", "李四").UpdateOne(update3)
	if err != nil {
		log.Fatal("递增失败:", err)
	}
	fmt.Println("✓ 李四年龄 +1")

	// 4.4 数组操作
	update4 := mgo.Update().
		Push("tags", "premium").
		Set("updated_at", time.Now())

	_, err = coll.Query(ctx).Eq("name", "王五").UpdateOne(update4)
	if err != nil {
		log.Fatal("数组操作失败:", err)
	}
	fmt.Println("✓ 给王五添加标签 'premium'")

	// 4.5 通过 ID 更新（快捷方法）
	var userToUpdate User
	coll.Query(ctx).Eq("name", "孙七").One(&userToUpdate)

	update5 := mgo.Update().Set("status", "vip")
	_, err = coll.UpdateByID(ctx, userToUpdate.ID, update5)
	if err != nil {
		log.Fatal("通过ID更新失败:", err)
	}
	fmt.Printf("✓ 通过 ID 更新孙七的状态为 vip\n\n")
}

// 示例5: 删除操作
func example5_Delete(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("🗑️  示例5: 删除操作")
	fmt.Println("----------------------------------------")

	// 5.1 删除单条文档
	result1, err := coll.Query(ctx).Eq("name", "王五").DeleteOne()
	if err != nil {
		log.Fatal("删除失败:", err)
	}
	fmt.Printf("✓ 删除王五: 删除了 %d 条文档\n", result1.DeletedCount)

	// 5.2 删除多条文档
	result2, err := coll.Query(ctx).Eq("status", "inactive").DeleteMany()
	if err != nil {
		log.Fatal("删除失败:", err)
	}
	fmt.Printf("✓ 删除所有 inactive 用户: 删除了 %d 条文档\n", result2.DeletedCount)

	// 5.3 查看剩余用户
	count, _ := coll.Count(ctx)
	fmt.Printf("✓ 剩余用户总数: %d\n\n", count)
}

// 示例6: 聚合操作
func example6_Aggregation(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("📊 示例6: 聚合操作")
	fmt.Println("----------------------------------------")

	// 6.1 按城市分组统计
	type CityCount struct {
		City  string `bson:"_id"`
		Count int    `bson:"count"`
	}
	var cityCounts []CityCount
	err := coll.Aggs(ctx).
		Group("$city", mgo.M{
			"count": mgo.Sum(1),
		}).
		SortDesc("count").
		All(&cityCounts)
	if err != nil {
		log.Fatal("聚合失败:", err)
	}
	fmt.Println("✓ 按城市分组统计用户数:")
	for _, cc := range cityCounts {
		fmt.Printf("  - %s: %d 人\n", cc.City, cc.Count)
	}

	// 6.2 按城市分组计算平均年龄
	type CityStats struct {
		City      string  `bson:"_id"`
		AvgAge    float64 `bson:"avg_age"`
		MinAge    int     `bson:"min_age"`
		MaxAge    int     `bson:"max_age"`
		UserCount int     `bson:"user_count"`
	}
	var cityStats []CityStats
	err = coll.Aggs(ctx).
		Match(mgo.Filter().Eq("status", "active")).
		Group("$city", mgo.M{
			"avg_age":    mgo.Avg("$age"),
			"min_age":    mgo.Min("$age"),
			"max_age":    mgo.Max("$age"),
			"user_count": mgo.Sum(1),
		}).
		SortDesc("avg_age").
		All(&cityStats)
	if err != nil {
		log.Fatal("聚合失败:", err)
	}
	fmt.Println("✓ 按城市统计年龄信息 (仅 active 用户):")
	for _, cs := range cityStats {
		fmt.Printf("  - %s: 平均 %.1f 岁, 最小 %d, 最大 %d (共 %d 人)\n",
			cs.City, cs.AvgAge, cs.MinAge, cs.MaxAge, cs.UserCount)
	}

	// 6.3 按状态分组
	type StatusGroup struct {
		Status string   `bson:"_id"`
		Count  int      `bson:"count"`
		Names  []string `bson:"names"`
	}
	var statusGroups []StatusGroup
	err = coll.Aggs(ctx).
		Group(mgo.F("status"), mgo.M{
			"count": mgo.Sum(1),
			"names": mgo.Push("$name"),
		}).
		All(&statusGroups)
	if err != nil {
		log.Fatal("聚合失败:", err)
	}
	fmt.Println("✓ 按状态分组:")
	for _, sg := range statusGroups {
		fmt.Printf("  - %s: %d 人 - %v\n", sg.Status, sg.Count, sg.Names)
	}

	// 6.4 复杂聚合 - 筛选、投影、排序、限制
	var topUsers []User
	err = coll.Aggs(ctx).
		Match(mgo.Filter().Eq("status", "active")).
		Project(mgo.NewProjection().Include("name", "age", "city")).
		SortDesc("age").
		Limit(2).
		All(&topUsers)
	if err != nil {
		log.Fatal("聚合失败:", err)
	}
	fmt.Println("✓ Active 用户中年龄最大的前2名:")
	for i, u := range topUsers {
		fmt.Printf("  %d. %s - %d岁 (%s)\n", i+1, u.Name, u.Age, u.City)
	}
	fmt.Println()
}

// 示例7: 查询并修改
func example7_FindAndModify(ctx context.Context, coll *mgo.Collection) {
	fmt.Println("🔄 示例7: 查询并修改")
	fmt.Println("----------------------------------------")

	// 7.1 查询并更新
	var updatedUser User
	update := mgo.Update().
		Set("status", "premium").
		Inc("age", 1).
		Set("updated_at", time.Now())

	err := coll.Query(ctx).
		Eq("name", "李四").
		FindAndUpdate(update, &updatedUser)
	if err != nil {
		log.Fatal("查询并更新失败:", err)
	}
	fmt.Printf("✓ 查询并更新李四: 新状态=%s, 新年龄=%d\n", updatedUser.Status, updatedUser.Age)

	// 7.2 查询并删除
	var deletedUser User
	err = coll.Query(ctx).
		Eq("name", "孙七").
		FindAndDelete(&deletedUser)
	if err != nil {
		log.Fatal("查询并删除失败:", err)
	}
	fmt.Printf("✓ 查询并删除孙七: 删除的用户邮箱=%s\n", deletedUser.Email)

	// 7.3 使用游标（高级用法）
	cursor, err := coll.Query(ctx).
		Eq("status", "active").
		Cursor()
	if err != nil {
		log.Fatal("获取游标失败:", err)
	}
	defer cursor.Close(ctx)

	fmt.Println("✓ 使用游标遍历所有 active 用户:")
	count := 0
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Fatal("解码失败:", err)
		}
		count++
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}
	fmt.Printf("  共遍历 %d 条记录\n", count)
}
