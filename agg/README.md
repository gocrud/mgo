# AGG - 聚合查询子包

## 功能概述

提供流畅的 MongoDB 聚合查询 API，支持复杂的数据统计和分析。

## 安装

```go
import "github.com/gocrud/mgo/agg"
```

## 基础用法

### 1. 基础聚合

```go
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
        Max("max_age", "$age").
    SortDesc("count").
    Limit(10).
    All()
```

### 2. 关联查询（Lookup）

```go
type UserWithOrders struct {
    User
    Orders []Order `bson:"orders"`
}

results, err := agg.Aggregate[UserWithOrders](users).
    Lookup("orders", "_id", "user_id", "orders").
    Match(mgo.Eq("status", "active")).
    All()
```

### 3. 使用 Pipeline 构建器

```go
pipeline := agg.NewPipeline().
    Match(mgo.Eq("status", "active")).
    Group("$city", mgo.M{
        "count":   agg.Sum(1),
        "avg_age": agg.Avg("$age"),
    }).
    Sort(mgo.M{"count": -1}).
    Limit(10).
    Build()

var results []CityStats
err := users.Aggregate(pipeline, &results)
```

## 高级用法

### 复杂聚合 Pipeline

```go
results, err := agg.Aggregate[OrderStats](users).
    Match(mgo.Eq("status", "active")).
    Lookup("orders", "_id", "user_id", "orders").
    Unwind("$orders").
    Match(mgo.Eq("orders.status", "completed")).
    GroupBy("$city").
        Count("total_orders").
        Sum("total_amount", "$orders.amount").
        Avg("avg_order", "$orders.amount").
    SortDesc("total_amount").
    Limit(20).
    All()
```

### 使用累加器

```go
results, err := agg.Aggregate[Stats](users).
    GroupBy("$city").
        Count("total").                    // 计数
        Sum("total_balance", "$balance").  // 求和
        Avg("avg_age", "$age").           // 平均值
        Max("max_age", "$age").           // 最大值
        Min("min_age", "$age").           // 最小值
        First("first_user", "$name").     // 第一个
        Last("last_user", "$name").       // 最后一个
        Push("users", "$name").           // 收集到数组
        AddToSet("tags", "$tag").         // 收集到数组（去重）
    All()
```

### 条件累加器

```go
results, err := agg.Aggregate[Stats](users).
    GroupBy("$city").
        // 条件计数
        Custom("adult_count", agg.CountIf(
            mgo.M{"$gt": []string{"$age", "18"}},
        )).
        // 条件求和
        Custom("vip_balance", agg.SumIf(
            mgo.M{"$eq": []string{"$vip", true}},
            "$balance",
        )).
    All()
```

### 游标分页（大数据量场景）

```go
// 第一页
page, err := agg.Aggregate[CityStats](users).
    Match(mgo.M{"status": "active"}).
    GroupBy("$city").
        Count("count").
        Avg("avg_age", "$age").
    SortDesc("count").
    CursorPage("", 20) // 空游标表示第一页

// 下一页
nextPage, err := agg.Aggregate[CityStats](users).
    Match(mgo.M{"status": "active"}).
    GroupBy("$city").
        Count("count").
        Avg("avg_age", "$age").
    SortDesc("count").
    CursorPage(page.NextCursor, 20)

// 上一页（双向翻页）
prevPage, err := agg.Aggregate[CityStats](users).
    Match(mgo.M{"status": "active"}).
    GroupBy("$city").
        Count("count").
        Avg("avg_age", "$age").
    SortDesc("count").
    CursorPage(page.PrevCursor, 20)

// 遍历所有页
cursor := ""
for {
    page, err := agg.Aggregate[CityStats](users).
        Match(mgo.M{"status": "active"}).
        GroupBy("$city").
            Count("count").
            Avg("avg_age", "$age").
        SortDesc("count").
        CursorPage(cursor, 50)
    
    if err != nil {
        return err
    }
    
    // 处理当前页数据
    for _, item := range page.Items {
        fmt.Printf("城市: %s, 用户数: %d\n", item.City, item.Count)
    }
    
    // 没有更多数据，退出
    if !page.HasMore {
        break
    }
    
    // 使用下一页游标
    cursor = page.NextCursor
}
```

**游标分页特性**：
- ✅ 适合大数据量聚合结果的分页
- ✅ 性能稳定，不受数据量影响
- ✅ 支持多字段排序
- ✅ 支持双向翻页（前一页/后一页）
- ✅ 自动处理游标编解码
- ✅ 游标解析失败时自动返回第一页
- ⚠️ 无总页数信息（适合无限滚动场景）
- ⚠️ 需要保持聚合管道的一致性（每次调用使用相同的 Match、GroupBy、排序条件）

### 日期分组

```go
results, err := agg.Aggregate[DailyStats](users).
    GroupBy(mgo.M{
        "year":  agg.Year("$created_at"),
        "month": agg.Month("$created_at"),
        "day":   agg.DayOfMonth("$created_at"),
    }).
    Count("count").
    All()
```

## API 参考

### Builder 方法

- `Match(filter)` - 匹配条件
- `GroupBy(field)` - 分组
- `Sort(sort)` - 排序
- `SortAsc(fields...)` - 升序排序
- `SortDesc(fields...)` - 降序排序
- `Limit(n)` - 限制数量
- `Skip(n)` - 跳过数量
- `Project(projection)` - 投影
- `Unwind(path)` - 展开数组
- `Lookup(from, local, foreign, as)` - 关联查询
- `AddFields(fields)` - 添加字段
- `ReplaceRoot(newRoot)` - 替换根文档
- `Sample(size)` - 随机抽样
- `CursorPage(cursor, perPage)` - 游标分页（推荐大数据量场景）
- `All()` - 执行并返回所有结果
- `One()` - 执行并返回第一条结果
- `Count()` - 统计聚合结果数量

