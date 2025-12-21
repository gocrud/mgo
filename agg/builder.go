package agg

import (
	"context"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 聚合构建器 ====================

// Builder 聚合构建器
//
// 提供流畅的聚合查询 API
//
// 示例：
//
//	type CityStats struct {
//	    City   string  `bson:"_id"`
//	    Count  int     `bson:"count"`
//	    AvgAge float64 `bson:"avg_age"`
//	}
//
//	results, err := agg.Aggregate[CityStats](users).
//	    Match(mgo.Eq("status", "active")).
//	    GroupBy("$city").
//	        Count("count").
//	        Avg("avg_age", "$age").
//	    SortDesc("count").
//	    Limit(10).
//	    All()
type Builder[T any] struct {
	coll     *mongo.Collection
	pipeline mgo.Pipeline
	ctx      context.Context
	sort     mgo.D // 排序信息（用于游标分页）
}

// Aggregate 创建聚合构建器
//
// 示例：
//
//	builder := agg.Aggregate[Stats](users)
func Aggregate[T any](source interface{}) *Builder[T] {
	var coll *mongo.Collection
	var ctx context.Context

	switch src := source.(type) {
	case interface{ Native() *mongo.Collection }:
		coll = src.Native()
		if ctxGetter, ok := source.(interface{ Context() context.Context }); ok {
			ctx = ctxGetter.Context()
		}
	case *mongo.Collection:
		coll = src
		ctx = context.Background()
	default:
		panic("agg: invalid source type")
	}

	return &Builder[T]{
		coll:     coll,
		pipeline: mgo.Pipeline{},
		ctx:      ctx,
	}
}

// ==================== Pipeline 构建方法 ====================

// Match 添加匹配阶段
//
// 示例：
//
//	builder.Match(mgo.Eq("status", "active"))
func (b *Builder[T]) Match(filter mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$match", Value: filter}})
	return b
}

// GroupBy 添加分组阶段
//
// 示例：
//
//	builder.GroupBy("$city")
func (b *Builder[T]) GroupBy(field string) *GroupStage[T] {
	return &GroupStage[T]{
		builder: b,
		groupID: field,
		fields:  mgo.M{},
	}
}

// Sort 添加排序阶段
//
// 示例：
//
//	builder.Sort(mgo.D{{Key: "count", Value: -1}, {Key: "name", Value: 1}})
func (b *Builder[T]) Sort(sort mgo.D) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$sort", Value: sort}})
	b.sort = sort // 记录排序信息
	return b
}

// SortAsc 升序排序
//
// 示例：
//
//	builder.SortAsc("age")
func (b *Builder[T]) SortAsc(fields ...string) *Builder[T] {
	sort := make(mgo.D, 0, len(fields))
	for _, field := range fields {
		sort = append(sort, mgo.E{Key: field, Value: 1})
	}
	b.sort = sort // 记录排序信息
	return b.Sort(sort)
}

// SortDesc 降序排序
//
// 示例：
//
//	builder.SortDesc("count")
func (b *Builder[T]) SortDesc(fields ...string) *Builder[T] {
	sort := make(mgo.D, 0, len(fields))
	for _, field := range fields {
		sort = append(sort, mgo.E{Key: field, Value: -1})
	}
	b.sort = sort // 记录排序信息
	return b.Sort(sort)
}

// Limit 限制结果数量
//
// 示例：
//
//	builder.Limit(10)
func (b *Builder[T]) Limit(n int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$limit", Value: n}})
	return b
}

// Skip 跳过指定数量
//
// 示例：
//
//	builder.Skip(20)
func (b *Builder[T]) Skip(n int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$skip", Value: n}})
	return b
}

// Project 添加投影阶段
//
// 示例：
//
//	builder.Project(mgo.M{"name": 1, "email": 1})
func (b *Builder[T]) Project(projection mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$project", Value: projection}})
	return b
}

// Unwind 展开数组
//
// 示例：
//
//	builder.Unwind("$orders")
func (b *Builder[T]) Unwind(path string) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$unwind", Value: path}})
	return b
}

// Lookup 关联查询
//
// 示例：
//
//	builder.Lookup("orders", "_id", "user_id", "orders")
func (b *Builder[T]) Lookup(from, localField, foreignField, as string) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$lookup", Value: mgo.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}}})
	return b
}

// AddFields 添加字段
//
// 示例：
//
//	builder.AddFields(mgo.M{"totalAmount": mgo.M{"$sum": "$items.price"}})
func (b *Builder[T]) AddFields(fields mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$addFields", Value: fields}})
	return b
}

// ReplaceRoot 替换根文档
//
// 示例：
//
//	builder.ReplaceRoot("$profile")
func (b *Builder[T]) ReplaceRoot(newRoot string) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$replaceRoot", Value: mgo.M{
		"newRoot": newRoot,
	}}})
	return b
}

