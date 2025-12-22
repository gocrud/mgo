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
	CreatedAt time.Time    `bson:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at"`
}

func main() {
	fmt.Println("🚀 非泛型 Collection 示例 - 统一 API")
	fmt.Println("=======================================")

	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_example")
	defer db.Close()
	fmt.Println("✅ 已连接到数据库")

	// 2. 获取非泛型集合
	coll := db.Collection("users")
	fmt.Println("✅ 已创建非泛型集合")

	// 清空集合（测试用）
	coll.Drop()

	// ==================== 插入方法示例 ====================
	fmt.Println("\n📝 1. 插入方法（与泛型一致）")
	fmt.Println("-----------------------------------")

	// Insert - 插入单条文档
	user1 := &User{
		Name:   "张三",
		Email:  "zhangsan@example.com",
		Age:    25,
		Status: "active",
	}

	id1, err := coll.Insert(user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Insert 单条: %s (ID: %s)\n", user1.Name, id1.Hex())

	// InsertMany - 可变参数，与泛型一致
	user2 := &User{Name: "李四", Email: "lisi@example.com", Age: 30, Status: "active"}
	user3 := &User{Name: "王五", Email: "wangwu@example.com", Age: 22, Status: "pending"}
	user4 := &User{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: "inactive"}

	ids, err := coll.InsertMany(user2, user3, user4)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ InsertMany 批量插入: 成功插入 %d 条记录\n", len(ids))

	// ==================== 分页查询示例 ====================
	fmt.Println("\n📄 2. 分页查询")
	fmt.Println("-----------------------------------")

	// 添加更多测试数据
	for i := 1; i <= 50; i++ {
		coll.Insert(&User{
			Name:   fmt.Sprintf("用户%d", i),
			Email:  fmt.Sprintf("user%d@example.com", i),
			Age:    20 + (i % 30),
			Status: []string{"active", "pending", "inactive"}[i%3],
		})
	}
	fmt.Println("✅ 已添加 50 条测试数据")

	// PageList - 标准分页（带总数统计）
	fmt.Println("\n📊 标准分页（PageList）:")
	var users1 []User
	page1, err := coll.Find().
		Where("status", "active").
		OrderBy("age").Desc().
		PageList(1, 10, &users1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   总记录数: %d\n", page1.Total)
	fmt.Printf("   总页数: %d\n", page1.Pages)
	fmt.Printf("   当前页: %d\n", page1.Page)
	fmt.Printf("   每页数量: %d\n", page1.PerPage)
	fmt.Printf("   当前页记录: %d 条\n", len(users1))

	// SimplePageList - 简化分页（不统计总数，性能更好）
	fmt.Println("\n⚡ 简化分页（SimplePageList）:")
	var users2 []User

	page2, err := coll.Find().
		Where("status", "active").
		OrderBy("age").
		SimplePageList(1, 10, &users2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   当前页: %d\n", page2.Page)
	fmt.Printf("   每页数量: %d\n", page2.PerPage)
	fmt.Printf("   是否有下一页: %v\n", page2.HasMore)
	fmt.Printf("   当前页记录: %d 条\n", len(users2))

	// CursorPage - 游标分页（适用于大数据量）
	fmt.Println("\n🔄 游标分页（CursorPage）:")
	var users3 []User
	page3, err := coll.Find().
		Where("status", "active").
		OrderBy("age").Desc().
		CursorPage("", 10, &users3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   当前页记录: %d 条\n", len(users3))
	fmt.Printf("   是否有下一页: %v\n", page3.HasMore)
	fmt.Printf("   下一页游标: %s\n", page3.NextCursor[:30]+"...")

	// 获取下一页
	var users4 []User
	page4, err := coll.Find().
		Where("status", "active").
		OrderBy("age").Desc().
		CursorPage(page3.NextCursor, 10, &users4)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   下一页记录: %d 条\n", len(users4))
	fmt.Printf("   是否有更多: %v\n", page4.HasMore)

	// ==================== API 对比 ====================
	fmt.Println("\n📋 3. API 对比总结")
	fmt.Println("-----------------------------------")
	fmt.Println("插入方法：")
	fmt.Println("  ✅ Insert(doc)       - 单条插入（与泛型一致）")
	fmt.Println("  ✅ InsertMany(docs...) - 批量插入，可变参数（与泛型一致）")
	fmt.Println("\n分页方法：")
	fmt.Println("  ✅ PageList(page, perPage, &results)")
	fmt.Println("     - 标准分页，返回总数和总页数")
	fmt.Println("  ✅ SimplePageList(page, perPage, &results)")
	fmt.Println("     - 简化分页，不统计总数（性能更好）")
	fmt.Println("  ✅ CursorPage(cursor, perPage, &results)")
	fmt.Println("     - 游标分页，适用于大数据量，支持双向翻页")

	// ==================== 与泛型版本对比 ====================
	fmt.Println("\n🔄 4. 与泛型版本对比")
	fmt.Println("-----------------------------------")
	fmt.Println("非泛型 Collection:")
	fmt.Println("  coll.Insert(doc)           // 返回 (ObjectID, error)")
	fmt.Println("  coll.InsertMany(doc1, doc2) // 返回 ([]ObjectID, error)")
	fmt.Println("  coll.Find().PageList(1, 20, &users)")
	fmt.Println("\n泛型 TypedCollection:")
	fmt.Println("  users.Insert(doc)           // 返回 (ObjectID, error)")
	fmt.Println("  users.InsertMany(doc1, doc2) // 返回 ([]ObjectID, error)")
	fmt.Println("  users.Find().PageList(1, 20)    // 返回 (*PageResult[T], error)")

	fmt.Println("\n✨ 完成！所有功能演示完毕")
}