### GroupStage 累加器

- `Count(field)` - 计数
- `Sum(field, expr)` - 求和
- `Avg(field, expr)` - 平均值
- `Max(field, expr)` - 最大值
- `Min(field, expr)` - 最小值
- `First(field, expr)` - 第一个
- `Last(field, expr)` - 最后一个
- `Push(field, expr)` - 收集到数组
- `AddToSet(field, expr)` - 收集到数组（去重）

### 累加器函数

- `agg.Sum(expr)` - 求和
- `agg.Avg(expr)` - 平均值
- `agg.Max(expr)` - 最大值
- `agg.Min(expr)` - 最小值
- `agg.CountIf(condition)` - 条件计数
- `agg.SumIf(condition, expr)` - 条件求和

### 操作符函数

**算术操作符**：
- `agg.Add(exprs...)` - 加法
- `agg.Subtract(expr1, expr2)` - 减法
- `agg.Multiply(exprs...)` - 乘法
- `agg.Divide(expr1, expr2)` - 除法
- `agg.Mod(expr, divisor)` - 取模

**字符串操作符**：
- `agg.Concat(exprs...)` - 连接
- `agg.Substr(expr, start, length)` - 截取
- `agg.ToLower(expr)` - 转小写
- `agg.ToUpper(expr)` - 转大写

**数组操作符**：
- `agg.Size(expr)` - 数组大小
- `agg.ArrayElemAt(expr, index)` - 获取元素
- `agg.Slice(expr, start, length)` - 数组切片
- `agg.Filter(input, as, cond)` - 过滤数组
- `agg.Map(input, as, in)` - 映射数组
- `agg.Reduce(input, initial, in)` - 归约数组

**条件操作符**：
- `agg.Cond(condition, ifTrue, ifFalse)` - 条件表达式
- `agg.IfNull(expr, replacement)` - 空值替换
- `agg.Switch(branches, default)` - 多条件分支

**比较操作符**：
- `agg.Eq(expr1, expr2)` - 等于
- `agg.Ne(expr1, expr2)` - 不等于
- `agg.Gt(expr1, expr2)` - 大于
- `agg.Gte(expr1, expr2)` - 大于等于
- `agg.Lt(expr1, expr2)` - 小于
- `agg.Lte(expr1, expr2)` - 小于等于

**日期操作符**：
- `agg.Year(expr)` - 年份
- `agg.Month(expr)` - 月份
- `agg.DayOfMonth(expr)` - 日期
- `agg.DayOfWeek(expr)` - 星期
- `agg.Hour(expr)` - 小时
- `agg.DateToString(expr, format)` - 日期转字符串

## 完整示例

```go
package main

import (
    "fmt"
    "github.com/gocrud/mgo"
    "github.com/gocrud/mgo/agg"
)

type User struct {
    ID     mgo.ObjectID `bson:"_id,omitempty"`
    Name   string       `bson:"name"`
    City   string       `bson:"city"`
    Age    int          `bson:"age"`
    Status string       `bson:"status"`
}

type CityStats struct {
    City      string  `bson:"_id"`
    Count     int     `bson:"count"`
    AvgAge    float64 `bson:"avg_age"`
    MaxAge    int     `bson:"max_age"`
    MinAge    int     `bson:"min_age"`
    TotalAge  int     `bson:"total_age"`
}

func main() {
    db := mgo.MustOpen("mongodb://localhost/myapp")
    defer db.Close()

    users := mgo.Model[User](db)

    // 按城市统计用户
    stats, err := agg.Aggregate[CityStats](users).
        Match(mgo.Eq("status", "active")).
        GroupBy("$city").
            Count("count").
            Avg("avg_age", "$age").
            Max("max_age", "$age").
            Min("min_age", "$age").
            Sum("total_age", "$age").
        SortDesc("count").
        Limit(10).
        All()
    
    if err != nil {
        panic(err)
    }

    for _, stat := range stats {
        fmt.Printf("%s: %d 人，平均年龄 %.1f\n", 
            stat.City, stat.Count, stat.AvgAge)
    }
}
```

## 性能建议

1. **Match 前置**：尽早过滤数据
2. **Project 优化**：只投影需要的字段
3. **索引支持**：确保 Match 条件有索引
4. **限制结果**：使用 Limit 限制结果数量
5. **避免 $lookup**：Join 操作开销大，能在应用层做就不要在数据库层做
6. **大数据量分页**：使用 `CursorPage` 代替 `Skip/Limit` 组合

## 分页对比

| 场景 | 推荐方案 | 说明 |
|------|----------|------|
| 小数据量（< 1000条） | `Limit` + `Skip` | 简单直接 |
| 大数据量分页 | `CursorPage` | 性能稳定，不受数据量影响 |
| 需要总页数 | 先 `Count`，再 `Limit/Skip` | 有性能开销 |
| 无限滚动 | `CursorPage` | 最佳选择 |
| 双向翻页 | `CursorPage` | 原生支持前后翻页 |

## 更多示例

完整示例代码请参考：[examples/agg_cursor_pagination/main.go](../examples/agg_cursor_pagination/main.go)
