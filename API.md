# MGO API 完整参考

## 📦 导入

```go
import "github.com/gocrud/mgo"
```

## 🎯 核心类型

### Model[T]

`Model[T]` 是泛型集合的入口，提供类型安全的 CRUD 操作。

```go
users := mgo.Model[User](db)
```

#### 配置方法

- `WithTimestamps()`: 启用自动时间戳（`created_at`, `updated_at`）。
- `WithSoftDelete()`: 启用软删除（`deleted_at`）。
- `WithContext(ctx)`: 设置上下文。

### Query[T]

`Query[T]` 是查询构建器，支持链式调用。

#### 基础查询

- `Find()`: 开始构建查询。
- `FindByID(id)`: 按 ID 查询。
- `FindByIDs(ids)`: 按多个 ID 查询。

#### 条件过滤

- `Where(filter M)`: 添加过滤条件（支持智能合并）。
- `Eq(field, value)`: 等于。
- `Ne(field, value)`: 不等于。
- `Gt(field, value)`: 大于。
- `Gte(field, value)`: 大于等于。
- `Lt(field, value)`: 小于。
- `Lte(field, value)`: 小于等于。
- `In(field, values...)`: 包含。
- `Nin(field, values...)`: 不包含。
- `Regex(field, pattern, options)`: 正则匹配。

#### 条件执行 (When)

```go
q.When(condition, func(q *Query[T]) {
    q.Where(...)
})
```

#### 时间查询

- `WhereDateAfter(field, date)`: 在日期之后。
- `WhereDateBefore(field, date)`: 在日期之前。
- `WhereDateBetween(field, start, end)`: 在日期之间。
- `WhereToday(field)`: 今天。
- `WhereYesterday(field)`: 昨天。
- `WhereThisWeek(field)`: 本周。
- `WhereThisMonth(field)`: 本月。
- `WhereThisYear(field)`: 本年。

#### 排序与分页

- `OrderBy(field)`: 排序（默认降序）。
- `Asc(fields...)`: 升序排序。
- `Desc(fields...)`: 降序排序。
- `Limit(n)`: 限制数量。
- `Skip(n)`: 跳过数量。
- `Page(page, pageSize)`: 分页。

#### 字段选择

- `Select(fields...)`: 选择字段。
- `Omit(fields...)`: 排除字段。

#### 执行查询

- `One()`: 返回单条记录 `(*T, error)`。
- `All()`: 返回所有记录 `([]*T, error)`。
- `Count()`: 返回数量 `(int64, error)`。
- `Exists()`: 判断是否存在 `(bool, error)`。
- `First()`: 返回第一条记录（同 `One`）。
- `Last()`: 返回最后一条记录。

#### 更新与删除

- `Update(update M)`: 更新匹配的文档。
- `UpdateSet(update M)`: 更新指定字段（自动添加 `$set`）。
- `Delete()`: 删除匹配的文档。
- `SoftDelete()`: 软删除匹配的文档。

## 🛠 辅助类型

### M (Map)

`mgo.M` 是 `map[string]interface{}` 的别名，用于构建 BSON 文档。

```go
mgo.M{"name": "张三", "age": 18}
```

### D (Document)

`mgo.D` 是 `[]E` 的别名，用于构建有序的 BSON 文档。

```go
mgo.D{{Key: "name", Value: "张三"}}
```

### E (Element)

`mgo.E` 是 BSON 元素。

```go
mgo.E{Key: "name", Value: "张三"}
```

### ObjectID

`mgo.ObjectID` 是 MongoDB 的 ObjectID 类型。

## 📚 子包参考

### agg (聚合)

提供类型安全的聚合管道构建器。

- `agg.Aggregate[T](source)`: 创建聚合构建器。
- `Match(filter)`: 匹配阶段。
- `Group(id, fields)`: 分组阶段。
- `Lookup(...)`: 关联查询。

### batch (批量处理)

提供高效的批量操作和流式处理。

- `batch.InsertBatch(coll, docs)`: 批量插入。
- `batch.UpdateBatch(coll, updates)`: 批量更新.
- `batch.Each(query, fn)`: 流式遍历。

### tx (事务)

提供简化的事务管理。

- `tx.Transaction(db, fn)`: 自动事务。
- `tx.WithRetry(db, retries, fn)`: 带重试的事务。

## 📝 完整示例

```go
// 查询示例
users, err := mgo.Model[User](db).Find().
    Where(mgo.M{"status": "active"}).
    When(age > 0, func(q *mgo.Query[User]) {
        q.Gt("age", age)
    }).
    Desc("created_at").
    Limit(10).
    All()
```
