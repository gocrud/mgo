# MGO API 完整参考

## 📦 导入

```go
import "github.com/gocrud/mgo"
```

## 🔌 连接数据库

### 连接方式

```go
// 方式 1：最简单（推荐）
db := mgo.MustOpen("mongodb://localhost/myapp")
defer db.Close()

// 方式 2：传统错误处理
db, err := mgo.Open("mongodb://localhost/myapp")
if err != nil {
    return err
}
defer db.Close()

// 方式 3：带配置
db, err := mgo.Connect(
    mgo.URI("mongodb://localhost/myapp"),
    mgo.Timeout(10*time.Second),
    mgo.MaxPoolSize(100),
    mgo.RetryWrites(true),
)

// 方式 4：从现有客户端
db := mgo.From(nativeClient, "myapp")
```

### 客户端选项

- `mgo.URI(uri)` - 设置连接 URI
- `mgo.Timeout(duration)` - 设置连接超时
- `mgo.MaxPoolSize(size)` - 最大连接池大小
- `mgo.MinPoolSize(size)` - 最小连接池大小
- `mgo.RetryWrites(bool)` - 重试写操作
- `mgo.RetryReads(bool)` - 重试读操作
- `mgo.WithContext(ctx)` - 设置默认上下文

## 📋 模型定义

### 基础模型

```go
type User struct {
    ID        mgo.ObjectID `bson:"_id,omitempty"`
    Name      string       `bson:"name"`
    Email     string       `bson:"email"`
    Age       int          `bson:"age"`
    Status    string       `bson:"status"`
    CreatedAt time.Time    `bson:"created_at"`
    UpdatedAt time.Time    `bson:"updated_at"`
    DeletedAt *time.Time   `bson:"deleted_at,omitempty"`
}
```

### 获取集合

```go
// 自动推断集合名为 "users"
users := mgo.Model[User](db)

// 显式指定集合名
users := mgo.Model[User](db, "app_users")

// 链式配置
users := mgo.Model[User](db).
    WithTimestamps().        // 自动时间戳
    WithSoftDelete()         // 软删除

// 自定义字段名
users := mgo.Model[User](db).
    WithTimestamps("created_at", "updated_at").
    WithSoftDelete("deleted_at")
```

## 📖 查询 API

### 基础查询

```go
// 按 ID 查询
user, err := users.FindByID(id)

// 单条查询
user, err := users.Find().Where("email", email).One()

// 多条查询
results, err := users.Find().
    Where("status", "active").
    Where("age", ">", 18).
    OrderBy("created_at").
    Limit(10).
    All()
```

### Where 条件

```go
// 2 参数：等于
query.Where("status", "active")

// 3 参数：带操作符
query.Where("age", ">", 18)
query.Where("age", ">=", 18)
query.Where("age", "<", 60)
query.Where("age", "<=", 60)
query.Where("status", "!=", "deleted")
query.Where("city", "in", []string{"北京", "上海"})
query.Where("name", "like", "张%")
```

### 复杂条件

```go
results, err := users.Find().
    Filter(
        mgo.And(
            mgo.Eq("status", "active"),
            mgo.Gt("age", 18),
            mgo.Or(
                mgo.Eq("vip", true),
                mgo.Gte("score", 90),
            ),
        ),
    ).
    All()
```

### 条件构建函数

- `mgo.Eq(field, value)` - 等于
- `mgo.Ne(field, value)` - 不等于
- `mgo.Gt(field, value)` - 大于
- `mgo.Gte(field, value)` - 大于等于
- `mgo.Lt(field, value)` - 小于
- `mgo.Lte(field, value)` - 小于等于
- `mgo.In(field, values...)` - 包含于
- `mgo.Nin(field, values...)` - 不包含于
- `mgo.Exists(field, bool)` - 字段存在
- `mgo.Like(field, value)` - 模糊匹配
- `mgo.StartsWith(field, value)` - 以...开头
- `mgo.EndsWith(field, value)` - 以...结尾
- `mgo.Between(field, min, max)` - 范围
- `mgo.IsNull(field)` - 为空
- `mgo.IsNotNull(field)` - 不为空