// Sample 随机抽样
//
// 示例：
//
//	builder.Sample(10)
func (b *Builder[T]) Sample(size int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$sample", Value: mgo.M{
		"size": size,
	}}})
	return b
}

// ==================== 执行方法 ====================

// All 执行聚合并返回所有结果
//
// 示例：
//
//	results, err := builder.All()
func (b *Builder[T]) All() ([]*T, error) {
	cursor, err := b.coll.Aggregate(b.ctx, b.pipeline)
	if err != nil {
		return nil, mgo.WrapError(err, "failed to aggregate")
	}
	defer cursor.Close(b.ctx)

	var results []*T
	if err := cursor.All(b.ctx, &results); err != nil {
		return nil, mgo.WrapError(err, "failed to decode aggregate results")
	}

	return results, nil
}

// One 执行聚合并返回第一条结果
//
// 示例：
//
//	result, err := builder.One()
func (b *Builder[T]) One() (*T, error) {
	b.Limit(1)
	results, err := b.All()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, mgo.ErrNoDocuments
	}

	return results[0], nil
}

// Count 统计聚合结果数量
//
// 示例：
//
//	count, err := builder.Count()
func (b *Builder[T]) Count() (int64, error) {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$count", Value: "count"}})

	type countResult struct {
		Count int64 `bson:"count"`
	}

	cursor, err := b.coll.Aggregate(b.ctx, b.pipeline)
	if err != nil {
		return 0, mgo.WrapError(err, "failed to count aggregate")
	}
	defer cursor.Close(b.ctx)

	var result countResult
	if cursor.Next(b.ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, mgo.WrapError(err, "failed to decode count result")
		}
		return result.Count, nil
	}

	return 0, nil
}

// Pipeline 获取构建的 pipeline
//
// 示例：
//
//	pipeline := builder.Pipeline()
func (b *Builder[T]) Pipeline() mgo.Pipeline {
	return b.pipeline
}

// Ctx 设置上下文
//
// 示例：
//
//	builder.Ctx(ctx).All()
func (b *Builder[T]) Ctx(ctx context.Context) *Builder[T] {
	b.ctx = ctx
	return b
}

