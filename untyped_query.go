package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ==================== UntypedQuery 非泛型查询构建器 ====================

// UntypedQuery 非泛型查询构建器
//
// 为传统 Collection 提供类似 TypedCollection 的链式查询API
//
// 示例：
//
//	var users []User
//	err := coll.Query().
//	    Where("status", "active").
//	    Where("age", ">", 18).
//	    OrderBy("created_at").
//	    Limit(10).
//	    All(&users)
type UntypedQuery struct {
	coll        *Collection
	filter      M
	sort        D
	skip        *int64
	limit       *int64
	projection  M
	omit        M
	ctx         context.Context
	withTrashed bool
	onlyTrashed bool
}

// newUntypedQuery 创建新的非泛型查询构建器
func newUntypedQuery(coll *Collection) *UntypedQuery {
	return &UntypedQuery{
		coll:   coll,
		filter: M{},
		sort:   D{},
		ctx:    coll.Context(),
	}
}

// ==================== 条件方法 ====================

// Where 添加查询条件
//
// 支持两种形式：
//  1. Where(field, value) - 等于查询
//  2. Where(field, operator, value) - 操作符查询
//
// 示例：
//
//	query.Where("status", "active")             // status = "active"
//	query.Where("age", ">", 18)                 // age > 18
//	query.Where("city", "in", []string{"北京"})  // city in ["北京"]
func (q *UntypedQuery) Where(args ...interface{}) *UntypedQuery {
	if len(args) < 2 {
		return q
	}

	field := args[0].(string)

	var op string
	var value interface{}

	// 两个参数：Where(field, value)
	if len(args) == 2 {
		op = "$eq"
		value = args[1]
	} else {
		// 三个参数：Where(field, op, value)
		opStr, ok := args[1].(string)
		if !ok {
			return q
		}
		op = ParseOperator(opStr)
		value = NormalizeValue(args[2])
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

// WhereIf 条件性添加查询条件
//
// 示例：
//
//	query.WhereIf(keyword != "", "name", keyword)
//	query.WhereIf(minAge > 0, "age", ">=", minAge)
func (q *UntypedQuery) WhereIf(condition bool, args ...interface{}) *UntypedQuery {
	if condition {
		return q.Where(args...)
	}
	return q
}

// ID 根据 _id 查询
//
// 示例：
//
//	query.ID(id)
func (q *UntypedQuery) ID(id interface{}) *UntypedQuery {
	q.filter["_id"] = id
	return q
}

// Filter 使用复杂过滤器
//
// 示例：
//
//	query.Filter(mgo.And(
//	    mgo.Eq("status", "active"),
//	    mgo.Gt("age", 18),
//	))
func (q *UntypedQuery) Filter(filter M) *UntypedQuery {
	for k, v := range filter {
		q.filter[k] = v
	}
	return q
}

// FilterIf 条件性使用过滤器
//
// 示例：
//
//	query.FilterIf(needFilter, mgo.Eq("status", "active"))
func (q *UntypedQuery) FilterIf(condition bool, filter M) *UntypedQuery {
	if condition {
		return q.Filter(filter)
	}
	return q
}

// ==================== 排序方法 ====================

// OrderBy 设置排序字段
//
// 示例：
//
//	query.OrderBy("created_at")       // 默认升序
//	query.OrderBy("age").Desc()       // 配合 Desc() 降序
func (q *UntypedQuery) OrderBy(field string) *UntypedQuery {
	q.sort = append(q.sort, E{Key: field, Value: 1})
	return q
}

// Desc 设置为降序排序
//
// 示例：
//
//	query.OrderBy("created_at").Desc()
func (q *UntypedQuery) Desc() *UntypedQuery {
	if len(q.sort) > 0 {
		q.sort[len(q.sort)-1].Value = -1
	}
	return q
}

// Asc 设置为升序排序
//
// 示例：
//
//	query.OrderBy("created_at").Asc()
func (q *UntypedQuery) Asc() *UntypedQuery {
	if len(q.sort) > 0 {
		q.sort[len(q.sort)-1].Value = 1
	}
	return q
}

// Sort 使用复杂排序
//
// 示例：
//
//	query.Sort(mgo.D{{Key: "age", Value: -1}, {Key: "name", Value: 1}})
func (q *UntypedQuery) Sort(sort D) *UntypedQuery {
	q.sort = sort
	return q
}

// ==================== 分页方法 ====================

// Skip 跳过指定数量
//
// 示例：
//
//	query.Skip(10)
func (q *UntypedQuery) Skip(n int64) *UntypedQuery {
	q.skip = &n
	return q
}

// Limit 限制数量
//
// 示例：
//
//	query.Limit(10)
func (q *UntypedQuery) Limit(n int64) *UntypedQuery {
	q.limit = &n
	return q
}

// Offset Offset 是 Skip 的别名
//
// 示例：
//
//	query.Offset(20)
func (q *UntypedQuery) Offset(n int64) *UntypedQuery {
	return q.Skip(n)
}

// ==================== 字段选择方法 ====================

// Select 选择要返回的字段
//
// 示例：
//
//	query.Select("name", "email")
func (q *UntypedQuery) Select(fields ...string) *UntypedQuery {
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
func (q *UntypedQuery) Omit(fields ...string) *UntypedQuery {
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
//	query.Ctx(ctx).All(&results)
func (q *UntypedQuery) Ctx(ctx context.Context) *UntypedQuery {
	q.ctx = ctx
	return q
}

// Context 获取查询上下文
func (q *UntypedQuery) Context() context.Context {
	return getContext(q.ctx)
}

// ==================== 软删除相关方法 ====================

// WithTrashed 包含已删除的记录
//
// 示例：
//
//	query.WithTrashed().All(&results)
func (q *UntypedQuery) WithTrashed() *UntypedQuery {
	q.withTrashed = true
	return q
}

// OnlyTrashed 只查询已删除的记录
//
// 示例：
//
//	query.OnlyTrashed().All(&results)
func (q *UntypedQuery) OnlyTrashed() *UntypedQuery {
	q.onlyTrashed = true
	return q
}

// ==================== 执行方法 ====================

// One 查询单条文档
//
// 示例：
//
//	var user User
//	err := coll.Query().Where("email", email).One(&user)
func (q *UntypedQuery) One(result interface{}) error {
	ctx := q.Context()
	filter := q.buildFilter()

	err := q.coll.coll.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if IsNoDocuments(err) {
			return ErrNoDocuments
		}
		return WrapError(err, "failed to find one")
	}

	return nil
}

// All 查询所有匹配的文档
//
// 示例：
//
//	var users []User
//	err := coll.Query().Where("status", "active").All(&users)
func (q *UntypedQuery) All(results interface{}) error {
	ctx := q.Context()
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(ctx, filter, opts)
	if err != nil {
		return WrapError(err, "failed to find")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return WrapError(err, "failed to decode results")
	}

	return nil
}

// Count 统计匹配的文档数量
//
// 示例：
//
//	count, err := coll.Query().Where("status", "active").Count()
func (q *UntypedQuery) Count() (int64, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	count, err := q.coll.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count")
	}

	return count, nil
}

// Exists 检查是否存在匹配的文档
//
// 示例：
//
//	exists, err := coll.Query().Where("email", email).Exists()
func (q *UntypedQuery) Exists() (bool, error) {
	count, err := q.Limit(1).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Distinct 获取字段的不重复值
//
// 示例：
//
//	var cities []string
//	err := coll.Query().Where("status", "active").Distinct("city", &cities)
func (q *UntypedQuery) Distinct(field string, results interface{}) error {
	ctx := q.Context()
	filter := q.buildFilter()

	distinctResult := q.coll.coll.Distinct(ctx, field, filter)
	if distinctResult.Err() != nil {
		return WrapError(distinctResult.Err(), "failed to get distinct values")
	}

	var values []interface{}
	if err := distinctResult.Decode(&values); err != nil {
		return WrapError(err, "failed to decode distinct values")
	}

	// 使用反射将 []interface{} 转换为目标类型
	return decodeInterfaceSlice(values, results)
}

// ==================== 游标方法 ====================

// Cursor 返回游标用于迭代
//
// 示例：
//
//	cursor, err := coll.Query().Where("status", "active").Cursor()
//	if err != nil {
//	    return err
//	}
//	defer cursor.Close(ctx)
//
//	for cursor.Next(ctx) {
//	    var user User
//	    if err := cursor.Decode(&user); err != nil {
//	        return err
//	    }
//	    // 处理 user
//	}
func (q *UntypedQuery) Cursor() (*mongo.Cursor, error) {
	ctx := q.Context()
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, WrapError(err, "failed to create cursor")
	}

	return cursor, nil
}

// Each 遍历每一条记录
//
// 示例：
//
//	err := coll.Query().Each(func(doc M) error {
//	    // 处理文档
//	    return nil
//	})
func (q *UntypedQuery) Each(fn func(M) error) error {
	ctx := q.Context()
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(ctx, filter, opts)
	if err != nil {
		return WrapError(err, "failed to find")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc M
		if err := cursor.Decode(&doc); err != nil {
			return WrapError(err, "failed to decode document")
		}

		if err := fn(doc); err != nil {
			return err
		}
	}

	if err := cursor.Err(); err != nil {
		return WrapError(err, "cursor error")
	}

	return nil
}

// ==================== 构建方法 ====================

// buildFilter 构建最终的过滤条件
func (q *UntypedQuery) buildFilter() M {
	filter := CopyMap(q.filter)

	// 处理软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		deletedField := q.coll.opts.SoftDelete.Field

		if q.onlyTrashed {
			// 只查询已删除的
			filter[deletedField] = M{"$ne": nil}
		} else if !q.withTrashed {
			// 排除已删除的（默认行为）
			filter[deletedField] = nil
		}
		// withTrashed 为 true 时不添加任何过滤条件
	}

	return filter
}

// buildOptions 构建查询选项
func (q *UntypedQuery) buildOptions() *options.FindOptionsBuilder {
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

// Clone 克隆查询构建器
//
// 示例：
//
//	baseQuery := coll.Query().Where("status", "active")
//	query1 := baseQuery.Clone().Where("age", ">", 18)
//	query2 := baseQuery.Clone().Where("city", "北京")
func (q *UntypedQuery) Clone() *UntypedQuery {
	cloned := &UntypedQuery{
		coll:        q.coll,
		filter:      CopyMap(q.filter),
		sort:        append(D{}, q.sort...),
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
