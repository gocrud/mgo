package main

import (
	"context"
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

func (User) CollName() string {
	return "users"
}

func main() {
	ctx := context.Background()

	fmt.Println("🚀 MGO 完整示例")
	fmt.Println("================")

	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_example")
	defer db.Close(ctx)
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

	id1, err := users.WithContext(ctx).Insert(user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 插入用户 1: %s\n", id1.Hex())

	// 批量插入
	users.WithContext(ctx).InsertMany([]*User{
		{Name: "李四", Email: "lisi@example.com", Age: 30, Status: "active", Balance: 500},
		{Name: "王五", Email: "wangwu@example.com", Age: 22, Status: "pending", Balance: 200},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: "inactive", Balance: 800},
	})
	fmt.Println("✅ 批量插入完成")

	// 4. 查询数据
	fmt.Println("\n🔍 查询数据...")

	// 按 ID 查询
	user, err := users.WithContext(ctx).FindByID(id1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 按 ID 查询: %s (%s)\n", user.Name, user.Email)

	// 条件查询
	activeUsers, _ := users.Find().WithContext(ctx).
		Eq("status", "active").
		Gt("age", 20).
		OrderBy("age").
		All()
	fmt.Printf("✅ 活跃用户数: %d\n", len(activeUsers))

	// 统计
	count, _ := users.Find().WithContext(ctx).
		Eq("status", "active").
		Count()
	fmt.Printf("✅ 统计数量: %d\n", count)

	// 5. 时间查询
	fmt.Println("\n📅 时间查询...")
	todayUsers, _ := users.Find().WithContext(ctx).
		WhereToday("created_at").
		All()
	fmt.Printf("✅ 今天创建的用户: %d 个\n", len(todayUsers))
}
