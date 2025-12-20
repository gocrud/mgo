# TX - 事务管理子包

## 功能概述

提供增强的 MongoDB 事务管理功能，支持自动事务、手动事务、重试机制等。

## 安装

```go
import "github.com/gocrud/mgo/tx"
```

## 基础用法

### 1. 自动事务（推荐）

```go
err := tx.Transaction(db, func(sess *tx.Session) error {
    users := mgo.Model[User](sess)
    orders := mgo.Model[Order](sess)
    
    // 扣减余额
    if err := users.Find().ID(userID).Inc("balance", -100).Update(); err != nil {
        return err  // 自动回滚
    }
    
    // 创建订单
    if _, err := orders.Insert(order); err != nil {
        return err  // 自动回滚
    }
    
    return nil  // 自动提交
})
```

### 2. 手动事务

```go
sess, err := tx.Begin(db)
if err != nil {
    return err
}
defer sess.Rollback()

users := mgo.Model[User](sess)
orders := mgo.Model[Order](sess)

// 执行操作
if err := users.Find().ID(userID).Update(...); err != nil {
    return err
}

if _, err := orders.Insert(order); err != nil {
    return err
}

// 手动提交
return sess.Commit()
```

### 3. 带重试的事务

```go
err := tx.WithRetry(db, 3, func(sess *tx.Session) error {
    // 事务操作
    return nil
})
```

### 4. Must 版本（panic on error）

```go
tx.MustTransaction(db, func(sess *tx.Session) error {
    // 事务操作
    return nil
})
```

## 跨库事务

### 1. 跨库自动事务

```go
// 创建 Client 实例
client, err := mgo.OpenClient("mongodb://localhost")
if err != nil {
    return err
}
defer client.Close()

// 跨库事务
err = tx.CrossDBTransaction(client, func(sess *mgo.ClientSession) error {
    // 访问多个数据库
    accountsDB := sess.Database("accounts")
    logsDB := sess.Database("logs")
    ordersDB := sess.Database("orders")
    
    users := mgo.Model[User](accountsDB)
    logs := mgo.Model[Log](logsDB)
    orders := mgo.Model[Order](ordersDB)
    
    // 跨库操作
    if err := users.Find().ID(userID).Inc("balance", -amount).Update(); err != nil {
        return err  // 自动回滚
    }
    
    if _, err := logs.Insert(&Log{Type: "payment"}); err != nil {
        return err  // 自动回滚
    }
    
    if _, err := orders.Insert(order); err != nil {
        return err  // 自动回滚
    }
    
    return nil  // 自动提交
})
```

### 2. 跨库手动事务

```go
sess, err := tx.BeginCrossDBTransaction(client)
if err != nil {
    return err
}
defer sess.Rollback()

accountsDB := sess.Database("accounts")
logsDB := sess.Database("logs")

// 执行跨库操作...

return sess.Commit()
```

### 3. 跨库事务重试

```go
err := tx.CrossDBWithRetry(client, 3, func(sess *mgo.ClientSession) error {
    // 跨库事务操作
    return nil
})
```

### 4. 使用 Databases 辅助方法

```go
err := client.Transaction(func(sess *mgo.ClientSession) error {
    // 同时获取多个数据库
    dbs := sess.Databases("accounts", "logs", "orders")
    accountsDB := dbs[0]
    logsDB := dbs[1]
    ordersDB := dbs[2]
    
    // 跨库操作
    return nil
})
```

## 高级用法

### 批量事务

```go
operations := []func(*tx.Session) error{
    func(sess *tx.Session) error { return operation1(sess) },
    func(sess *tx.Session) error { return operation2(sess) },
    func(sess *tx.Session) error { return operation3(sess) },
}

results := tx.BatchTransaction(db, operations, 10)
for i, result := range results {
    if result.Success {
        fmt.Printf("操作 %d 成功\n", i)
    } else {
        fmt.Printf("操作 %d 失败: %v\n", i, result.Error)
    }
}
```

### 嵌套事务（模拟）

```go
err := tx.Transaction(db, func(sess *tx.Session) error {
    // 外层事务
    
    err := tx.NestedTransaction(sess, func(nested *tx.Session) error {
        // 内层事务（实际使用同一事务）
        return nil
    })
    
    return err
})
```

### 事务统计

```go
stats := tx.NewTransactionStats()

for i := 0; i < 100; i++ {
    err := tx.Transaction(db, func(sess *tx.Session) error {
        // 事务操作
        return nil
    })
    
    if err != nil {
        stats.RecordFailure()
    } else {
        stats.RecordSuccess()
    }
}

fmt.Printf("成功率: %.2f%%\n", stats.SuccessRate())
```

## API 参考

### 单库事务函数

- `Transaction(db, fn)` - 自动事务
- `MustTransaction(db, fn)` - 自动事务（panic on error）
- `Begin(db)` - 开始手动事务
- `MustBegin(db)` - 开始手动事务（panic on error）
- `WithRetry(db, maxRetries, fn)` - 带重试的事务

