package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ==================== Query 泛型查询构建器 ====================

// Query 泛型查询构建器
//
// 提供流畅的链式查询API
//
// 示例：
//
//	results, err := users.Find().
//	    Where("status", "active").
//	    Where("age", ">", 18).
//	    OrderBy("created_at").
//	    Limit(10).
//	    All()
type Query[T any] struct {
	coll        *TypedCollection[T]
	filter      M
	sort        M
	skip        *int64
	limit       *int64
	projection  M
	omit        M
	ctx         context.Context
	withTrashed bool
	onlyTrashed bool
}

// newQuery 创建新的查询构建器
func newQuery[T any](coll *TypedCollection[T]) *Query[T] {
	return &Query[T]{
		coll:   coll,
		filter: M{},
		sort:   M{},
		ctx:    coll.Context(),
	}
}

// ==================== 条件方法 ====================

// Where 添加查询条件
//
// 支持两种形式：
// 1. Where(field, value) - 等于条件
// 2. Where(field, operator, value) - 带操作符的条件
//
// 示例：
//
//	query.Where("status", "active")  // status = "active"
//	query.Where("age", ">", 18)      // age > 18
//	query.Where("name", "like", "张%")  // name LIKE "张%"
func (q *Query[T]) Where(field string, args ...interface{}) *Query[T] {
	if len(args) == 0 {
		return q
	}

	var op string
	var value interface{}

	if len(args) == 1 {
		// Where("field", value) - 等于
		op = "$eq"
		value = args[0]
	} else {
		// Where("field", "op", value) - 带操作符
		opStr, ok := args[0].(string)
		if !ok {
			return q
		}
		op = ParseOperator(opStr)
		value = NormalizeValue(args[1])
	}

	// 特殊处理等于操作符
	if op == "$eq" {
		q.filter[field] = NormalizeValue(value)
	} else {
		// 其他操作符
		if existing, ok := q.filter[field].(M); ok {
			existing[op] = value
		} else {
			q.filter[field] = M{op: value}
		}
	}

	return q
}

// Filter 使用复杂过滤条件
//
// 示例：
//
//	query.Filter(mgo.And(
//	    mgo.Eq("status", "active"),
//	    mgo.Gt("age", 18),
//	))
func (q *Query[T]) Filter(filter M) *Query[T] {
	for k, v := range filter {
		q.filter[k] = v
	}
	return q
}

// ID 按 ID 查询
//
// 示例：
//
//	user, err := users.Find().ID(id).One()
func (q *Query[T]) ID(id interface{}) *Query[T] {
	q.filter["_id"] = id
	return q
}

// IDs 按多个 ID 查询
//
// 示例：
//
//	users, err := users.Find().IDs(id1, id2, id3).All()
func (q *Query[T]) IDs(ids ...interface{}) *Query[T] {
	q.filter["_id"] = M{"$in": ids}
	return q
}

// ==================== 排序方法 ====================

// OrderBy 排序（默认降序）
//
// 示例：
//
//	query.OrderBy("created_at")  // 降序
func (q *Query[T]) OrderBy(field string) *Query[T] {
	q.sort[field] = -1
	return q
}

// Asc 升序排序
//
// 示例：
//
//	query.Asc("age")
func (q *Query[T]) Asc(fields ...string) *Query[T] {
	for _, field := range fields {
		q.sort[field] = 1
	}
	return q
}

// Desc 降序排序
//
// 示例：
//
//	query.Desc("created_at", "updated_at")
func (q *Query[T]) Desc(fields ...string) *Query[T] {
	for _, field := range fields {
		q.sort[field] = -1
	}
	return q
}

// Sort 自定义排序
//
// 示例：
//
//	query.Sort(mgo.M{"age": 1, "created_at": -1})
func (q *Query[T]) Sort(sort M) *Query[T] {
	for k, v := range sort {
		q.sort[k] = v
	}
	return q
}

// ==================== 分页方法 ====================

// Skip 跳过指定数量
//
// 示例：
//
//	query.Skip(20)
func (q *Query[T]) Skip(n int64) *Query[T] {
	q.skip = &n
	return q
}

// Limit 限制数量
//
// 示例：
//
//	query.Limit(10)
func (q *Query[T]) Limit(n int64) *Query[T] {
	q.limit = &n
	return q
}

