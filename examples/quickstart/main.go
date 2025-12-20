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
	CreatedAt time.Time    `bson:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at"`
}

func main() {
	// 连接数据库
	db := mgo.MustOpen("mongodb://localhost/quickstart")
	defer db.Close()

	// 获取集合
	users := mgo.Model[User](db).WithTimestamps()

	// 插入
	user := &User{
		Name:  "张三",
		Email: "zhangsan@example.com",
	}
	id, err := users.Insert(user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("插入成功，ID: %s\n", id.Hex())

	// 查询
	found, err := users.FindByID(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("查询成功: %s (%s)\n", found.Name, found.Email)

	// 更新
	err = users.Find().ID(id).
		Set("name", "张三丰").
		Update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("更新成功")

	// 查询列表
	results, _ := users.Find().
		OrderBy("created_at").
		Limit(10).
		All()
	fmt.Printf("查询到 %d 个用户\n", len(results))

	// 删除
	err = users.Find().ID(id).Delete()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("删除成功")
}
