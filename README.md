# mgo: 极致 DX 的 Go MongoDB 库

`mgo` 是一个基于官方 `go.mongodb.org/mongo-driver/v2` 的轻量级封装库。它旨在提供极致的开发者体验 (DX)，通过链式调用、高性能的游标分页，让 MongoDB 的操作变得简单、类型安全且高效。

## 核心特性 (Features)

*   **基于 Driver V2**: 充分利用官方 v2 版本的性能优化（如 `bson.D` 默认解码、更高效的连接池管理）。
*   **链式调用 (Chainable API)**: 采用 Builder 模式，操作逻辑清晰流畅（`Find().Where().SortAsc().All()`）。
*   **操作分离**: 独立的 `Find`, `Insert`, `Update`, `Delete` 构建器，职责单一。
*   **高性能设计**: 底层全链路采用 `bson.D` (Slice) 而非 `bson.M` (Map)，大幅减少内存分配与 GC 压力。
*   **高性能分页**:
    *   **Offset Pagination**: 传统的 `Paginate(page, size)`，自动处理 Count 和 Skip。
    *   **Cursor Pagination**: 基于 `Seek()` 的游标分页，利用索引实现海量数据的高性能无限滚动。
*   **类型别名 (Type Aliases)**: 直接导出 `mgo.D`, `mgo.M`, `mgo.A`, `mgo.E`，无需引入 `go.mongodb.org/mongo-driver/bson`。
*   **开发辅助**: 内置 UTC 时间转换、时间范围生成等实用工具。

## 安装 (Installation)

```bash
go get github.com/gocrud/mgo
```

## 快速开始 (Quick Start)

### 1. 定义模型

```go
import (
    "time"
    "github.com/gocrud/mgo"
)

type User struct {
    ID        string    `bson:"_id,omitempty"`
    Name      string    `bson:"name"`
    Age       int       `bson:"age"`
    Role      string    `bson:"role"`
    CreatedAt time.Time `bson:"created_at"`
}
```

### 2. 初始化连接与集合

```go
func main() {
    // 1. 连接数据库 (mgo.Connect 封装了默认超时配置)
    // 返回 *mgo.Client (包装器，内嵌原生 Client)
    cli, err := mgo.Connect("mongodb://localhost:27017")
    if err != nil {
        panic(err)
    }

    // 2. 获取 Database 对象
    db := cli.Database("my_db")

    // 3. 创建 Collection
    // 传入 db 实例
    users := db.Collection("users")
}
```

## CRUD 操作示例

### 插入 (Insert)

利用 v2 特性，直接支持切片插入，无需 `[]any` 转换。

```go
ctx := context.Background()

// 单条插入
users.Insert().Doc(&User{Name: "Alice", Age: 20}).One()

// 批量插入
users.Insert().
    Ctx(ctx).
    Docs(
        &User{Name: "Bob", Age: 22},
        &User{Name: "Charlie", Age: 23},
    ).
    Many()
```

### 查询 (Find)

```go
// Where 接受变长参数 ...bson.E，内部直接构建 bson.D，零 Map 分配
var list []User
err := users.Find().
    Where(
        mgo.Gt("age", 18),
        mgo.Eq("status", "active"),
    ).
    SortDesc("created_at"). // 明确的排序方向
    Limit(10).
    All(&list) // 传入切片指针
```

### 更新 (Update)

提供 `Set`, `Inc` 等快捷方法，无需手动构建 `$set` map。

```go
// 将名为 Alice 的用户年龄改为 25，并增加登录次数
err := users.Update().
    Where(mgo.Eq("name", "Alice")).
    Set("age", 25).
    Inc("login_count", 1).
    One()
```

### 删除 (Delete)

```go
res, err := users.Delete().
    Where(mgo.Lt("age", 10)).
    Many()
```

## 高级特性

### 分页方案 (Pagination)

#### 方案 A: 传统分页 (Page/Size)
适用于后台管理列表，数据量中等。

```go
// 查询第 2 页，每页 20 条
var list []User
pageRes, err := users.Find().
    Where(mgo.F("role").Eq("admin")).
    SortDesc("created_at").
    Paginate(2, 20).
    All(&list)

fmt.Printf("总数: %d, 总页数: %d\n", pageRes.Total, pageRes.TotalPages)
for _, u := range list {
    fmt.Println(u.Name)
}
```

#### 方案 B: 游标分页 (Cursor/Seek) - **高性能推荐**
适用于移动端 Feed 流、无限滚动，数据量大。利用索引查找，无 `Skip` 开销。

