package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocrud/mgo"
)

// User 用户模型
type User struct {
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

func (User) TableName() string {
	return "users"
}

func main() {
	fmt.Println("🚀 MGO 完整示例")
	fmt.Println("================")

	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_example")
	defer db.Close()
	fmt.Println("✅ 已连接到数据库")

	// 2. 获取泛型集合（自动时间戳 + 软删除）
	users := mgo.Model[User](db).
		WithTimestamps().
		WithSoftDelete()
	fmt.Println("✅ 已创建泛型集合")

	// 清空集合（测试用）
	users.Truncate()

	// 3. 插入数据
	fmt.Println("\n📝 插入数据...")
	user1 := &User{
		Name:    "张三",
		Email:   "zhangsan@example.com",
		Age:     25,
		Status:  "active",
		Balance: 1000.50,
		Tags:    []string{"vip", "verified"},
	}

	id1, err := users.Insert(user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 插入用户 1: %s\n", id1.Hex())

	// 批量插入
	users.InsertMany(
		&User{Name: "李四", Email: "lisi@example.com", Age: 30, Status: "active", Balance: 500},
		&User{Name: "王五", Email: "wangwu@example.com", Age: 22, Status: "pending", Balance: 200},
		&User{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: "inactive", Balance: 800},
	)
	fmt.Println("✅ 批量插入完成")

	// 4. 查询数据
	fmt.Println("\n🔍 查询数据...")

	// 按 ID 查询
	user, err := users.FindByID(id1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 按 ID 查询: %s (%s)\n", user.Name, user.Email)

	// 条件查询
	activeUsers, _ := users.Find().
		Where("status", "active").
		Where("age", ">", 20).
		OrderBy("age").
		All()
	fmt.Printf("✅ 活跃用户数: %d\n", len(activeUsers))

	// 统计
	count, _ := users.Find().
		Where("status", "active").
		Count()
	fmt.Printf("✅ 统计数量: %d\n", count)

	// 5. 时间查询
	fmt.Println("\n📅 时间查询...")
	todayUsers, _ := users.Find().
		WhereToday("created_at").
		All()
	fmt.Printf("✅ 今天创建的用户: %d 个\n", len(todayUsers))

	// 6. 更新数据
	fmt.Println("\n✏️ 更新数据...")
	err = users.Find().ID(id1).
		Set("status", "premium").
		Inc("balance", 500).
		Push("tags", "gold").
		Update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 更新成功")

	// 批量更新
	n, _ := users.Find().
		Where("status", "pending").
		Set("status", "active").
		UpdateMany()
	fmt.Printf("✅ 批量更新: %d 条记录\n", n)

	// 7. 分页查询
	fmt.Println("\n📄 分页查询...")
	page, _ := users.Find().
		Where("status", "active").
		Page(1, 2)
	fmt.Printf("✅ 第 1 页，共 %d 页，总计 %d 条记录\n", page.Pages, page.Total)
	for _, u := range page.Items {
		fmt.Printf("   - %s (Age: %d)\n", u.Name, u.Age)
	}

	// 8. 软删除
	fmt.Println("\n🗑️ 软删除...")
	err = users.Find().ID(id1).Delete()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 软删除成功")

	// 验证软删除
	allUsers, _ := users.Find().All()
	fmt.Printf("✅ 默认查询（排除已删除）: %d 个用户\n", len(allUsers))

	withTrashed, _ := users.Find().WithTrashed().All()
	fmt.Printf("✅ 包含已删除: %d 个用户\n", len(withTrashed))

	// 恢复
	err = users.Find().ID(id1).WithTrashed().Restore()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 恢复成功")

	// 9. 复杂查询示例
	fmt.Println("\n🎯 复杂查询...")
	complexResults, _ := users.Find().
		Filter(
			mgo.And(
				mgo.Eq("status", "active"),
				mgo.Or(
					mgo.Gt("age", 25),
					mgo.Gt("balance", 500),
				),
			),
		).
		Desc("balance").
		Limit(5).
		All()
	fmt.Printf("✅ 复杂查询结果: %d 个用户\n", len(complexResults))

	// 10. 便捷方法
	fmt.Println("\n⚡ 便捷方法...")
	exists, _ := users.Find().
		Where("email", "zhangsan@example.com").
		Exists()
	fmt.Printf("✅ 邮箱存在性检查: %v\n", exists)

	// 查询或创建
	testEmail := "test@example.com"
	testUser, created, _ := users.Find().
		Where("email", testEmail).
		FirstOrCreate(&User{
			Name:   "测试用户",
			Email:  testEmail,
			Status: "active",
		})
	if created {
		fmt.Printf("✅ 创建新用户: %s\n", testUser.Name)
	} else {
		fmt.Printf("✅ 用户已存在: %s\n", testUser.Name)
	}

	fmt.Println("\n✨ 示例完成！")
}
