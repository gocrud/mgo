# MGO 渐进式使用指南

本指南旨在帮助开发者从零开始掌握 `mgo` 库的使用，涵盖从基础 CRUD 到高级聚合、事务和批量处理的全方位功能。

## 1. 安装

```bash
go get github.com/gocrud/mgo
```

## 2. 快速上手

### 2.1 定义模型

`mgo` 使用泛型设计，模型结构体需要实现 `mgo.Namer` 接口来指定集合名称。

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

### 2.2 连接数据库与基础 CRUD

```go
func main() {
	// 1. 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_demo")
	
	// 2. 获取泛型集合
	// WithTimestamps() 会自动处理 CreatedAt 和 UpdatedAt
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

## 3. 核心功能详解

### 3.1 连接管理

`mgo` 提供了简单的连接管理方式。

```go
// 方式一：MustOpen (连接失败会 panic)
db := mgo.MustOpen("mongodb://localhost/mydb")

// 方式二：Open (返回 error)
db, err := mgo.Open("mongodb://localhost/mydb")

// 方式三：OpenClient (获取 Client 实例，用于跨库操作)
client, err := mgo.OpenClient("mongodb://localhost")
db := client.Database("mydb")
```

### 3.2 模型选项

在创建集合引用时，可以配置一些自动化选项。

```go
// 启用自动时间戳
// 插入时自动设置 created_at, updated_at
// 更新时自动更新 updated_at
users := mgo.Model[User](db).WithTimestamps()

// 启用软删除
// 删除时不会物理删除，而是设置 deleted_at
// 查询时会自动过滤掉已删除的记录
users := mgo.Model[User](db).WithSoftDelete()

// 组合使用
users := mgo.Model[User](db).
    WithTimestamps().
    WithSoftDelete()
```

### 3.3 查询构建器 (Query Builder)

`mgo` 提供了丰富的链式调用方法来构建查询。

#### 基础条件

```go
q := users.Find()

q.Eq("status", "active")     // 等于
q.Ne("status", "banned")     // 不等于
q.Gt("age", 18)              // 大于
q.Gte("age", 18)             // 大于等于
q.Lt("age", 60)              // 小于
q.Lte("age", 60)             // 小于等于
q.In("role", "admin", "editor") // 包含
q.Nin("role", "guest")       // 不包含
```

#### 逻辑组合

```go
// AND 查询 (链式调用默认是 AND)
users.Find().Eq("status", "active").Gt("age", 18)

// OR 查询
users.Find().Where(mgo.Or(
    mgo.Eq("role", "admin"),
    mgo.Eq("role", "super_admin"),
))

// 复杂组合
users.Find().Where(mgo.And(
    mgo.Eq("status", "active"),
    mgo.Or(
        mgo.Gt("score", 80),
        mgo.Lt("complaints", 3),
    ),
))

// 条件分支 (When)
users.Find().When(isAdmin, func(q *mgo.Query[User]) {
    q.Eq("role", "admin")
})
```

#### 排序与分页

```go
users.Find().
    OrderBy("age").           // 升序
    OrderByDesc("created_at"). // 降序
    Skip(10).                 // 跳过前 10 条
    Limit(20).                // 取 20 条
    All()
```

#### 高级分页

`mgo` 提供了封装好的分页方法，自动计算总数和页码。

```go
// 标准分页 (包含总数统计)
// page: 1, perPage: 20
result, err := users.Find().Eq("status", "active").PageList(1, 20)
if err == nil {
    fmt.Printf("当前页: %d, 总页数: %d, 总记录: %d\n", 
        result.Page, result.Pages, result.Total)
    for _, user := range result.Items {
        // ...
    }
}

#### 时间存储说明

MongoDB 默认存储 **UTC 时间**。`mgo` 在处理自动时间戳和时间查询时遵循这一标准：

*   **存储时**：`WithTimestamps` 会自动将 `created_at` 和 `updated_at` 设置为 UTC 时间。
*   **查询时**：`WhereToday`、`WhereDateBetween` 等方法会自动将输入的本地时间转换为 UTC 时间进行查询。
*   **展示时**：从数据库取出的时间是 UTC 时间，在前端展示时请根据用户时区进行转换（例如使用 `t.In(time.Local)`）。

#### 简单分页 (不统计总数，性能更好，仅判断是否有下一页)
simpleResult, err := users.Find().SimplePageList(1, 20)
if simpleResult.HasMore {
    fmt.Println("还有下一页")
}
```

#### 时间查询辅助

提供了便捷的时间范围查询方法。

```go
// 查询今天注册的用户
users.Find().WhereToday("created_at").All()

