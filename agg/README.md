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