```go
// 假设 lastUser 是上一页最后一条数据
var lastUser User 
// ... (获取 lastUser)

query := users.Find().
    Where(mgo.Eq("role","user")).
    SortDesc("created_at"). // 必须指定排序，Seek 依赖此方向
    Limit(20)

// 如果不是第一页，传入上一条记录对象
if lastUser.ID != "" {
    // 库会自动根据 SortDesc("created_at") 生成 { created_at: { $lt: lastUser.CreatedAt } }
    // 支持单字段排序的自动推断
    query.Seek(lastUser)
}

var list []User
err := query.All(&list)
```

### 时间辅助函数 (Time Helpers)

解决 MongoDB 存储 UTC 时间导致的查询转换痛点。

```go
// 查询“今天”注册的所有用户
// mgo.DayRange 自动返回当天的 UTC 起止时间 (00:00:00 - 23:59:59)
start, end := mgo.DayRange(time.Now())

var list []User
err := users.Find().
    Where(
        mgo.Gte("created_at", start),
        mgo.Lt("created_at", end),
    ).
    All(&list)
```

## 聚合 (Aggregation)

`mgo` 提供了一套符合直觉的链式聚合 API，并通过 `agg` 子包提供类型安全的累加器支持。

### 1. 分组统计 (Group)

```go
import (
    "github.com/gocrud/mgo"
    "github.com/gocrud/mgo/agg" // 引入聚合子包
)

type RoleStat struct {
    Role      string  `bson:"_id"`
    Count     int     `bson:"count"`
    AvgAge    float64 `bson:"avg_age"`
}

var results []RoleStat

err := users.Aggregate().
    // 1. 筛选: 复用 mgo.Op，保持习惯一致
    Match(
        mgo.Eq("status", "active"),
        mgo.Gt("age", 18),
    ).
    // 2. 分组: 使用 agg 子包辅助函数
    Group(
        "$role", // _id: 按 role 字段分组
        agg.Count("count"),
        agg.Avg("avg_age", "$age"),
    ).
    // 3. 排序
    SortDesc("count").
    All(&results)
```

### 2. 多表关联 (Join & Lookup)

`mgo` 提供了 `Join` 系列快捷方法，针对最常见的 **1:1 关联 (Left Join)** 场景进行了极致简化（自动关联 `_id` 并 `Unwind`）。

```go
type OrderWithUser struct {
    OrderID string `bson:"_id"`
    Amount  int    `bson:"amount"`
    User    *User  `bson:"user_info"` // Join 自动 Unwind，直接映射为结构体指针
}

var orders []OrderWithUser

err := db.Collection("orders").Aggregate().
    Match(mgo.Gte("amount", 100)).
    // ✅ 推荐: 快捷方式 (1:1)
    // 自动关联 users._id，并执行 Unwind (Left Join)
    Join("users", "user_id", "user_info").
    All(&orders)

// 💡 场景: 自定义外键 (1:1)
// JoinBy("products", "sku_code", "code", "product_detail")

// 💡 场景: 原始 Lookup (1:N)
// 当你需要获取数组结果时使用
// Lookup("comments", "_id", "post_id", "comments_list")
```

### 3. 混合原生管道 (Pipeline Escape Hatch)

当遇到 `mgo` 尚未封装的高级操作（如 `$bucket`, `$facet`）时，使用 `.Pipeline()` 混合注入原生 BSON。
`mgo` 提供了 `mgo.D`, `mgo.A` 等类型别名，无需引入官方 bson 包。

```go
err := products.Aggregate().
    Match(mgo.Eq("category", "electronics")).
    // 插入原生管道处理复杂逻辑
    Pipeline(mgo.A{
        mgo.D{{"$bucket", mgo.D{
            {"groupBy", "$price"},
            {"boundaries", mgo.A{0, 100, 500, 1000}},
            {"default", "expensive"},
            {"output", mgo.D{
                {"count", mgo.D{{"$sum", 1}}},
            }},
        }}},
    }).
    All(&results)
```

### 4. 聚合分页与单条查询

#### 聚合分页 (Paginate)

`mgo` 的聚合分页采用 **Count + Query** 策略，自动处理总数统计和分页逻辑。

