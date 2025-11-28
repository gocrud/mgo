# mgo 使用文档

> `github.com/gocrud/mgo` 基于 MongoDB 官方 Go Driver v2 提供链式、可读性更高的 API，帮助开发者以更少的样板代码完成查询、聚合、更新、事务等日常操作。

## 目录

- [环境要求](#环境要求)
- [安装](#安装)
- [快速上手](#快速上手)
- [核心组件概览](#核心组件概览)
- [常用操作对比](#常用操作对比)
  - [连接与生命周期](#连接与生命周期)
  - [查询](#查询)
  - [更新](#更新)
  - [删除](#删除)
  - [聚合](#聚合)
  - [事务](#事务)
- [软删除功能](#软删除功能)
- [泛型支持](#泛型支持)
- [简化更新构建](#简化更新构建)
- [QueryBuilder 常用方法速查](#querybuilder-常用方法速查)
- [UpdateBuilder 常用方法速查](#updatebuilder-常用方法速查)
- [聚合阶段与表达式速查](#聚合阶段与表达式速查)
- [常见问题](#常见问题)
- [参考资料](#参考资料)
- [许可证](#许可证)

## 环境要求

- Go 1.25 及以上版本（`go.mod` 中指定 `go 1.25.0`）
- MongoDB 5.0 及以上版本（推荐 6.x，以便使用新的聚合阶段）
- 依赖 [`go.mongodb.org/mongo-driver/v2`](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2)

## 安装

```bash
go get github.com/gocrud/mgo
```

## 快速上手

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/gocrud/mgo"
    "go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
    ID        bson.ObjectID `bson:"_id,omitempty"`
    Name      string        `bson:"name"`
    Email     string        `bson:"email"`
    Age       int           `bson:"age"`
    City      string        `bson:"city"`
    CreatedAt time.Time     `bson:"created_at"`
}

func main() {
    ctx := context.Background()

    client, err := mgo.NewClient(ctx, "mongodb://localhost:27017")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    coll := client.Database("quickstart_db").Collection("users")

    // 插入
    coll.InsertOne(ctx, User{Name: "张三", Email: "zhangsan@example.com", Age: 25, City: "北京", CreatedAt: time.Now()})

    // 查询
    var user User
    if err := coll.Query(ctx).Eq("name", "张三").One(&user); err != nil {
        log.Fatal(err)
    }

    // 更新
    coll.Query(ctx).Eq("name", "张三").UpdateOne(mgo.Update().Inc("age", 1))

    // 聚合
    type CityCount struct {
        City  string `bson:"_id"`
        Count int    `bson:"count"`
    }
    var stats []CityCount
    coll.Aggs(ctx).
        Stage(mgo.Stage().Group("$city", mgo.M{"count": mgo.Sum(1)})).
        All(&stats)
}
```

完整示例可参见 `examples/quickstart/main.go`、`examples/complete/main.go` 与 `examples/client/main.go`。

## 核心组件概览

- **Client**：封装连接初始化、`Ping`、数据库访问与事务能力；提供 `NewClient`、`MustNewClient`、`WithTransaction`、`StartTransaction`、`WrapClient` 等方法。
- **Database**：对 `mongo.Database` 的轻量包装，提供 `Collection/Coll`、`ListCollectionNames`、`Drop`、`Native` 等常用操作。支持配置软删除功能。
- **Collection**：核心入口，`Query(ctx)` 返回 `QueryBuilder`，`Aggs(ctx)` 返回 `AggsBuilder`，同时保留 `InsertOne/InsertMany/UpdateByID/DeleteByID/Count/CreateIndex/BulkWrite` 等便捷方法。支持启用软删除。
- **构建器体系**：
  - **FilterBuilder**：查询条件构建器
  - **UpdateBuilder**：更新操作构建器
  - **Projection**：投影字段构建器
  - **Sort**：排序构建器
  - **StageBuilder**：聚合 Stage 构建器（纯构建，可复用）
  - **AggsBuilder**：聚合执行器（执行聚合操作）
  - **Expr**：表达式构建器
- **软删除**：可选功能，启用后 `DeleteOne/DeleteMany` 自动执行软删除，支持 `WithDeleted/OnlyDeleted/Restore/WithHardDelete` 等操作。

## 常用操作对比

### 连接与生命周期

#### 使用 mgo

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

client, err := mgo.NewClient(ctx, uri)
if err != nil {
    return err
}
defer client.Disconnect(ctx)

coll := client.DB("demo").Coll("users")
```

#### 原生 driver

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

clientOpts := options.Client().ApplyURI(uri)
client, err := mongo.Connect(clientOpts)
if err != nil {
    return err
}
if err := client.Ping(ctx, nil); err != nil {
    return err
}
defer client.Disconnect(ctx)

coll := client.Database("demo").Collection("users")
```

### 查询

目标：查询 `status = "active"` 且 `age > 18` 的用户，只返回姓名和邮箱，按创建时间倒序，限制 10 条。

#### 使用 mgo

```go
var users []User
err := coll.Query(ctx).
    Eq("status", "active").
    Gt("age", 18).
    Select("name", "email").
    Desc("created_at").
    Limit(10).
    All(&users)
```

#### 原生 driver

```go
filter := bson.M{
    "status": "active",
    "age": bson.M{"$gt": 18},
}
opts := options.Find().
    SetProjection(bson.M{"name": 1, "email": 1}).
    SetSort(bson.M{"created_at": -1}).
    SetLimit(10)

cursor, err := coll.Find(ctx, filter, opts)
if err != nil {
    return err
}
defer cursor.Close(ctx)

if err := cursor.All(ctx, &users); err != nil {
    return err
}
```

#### 组合条件复用

```go
vipFilter := mgo.Filter().
    Eq("status", "active").
    Or(
        mgo.Filter().Eq("vip", true),
        mgo.Filter().Gt("level", 5),
    )

if err := coll.Query(ctx).Filter(vipFilter).All(&users); err != nil {
    return err
}
```

原生写法需要显式构造嵌套 `bson.D`：

```go
filter := bson.D{
    {Key: "status", Value: "active"},
    {Key: "$or", Value: bson.A{
        bson.D{{Key: "vip", Value: true}},
        bson.D{{Key: "level", Value: bson.D{{Key: "$gt", Value: 5}}}},
    }},
}

cursor, err := coll.Find(ctx, filter)
```

### 更新

目标：将用户状态设置为 `active`，登录次数 +1，并刷新 `updated_at`。

#### 使用 mgo

```go
update := mgo.Update().
    Set("status", "active").
    Inc("login_count", 1).
    CurrentDate("updated_at", false)

if _, err := coll.Query(ctx).Eq("_id", id).UpdateOne(update); err != nil {
    return err
}
```

#### 原生 driver

```go
update := bson.D{
    {Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}},
    {Key: "$inc", Value: bson.D{{Key: "login_count", Value: 1}}},
    {Key: "$currentDate", Value: bson.D{{Key: "updated_at", Value: true}}},
}

if _, err := coll.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
    return err
}
```

更多数组操作（mgo 写法）：

```go
update := mgo.Update().
    Push("tags", "premium").
    PullFilter("orders", mgo.Filter().Lt("expire_at", time.Now())).
    AddToSetEach("roles", "editor", "auditor")

coll.Query(ctx).Eq("account", "demo").UpdateOne(update)
```

### 删除

目标：删除用户文档。

#### 使用 mgo

```go
// 硬删除单条
result, err := coll.Query(ctx).Eq("_id", id).DeleteOne()
if err != nil {
    return err
}
fmt.Printf("删除了 %d 条文档\n", result.DeletedCount)

// 批量删除
result, err = coll.Query(ctx).
    Eq("status", "expired").
    Lt("last_login", expireDate).
    DeleteMany()
```

#### 原生 driver

```go
// 硬删除单条
result, err := coll.DeleteOne(ctx, bson.M{"_id": id})
if err != nil {
    return err
}

// 批量删除
result, err = coll.DeleteMany(ctx, bson.M{
    "status": "expired",
    "last_login": bson.M{"$lt": expireDate},
})
```

**使用软删除**（需要启用，详见[软删除功能](#软删除功能)）：

```go
// 启用软删除
users := db.Collection("users", mgo.WithSoftDelete())

// 软删除（设置 deleted_at 字段）
result, err := users.Query(ctx).Eq("_id", id).DeleteOne()

// 强制硬删除
result, err = users.Query(ctx).
    Eq("_id", id).
    WithHardDelete().
    DeleteOne()
```

### 聚合

目标：统计 `status = "active"` 用户，按城市输出平均年龄和人数，按平均年龄降序。

#### 使用 mgo（StageBuilder）

```go
type CityStats struct {
    City      string  `bson:"_id"`
    AvgAge    float64 `bson:"avg_age"`
    UserCount int     `bson:"user_count"`
}

// 方式 1: 内联构建
var stats []CityStats
err := coll.Aggs(ctx).
    Stage(
        mgo.Stage().
            Match(mgo.Filter().Eq("status", "active")).
            Group("$city", mgo.M{
                "avg_age":    mgo.Avg("$age"),
                "user_count": mgo.Sum(1),
            }).
            SortDesc("avg_age"),
    ).
    All(&stats)

// 方式 2: 先构建后执行（可复用）
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Group("$city", mgo.M{
        "avg_age":    mgo.Avg("$age"),
        "user_count": mgo.Sum(1),
    }).
    SortDesc("avg_age")

err := coll.Aggs(ctx).Stage(pipeline).All(&stats)
```

#### 原生 driver

```go
pipeline := mongo.Pipeline{
    {{Key: "$match", Value: bson.M{"status": "active"}}},
    {{Key: "$group", Value: bson.M{
        "_id":       "$city",
        "avg_age":    bson.M{"$avg": "$age"},
        "user_count": bson.M{"$sum": 1},
    }}},
    {{Key: "$sort", Value: bson.M{"avg_age": -1}}},
}

cursor, err := coll.Aggregate(ctx, pipeline)
if err != nil {
    return err
}
defer cursor.Close(ctx)

if err := cursor.All(ctx, &stats); err != nil {
    return err
}
```

### 事务

#### 使用 mgo

```go
err := client.WithTransaction(ctx, func(txCtx context.Context) error {
    users := client.DB("demo").Coll("users")
    orders := client.DB("demo").Coll("orders")

    if _, err := users.InsertOne(txCtx, user); err != nil {
        return err
    }
    if _, err := orders.InsertOne(txCtx, order); err != nil {
        return err
    }
    return nil
})
```

#### 原生 driver

```go
session, err := client.StartSession()
if err != nil {
    return err
}
defer session.EndSession(ctx)

callback := func(sc mongo.SessionContext) (any, error) {
    if _, err := client.Database("demo").Collection("users").InsertOne(sc, user); err != nil {
        return nil, err
    }
    if _, err := client.Database("demo").Collection("orders").InsertOne(sc, order); err != nil {
        return nil, err
    }
    return nil, nil
}

_, err = session.WithTransaction(ctx, callback)
```

## 软删除功能

软删除允许标记文档为"已删除"而不是物理删除，支持保留历史记录和恢复操作。

### 启用软删除

```go
// 使用默认字段名 "deleted_at"
users := db.Collection("users", mgo.WithSoftDelete())

// 自定义字段名
users := db.Collection("users", mgo.WithSoftDelete("removed_at"))

// 不启用（默认）
users := db.Collection("users")
```

### 删除行为

`DeleteOne()` 和 `DeleteMany()` 的行为取决于是否启用软删除：

| 场景 | DeleteOne/DeleteMany 行为 |
|------|--------------------------|
| 未启用软删除 | 硬删除（物理删除） |
| 启用软删除 | 软删除（设置 deleted_at） |
| 启用软删除 + WithHardDelete() | 硬删除（物理删除） |

### 基本用法

```go
users := db.Collection("users", mgo.WithSoftDelete())

// 1. 软删除（自动）
result, err := users.Query(ctx).Eq("_id", id).DeleteOne()
// 实际执行：UpdateOne 设置 deleted_at 为当前时间

// 2. 批量软删除
result, err = users.Query(ctx).
    Eq("status", "expired").
    DeleteMany()

// 3. 查询（自动排除已删除）
var activeUsers []User
err = users.Query(ctx).Eq("status", "active").All(&activeUsers)
// 自动添加：deleted_at 不存在的条件

// 4. 包含已删除的文档
var allUsers []User
err = users.Query(ctx).WithDeleted().All(&allUsers)

// 5. 仅查询已删除的文档
var deletedUsers []User
err = users.Query(ctx).OnlyDeleted().All(&deletedUsers)

// 6. 恢复软删除的文档
result, err = users.Query(ctx).Eq("_id", id).Restore()
// 移除 deleted_at 字段

// 7. 强制硬删除
result, err = users.Query(ctx).
    Eq("_id", id).
    WithHardDelete().
    DeleteOne()
// 真正从数据库中删除
```

### 清理场景

```go
// 清理已软删除30天的数据
result, err := users.Query(ctx).
    OnlyDeleted().                                      // 仅已删除的
    Lt("deleted_at", time.Now().AddDate(0, 0, -30)).   // 30天前
    WithHardDelete().                                   // 强制硬删除
    DeleteMany()

fmt.Printf("清理了 %d 条数据\n", result.DeletedCount)
```

### 软删除 API

**查询相关**：
- `WithDeleted()` - 包含已软删除的文档
- `OnlyDeleted()` - 仅查询已软删除的文档

**删除相关**：
- `DeleteOne()` / `DeleteMany()` - 根据配置自动选择软删除或硬删除
- `WithHardDelete()` - 强制执行硬删除
- `FindAndDelete(result)` - 查找并删除（遵循软删除逻辑）

**恢复相关**：
- `Restore()` - 恢复软删除的文档（移除 deleted_at 字段）

**索引建议**：

```go
// 创建稀疏索引以优化性能
indexModel := mongo.IndexModel{
    Keys: bson.D{{Key: "deleted_at", Value: 1}},
    Options: options.Index().SetSparse(true),
}
_, err := users.Native().Indexes().CreateOne(ctx, indexModel)
```

更多详情请参考 [SOFT_DELETE.md](SOFT_DELETE.md)。

## 泛型支持

mgo 提供了强大的泛型支持，使代码更加简洁且类型安全。通过 `mgo.Model[T]` 方法，可以将普通集合转换为泛型集合。

### 1. 泛型集合初始化

```go
// 假设有一个 User 结构体
type User struct {
    ID   bson.ObjectID `bson:"_id"`
    Name string        `bson:"name"`
    Age  int           `bson:"age"`
}

// 1. 获取普通集合
coll := client.Database("db").Collection("users")

// 2. 转换为泛型集合
userColl := mgo.Model[User](coll)
```

### 2. 泛型查询 (Type-Safe Query)

使用泛型集合后，`One()` 和 `All()` 方法会直接返回目标类型，无需手动声明变量或传递指针。

```go
// 查询单条 (返回 User, error)
user, err := userColl.Query(ctx).Eq("name", "张三").One()
if err != nil {
    log.Fatal(err)
}
fmt.Println(user.Name)

// 查询列表 (返回 []User, error)
users, err := userColl.Query(ctx).Gte("age", 18).All()
for _, u := range users {
    fmt.Println(u.Name)
}

// 查找并更新 (返回 User, error)
updatedUser, err := userColl.Query(ctx).
    Eq("_id", id).
    FindAndUpdate(mgo.Set("age", 26))
```

## 简化更新构建

除了功能强大的 `UpdateBuilder`，mgo 现在支持更简洁的更新方式，允许直接使用 `map`、`bson.M` 或快捷函数。

### 1. 使用 mgo.Set 快捷函数

对于最常见的 `$set` 操作，提供了 `mgo.Set` 快捷函数：

```go
// 简单更新
coll.Query(ctx).Eq("_id", id).UpdateOne(mgo.Set("status", "active"))
```

### 2. 直接使用 map

支持直接传入 `map[string]any` 或 `bson.M`，使构建复杂更新更加灵活：

```go
// 混合操作符
update := bson.M{
    "$set": bson.M{"status": "active"},
    "$inc": bson.M{"login_count": 1},
}
coll.Query(ctx).Eq("_id", id).UpdateOne(update)
```

### 3. 泛型更新

泛型集合的 `FindAndUpdate`、`FindAndReplace` 等方法同样支持上述简化写法，并直接返回强类型的文档对象。

## QueryBuilder 常用方法速查

| 方法 | 描述 | MongoDB 原始写法 |
| --- | --- | --- |
| `Eq("field", value)` | 等于条件 | `{field: value}` |
| `Gt/Gte/Lt/Lte` | 数值比较 | `{field: {$gt: value}}` 等 |
| `Between("field", min, max)` | 范围（含端点） | `{field: {$gte: min, $lte: max}}` |
| `In("field", values...)` / `Nin` | 集合匹配 | `{field: {$in: [...]}}` |
| `Contains/StartsWith/EndsWith` | 模糊匹配（不区分大小写） | `{field: {$regex: pattern, $options: "i"}}` |
| `Exists/NotExists` | 字段存在性 | `{field: {$exists: true/false}}` |
| `And/Or/Nor` | 逻辑组合 | `{$and: [...]}` 等 |
| `Select("a", "b")` | 设置返回字段 | `projection: {a: 1, b: 1}` |
| `Omit("password")` | 排除字段 | `projection: {password: 0}` |
| `Asc/Desc` | 排序 | `sort: {field: 1/-1}` |
| `Limit/Skip/Page` | 分页控制 | `limit`/`skip` 选项 |
| `FindAndUpdate`/`FindAndReplace`/`FindAndDelete` | 查找并修改 | `FindOneAndUpdate`/`Replace`/`Delete` |
| `UpdateOne`/`UpdateMany` | 查询 + 更新 | `UpdateOne`/`UpdateMany` |
| `DeleteOne`/`DeleteMany` | 查询 + 删除 | `DeleteOne`/`DeleteMany` |
| `WithDeleted()` | 包含已软删除的文档 | 查询包含 deleted_at 的文档 |
| `OnlyDeleted()` | 仅查询已软删除的文档 | 仅查询 deleted_at 存在的文档 |
| `WithHardDelete()` | 强制硬删除 | 真正删除而非软删除 |
| `Restore()` | 恢复软删除的文档 | 移除 deleted_at 字段 |

`Hint`、`BatchSize`、`Cursor` 等方法与官方驱动保持一致，可用于索引优化与流式处理。

## UpdateBuilder 常用方法速查

| 方法 | 描述 | MongoDB 原始写法 |
| --- | --- | --- |
| `Set("field", value)` | 设置字段 | `{$set: {field: value}}` |
| `Unset("field")` | 删除字段 | `{$unset: {field: ""}}` |
| `Inc("field", n)` | 自增/自减 | `{$inc: {field: n}}` |
| `Mul("field", factor)` | 乘法更新 | `{$mul: {field: factor}}` |
| `Min/Max` | 最小值/最大值约束 | `{$min: {...}}` / `{$max: {...}}` |
| `CurrentDate("field", true/false)` | 写入当前时间或 timestamp | `{$currentDate: {field: true}}` |
| `SetOnInsert` | upsert 时设置初值 | `{$setOnInsert: {...}}` |
| `Push` / `PushEach` / `PushSlice` / `PushPosition` | 数组追加、截断或指定位置插入 | `{$push: {...}}` |
| `Pull` / `PullAll` / `PullFilter` | 数组删除 | `{$pull: {...}}` / `{$pullAll: {...}}` |
| `AddToSet` / `AddToSetEach` | 数组去重追加 | `{$addToSet: {...}}` |
| `Bit("field", "or", value)` | 位运算 | `{$bit: {field: {or: value}}}` |

`Build()` 返回 `bson.D`，`BuildM()` 返回 `bson.M`，在需要原生 API 时可以直接复用。

## 聚合阶段与表达式速查

### AggsBuilder 与 StageBuilder

聚合 API 采用**职责分离**设计：

- **`StageBuilder`**：纯粹的 stage 构建器，用于构建聚合管道
- **`AggsBuilder`**：执行器，负责执行聚合操作

#### 基本使用

```go
// 创建 StageBuilder
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Group("$city", mgo.M{
        "count": mgo.Sum(1),
        "avgAge": mgo.Avg("$age"),
    }).
    SortDesc("count").
    Limit(10)

// 执行聚合
var results []Result
err := coll.Aggs(ctx).Stage(pipeline).All(&results)
```

#### StageBuilder 支持的方法

**组合方法**：
- `Clone()` - 克隆 StageBuilder（用于复用）
- `Append(other)` - 追加另一个 StageBuilder
- `Prepend(other)` - 前置另一个 StageBuilder
- `AddStage(bson.D)` - 添加自定义 stage
- `AddStages(...bson.D)` - 添加多个自定义 stages

**Stage 构建方法**：
- `Match(filter)` / `MatchDoc(M)` - 匹配阶段
- `Project(projection)` / `ProjectDoc(M)` - 投影阶段
- `Group(id, accumulators)` / `GroupBy(field, accumulators)` - 分组阶段
- `Sort(field, direction)` / `SortBy(sort)` / `SortDoc(M)` / `SortAsc(...)` / `SortDesc(...)` - 排序
- `Limit(n)` / `Skip(n)` / `Page(page, size)` - 分页
- `Unwind(path)` / `UnwindPreserveEmpty(path)` - 展开数组
- `Lookup(from, local, foreign, as)` / `LookupStage(from, as, let, pipeline)` - 关联查询
- `AddFields(M)` / `ReplaceRoot(newRoot)` - 字段操作
- `Count(field)` / `Sample(size)` - 计数、抽样
- `Facet(facets)` / `Bucket(...)` - 多面查询、分桶
- `Out(collection)` / `Merge(...)` - 输出到集合
- `GeoNear(...)` / `Redact(...)` - 地理位置、条件过滤

**累加器函数**：
- `Sum(expr)` - 求和/计数
- `Avg(expr)` - 平均值
- `Max(expr)` / `Min(expr)` - 最大/最小值
- `First(expr)` / `Last(expr)` - 第一个/最后一个值
- `Push(expr)` / `AddToSet(expr)` - 数组追加/去重追加
- `StdDevPop(expr)` / `StdDevSamp(expr)` - 标准差

#### AggsBuilder 执行方法

```go
// 设置 pipeline
.Stage(stageBuilder)          // 使用 StageBuilder
.Pipes([]bson.D)              // 使用原始 stages

// 追加 pipeline
.AppendStage(stageBuilder)    // 追加 StageBuilder
.AppendPipes([]bson.D)        // 追加原始 stages
.AddPipe(bson.D)              // 添加单个 stage

// 执行
.One(&result)                 // 返回单条结果
.All(&results)                // 返回所有结果
.Cursor()                     // 获取游标
.Build()                      // 构建原始 pipeline
```

#### 高级用法

**1. 复用 Pipeline**：

```go
// 定义可复用的 stage 组合
activeUsers := mgo.Stage().Match(mgo.Filter().Eq("status", "active"))
recentSort := mgo.Stage().SortDesc("created_at")

// 组合使用
pipeline := mgo.Stage().
    Append(activeUsers).
    Append(recentSort).
    Limit(10)

err := coll.Aggs(ctx).Stage(pipeline).All(&results)
```

**2. Clone 用于并行查询**：

```go
base := mgo.Stage().Match(mgo.Filter().Eq("status", "active"))

// 查询1：前10条
pipeline1 := base.Clone().Limit(10)
coll.Aggs(ctx).Stage(pipeline1).All(&results1)

// 查询2：前20条
pipeline2 := base.Clone().Limit(20)
coll.Aggs(ctx).Stage(pipeline2).All(&results2)
```

**3. Match 条件过滤**：

```go
// 单条件
pipeline := mgo.Stage().Match(mgo.Filter().Eq("status", "active"))

// 多条件组合
pipeline := mgo.Stage().Match(
    mgo.Filter().
        Eq("status", "active").
        Gt("age", 18).
        In("city", "北京", "上海", "深圳"),
)

// 使用文档方式
pipeline := mgo.Stage().MatchDoc(mgo.M{
    "status": "active",
    "age": mgo.M{"$gte": 18},
})
```

**4. Group 聚合统计**：

```go
// 按字段分组
pipeline := mgo.Stage().GroupBy("city", mgo.M{
    "count":      mgo.Sum(1),
    "avgAge":     mgo.Avg("$age"),
    "maxSalary":  mgo.Max("$salary"),
    "minSalary":  mgo.Min("$salary"),
    "totalSales": mgo.Sum("$sales"),
})

// 按多字段分组
pipeline := mgo.Stage().Group(
    mgo.M{"city": "$city", "status": "$status"},
    mgo.M{"count": mgo.Sum(1)},
)

// 数组聚合
pipeline := mgo.Stage().GroupBy("department", mgo.M{
    "employees": mgo.Push("$name"),           // 收集所有名字
    "skills":    mgo.AddToSet("$skill"),      // 去重收集技能
    "firstHire": mgo.First("$hire_date"),     // 第一个入职日期
    "lastHire":  mgo.Last("$hire_date"),      // 最后一个入职日期
})
```

**5. Sort 排序**：

```go
// 单字段排序
pipeline := mgo.Stage().Sort("created_at", -1)  // 降序
pipeline := mgo.Stage().Sort("age", 1)          // 升序

// 多字段排序
pipeline := mgo.Stage().
    SortBy(mgo.NewSort().Desc("priority").Asc("name"))

// 快捷方法
pipeline := mgo.Stage().SortDesc("created_at", "updated_at")
pipeline := mgo.Stage().SortAsc("name", "age")

// 使用文档
pipeline := mgo.Stage().SortDoc(mgo.M{
    "priority": -1,
    "name":     1,
})
```

**6. Project 投影**：

```go
// 选择字段
pipeline := mgo.Stage().Project(
    mgo.NewProjection().
        Include("name", "email", "age").
        Exclude("_id"),
)

// 计算字段
pipeline := mgo.Stage().ProjectDoc(mgo.M{
    "name":     1,
    "age":      1,
    "isAdult":  mgo.M{"$gte": []any{"$age", 18}},
    "fullName": mgo.M{"$concat": []any{"$firstName", " ", "$lastName"}},
})

// 嵌套字段
pipeline := mgo.Stage().Project(
    mgo.NewProjection().
        Include("name").
        IncludeDot("address.city", "address.country"),
)
```

**7. Unwind 展开数组**：

```go
// 基本展开
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Unwind("$tags").  // 展开 tags 数组
    GroupBy("tags", mgo.M{"count": mgo.Sum(1)})

// 保留空数组
pipeline := mgo.Stage().
    UnwindPreserveEmpty("$orders").  // 即使 orders 为空也保留文档
    Match(mgo.Filter().Exists("orders"))
```

**8. Lookup 关联查询**：

```go
// 简单关联
pipeline := mgo.Stage().
    Lookup("orders", "user_id", "_id", "user_orders").
    Unwind("$user_orders")

// 带条件的关联（使用嵌套 pipeline）
pipeline := mgo.Stage().
    LookupStage("orders", "user_orders",
        mgo.M{"userId": "$_id"},
        mgo.Stage().
            Match(mgo.Filter().
                Eq("status", "completed").
                Gte("total", 100),
            ).
            SortDesc("created_at").
            Limit(5),
    )

// 多层关联
pipeline := mgo.Stage().
    Lookup("orders", "user_id", "_id", "orders").
    Unwind("$orders").
    LookupStage("products", "order_products",
        mgo.M{"orderId": "$orders._id"},
        mgo.Stage().Match(mgo.Filter().Eq("available", true)),
    )
```

**9. AddFields 添加字段**：

```go
// 添加计算字段
pipeline := mgo.Stage().AddFields(mgo.M{
    "fullName": mgo.M{"$concat": []any{"$firstName", " ", "$lastName"}},
    "isVIP":    mgo.M{"$gte": []any{"$points", 1000}},
    "discount": mgo.M{"$multiply": []any{"$price", 0.9}},
})

// 添加日期字段
pipeline := mgo.Stage().AddFields(mgo.M{
    "year":  mgo.M{"$year": "$created_at"},
    "month": mgo.M{"$month": "$created_at"},
    "age":   mgo.M{"$subtract": []any{2024, mgo.M{"$year": "$birth_date"}}},
})
```

**10. Facet 多维度统计**：

```go
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Facet(map[string]*mgo.StageBuilder{
        // 按城市统计
        "byCity": mgo.Stage().
            GroupBy("city", mgo.M{"count": mgo.Sum(1)}).
            SortDesc("count").
            Limit(5),
        
        // 按年龄段统计
        "byAge": mgo.Stage().
            Bucket("$age", 
                []any{0, 18, 30, 50, 100},
                "其他",
                mgo.M{"count": mgo.Sum(1), "avgSalary": mgo.Avg("$salary")},
            ),
        
        // 总体统计
        "summary": mgo.Stage().
            Group(nil, mgo.M{
                "total":      mgo.Sum(1),
                "avgAge":     mgo.Avg("$age"),
                "totalSales": mgo.Sum("$sales"),
            }),
    })

var result struct {
    ByCity  []CityStats  `bson:"byCity"`
    ByAge   []AgeStats   `bson:"byAge"`
    Summary SummaryStats `bson:"summary"`
}
err := coll.Aggs(ctx).Stage(pipeline).One(&result)
```

**11. Page 分页**：

```go
// 基本分页
page := 2
pageSize := 20
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    SortDesc("created_at").
    Page(page, pageSize)  // 自动计算 skip 和 limit

// 手动分页
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    SortDesc("created_at").
    Skip((page - 1) * pageSize).
    Limit(pageSize)
```

**12. Sample 随机抽样**：

```go
// 随机抽取10条记录
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Sample(10)

err := coll.Aggs(ctx).Stage(pipeline).All(&results)
```

**13. Count 计数**：

```go
// 统计符合条件的文档数
pipeline := mgo.Stage().
    Match(mgo.Filter().Eq("status", "active")).
    Count("total")

var result struct {
    Total int `bson:"total"`
}
err := coll.Aggs(ctx).Stage(pipeline).One(&result)
```

**14. Bucket 分桶统计**：

```go
// 按年龄段分桶
pipeline := mgo.Stage().
    Match(mgo.Filter().Exists("age")).
    Bucket("$age",
        []any{0, 18, 30, 45, 60, 100},  // 边界值
        "其他",  // 默认桶名
        mgo.M{
            "count":     mgo.Sum(1),
            "avgSalary": mgo.Avg("$salary"),
            "names":     mgo.Push("$name"),
        },
    )

type AgeBucket struct {
    ID         int      `bson:"_id"`
    Count      int      `bson:"count"`
    AvgSalary  float64  `bson:"avgSalary"`
    Names      []string `bson:"names"`
}
var buckets []AgeBucket
err := coll.Aggs(ctx).Stage(pipeline).All(&buckets)
```

**15. 组合使用（完整示例）**：

```go
// 复杂的业务场景：统计活跃用户的订单情况
pipeline := mgo.Stage().
    // 1. 筛选活跃用户
    Match(mgo.Filter().
        Eq("status", "active").
        Gte("last_login", time.Now().AddDate(0, -1, 0)),
    ).
    // 2. 关联订单
    LookupStage("orders", "user_orders",
        mgo.M{"userId": "$_id"},
        mgo.Stage().
            Match(mgo.Filter().Eq("status", "completed")).
            SortDesc("created_at"),
    ).
    // 3. 添加计算字段
    AddFields(mgo.M{
        "orderCount": mgo.M{"$size": "$user_orders"},
        "totalAmount": mgo.M{
            "$sum": "$user_orders.amount",
        },
    }).
    // 4. 筛选有订单的用户
    Match(mgo.Filter().Gt("orderCount", 0)).
    // 5. 按城市分组统计
    GroupBy("city", mgo.M{
        "userCount":     mgo.Sum(1),
        "avgOrderCount": mgo.Avg("$orderCount"),
        "totalRevenue":  mgo.Sum("$totalAmount"),
        "topUsers":      mgo.Push(mgo.M{
            "name":        "$name",
            "orderCount":  "$orderCount",
            "totalAmount": "$totalAmount",
        }),
    }).
    // 6. 按收入排序
    SortDesc("totalRevenue").
    // 7. 限制返回前10个城市
    Limit(10)

type CityReport struct {
    City          string  `bson:"_id"`
    UserCount     int     `bson:"userCount"`
    AvgOrderCount float64 `bson:"avgOrderCount"`
    TotalRevenue  float64 `bson:"totalRevenue"`
    TopUsers      []struct {
        Name        string  `bson:"name"`
        OrderCount  int     `bson:"orderCount"`
        TotalAmount float64 `bson:"totalAmount"`
    } `bson:"topUsers"`
}

var reports []CityReport
err := coll.Aggs(ctx).Stage(pipeline).All(&reports)
```

**16. 动态构建 Pipeline**：

```go
// 根据条件动态添加 stages
pipeline := mgo.Stage()

// 基础筛选
if status != "" {
    pipeline = pipeline.Match(mgo.Filter().Eq("status", status))
}

// 可选的时间范围
if !startDate.IsZero() {
    pipeline = pipeline.Match(mgo.Filter().Gte("created_at", startDate))
}
if !endDate.IsZero() {
    pipeline = pipeline.Match(mgo.Filter().Lte("created_at", endDate))
}

// 可选的城市筛选
if len(cities) > 0 {
    pipeline = pipeline.Match(mgo.Filter().In("city", cities...))
}

// 分组统计
pipeline = pipeline.GroupBy("category", mgo.M{
    "count": mgo.Sum(1),
    "total": mgo.Sum("$amount"),
})

// 排序和分页
if sortField != "" {
    pipeline = pipeline.Sort(sortField, sortOrder)
}
if pageSize > 0 {
    pipeline = pipeline.Page(page, pageSize)
}

err := coll.Aggs(ctx).Stage(pipeline).All(&results)
```

### 表达式构建器

`mgo.F("field")` 创建字段引用，可链式 `.Dot()` / `.Index()`；`mgo.Exp` 提供算术、逻辑、条件、字符串、数组、日期等丰富表达式。

示例：

```go
proj := mgo.NewProjection().
    Include("name").
    SetExpr("age_group", mgo.Exp.Cond(
        mgo.Exp.Gte(mgo.F("age"), 30),
        "adult",
        "youth",
    ))

pipeline := mgo.Stage().Project(proj)
coll.Aggs(ctx).Stage(pipeline).All(&docs)
```

## 常见问题

- **如何启用认证或 TLS？**
  ```go
  cred := options.Credential{Username: "app", Password: "secret"}
  client, err := mgo.NewClient(ctx, uri, options.Client().SetAuth(cred))
  ```
- **已有 `mongo.Client` 如何复用？**使用 `mgo.WrapClient(existingClient)`。
- **如何执行原生操作？**调用 `collection.Native()`、`database.Native()` 或 `client.Native()` 取得底层类型。
- **如何控制超时和取消？**使用 `context.WithTimeout` / `context.WithCancel` 对上下文进行包裹,再传入 `Query`/`Aggs` 等方法。
- **如何调试生成的查询？**`FilterBuilder.Build()`、`UpdateBuilder.Build()`、`AggsBuilder.Build()` 均可返回原生 `bson` 文档进行打印或测试。
- **软删除是否影响性能？** 软删除会在查询时自动添加 `deleted_at` 过滤条件。建议为该字段创建索引以确保查询性能。对于大量数据，定期清理已删除数据可以优化存储空间。
- **如何迁移现有数据以支持软删除？** 现有数据无需修改即可使用软删除功能。未被软删除的文档不包含 `deleted_at` 字段（或该字段为 `nil`），会被正常查询。只有执行 `DeleteOne`/`DeleteMany` 后才会添加删除时间戳。
- **软删除对聚合查询有什么影响？** 聚合查询（`Aggs`）也会自动应用软删除过滤。如需在聚合中包含已删除文档，使用 `WithDeleted()` 方法。
- **何时应该使用软删除？** 软删除适用于需要审计追踪、支持数据恢复、或希望保留历史记录的场景。如果数据确实不再需要且不需要恢复能力，使用 `WithHardDelete()` 执行硬删除会更高效。
- **软删除的文档如何永久清理？** 使用 `OnlyDeleted().WithHardDelete().DeleteMany()` 可以永久删除已软删除的文档。建议定期执行清理任务，删除超过保留期限的已删除数据。

## 参考资料

- MongoDB Go Driver 文档：https://www.mongodb.com/docs/drivers/go/current/
- 项目示例：`examples/` 目录
- 单元测试：`document_test.go`、`filter_test.go`、`transaction_test.go`、`update_test.go`
