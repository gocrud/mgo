package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gocrud/mgo"
)

func main() {
	ctx := context.Background()

	fmt.Println("🚀 非泛型 Collection 示例 (使用 mgo.M)")
	fmt.Println("=======================================")

	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_example")
	defer db.Close(ctx)
	fmt.Println("✅ 已连接到数据库")

	// 2. 获取非泛型集合 (默认返回 Collection[M])
	coll := db.Collection("users")
	fmt.Println("✅ 已创建非泛型集合")

	// 清空集合（测试用）
	coll.Drop()

	// ==================== 插入方法示例 ====================
	fmt.Println("\n📝 1. 插入方法")
	fmt.Println("-----------------------------------")

	// Insert - 插入单条文档
	user1 := mgo.M{
		"name":   "张三",
		"email":  "zhangsan@example.com",
		"age":    25,
		"status": "active",
	}

	id1, err := coll.WithContext(ctx).Insert(&user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Insert 单条: %s (ID: %s)\n", user1["name"], id1.Hex())

	// InsertMany - 批量插入
	user2 := mgo.M{"name": "李四", "email": "lisi@example.com", "age": 30, "status": "active"}
	user3 := mgo.M{"name": "王五", "email": "wangwu@example.com", "age": 22, "status": "pending"}
	user4 := mgo.M{"name": "赵六", "email": "zhaoliu@example.com", "age": 35, "status": "inactive"}

	// InsertMany 接受 []*T
	users := []*mgo.M{&user2, &user3, &user4}
	ids, err := coll.WithContext(ctx).InsertMany(users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ InsertMany 批量插入: 成功插入 %d 条记录\n", len(ids))

	// ==================== 分页查询示例 ====================
	fmt.Println("\n📄 2. 分页查询")
	fmt.Println("-----------------------------------")

	// 添加更多测试数据
	moreUsers := make([]*mgo.M, 0, 50)
	for i := 1; i <= 50; i++ {
		u := mgo.M{
			"name":   fmt.Sprintf("用户%d", i),
			"email":  fmt.Sprintf("user%d@example.com", i),
			"age":    20 + (i % 30),
			"status": []string{"active", "pending", "inactive"}[i%3],
		}
		moreUsers = append(moreUsers, &u)
	}
	coll.WithContext(ctx).InsertMany(moreUsers)
	fmt.Println("✅ 已添加 50 条测试数据")

	// PageList - 标准分页（带总数统计）
	fmt.Println("\n📊 标准分页（PageList）:")

	// 暂时使用 All 和 Limit 模拟
	results, err := coll.Find().WithContext(ctx).
		Eq("status", "active").
		Limit(10).
		All()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   查询到 %d 条记录\n", len(results))
}