### 逻辑组合

- `mgo.And(conditions...)` - 逻辑与
- `mgo.Or(conditions...)` - 逻辑或
- `mgo.Not(field, condition)` - 逻辑非
- `mgo.Nor(conditions...)` - 逻辑非或

### 时间查询

```go
// 自动转换本地时间为 UTC
users.Find().Where("created_at", ">=", localTime).All()

// 专门的时间方法
users.Find().WhereToday("created_at").All()
users.Find().WhereYesterday("created_at").All()
users.Find().WhereThisWeek("created_at").All()
users.Find().WhereThisMonth("created_at").All()
users.Find().WhereThisYear("created_at").All()
users.Find().WhereLastDays("created_at", 7).All()
users.Find().WhereLastHours("created_at", 24).All()
users.Find().WhereDateBetween("created_at", "2024-01-01", "2024-12-31").All()
users.Find().WhereDateAfter("created_at", "2024-01-01").All()
users.Find().WhereDateBefore("created_at", "2024-12-31").All()
```

### 排序

```go
query.OrderBy("created_at")        // 降序（默认）
query.Asc("age")                   // 升序
query.Desc("created_at")           // 降序
query.Sort(mgo.M{"age": 1, "created_at": -1})
```

### 字段选择

```go
query.Select("name", "email")      // 只返回指定字段
query.Omit("password", "secret")   // 排除指定字段
```

### 分页

```go
query.Skip(20)                     // 跳过
query.Limit(10)                    // 限制数量
query.Offset(20)                   // Skip 的别名
```

### 查询执行

```go
// 基础方法
user, err := query.One()           // 单条
users, err := query.All()          // 所有
count, err := query.Count()        // 统计
exists, err := query.Exists()      // 是否存在

// 便捷方法
user := query.MustOne()            // panic on error
user := query.OneOrNil()           // 未找到返回 nil
users := query.AllOrEmpty()        // 失败返回空切片
user := query.OneOr(defaultUser)   // 未找到返回默认值

// 高级方法
user, err := query.First()         // 第一条
user, err := query.Last()          // 最后一条
user, created, err := query.FirstOrCreate(doc)  // 查询或创建
```

### 分页查询

```go
// 标准分页
page, err := users.Find().Page(1, 20)
fmt.Printf("Total: %d, Pages: %d\n", page.Total, page.Pages)

// 简化分页（不统计总数）
page, err := users.Find().SimplePaginate(1, 20)

// 游标分页
page, err := users.Find().CursorPaginate("", 20)
```

### 批量处理

```go
// 分块处理
err := users.Find().Chunk(100, func(users []*User) error {
    for _, user := range users {
        process(user)
    }
    return nil
})

// 遍历每条
err := users.Find().Each(func(user *User) error {
    return process(user)
})
```

### 去重

```go
cities, err := users.Find().Distinct("city")
```

## ✏️ 插入 API

### 单条插入

```go
user := &User{Name: "张三", Email: "test@example.com"}
id, err := users.Insert(user)
// user.ID 自动填充
```

### 批量插入

```go
ids, err := users.InsertMany(user1, user2, user3)

// 或使用切片
userList := []*User{user1, user2, user3}
ids, err := users.InsertMany(userList...)
```

## 🔄 更新 API

### 字段更新

```go
err := users.Find().ID(id).
    Set("status", "inactive").
    Inc("login_count", 1).
    Mul("score", 2).
    SetMin("age", 18).
    SetMax("age", 60).
    Unset("temp_field").
    Rename("old_name", "new_name").
    Update()
```

### 数组操作