// CursorPage 使用游标的分页（适用于大数据量聚合，支持双向翻页）
//
// 参数：
//   - cursor: 游标字符串（空字符串表示第一页）
//   - perPage: 每页数量
//
// 特性：
//   - 自动使用已设置的排序
//   - 无排序时默认按 _id 降序
//   - 支持多字段排序
//   - 提供前后游标实现双向翻页
//   - 游标解析失败时返回第一页
//
// 示例：
//
//	// 第一页（按 count 降序）
//	page, err := agg.Aggregate[Stats](users).
//	    Match(mgo.M{"status": "active"}).
//	    GroupBy("$city").
//	        Count("count").
//	        Avg("avg_age", "$age").
//	    SortDesc("count").
//	    CursorPage("", 20)
//
//	// 下一页
//	nextPage, _ := agg.Aggregate[Stats](users).
//	    Match(mgo.M{"status": "active"}).
//	    GroupBy("$city").
//	        Count("count").
//	        Avg("avg_age", "$age").
//	    SortDesc("count").
//	    CursorPage(page.NextCursor, 20)
//
//	// 上一页
//	prevPage, _ := agg.Aggregate[Stats](users).
//	    Match(mgo.M{"status": "active"}).
//	    GroupBy("$city").
//	        Count("count").
//	        Avg("avg_age", "$age").
//	    SortDesc("count").
//	    CursorPage(page.PrevCursor, 20)
func (b *Builder[T]) CursorPage(cursor string, perPage int) (*mgo.CursorPage[T], error) {
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}

	// 克隆 pipeline 避免修改原始 pipeline
	pipeline := make(mgo.Pipeline, len(b.pipeline))
	copy(pipeline, b.pipeline)

	// 确定排序
	sortDoc := b.sort
	if len(sortDoc) == 0 {
		sortDoc = mgo.D{{Key: "_id", Value: -1}}
	}

	// 解析游标并添加过滤条件
	var cursorDir string
	if cursor != "" {
		data, err := mgo.DecodeCursorData(cursor)
		if err != nil {
			// 游标解析失败，忽略并返回第一页（用户友好）
		} else {
			cursorDir = data.Direction

			// 根据游标方向反转排序（prev时需要反向查询）
			currentSort := sortDoc
			if data.Direction == "prev" {
				newSort := make(mgo.D, len(sortDoc))
				for i, elem := range sortDoc {
					if o, ok := elem.Value.(int); ok {
						newSort[i] = mgo.E{Key: elem.Key, Value: -o}
					} else {
						newSort[i] = elem
					}
				}
				currentSort = newSort
			}

			// 构建游标过滤条件
			filter := buildAggCursorFilter(data, sortDoc)
			if len(filter) > 0 {
				// 在 pipeline 中插入 $match 阶段
				// 需要在最后一个 $match 之后、$sort 之前插入
				insertPos := len(pipeline)
				for i := len(pipeline) - 1; i >= 0; i-- {
					if len(pipeline[i]) > 0 {
						if pipeline[i][0].Key == "$sort" {
							insertPos = i
						}
					}
				}

				// 插入游标过滤条件
				matchStage := mgo.D{{Key: "$match", Value: filter}}
				pipeline = append(pipeline[:insertPos], append(mgo.Pipeline{matchStage}, pipeline[insertPos:]...)...)
			}

			// 确保排序在正确位置
			// 移除现有的 $sort 并在游标 $match 后添加
			var newPipeline mgo.Pipeline
			for _, stage := range pipeline {
				if len(stage) > 0 && stage[0].Key != "$sort" {
					newPipeline = append(newPipeline, stage)
				}
			}
			newPipeline = append(newPipeline, mgo.D{{Key: "$sort", Value: currentSort}})
			pipeline = newPipeline
		}
	} else {
		// 第一页，确保有排序
		hasSort := false
		for _, stage := range pipeline {
			if len(stage) > 0 && stage[0].Key == "$sort" {
				hasSort = true
				break
			}
		}
		if !hasSort {
			pipeline = append(pipeline, mgo.D{{Key: "$sort", Value: sortDoc}})
		}
	}

	// 添加 limit（多查一条判断是否有下一页）
	limit := int64(perPage + 1)
	pipeline = append(pipeline, mgo.D{{Key: "$limit", Value: limit}})

	// 执行聚合
	aggCursor, err := b.coll.Aggregate(b.ctx, pipeline)
	if err != nil {
		return nil, mgo.WrapError(err, "failed to aggregate for cursor page")
	}
	defer aggCursor.Close(b.ctx)

	var items []*T
	if err := aggCursor.All(b.ctx, &items); err != nil {
		return nil, mgo.WrapError(err, "failed to decode cursor page results")
	}

	// 如果是反向查询，需要反转结果顺序
	if cursorDir == "prev" {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	hasMore := len(items) > perPage
	if hasMore {
		items = items[:perPage]
	}

	// 生成游标
	var nextCursor, prevCursor string

	// 生成下一页游标（从最后一条记录）
	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = mgo.EncodeCursorFromItem(lastItem, sortDoc, "next")
	}

	// 生成上一页游标（从第一条记录）
	// 只有在不是第一页时才生成
	if cursor != "" && len(items) > 0 {
		firstItem := items[0]
		prevCursor = mgo.EncodeCursorFromItem(firstItem, sortDoc, "prev")
	}

	return &mgo.CursorPage[T]{
		Items:      items,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasMore:    hasMore,
	}, nil
}

// ==================== GroupStage 分组阶段 ====================

// GroupStage 分组阶段构建器
type GroupStage[T any] struct {
	builder *Builder[T]
	groupID string
	fields  mgo.M
}

// Count 统计数量
//
// 示例：
//
//	builder.GroupBy("$city").Count("total")
func (g *GroupStage[T]) Count(field string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$sum": 1}
	return g
}

// Sum 求和
//
// 示例：
//
//	builder.GroupBy("$city").Sum("total_amount", "$amount")
func (g *GroupStage[T]) Sum(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$sum": expr}
	return g
}

// Avg 平均值
//
// 示例：
//
//	builder.GroupBy("$city").Avg("avg_age", "$age")
func (g *GroupStage[T]) Avg(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$avg": expr}
	return g
}

// Max 最大值
//
// 示例：
//
//	builder.GroupBy("$city").Max("max_age", "$age")
func (g *GroupStage[T]) Max(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$max": expr}
	return g
}

// Min 最小值
//
// 示例：
//
//	builder.GroupBy("$city").Min("min_age", "$age")
func (g *GroupStage[T]) Min(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$min": expr}
	return g
}

// First 第一个值
//
// 示例：
//
//	builder.GroupBy("$city").First("first_user", "$name")
func (g *GroupStage[T]) First(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$first": expr}
	return g
}

// Last 最后一个值
//
// 示例：
//
//	builder.GroupBy("$city").Last("last_user", "$name")
func (g *GroupStage[T]) Last(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$last": expr}
	return g
}

// Push 收集到数组
//
// 示例：
//
//	builder.GroupBy("$city").Push("users", "$name")
func (g *GroupStage[T]) Push(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$push": expr}
	return g
}

