# 软删除功能使用指南

## 功能概述

软删除功能允许您标记文档为"已删除"状态，而不是从数据库中物理删除。这对于需要保留历史记录、支持恢复操作的场景非常有用。

## 核心设计

### 删除行为

`DeleteOne()` 和 `DeleteMany()` 的行为取决于是否启用软删除：

1. **未启用软删除**：执行硬删除（物理删除）
2. **启用软删除**：执行软删除（设置 `deleted_at` 字段）
3. **强制硬删除**：使用 `WithHardDelete()` 强制执行硬删除

## 启用软删除

### 方式1：使用默认字段名

```go
// 使用默认字段名 "deleted_at"
users := db.Collection("users", mgo.WithSoftDelete())
```

### 方式2：自定义字段名

```go
// 使用自定义字段名
users := db.Collection("users", mgo.WithSoftDelete("removed_at"))
```

### 方式3：未启用软删除（默认）

```go
// 不启用软删除，保持原有行为
users := db.Collection("users")
```

## 删除操作

### 未启用软删除时

```go
users := db.Collection("users") // 未启用软删除

// 执行硬删除（物理删除）
result, err := users.Query(ctx).Eq("_id", userId).DeleteOne()
fmt.Printf("删除了 %d 条文档\n", result.DeletedCount)

// 批量硬删除
result, err := users.Query(ctx).
    Eq("status", "expired").
    DeleteMany()
```

### 启用软删除时

```go
users := db.Collection("users", mgo.WithSoftDelete()) // 启用软删除

// 执行软删除（设置 deleted_at 字段为当前时间）
result, err := users.Query(ctx).Eq("_id", userId).DeleteOne()
fmt.Printf("软删除了 %d 条文档\n", result.DeletedCount)

// 批量软删除
result, err := users.Query(ctx).
    Eq("status", "expired").
    DeleteMany()
```

### 强制硬删除

```go
users := db.Collection("users", mgo.WithSoftDelete()) // 启用软删除

// 使用 WithHardDelete() 强制执行硬删除
result, err := users.Query(ctx).
    Eq("_id", userId).
    WithHardDelete().  // 强制硬删除
    DeleteOne()
fmt.Printf("永久删除了 %d 条文档\n", result.DeletedCount)

// 清理已软删除的数据
result, err := users.Query(ctx).
    OnlyDeleted().      // 仅查询已软删除的
    WithHardDelete().   // 强制硬删除
    DeleteMany()
```

## 查询操作

### 默认查询（自动排除已删除）

```go
// 查询时自动排除 deleted_at 存在的文档
var users []User
err := users.Query(ctx).
    Eq("status", "active").
    All(&users)
// 相当于添加了: deleted_at 字段不存在的条件
```

### 包含已删除的文档

```go
// 使用 WithDeleted() 包含已删除的文档
var allUsers []User
err := users.Query(ctx).
    Eq("status", "active").
    WithDeleted().  // 包含软删除的文档
    All(&allUsers)
```

### 仅查询已删除的文档

```go
// 使用 OnlyDeleted() 仅查询已删除的文档
var deletedUsers []User
err := users.Query(ctx).
    OnlyDeleted().  // 仅查询已删除的
    All(&deletedUsers)
```

## 恢复操作

### 恢复单条文档

```go
// 恢复软删除的用户（移除 deleted_at 字段）
result, err := users.Query(ctx).
    Eq("_id", userId).
    Restore()

if err != nil {
    log.Fatal(err)
}
fmt.Printf("恢复了 %d 条文档\n", result.ModifiedCount)
```

### 批量恢复

```go
// 批量恢复
result, err := users.Query(ctx).
    In("_id", userIds...).
    Restore()
```

## FindAndDelete

`FindAndDelete` 也遵循同样的逻辑：

```go
var deletedUser User

// 未启用软删除：硬删除
err := users.Query(ctx).Eq("_id", userId).FindAndDelete(&deletedUser)

// 启用软删除：软删除（返回删除前的文档）
err := users.Query(ctx).Eq("_id", userId).FindAndDelete(&deletedUser)

// 强制硬删除
err := users.Query(ctx).
    Eq("_id", userId).
    WithHardDelete().
    FindAndDelete(&deletedUser)
```

## 完整示例

### 示例1：用户管理

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/gocrud/mgo"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email"`
    Status    string             `bson:"status"`
    DeletedAt *time.Time         `bson:"deleted_at,omitempty"`
    CreatedAt time.Time          `bson:"created_at"`
}

func main() {
    ctx := context.Background()

    // 连接数据库
    client, err := mongo.Connect(options.Client().
        ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    // 创建启用软删除的集合
    db := mgo.NewDatabase(client.Database("testdb"))
    users := db.Collection("users", mgo.WithSoftDelete())

    // 1. 插入用户
    user := User{
        Name:      "张三",
        Email:     "zhangsan@example.com",
        Status:    "active",
        CreatedAt: time.Now(),
    }
    result, err := users.InsertOne(ctx, user)
    if err != nil {
        log.Fatal(err)
    }
    userId := result.InsertedID

    // 2. 查询用户（自动排除已删除）
    var foundUser User
    err = users.Query(ctx).Eq("_id", userId).One(&foundUser)
    fmt.Printf("找到用户: %s\n", foundUser.Name)

    // 3. 软删除用户
    _, err = users.Query(ctx).Eq("_id", userId).DeleteOne()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("用户已软删除")

    // 4. 默认查询找不到（已被软删除）
    err = users.Query(ctx).Eq("_id", userId).One(&foundUser)
    if err == mongo.ErrNoDocuments {
        fmt.Println("查询不到用户（已被软删除）")
    }

    // 5. 使用 WithDeleted 可以查询到
    err = users.Query(ctx).Eq("_id", userId).WithDeleted().One(&foundUser)
    if err == nil {
        fmt.Printf("使用 WithDeleted 找到用户: %s, 删除时间: %v\n",
            foundUser.Name, foundUser.DeletedAt)
    }

    // 6. 恢复用户
    _, err = users.Query(ctx).Eq("_id", userId).Restore()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("用户已恢复")

    // 7. 恢复后可以正常查询到
    err = users.Query(ctx).Eq("_id", userId).One(&foundUser)
    if err == nil {
        fmt.Printf("恢复后找到用户: %s\n", foundUser.Name)
    }

    // 8. 永久删除（使用 WithHardDelete）
    _, err = users.Query(ctx).Eq("_id", userId).WithHardDelete().DeleteOne()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("用户已永久删除")
}
```