```go
err := users.Find().ID(id).
    Push("tags", "new_tag").
    PushAll("tags", []interface{}{"tag1", "tag2"}).
    Pull("tags", "old_tag").
    PullAll("tags", []interface{}{"tag1", "tag2"}).
    AddToSet("roles", "admin").
    Pop("tags", 1).  // 1: 移除最后一个, -1: 移除第一个
    Update()
```

### 批量更新

```go
n, err := users.Find().
    Where("status", "pending").
    Set("status", "active").
    UpdateMany()
```

### 部分更新

```go
err := users.Find().ID(id).
    Patch(&User{Status: "inactive", Age: 30})
```

### 完整替换

```go
err := users.Find().ID(id).
    Replace(newUser)
```

### FindAndModify

```go
user, err := users.Find().ID(id).
    Set("status", "processing").
    UpdateAndReturn()  // 返回更新后的文档

oldUser, err := users.Find().ID(id).
    Set("status", "processing").
    UpdateAndReturnOld()  // 返回更新前的文档
```

### Upsert

```go
err := users.Find().
    Where("email", email).
    Upsert(&user)
```

## 🗑️ 删除 API

### 基础删除

```go
// 删除单条
err := users.Find().ID(id).Delete()

// 批量删除
n, err := users.Find().
    Where("status", "expired").
    DeleteMany()
```

### 软删除

```go
// 启用软删除
users := mgo.Model[User](db).WithSoftDelete()

// 删除（设置 deleted_at）
err := users.Find().ID(id).Delete()

// 物理删除
err := users.Find().ID(id).ForceDelete()

// 查询控制
users.Find().All()               // 自动排除已删除
users.Find().WithTrashed().All() // 包含已删除
users.Find().OnlyTrashed().All() // 仅已删除

// 恢复
err := users.Find().ID(id).WithTrashed().Restore()

// 批量恢复
n, err := users.Find().WithTrashed().RestoreMany()
```

### 删除并返回

```go
user, err := users.Find().ID(id).DeleteAndReturn()
```

### 清空集合

```go
err := users.Truncate()
```

## 🔒 类型安全字段引用

```go
// 获取字段引用
f := users.Field()

// 查询使用
users.Find().
    Where(f.Status, "active").
    Where(f.Age, ">", 18).
    OrderBy(f.CreatedAt).
    Select(f.Name, f.Email).
    All()

// 更新使用
users.Find().ID(id).
    Set(f.Status, "inactive").
    Inc(f.LoginCount, 1).
    Update()
```

## 📊 统计 API

```go
// 统计数量
count, err := users.Count(mgo.M{"status": "active"})
count, err := users.CountAll()

// 是否存在
exists, err := users.Exists(mgo.M{"email": email})

// 去重
values, err := users.Find().Distinct("city")

// 求和（未实现）
total, err := users.Find().Sum("balance")

// 平均值（未实现）
avg, err := users.Find().Avg("age")

// 最大值（未实现）
max, err := users.Find().Max("age")

// 最小值（未实现）
min, err := users.Find().Min("age")
```

## 💼 事务 API

```go
// 自动事务（推荐）
err := db.Transaction(func(sess *mgo.Session) error {
    users := mgo.Model[User](sess)
    orders := mgo.Model[Order](sess)
    
    if err := users.Find().ID(userID).Inc("balance", -100).Update(); err != nil {
        return err  // 自动回滚
    }
    
    if _, err := orders.Insert(order); err != nil {
        return err  // 自动回滚
    }
    
    return nil  // 自动提交
})
```

## 🛠️ 辅助函数

### ObjectID

```go
id := mgo.NewObjectID()
id, err := mgo.ObjectIDFromHex("507f1f77bcf86cd799439011")
id := mgo.MustObjectIDFromHex("507f1f77bcf86cd799439011")
valid := mgo.IsValidObjectID(hexString)
id, err := mgo.ToObjectID(anyValue)
```

### DateTime

```go
dt := mgo.NewDateTime(time.Now())
dt := mgo.Now()
```