// AddToSet 收集到数组（去重）
//
// 示例：
//
//	builder.GroupBy("$city").AddToSet("tags", "$tag")
func (g *GroupStage[T]) AddToSet(field, expr string) *GroupStage[T] {
	g.fields[field] = mgo.M{"$addToSet": expr}
	return g
}

// 完成分组并返回 Builder
func (g *GroupStage[T]) build() *Builder[T] {
	group := mgo.M{"_id": g.groupID}
	for k, v := range g.fields {
		group[k] = v
	}

	g.builder.pipeline = append(g.builder.pipeline, mgo.D{{Key: "$group", Value: group}})
	return g.builder
}

// Match 继续添加 Match 阶段
func (g *GroupStage[T]) Match(filter mgo.M) *Builder[T] {
	return g.build().Match(filter)
}

// Sort 继续添加 Sort 阶段
func (g *GroupStage[T]) Sort(sort mgo.D) *Builder[T] {
	return g.build().Sort(sort)
}

// SortAsc 升序排序
func (g *GroupStage[T]) SortAsc(fields ...string) *Builder[T] {
	return g.build().SortAsc(fields...)
}

// SortDesc 降序排序
func (g *GroupStage[T]) SortDesc(fields ...string) *Builder[T] {
	return g.build().SortDesc(fields...)
}

// Limit 限制数量
func (g *GroupStage[T]) Limit(n int64) *Builder[T] {
	return g.build().Limit(n)
}

// Skip 跳过数量
func (g *GroupStage[T]) Skip(n int64) *Builder[T] {
	return g.build().Skip(n)
}

// All 执行聚合
func (g *GroupStage[T]) All() ([]*T, error) {
	return g.build().All()
}

// One 执行聚合返回一条
func (g *GroupStage[T]) One() (*T, error) {
	return g.build().One()
}

// CursorPage 游标分页
func (g *GroupStage[T]) CursorPage(cursor string, perPage int) (*mgo.CursorPage[T], error) {
	return g.build().CursorPage(cursor, perPage)
}

// ==================== 游标分页辅助函数 ====================

// buildAggCursorFilter 根据游标数据构建聚合的过滤条件
func buildAggCursorFilter(data *mgo.CursorData, sortDoc mgo.D) mgo.M {
	if data == nil || len(data.Values) == 0 {
		return mgo.M{}
	}

	// 获取排序字段和方向
	sortFields := make([]string, 0, len(sortDoc))
	sortOrders := make(map[string]int)
	for _, elem := range sortDoc {
		sortFields = append(sortFields, elem.Key)
		if order, ok := elem.Value.(int); ok {
			sortOrders[elem.Key] = order
		} else {
			sortOrders[elem.Key] = 1 // 默认升序
		}
	}

	// 如果只有一个排序字段（常见情况），使用简化查询
	if len(sortFields) == 1 {
		field := sortFields[0]
		value := data.Values[field]
		order := sortOrders[field]

		// 根据排序方向选择操作符
		var op string
		if data.Direction == "prev" {
			// 反向查询已在外层反转排序，这里保持一致
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		} else {
			// 正向查询
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		}

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := mgo.ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if data.ID != "" {
				if oid, err := mgo.ObjectIDFromHex(data.ID); err == nil {
					value = oid
				}
			}
		}

		return mgo.M{field: mgo.M{op: value}}
	}

	// 多字段排序：构建复杂的 $or 条件
	// 例如：ORDER BY age ASC, created_at DESC
	// 游标值：{age: 25, created_at: "2024-01-01"}
	// 条件：(age > 25) OR (age = 25 AND created_at < "2024-01-01")

	conditions := make([]mgo.M, 0)

	for i, field := range sortFields {
		value := data.Values[field]
		order := sortOrders[field]

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := mgo.ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if data.ID != "" {
				if oid, err := mgo.ObjectIDFromHex(data.ID); err == nil {
					value = oid
				}
			}
		}

		// 根据排序方向选择操作符
		var op string
		if data.Direction == "prev" {
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		} else {
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		}

		// 构建条件
		condition := mgo.M{}

		// 前面的字段必须相等
		for j := 0; j < i; j++ {
			prevField := sortFields[j]
			prevValue := data.Values[prevField]

			// 特殊处理 _id
			if prevField == "_id" {
				if hexID, ok := prevValue.(string); ok {
					if oid, err := mgo.ObjectIDFromHex(hexID); err == nil {
						prevValue = oid
					}
				}
			}

			condition[prevField] = prevValue
		}

		// 当前字段使用比较操作符
		condition[field] = mgo.M{op: value}

		conditions = append(conditions, condition)
	}

	// 如果只有一个条件，直接返回
	if len(conditions) == 1 {
		return conditions[0]
	}

	// 多个条件用 $or 连接
	return mgo.M{"$or": conditions}
}