### 示例2：订单管理（定期清理）

```go
// 定期清理已软删除30天的订单
func CleanupOldOrders(ctx context.Context, orders *mgo.Collection) error {
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
    
    // 清理已软删除超过30天的订单（硬删除）
    result, err := orders.Query(ctx).
        OnlyDeleted().
        Lt("deleted_at", thirtyDaysAgo).
        WithHardDelete().  // 强制硬删除
        DeleteMany()
    
    if err != nil {
        return err
    }
    
    log.Printf("清理了 %d 条已删除超过30天的订单", result.DeletedCount)
    return nil
}
```

### 示例3：统计（包含已删除数据）

```go
// 统计所有用户（包括已删除的）
func GetUserStats(ctx context.Context, users *mgo.Collection) (int64, int64, error) {
    // 活跃用户数
    activeCount, err := users.Query(ctx).
        Eq("status", "active").
        Count()
    if err != nil {
        return 0, 0, err
    }
    
    // 包含已删除的总用户数
    totalCount, err := users.Query(ctx).
        WithDeleted().
        Count()
    if err != nil {
        return 0, 0, err
    }
    
    return activeCount, totalCount, nil
}
```

### 示例3：事务中使用软删除

```go
func TransferUserData(ctx context.Context, db *mgo.Database,
    fromUserId, toUserId primitive.ObjectID) error {
    
    users := db.Collection("users", mgo.WithSoftDelete())
    orders := db.Collection("orders", mgo.WithSoftDelete())
    
    return client.Transaction(ctx, func(ctx context.Context) error {
        // 1. 转移订单
        _, err := orders.Query(ctx).
            Eq("user_id", fromUserId).
            UpdateMany(mgo.Update().Set("user_id", toUserId))
        if err != nil {
            return err
        }
        
        // 2. 软删除原用户
        _, err = users.Query(ctx).
            Eq("_id", fromUserId).
            DeleteOne()  // 软删除
        
        return err
    })
}
```

## 索引建议

为了优化软删除查询性能，建议在软删除字段上创建索引：

```go
// 创建 deleted_at 字段的索引
indexModel := mongo.IndexModel{
    Keys: bson.D{{Key: "deleted_at", Value: 1}},
    Options: options.Index().SetSparse(true), // 稀疏索引，只索引存在该字段的文档
}
_, err := users.Native().Indexes().CreateOne(ctx, indexModel)
```

## 注意事项

1. **性能考虑**：启用软删除后，所有查询都会自动添加 `deleted_at` 过滤条件，建议在 `deleted_at` 字段上创建稀疏索引

2. **删除行为自适应**：
   - 未启用软删除：`DeleteOne/DeleteMany` = 硬删除
   - 启用软删除：`DeleteOne/DeleteMany` = 软删除
   - 强制硬删除：`WithHardDelete().DeleteOne/DeleteMany()` = 硬删除

3. **聚合操作**：目前聚合操作复杂性暂不支持自动过滤软删除，需要手动过滤

4. **字段类型**：软删除字段默认使用 `time.Time` 类型，记录删除时间

5. **恢复操作**：`Restore()` 会移除 `deleted_at` 字段，使文档恢复到未删除状态

6. **FindAndDelete**：也遵循同样的软删除/硬删除逻辑

## 错误处理

```go
// 未启用软删除时调用 Restore 会返回错误
_, err := users.Query(ctx).Eq("_id", userId).Restore()
if err == mgo.ErrSoftDeleteNotEnabled {
    log.Println("集合未启用软删除功能")
}
```

## API 参考

### Collection 配置

- `WithSoftDelete(field ...string)` - 启用软删除，可选自定义字段名（默认 "deleted_at"）

### QueryBuilder 方法

**查询相关**：
- `WithDeleted()` - 包含已软删除的文档
- `OnlyDeleted()` - 仅查询已软删除的文档

**删除相关**：
- `DeleteOne()` - 删除单条文档（根据配置自动选择软删除或硬删除）
- `DeleteMany()` - 删除多条文档（根据配置自动选择软删除或硬删除）
- `WithHardDelete()` - 强制执行硬删除
- `FindAndDelete(result)` - 查找并删除文档（遵循软删除逻辑）

**恢复相关**：
- `Restore()` - 恢复软删除的文档

### 错误常量

- `ErrSoftDeleteNotEnabled` - 软删除未启用错误