// 查询昨天
users.Find().WhereYesterday("created_at").All()

// 日期范围 (字符串格式)
users.Find().WhereDateBetween("created_at", "2024-01-01", "2024-12-31").All()

// 指定日期之后
users.Find().WhereDateAfter("created_at", "2024-01-01").All()
```

#### 结果获取

```go
// 获取所有结果
var list []*User
list, err := users.Find().All()

// 获取单条结果
var one *User
one, err := users.Find().Eq("email", "test@example.com").One()

// 获取数量
count, err := users.Find().Eq("status", "active").Count()

// 判断是否存在
exists, err := users.Find().Eq("email", "test@example.com").Exists()
```

### 3.4 更新操作

支持链式更新和原子操作。

```go
// 设置字段
users.Find().ID(id).Set("name", "李四").Update()

// 数值增减
users.Find().ID(id).Inc("views", 1).Update()

// 数组操作
users.Find().ID(id).Push("tags", "golang").Update()      // 添加元素
users.Find().ID(id).Pull("tags", "deprecated").Update()  // 移除元素
users.Find().ID(id).AddToSet("tags", "unique").Update()  // 添加不重复元素
```

### 3.5 删除操作

```go
// 根据 ID 删除
users.Find().ID(id).Delete()

// 根据条件删除
users.Find().Eq("status", "inactive").Delete()

// 如果开启了 WithSoftDelete()，上述操作会执行软删除
// 要强制物理删除：
users.Find().ID(id).ForceDelete()
```

## 4. 进阶功能

### 4.1 聚合查询 (Aggregation)

使用 `agg` 子包进行复杂的聚合分析。

```go
import "github.com/gocrud/mgo/agg"

type UserStats struct {
    City      string  `bson:"_id"`
    UserCount int     `bson:"count"`
    AvgAge    float64 `bson:"avg_age"`
}

results, err := agg.Aggregate[UserStats](users).
    Match(mgo.Eq("status", "active")). // 过滤
    GroupBy("$city").                  // 分组
        Count("count").                // 统计数量
        Avg("avg_age", "$age").        // 计算平均年龄
    SortDesc("count").                 // 排序
    Limit(5).                          // 限制数量
    All()
```

### 4.2 批量处理 (Batch)

使用 `batch` 子包处理大量数据，提高性能。

```go
import "github.com/gocrud/mgo/batch"

// 批量插入
largeData := make([]*User, 1000)
// ... 填充数据
// 自动分批插入，默认每批 1000 条
err := batch.InsertBatch(users, largeData)

// 流式遍历 (处理海量数据，避免内存溢出)
err := batch.Each(users.Find(), func(u *User) error {
    // 处理每个用户
    return nil
})
```

### 4.3 事务管理 (Transactions)

使用 `tx` 子包管理事务，支持自动回滚。

```go
import "github.com/gocrud/mgo/tx"

err := tx.Transaction(db, func(sess *tx.Session) error {
    // 在事务会话中创建集合引用
    txUsers := mgo.Model[User](sess)
    txOrders := mgo.Model[Order](sess)

    // 操作 1: 扣减余额
    if err := txUsers.Find().ID(uid).Inc("balance", -100).Update(); err != nil {
        return err // 返回错误自动回滚
    }

    // 操作 2: 创建订单
    if _, err := txOrders.Insert(newOrder); err != nil {
        return err // 返回错误自动回滚
    }

    return nil // 返回 nil 自动提交
})
```

## 5. 最佳实践

1.  **Context 传递**: 始终使用 `WithContext` 传递上下文，以便控制超时和取消。
    ```go
    users.WithContext(ctx).Find()...
    ```

2.  **错误处理**: 
    `mgo` 提供了辅助函数来判断常见错误类型。
    ```go
    user, err := users.FindByID(id)
    if err != nil {
        if mgo.IsNoDocuments(err) {
            fmt.Println("用户不存在")
        } else if mgo.IsDuplicateKey(err) {
            fmt.Println("用户已存在 (唯一索引冲突)")
        } else if mgo.IsTimeout(err) {
            fmt.Println("查询超时")
        } else {
            log.Fatal(err)
        }
    }
    ```

3.  **索引管理**: 虽然 `mgo` 简化了查询，但 MongoDB 的性能依赖于索引。请确保为常用查询字段创建索引（可以使用原生驱动或 `mgo` 的辅助方法，如果存在）。

4.  **结构体标签**: 确保结构体字段有正确的 `bson` 标签，特别是 `_id` 字段。

```go
type Model struct {
    ID mgo.ObjectID `bson:"_id,omitempty"`
}
```
