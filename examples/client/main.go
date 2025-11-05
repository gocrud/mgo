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
	ID        any       `bson:"_id,omitempty"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Age       int       `bson:"age"`
	Status    string    `bson:"status"`
	CreatedAt time.Time `bson:"created_at"`
}

func main() {
	ctx := context.Background()

	fmt.Println("=== mgo Client 使用示例 ===\n")

	// 连接数据库
	client, err := mgo.NewClient(ctx, "mongodb://example:example@localhost:27017/?directConnection=true")
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer client.Disconnect(ctx)
	fmt.Println("✓ 连接成功")

	// 访问集合
	users := client.DB("testdb").Collection("users")

	// 插入用户
	user := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	result, err := users.InsertOne(ctx, user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 插入成功，ID: %v\n", result.InsertedID)

	// 查询用户
	var foundUser User
	err = users.Query(ctx).Eq("name", "张三").One(&foundUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 查询成功: %s (年龄: %d)\n", foundUser.Name, foundUser.Age)

	// 使用事务
	err = client.WithTransaction(ctx, func(ctx context.Context) error {
		_, err := users.InsertOne(ctx, User{
			Name:      "李四",
			Email:     "lisi@example.com",
			Age:       30,
			Status:    "active",
			CreatedAt: time.Now(),
		})
		return err
	})
	if err != nil {
		log.Fatal("事务失败:", err)
	}
	fmt.Println("✓ 事务执行成功")

	// 清理
	_, err = users.Query(ctx).In("name", "张三", "李四").DeleteMany()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ 清理完成")

	fmt.Println("\n所有示例执行完毕！")
}