// Offset Offset 是 Skip 的别名
//
// 示例：
//
//	query.Offset(20)
func (q *Query[T]) Offset(n int64) *Query[T] {
	return q.Skip(n)
}

// ==================== 字段选择方法 ====================

// Select 选择要返回的字段
//
// 示例：
//
//	query.Select("name", "email")
func (q *Query[T]) Select(fields ...string) *Query[T] {
	if q.projection == nil {
		q.projection = M{}
	}
	for _, field := range fields {
		q.projection[field] = 1
	}
	return q
}

// Omit 排除指定字段
//
// 示例：
//
//	query.Omit("password", "secret")
func (q *Query[T]) Omit(fields ...string) *Query[T] {
	if q.omit == nil {
		q.omit = M{}
	}
	for _, field := range fields {
		q.omit[field] = 0
	}
	return q
}

// ==================== 上下文方法 ====================

// Ctx 设置查询上下文
//
// 示例：
//
//	ctx := context.WithTimeout(context.Background(), 5*time.Second)
//	query.Ctx(ctx).All()
func (q *Query[T]) Ctx(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

// Context 获取查询上下文
func (q *Query[T]) Context() context.Context {
	return getContext(q.ctx)
}

// ==================== 软删除相关方法 ====================

// WithTrashed 包含已删除的记录
//
// 示例：
//
//	query.WithTrashed().All()
func (q *Query[T]) WithTrashed() *Query[T] {
	q.withTrashed = true
	q.onlyTrashed = false
	return q
}

// OnlyTrashed 只查询已删除的记录
//
// 示例：
//
//	query.OnlyTrashed().All()
func (q *Query[T]) OnlyTrashed() *Query[T] {
	q.onlyTrashed = true
	q.withTrashed = false
	return q
}

// ==================== 构建方法 ====================

// buildFilter 构建最终的过滤条件
func (q *Query[T]) buildFilter() M {
	filter := M{}

	// 复制 filter，但排除更新操作符
	updateOps := map[string]bool{
		"$set": true, "$inc": true, "$mul": true, "$min": true, "$max": true,
		"$unset": true, "$rename": true, "$push": true, "$pull": true,
		"$pullAll": true, "$addToSet": true, "$pop": true,
	}

	for k, v := range q.filter {
		if !updateOps[k] {
			filter[k] = v
		}
	}

	// 应用软删除过滤
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		field := q.coll.opts.SoftDelete.Field

		if q.onlyTrashed {
			// 只查询已删除的记录
			filter[field] = M{"$ne": nil}
		} else if !q.withTrashed {
			// 默认：只查询未删除的记录
			// 如果 filter 中没有显式设置 deleted_at，添加默认过滤
			if _, exists := filter[field]; !exists {
				filter[field] = nil
			}
		}
		// withTrashed == true: 不添加任何过滤，包含所有记录
	}

	return filter
}

// buildOptions 构建查询选项
func (q *Query[T]) buildOptions() *options.FindOptionsBuilder {
	opts := options.Find()

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}

	if q.skip != nil {
		opts.SetSkip(*q.skip)
	}

	if q.limit != nil {
		opts.SetLimit(*q.limit)
	}

	// 构建投影
	projection := M{}
	if len(q.projection) > 0 {
		projection = q.projection
	}
	if len(q.omit) > 0 {
		for k, v := range q.omit {
			projection[k] = v
		}
	}
	if len(projection) > 0 {
		opts.SetProjection(projection)
	}

	return opts
}

// ==================== 克隆方法 ====================

// Clone 克隆查询构建器
//
// 示例：
//
//	baseQuery := users.Find().Where("status", "active")
//	query1 := baseQuery.Clone().Where("age", ">", 18)
//	query2 := baseQuery.Clone().Where("city", "北京")
func (q *Query[T]) Clone() *Query[T] {
	cloned := &Query[T]{
		coll:        q.coll,
		filter:      CopyMap(q.filter),
		sort:        CopyMap(q.sort),
		ctx:         q.ctx,
		withTrashed: q.withTrashed,
		onlyTrashed: q.onlyTrashed,
	}

	if q.skip != nil {
		skip := *q.skip
		cloned.skip = &skip
	}

	if q.limit != nil {
		limit := *q.limit
		cloned.limit = &limit
	}

	if q.projection != nil {
		cloned.projection = CopyMap(q.projection)
	}

	if q.omit != nil {
		cloned.omit = CopyMap(q.omit)
	}

	return cloned
}
