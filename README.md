# mgo - 泛型 MongoDB Go 客户端

`mgo` 是一个基于 Go 泛型的 MongoDB 客户端封装，旨在提供类型安全、简洁且功能强大的数据库操作体验。它在官方 `mongo-driver` 的基础上进行了封装，简化了常见的 CRUD 操作，并提供了聚合、事务、批量处理等高级功能。

## ✨ 特性

- **泛型支持**：完全基于 Go 1.18+ 泛型，提供类型安全的 CRUD 操作。
- **链式调用**：流畅的查询构建器 API，支持复杂的过滤、排序和分页。
- **自动化钩子**：内置 `Created` / `Updated` 时间戳自动管理和软删除支持。
- **高级聚合**：提供类型安全的聚合管道构建器 (`agg` 包)。
- **事务管理**：简化事务操作，支持自动回滚和重试 (`tx` 包)。
- **批量处理**：高效的批量插入和流式处理 (`batch` 包)。
- **上下文集成**：原生支持 `context.Context`，便于超时控制和链路追踪。

## 📦 安装

```bash
go get github.com/gocrud/mgo
```

## 🚀 快速开始

### 1. 定义模型

模型结构体需要实现 `mgo.Namer` 接口来指定集合名称。

```go
package main

import (
	"time"
	"github.com/gocrud/mgo"
)

type User struct {
	ID        mgo.ObjectID `bson:"_id,omitempty"`
	Name      string       `bson:"name"`
	Email     string       `bson:"email"`
	Age       int          `bson:"age"`
	CreatedAt time.Time    `bson:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at"`
}

// 实现 mgo.Namer 接口
func (User) CollName() string {
	return "users"
}
```

### 2. 连接与基本操作

```go
func main() {
	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_demo")
	
	// 2. 获取泛型集合 (启用自动时间戳)
	users := mgo.Model[User](db).WithTimestamps()

	// 3. 插入数据
	user := &User{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}
	id, _ := users.Insert(user)

	// 4. 查询数据
	// 根据 ID 查询
	foundUser, _ := users.FindByID(id)
	
	// 链式查询
	activeUsers, _ := users.Find().
		Eq("age", 25).
		Limit(10).
		All()

	// 5. 更新数据
	users.Find().ID(id).Set("age", 26).Update()

	// 6. 删除数据
	users.Find().ID(id).Delete()
}
```

## 📖 详细文档

### 连接数据库

`mgo` 提供了多种连接方式：

```go
// 方式一：MustOpen (连接失败会 panic，适合初始化)
db := mgo.MustOpen("mongodb://localhost/mydb")

// 方式二：Open (返回 error)
db, err := mgo.Open("mongodb://localhost/mydb")

// 方式三：OpenClient (获取 Client 实例，用于跨库操作)
client, err := mgo.OpenClient("mongodb://localhost")
db := client.Database("mydb")

// 配置选项
db, err := mgo.Open("mongodb://localhost/mydb", 
    mgo.MaxPoolSize(100),
    mgo.Timeout(10*time.Second),
)
```

### 模型定义与选项

使用 `mgo.Model[T](db)` 获取集合操作对象。

```go
// 基础用法
users := mgo.Model[User](db)

// 启用自动时间戳 (默认字段: created_at, updated_at)
users := mgo.Model[User](db).WithTimestamps()

// 自定义时间戳字段
users := mgo.Model[User](db).WithTimestamps("create_time", "update_time")

// 启用软删除 (默认字段: deleted_at)
users := mgo.Model[User](db).WithSoftDelete()
```

### 查询 (Query)

`mgo` 提供了丰富的链式查询方法。

#### 基础查询

```go
// 查询单条
user, err := users.Find().Eq("name", "张三").One()

// 查询列表
list, err := users.Find().Gt("age", 18).All()

// 统计数量
count, err := users.Find().Eq("status", "active").Count()
```

#### 过滤条件

```go
q := users.Find()

q.Eq("name", "张三")       // 等于
q.Ne("status", "banned")   // 不等于
q.Gt("age", 18)            // 大于
q.Gte("age", 18)           // 大于等于
q.Lt("age", 60)            // 小于
q.Lte("age", 60)           // 小于等于
q.In("role", []string{"admin", "editor"}) // 包含
q.Regex("email", "@gmail.com$") // 正则匹配

// 复杂条件 (支持 Map 合并)
q.Where(mgo.M{
    "age": mgo.M{"$gt": 18},
    "status": "active",
})

// 条件分支
q.When(isAdult, func(q *mgo.Query[User]) {
    q.Gt("age", 18)
})
```

#### 分页与排序

```go
// 排序 (字段名前加 - 表示降序)
users.Find().Sort("-created_at", "name").All()

// 分页
users.Find().Skip(10).Limit(20).All()

// 分页列表 (返回分页信息)
page, err := users.Find().PageList(1, 20)
fmt.Printf("总数: %d, 总页数: %d\n", page.Total, page.TotalPages)
```

### 更新 (Update)

```go
// 更新单条
users.Find().ID(id).Set("name", "李四").Update()

// 批量更新
users.Find().Lt("age", 18).Set("status", "minor").UpdateMany()

// 原子操作
users.Find().ID(id).Inc("balance", 100).Update() // 增加
users.Find().ID(id).Push("tags", "new_tag").Update() // 数组追加
users.Find().ID(id).Pull("tags", "old_tag").Update() // 数组移除
```

### 删除 (Delete)

```go
// 删除单条
users.Find().ID(id).Delete()

// 批量删除
users.Find().Eq("status", "deleted").DeleteMany()

// 如果启用了软删除，Delete() 会执行软删除，ForceDelete() 执行物理删除
```

### 聚合 (Aggregation)

使用 `agg` 子包进行聚合查询。

```go
import "github.com/gocrud/mgo/agg"

type CityStats struct {
    City   string  `bson:"_id"`
    Count  int     `bson:"count"`
    AvgAge float64 `bson:"avg_age"`
}

results, err := agg.Aggregate[CityStats](users).
    Match(mgo.Eq("status", "active")).
    GroupBy("$city").
        Count("count").
        Avg("avg_age", "$age").
    SortDesc("count").
    All()
```

### 事务 (Transactions)

使用 `tx` 子包管理事务。

```go
import "github.com/gocrud/mgo/tx"

err := tx.Transaction(db, func(sess *tx.Session) error {
    users := mgo.Model[User](sess)
    orders := mgo.Model[Order](sess)
    
    if err := users.Find().ID(uid).Inc("balance", -100).Update(); err != nil {
        return err
    }
    
    if _, err := orders.Insert(newOrder); err != nil {
        return err
    }
    
    return nil
})
```

### 批量处理 (Batch)

使用 `batch` 子包进行高效批量操作。

```go
import "github.com/gocrud/mgo/batch"

// 自动分批插入
err := batch.InsertBatch(users, largeUserList, batch.Size(500))
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
