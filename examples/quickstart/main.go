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
	CreatedAt time.Time     `bson:"created_at"`
}

func main() {
	// 连接 MongoDB
	uri := "mongodb://example:example@localhost:27017/?directConnection=true"
	ctx := context.Background()

	client, err := mgo.Connect(ctx, uri)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("✅ 连接成功\n")

	// 获取集合
	coll := client.Database("quickstart_db").Collection("users")

	// 清空集合
	coll.Drop(ctx)

	// ==================== 1. 插入 ====================
	fmt.Println("📝 1. 插入数据")
	user := User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		City:      "北京",
		CreatedAt: time.Now(),
	}
	result, _ := coll.InsertOne(ctx, user)
	fmt.Printf("   插入成功，ID: %v\n\n", result.InsertedID)

	// ==================== 2. 查询 ====================
	fmt.Println("📖 2. 查询数据")

	// 查询单条
	var foundUser User
	err = coll.Query(ctx).Eq("name", "张三").One(&foundUser)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   查到: %s, 邮箱: %s\n", foundUser.Name, foundUser.Email)

	// 计数
	count, _ := coll.Query(ctx).Eq("city", "北京").Count()
	fmt.Printf("   北京用户数: %d\n\n", count)

	// ==================== 3. 更新 ====================
	fmt.Println("✏️  3. 更新数据")
	update := mgo.Update().Set("age", 26)
	updateResult, _ := coll.Query(ctx).
		Eq("name", "张三").
		UpdateOne(update)
	fmt.Printf("   更新了 %d 条记录\n\n", updateResult.ModifiedCount)

	// ==================== 4. 聚合 ====================
	fmt.Println("📊 4. 聚合查询")

	// 先插入更多数据
	coll.InsertMany(ctx, []any{
		User{Name: "李四", Email: "lisi@example.com", Age: 30, City: "上海", CreatedAt: time.Now()},
		User{Name: "王五", Email: "wangwu@example.com", Age: 22, City: "北京", CreatedAt: time.Now()},
	})

	// 按城市分组统计
	type CityCount struct {
		City  string `bson:"_id"`
		Count int    `bson:"count"`
	}
	var cityCounts []CityCount
	coll.Aggs(ctx).
		Stage(mgo.Stage().Group("$city", mgo.M{"count": mgo.Sum(1)})).
		All(&cityCounts)

	for _, cc := range cityCounts {
		fmt.Printf("   %s: %d 人\n", cc.City, cc.Count)
	}
	fmt.Println()

	// ==================== 5. 删除 ====================
	fmt.Println("🗑️  5. 删除数据")
	deleteResult, _ := coll.Query(ctx).
		Eq("name", "王五").
		DeleteOne()
	fmt.Printf("   删除了 %d 条记录\n\n", deleteResult.DeletedCount)

	fmt.Println("✅ 快速入门示例完成！")
}