### 跨库事务函数

- `CrossDBTransaction(client, fn)` - 跨库自动事务
- `MustCrossDBTransaction(client, fn)` - 跨库事务（panic on error）
- `BeginCrossDBTransaction(client)` - 开始跨库手动事务
- `MustBeginCrossDBTransaction(client)` - 开始跨库手动事务（panic on error）
- `CrossDBWithRetry(client, maxRetries, fn)` - 带重试的跨库事务

### Session 方法（单库）

- `sess.Commit()` - 提交事务
- `sess.Rollback()` - 回滚事务
- `sess.Abort()` - 回滚事务（别名）
- `sess.Context()` - 获取上下文
- `sess.Database()` - 获取数据库
- `sess.Collection(name)` - 获取集合
- `sess.IsActive()` - 检查事务是否活跃

### ClientSession 方法（跨库）

- `sess.Database(name)` - 获取指定数据库
- `sess.Databases(names...)` - 同时获取多个数据库
- `sess.Commit()` - 提交事务
- `sess.Rollback()` - 回滚事务
- `sess.Abort()` - 回滚事务（别名）
- `sess.Context()` - 获取上下文
- `sess.IsActive()` - 检查事务是否活跃

### 高级函数

- `BatchTransaction(db, ops, size)` - 批量事务
- `NestedTransaction(sess, fn)` - 嵌套事务

## 注意事项

1. **自动回滚**：返回 error 自动回滚，返回 nil 自动提交
2. **defer Rollback**：手动事务建议使用 `defer sess.Rollback()`
3. **重试机制**：某些错误（如网络问题）可以重试
4. **嵌套限制**：MongoDB 不支持真正的嵌套事务
5. **跨库事务要求**：需要 MongoDB 4.0+ 和副本集或分片集群
6. **跨库性能**：跨库事务性能略低于单库事务，适度使用

## 最佳实践

1. **优先使用自动事务**：更简洁，自动管理
2. **保持事务短小**：避免长时间锁定
3. **合理使用重试**：网络错误可重试，业务错误不应重试
4. **错误处理**：明确区分业务错误和系统错误
5. **跨库事务使用场景**：
   - 需要保证多个数据库数据一致性时使用
   - 单库事务能满足需求时，优先使用单库事务
   - 考虑使用 Client 实例管理多数据库连接

## 完整示例

### 单库事务示例

```go
package main

import (
    "github.com/gocrud/mgo"
    "github.com/gocrud/mgo/tx"
)

type User struct {
    ID      mgo.ObjectID `bson:"_id,omitempty"`
    Balance float64      `bson:"balance"`
}

type Order struct {
    ID     mgo.ObjectID `bson:"_id,omitempty"`
    UserID mgo.ObjectID `bson:"user_id"`
    Amount float64      `bson:"amount"`
}

func createOrder(db *mgo.Database, userID mgo.ObjectID, amount float64) error {
    return tx.WithRetry(db, 3, func(sess *tx.Session) error {
        users := mgo.Model[User](sess)
        orders := mgo.Model[Order](sess)
        
        // 扣减余额
        if err := users.Find().ID(userID).Inc("balance", -amount).Update(); err != nil {
            return err
        }
        
        // 创建订单
        order := &Order{
            UserID: userID,
            Amount: amount,
        }
        if _, err := orders.Insert(order); err != nil {
            return err
        }
        
        return nil
    })
}
```

### 跨库事务示例

```go
package main

import (
    "github.com/gocrud/mgo"
    "github.com/gocrud/mgo/tx"
)

type User struct {
    ID      mgo.ObjectID `bson:"_id,omitempty"`
    Balance float64      `bson:"balance"`
}

type Log struct {
    ID     mgo.ObjectID `bson:"_id,omitempty"`
    Type   string       `bson:"type"`
    Amount float64      `bson:"amount"`
}

func transferWithLog(client *mgo.Client, fromUserID, toUserID mgo.ObjectID, amount float64) error {
    return tx.CrossDBWithRetry(client, 3, func(sess *mgo.ClientSession) error {
        // 访问不同数据库
        accountsDB := sess.Database("accounts")
        logsDB := sess.Database("logs")
        
        users := mgo.Model[User](accountsDB)
        logs := mgo.Model[Log](logsDB)
        
        // 扣款
        if err := users.Find().ID(fromUserID).Inc("balance", -amount).Update(); err != nil {
            return err
        }
        
        // 入账
        if err := users.Find().ID(toUserID).Inc("balance", amount).Update(); err != nil {
            return err
        }
        
        // 记录日志（不同数据库）
        log := &Log{
            Type:   "transfer",
            Amount: amount,
        }
        if _, err := logs.Insert(log); err != nil {
            return err
        }
        
        return nil
    })
}
```