### Decimal128

```go
dec, err := mgo.NewDecimal128("123.45")
dec := mgo.MustDecimal128("123.45")
dec := mgo.NewDecimal128FromFloat64(123.45)
```

### 其他类型

```go
regex := mgo.NewRegex("^test", "i")
js := mgo.NewJavaScript("code")
cws := mgo.NewCodeWithScope("code", scope)
bin := mgo.NewBinary(subtype, data)
ts := mgo.NewTimestamp(t, i)
```

### 时间解析

```go
t, err := mgo.ParseTime("2024-01-01 08:00:00")
t := mgo.MustParseTime("2024-01-01")
```

### 字符串转换

```go
snake := mgo.ToSnakeCase("UserProfile")  // "user_profile"
camel := mgo.ToCamelCase("user_profile") // "UserProfile"
plural := mgo.Pluralize("User")          // "users"
```

### 集合名推断

```go
name := mgo.InferCollectionName("User")            // "users"
name := mgo.InferCollectionNameFromType(User{})    // "users"
```

## ❌ 错误处理

### 标准错误

```go
mgo.ErrNoDocuments      // 未找到文档
mgo.ErrNilDocument      // 文档为 nil
mgo.ErrInvalidID        // ID 无效
mgo.ErrEmptyFilter      // 过滤条件为空
mgo.ErrEmptyUpdate      // 更新内容为空
mgo.ErrInvalidOperation // 无效操作
mgo.ErrAlreadyDeleted   // 文档已被删除
mgo.ErrNotFound         // 未找到
mgo.ErrDuplicateKey     // 重复键错误
```

### 错误检查

```go
if mgo.IsNoDocuments(err) {
    // 未找到
}

if mgo.IsDuplicateKey(err) {
    // 重复键
}

if mgo.IsNetworkError(err) {
    // 网络错误
}

if mgo.IsTimeout(err) {
    // 超时
}
```

### 错误包装

```go
err := mgo.WrapError(err, "failed to insert user")
err := mgo.WrapErrorf(err, "failed to find user with id %s", id)
err := mgo.NewValidationError("email", "email is required")
```

## 🎛️ Context 管理

### Context 创建

```go
ctx := mgo.WithTimeout(5 * time.Second)
ctx := mgo.WithDeadline(deadline)
ctx, cancel := mgo.WithCancel()
```

### Context 值

```go
ctx := mgo.WithUserID(ctx, "user123")
ctx := mgo.WithTraceID(ctx, "trace123")
ctx := mgo.WithRequestID(ctx, "req123")

userID, ok := mgo.UserIDFromContext(ctx)
traceID, ok := mgo.TraceIDFromContext(ctx)
requestID, ok := mgo.RequestIDFromContext(ctx)
```

### 查询中使用

```go
users.Find().Ctx(ctx).All()
```

## 📐 类型系统

### 支持的 MongoDB 类型

```go
mgo.ObjectID      // ObjectID
mgo.Decimal128    // Decimal128
mgo.DateTime      // DateTime
mgo.Regex         // 正则表达式
mgo.JavaScript    // JavaScript 代码
mgo.Binary        // 二进制数据
mgo.Timestamp     // 时间戳
mgo.M             // bson.M
mgo.D             // bson.D
mgo.A             // bson.A
mgo.E             // bson.E
```

### 特殊值

```go
mgo.NilObjectID   // 零值 ObjectID
mgo.MinKeyVal     // MinKey 实例
mgo.MaxKeyVal     // MaxKey 实例
mgo.UndefinedVal  // Undefined 实例
mgo.NullVal       // Null 实例
```

## 📊 完整示例

参见 [`examples/complete/main.go`](examples/complete/main.go)

## 🔜 未来功能

- 聚合子包 (`mgo/agg`)
- 批量子包 (`mgo/batch`)
- 事务子包 (`mgo/tx`)
- 索引管理
- Change Streams
- GridFS 支持
