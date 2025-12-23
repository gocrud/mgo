package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Query 泛型查询构建器
type Query[T any] struct {
	coll        *Collection[T]
	ctx         context.Context
	filter      M
	sort        D
	skip        *int64
	limit       *int64
	projection  M
	omit        M
	withTrashed bool
	onlyTrashed bool
	err         error
}

// newQuery 创建新的查询构建器
func newQuery[T any](coll *Collection[T]) *Query[T] {
	return &Query[T]{
		coll:   coll,
		ctx:    context.Background(),
		filter: M{},
		sort:   D{},
	}
}

// WithContext 设置上下文
func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

// Context 获取上下文
func (q *Query[T]) Context() context.Context {
	if q.ctx == nil {
		return context.Background()
	}
	return q.ctx
}

// ==================== 动态条件方法 ====================

// When 条件执行闭包
//
// 当 condition 为 true 时，执行 fn。
// 用于需要同时控制多个操作的场景。
//
// 示例：
//
//	q.When(age > 18, func(q *Query[T]) {
//	    q.Gt("age", 18).Eq("status", "adult")
//	})
func (q *Query[T]) When(condition bool, fn func(q *Query[T])) *Query[T] {
	if condition {
		fn(q)
	}
	return q
}

// ==================== 条件方法 ====================

// Eq 等于
func (q *Query[T]) Eq(field string, value interface{}) *Query[T] {
	q.mergeFilter(field, value)
	return q
}

// Gt 大于
func (q *Query[T]) Gt(field string, value interface{}) *Query[T] {
	return q.addOp(field, "$gt", value)
}

// Lt 小于
func (q *Query[T]) Lt(field string, value interface{}) *Query[T] {
	return q.addOp(field, "$lt", value)
}

// Gte 大于等于
func (q *Query[T]) Gte(field string, value interface{}) *Query[T] {
	return q.addOp(field, "$gte", value)
}

// Lte 小于等于
func (q *Query[T]) Lte(field string, value interface{}) *Query[T] {
	return q.addOp(field, "$lte", value)
}

// Ne 不等于
func (q *Query[T]) Ne(field string, value interface{}) *Query[T] {
	return q.addOp(field, "$ne", value)
}

// In 包含
func (q *Query[T]) In(field string, values ...interface{}) *Query[T] {
	// 如果 values 只有一个且是切片，则展开
	if len(values) == 1 {
		// 这里简化处理，直接传
	}
	return q.addOp(field, "$in", values)
}

// Nin 不包含
func (q *Query[T]) Nin(field string, values ...interface{}) *Query[T] {
	return q.addOp(field, "$nin", values)
}

// Regex 正则匹配
func (q *Query[T]) Regex(field string, pattern string, options ...string) *Query[T] {
	opts := ""
	if len(options) > 0 {
		opts = options[0]
	}
	q.filter[field] = M{"$regex": pattern, "$options": opts}
	return q
}

func (q *Query[T]) addOp(field, op string, value interface{}) *Query[T] {
	q.mergeFilter(field, M{op: NormalizeValue(value)})
	return q
}

// Where 使用复杂过滤条件（支持智能合并）
//
// 示例：
//
//	q.Where(mgo.M{"age": mgo.M{"$gt": 18}})
func (q *Query[T]) Where(filter M) *Query[T] {
	for k, v := range filter {
		q.mergeFilter(k, v)
	}
	return q
}

// mergeFilter 合并单个字段的过滤条件
func (q *Query[T]) mergeFilter(field string, value interface{}) {
	// 1. 如果该字段不存在，直接赋值
	oldVal, exists := q.filter[field]
	if !exists {
		q.filter[field] = NormalizeValue(value)
		return
	}

	// 2. 尝试将旧值和新值都标准化为 Map 形式以便合并
	oldMap, oldIsMap := oldVal.(M)
	if !oldIsMap {
		// 旧值是直接值 (Eq)，转换为 $eq Map
		oldMap = M{"$eq": oldVal}
	}

	newMap, newIsMap := value.(M)
	if !newIsMap {
		// 新值是直接值 (Eq)，转换为 $eq Map
		newMap = M{"$eq": NormalizeValue(value)}
	}

	// 3. 合并 Map
	for op, v := range newMap {
		oldMap[op] = v // 这里依然是覆盖同名操作符，但保留了不同操作符
	}

	q.filter[field] = oldMap
}

// ID 按 ID 查询
func (q *Query[T]) ID(id interface{}) *Query[T] {
	q.filter["_id"] = id
	return q
}

// IDs 按多个 ID 查询
func (q *Query[T]) IDs(ids ...interface{}) *Query[T] {
	q.filter["_id"] = M{"$in": ids}
	return q
}

// ==================== 排序方法 ====================

// OrderBy 排序（默认降序）
func (q *Query[T]) OrderBy(field string) *Query[T] {
	q.sort = removeField(q.sort, field)
	q.sort = append(q.sort, E{Key: field, Value: -1})
	return q
}

// Asc 升序排序
func (q *Query[T]) Asc(fields ...string) *Query[T] {
	for _, field := range fields {
		q.sort = removeField(q.sort, field)
		q.sort = append(q.sort, E{Key: field, Value: 1})
	}
	return q
}

// Desc 降序排序
func (q *Query[T]) Desc(fields ...string) *Query[T] {
	for _, field := range fields {
		q.sort = removeField(q.sort, field)
		q.sort = append(q.sort, E{Key: field, Value: -1})
	}
	return q
}

// Sort 自定义排序
func (q *Query[T]) Sort(sort D) *Query[T] {
	for _, elem := range sort {
		q.sort = removeField(q.sort, elem.Key)
		q.sort = append(q.sort, elem)
	}
	return q
}

// ==================== 分页方法 ====================

// Skip 跳过指定数量
func (q *Query[T]) Skip(n int64) *Query[T] {
	q.skip = &n
	return q
}

// Limit 限制数量
func (q *Query[T]) Limit(n int64) *Query[T] {
	q.limit = &n
	return q
}

// ==================== 字段选择方法 ====================

// Select 选择要返回的字段
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
func (q *Query[T]) Omit(fields ...string) *Query[T] {
	if q.omit == nil {
		q.omit = M{}
	}
	for _, field := range fields {
		q.omit[field] = 0
	}
	return q
}

// ==================== 软删除相关方法 ====================

// WithTrashed 包含已删除的记录
func (q *Query[T]) WithTrashed() *Query[T] {
	q.withTrashed = true
	q.onlyTrashed = false
	return q
}

// OnlyTrashed 只查询已删除的记录
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

// buildFindOneOptions 构建单条查询选项
func (q *Query[T]) buildFindOneOptions() *options.FindOneOptionsBuilder {
	opts := options.FindOne()

	if len(q.sort) > 0 {
		opts.SetSort(q.sort)
	}

	if q.skip != nil {
		opts.SetSkip(*q.skip)
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
func (q *Query[T]) Clone() *Query[T] {
	cloned := &Query[T]{
		coll:        q.coll,
		filter:      CopyMap(q.filter),
		sort:        append(D{}, q.sort...),
		withTrashed: q.withTrashed,
		onlyTrashed: q.onlyTrashed,
		err:         q.err,
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

// ==================== 辅助函数 ====================

// removeField 从有序文档中移除指定字段（用于处理重复字段）
func removeField(d D, field string) D {
	result := make(D, 0, len(d))
	for _, elem := range d {
		if elem.Key != field {
			result = append(result, elem)
		}
	}
	return result
}
