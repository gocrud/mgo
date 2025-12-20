# BATCH - 批量和流式处理子包

## 功能概述

提供高效的批量操作和流式处理功能，适用于大数据量场景。

## 安装

```go
import "github.com/gocrud/mgo/batch"
```

## 基础用法

### 1. 批量插入

```go
largeUserList := make([]*User, 10000)
// ... 填充数据

// 自动分批（默认 1000 条/批）
err := batch.InsertBatch(users, largeUserList)

// 自定义批次大小
err := batch.InsertBatch(users, largeUserList, batch.Size(500))

// 无序执行（更快）
err := batch.InsertBatch(users, largeUserList, 
    batch.Size(500), 
    batch.Ordered(false))
```

### 2. 带回调的批量插入

```go
err := batch.InsertBatchWithCallback(users, largeList, 
    func(inserted int) error {
        fmt.Printf("已插入 %d 条记录\n", inserted)
        return nil
    })
```

### 3. 并行批量插入

```go
// 4 个并发协程
err := batch.InsertBatchParallel(users, largeList, 4)
```

### 4. 批量插入统计

```go
stats, err := batch.InsertBatchWithStats(users, largeList)
fmt.Printf("总数: %d, 成功: %d, 失败: %d\n", 
    stats.Total, stats.Success, stats.Failed)

for _, err := range stats.Errors {
    log.Error(err)
}
```

## 流式处理

### 1. Each 遍历

```go
err := batch.Each(users.Find().Where("status", "active"),
    func(user *User) error {
        return process(user)
    })
```

### 2. Stream Channel

```go
for user := range batch.Stream(users.Find(), 100) {
    process(user)
}
```

### 3. Stream 带错误处理

```go
dataCh, errCh := batch.StreamWithError(users.Find(), 100)

for {
    select {
    case user, ok := <-dataCh:
        if !ok {
            return
        }
        process(user)
    case err := <-errCh:
        if err != nil {
            log.Error(err)
            return
        }
    }
}
```

### 4. Chunk 分块处理

```go
err := batch.Chunk(users.Find(), 100, 
    func(users []*User) error {
        for _, user := range users {
            process(user)
        }
        return nil
    })
```

## 批量更新

```go
updates := []batch.UpdateDoc{
    {
        Filter: mgo.M{"_id": id1},
        Update: mgo.M{"$set": mgo.M{"status": "active"}},
    },
    {
        Filter: mgo.M{"_id": id2},
        Update: mgo.M{"$set": mgo.M{"status": "inactive"}},
    },
}

err := batch.UpdateBatch(users, updates)
```

## 批量删除

```go
filters := []mgo.M{
    {"_id": id1},
    {"_id": id2},
    {"_id": id3},
}

n, err := batch.DeleteBatch(users, filters)
fmt.Printf("删除了 %d 条记录\n", n)
```

## 缓冲区插入

### 自动刷新缓冲区

```go
// 创建缓冲区（大小 100，5 秒超时）
buffer := batch.NewBuffer(users, 100, 5*time.Second)
defer buffer.Close()

// 添加文档
for _, user := range largeList {
    buffer.Add(user)  // 自动刷新
}

// 手动刷新
buffer.Flush()
```

### 更新缓冲区

```go
buffer := batch.NewUpdateBuffer(users, 100, 5*time.Second)
defer buffer.Close()

for _, item := range updateList {
    buffer.Add(
        mgo.M{"_id": item.ID},
        mgo.M{"$set": mgo.M{"status": item.Status}},
    )
}
```

## API 参考

### 批量操作

- `InsertBatch[T](coll, docs, opts...)` - 批量插入
- `InsertBatchWithCallback[T](coll, docs, callback, opts...)` - 带回调的批量插入
- `InsertBatchParallel[T](coll, docs, concurrency, opts...)` - 并行批量插入
- `InsertBatchWithStats[T](coll, docs, opts...)` - 批量插入（返回统计）
- `UpdateBatch(coll, updates, opts...)` - 批量更新
- `DeleteBatch(coll, filters, opts...)` - 批量删除

### 流式处理

- `Each[T](query, fn)` - 遍历处理
- `Stream[T](query, bufferSize)` - Channel 流式
- `StreamWithError[T](query, bufferSize)` - 带错误的流式
- `Chunk[T](query, size, fn)` - 分块处理

### 缓冲区

- `NewBuffer[T](coll, size, flushTime)` - 创建插入缓冲区
- `buffer.Add(doc)` - 添加文档
- `buffer.Flush()` - 刷新缓冲区
- `buffer.Close()` - 关闭缓冲区
- `NewUpdateBuffer(coll, size, flushTime)` - 创建更新缓冲区

### 选项

- `batch.Size(n)` - 设置批次大小
- `batch.Ordered(bool)` - 设置是否有序

## 完整示例

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/gocrud/mgo"
    "github.com/gocrud/mgo/batch"
)

type User struct {
    ID     mgo.ObjectID `bson:"_id,omitempty"`
    Name   string       `bson:"name"`
    Status string       `bson:"status"`
}

func main() {
    db := mgo.MustOpen("mongodb://localhost/myapp")
    defer db.Close()

    users := mgo.Model[User](db)

    // 1. 批量插入
    largeList := make([]*User, 10000)
    for i := 0; i < 10000; i++ {
        largeList[i] = &User{
            Name:   fmt.Sprintf("User%d", i),
            Status: "active",
        }
    }

    err := batch.InsertBatch(users, largeList, batch.Size(500))
    if err != nil {
        panic(err)
    }
    fmt.Println("批量插入完成")

    // 2. 流式处理
    count := 0
    err = batch.Each(users.Find(), func(user *User) error {
        count++
        if count%1000 == 0 {
            fmt.Printf("已处理 %d 条\n", count)
        }
        return nil
    })
    fmt.Printf("总共处理 %d 条\n", count)

    // 3. 使用缓冲区
    buffer := batch.NewBuffer(users, 100, 5*time.Second)
    defer buffer.Close()

    for i := 0; i < 1000; i++ {
        buffer.Add(&User{
            Name:   fmt.Sprintf("BufferedUser%d", i),
            Status: "active",
        })
    }
}
```

## 性能建议

1. **合适的批次大小**：500-1000 条为最佳
2. **无序执行**：不需要保证顺序时使用 `Ordered(false)`
3. **并行插入**：数据量超大时使用并行插入
4. **流式处理**：避免一次性加载所有数据到内存
5. **使用缓冲区**：高频小批量写入时使用缓冲区

## 注意事项

1. **内存管理**：流式处理避免 OOM
2. **错误处理**：无序执行会继续处理错误后的批次
3. **并发控制**：并行插入时注意数据库负载
4. **缓冲区刷新**：记得 Close 或 Flush 缓冲区
