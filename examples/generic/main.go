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
	Age       int           `bson:"age"`
	City      string        `bson:"city"`
	CreatedAt time.Time     `bson:"created_at"`
}

func main() {
	// 1. 连接 MongoDB
	ctx := context.Background()
	client, err := mgo.Connect(ctx, "mongodb://example:example@localhost:27017/?directConnection=true")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// 2. 获取泛型集合 (Type Safe)
	// mgo.Model[T] 将普通 Collection 包装为 TypedCollection[T]
	coll := mgo.Model[User](client.Database("test_db").Collection("users"))

	// 3. 插入数据 (使用原生 InsertOne，因为它是泛型的)
	user := User{
		Name:      "泛型用户",
		Age:       25,
		City:      "深圳",
		CreatedAt: time.Now(),
	}
	_, err = coll.InsertOne(ctx, user)
	if err != nil {
		log.Printf("插入失败: %v (请确保 MongoDB 正在运行且认证正确)", err)
		return
	}
	fmt.Println("✅ 插入成功")

	// 4. 泛型查询 (直接返回 User 对象，无需传指针)
	foundUser, err := coll.Query(ctx).Eq("name", "泛型用户").One()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 查询成功: %+v\n", foundUser)

	// 5. 泛型列表查询
	users, err := coll.Query(ctx).Gte("age", 18).All()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 找到 %d 个成年用户\n", len(users))

	// 6. 简化版 Update (无需 UpdateBuilder)
	// 使用 mgo.Set 快速构建 map
	updateResult, err := coll.Query(ctx).
		Eq("name", "泛型用户").
		UpdateOne(mgo.Set("age", 26))
	
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ 更新成功，修改了 %d 条\n", updateResult.ModifiedCount)

	// 7. 泛型 FindAndUpdate
	// 直接返回更新后的对象
	updatedUser, err := coll.Query(ctx).
		Eq("name", "泛型用户").
		FindAndUpdate(mgo.Set("city", "北京"))
	
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ FindAndUpdate 成功: 城市变更为 %s\n", updatedUser.City)
}