```go
type OrderDetail struct {
    OrderID string `bson:"_id"`
    Amount  int    `bson:"amount"`
    User    *User  `bson:"user_info"`
}

var list []OrderDetail

// Paginate(page, size)
// 返回 *PaginatedResult (Total, Page, Size)
res, err := orders.Aggregate().
    Match(mgo.Eq("status", "paid")).
    Join("users", "user_id", "user_info").
    SortDesc("amount").
    Paginate(1, 20). // 第 1 页，每页 20 条
    All(&list)       // 将当前页数据解码到 list

if err == nil {
    fmt.Printf("Total: %d, List Size: %d\n", res.Total, len(list))
}
```

#### 单条查询 (One)

用于统计结果或获取特定详情，自动追加 `$limit: 1` 并解码为单个对象。

```go
type StatResult struct {
    AvgPrice   float64 `bson:"avg_price"`
    TotalStock int     `bson:"total_stock"`
}

var stat StatResult

err := products.Aggregate().
    Match(mgo.Eq("category", "electronics")).
    Group(
        nil, // _id: null (全局统计)
        agg.Avg("avg_price", "$price"),
        agg.Sum("total_stock", "$stock"),
    ).
    One(&stat) // 直接解码为结构体
```

## 自动化时间管理 (Auto Time Management)

`mgo` 提供了一套零样板代码的方案，自动维护 `created_at` 和 `updated_at`。

### 1. 开启自动化

在初始化集合时，调用 `.AutoTime()`。

```go
// 开启后：
// 1. Insert 操作会自动检测 TimeHook 接口并填充时间
// 2. Update 操作会自动注入 $set: { updated_at: now }
users := db.Collection("users").AutoTime()
```

### 2. 内嵌公共模型

在你的结构体中内嵌 `mgo.TimeFields`，它已实现了 `TimeHook` 接口。

```go
type User struct {
    ID             string         `bson:"_id,omitempty"`
    Name           string         `bson:"name"`
    mgo.TimeFields `bson:",inline"` // ✅ 内嵌，自动获得时间管理能力
}
```

### 3. 使用效果

```go
// Insert: 自动填充 CreatedAt 和 UpdatedAt
users.Insert().Doc(&User{Name: "Alice"}).One()

// Update: 自动更新 UpdatedAt
err := users.Update().
    Where(mgo.Eq("name", "Alice")).
    Set("age", 26).
    One() // 实际执行: { $set: { age: 26, updated_at: <now> } }
```

## 软删除 (Soft Delete)

`mgo` 支持无侵入式的软删除，通过标记 `deleted_at` 字段来替代物理删除。

### 1. 开启软删除

```go
// 开启后：
// 1. Delete 操作变为 Update 设置 deleted_at
// 2. Find/Count/Paginate 操作自动过滤 deleted_at != null
users := db.Collection("users").SoftDelete()
```

### 2. 内嵌模型

```go
type User struct {
    ID             string         `bson:"_id,omitempty"`
    mgo.SoftDelete `bson:",inline"` // ✅ 内嵌，提供 DeletedAt 字段
}
```

### 3. 使用效果

```go
// 逻辑删除 (Soft Delete)
// 实际执行: UPDATE users SET deleted_at = <now> WHERE ...
res, err := users.Delete().Where(mgo.Eq("name", "Alice")).Many()

// 物理删除 (Hard Delete)
// 实际执行: DELETE FROM users WHERE ...
res, err = users.Delete().Where(mgo.Eq("name", "Bob")).Hard().Many()

// 查询 (自动过滤已删除)
var list []User
err = users.Find().All(&list)

// 查询包含已删除 (Unscoped)
err = users.Find().Unscoped().All(&list)

// 恢复数据 (Restore)
// 自动隐含 Unscoped，并将 deleted_at 置为 null
err = users.Update().
    Where(mgo.Eq("name", "Alice")).
    Restore().
    One()
```

## 事务 (Transactions)

`mgo` 提供了闭包式的事务支持，自动处理重试和提交/回滚。

### 使用 `client.Tx`

```go
func Transfer(client *mgo.Client, users *mgo.Collection, fromID, toID string, amount int) error {
    // 开启事务
    return client.Tx(context.Background(), func(txCtx context.Context) error {
        
        // 1. 扣款 (A)
        // ⚠️ 关键: 必须调用 .Tx(txCtx) 绑定事务上下文
        if err := users.Update().
            Tx(txCtx). 
            Where(mgo.Eq("_id", fromID)).
            Inc("balance", -amount).
            One(); err != nil {
            return err // 返回错误，自动回滚
        }

        // 2. 加款 (B)
        if err := users.Update().
            Tx(txCtx).
            Where(mgo.Eq("_id", toID)).
            Inc("balance", amount).
            One(); err != nil {
            return err
        }

        return nil // 返回 nil，自动提交
    })
}
```
