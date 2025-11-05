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
  - [聚合](#聚合)
  - [事务](#事务)
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
        Group("$city", mgo.M{"count": mgo.Sum(1)}).
        All(&stats)
}
```

完整示例可参见 `examples/quickstart/main.go`、`examples/complete/main.go` 与 `examples/client/main.go`。

## 核心组件概览

- **Client**：封装连接初始化、`Ping`、数据库访问与事务能力；提供 `NewClient`、`MustNewClient`、`WithTransaction`、`StartTransaction`、`WrapClient` 等方法。
- **Database**：对 `mongo.Database` 的轻量包装，提供 `Collection/Coll`、`ListCollectionNames`、`Drop`、`Native` 等常用操作。
- **Collection**：核心入口，`Query(ctx)` 返回 `QueryBuilder`，`Aggs(ctx)` 返回 `AggsBuilder`，同时保留 `InsertOne/InsertMany/UpdateByID/DeleteByID/Count/CreateIndex/BulkWrite` 等便捷方法。
- **构建器体系**：`FilterBuilder`、`UpdateBuilder`、`Projection`、`Sort`、`AggsBuilder` 与 `Expr` 提供链式 API，避免手写 `$gt`、`$push` 等原始操作符。

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

### 聚合

目标：统计 `status = "active"` 用户，按城市输出平均年龄和人数，按平均年龄降序。

#### 使用 mgo

```go
type CityStats struct {
    City      string  `bson:"_id"`
    AvgAge    float64 `bson:"avg_age"`
    UserCount int     `bson:"user_count"`
}

var stats []CityStats
err := coll.Aggs(ctx).
    Match(mgo.Filter().Eq("status", "active")).
    Group("$city", mgo.M{
        "avg_age":    mgo.Avg("$age"),
        "user_count": mgo.Sum(1),
    }).
    SortDesc("avg_age").
    All(&stats)
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

- `AggsBuilder` 支持 `Match`、`Project`、`Group`、`Sort`、`Limit`、`Skip`、`Page`、`Unwind`、`Lookup`、`AddFields`、`ReplaceRoot`、`Facet`、`Bucket`、`Sample`、`GeoNear`、`Merge`、`Out`、`AddStage` 等常规阶段。
- 累加器快捷方法：`Sum`、`Avg`、`Min`、`Max`、`First`、`Last`、`Push`、`AddToSet`、`StdDevPop`、`StdDevSamp`。
- `mgo.F("field")` 创建字段引用，可链式 `.Dot()` / `.Index()`；`mgo.Exp` 提供算术、逻辑、条件、字符串、数组、日期等丰富表达式。

示例：

```go
proj := mgo.NewProjection().
    Include("name").
    SetExpr("age_group", mgo.Exp.Cond(
        mgo.Exp.Gte(mgo.F("age"), 30),
        "adult",
        "youth",
    ))

coll.Aggs(ctx).Project(proj).All(&docs)
```

建议：

1. 应用启动时创建单例 `mgo.Client`，在进程结束时统一 `Disconnect`。
2. 在仓储层注入 `*mgo.Collection` 并封装查询/更新逻辑，保持服务层简洁。
3. 根据业务需要将通用 `Filter`、`Update` 片段提取为方法或常量，提升复用。

## 常见问题

- **如何启用认证或 TLS？**
  ```go
  cred := options.Credential{Username: "app", Password: "secret"}
  client, err := mgo.NewClient(ctx, uri, options.Client().SetAuth(cred))
  ```
- **已有 `mongo.Client` 如何复用？**使用 `mgo.WrapClient(existingClient)`。
- **如何执行原生操作？**调用 `collection.Native()`、`database.Native()` 或 `client.Native()` 取得底层类型。
- **如何控制超时和取消？**使用 `context.WithTimeout` / `context.WithCancel` 对上下文进行包裹，再传入 `Query`/`Aggs` 等方法。
- **如何调试生成的查询？**`FilterBuilder.Build()`、`UpdateBuilder.Build()`、`AggsBuilder.Build()` 均可返回原生 `bson` 文档进行打印或测试。

## 参考资料

- MongoDB Go Driver 文档：https://www.mongodb.com/docs/drivers/go/current/
- 项目示例：`examples/` 目录
- 单元测试：`document_test.go`、`filter_test.go`、`transaction_test.go`、`update_test.go`
