package agg

import (
	"context"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 聚合构建器 ====================

// Builder 聚合构建器
type Builder[T any] struct {
	coll     *mongo.Collection
	ctx      context.Context
	pipeline mgo.Pipeline
	sort     mgo.D // 排序信息（用于游标分页）
}

// Aggregate 创建聚合构建器
func Aggregate[T any](source interface{}) *Builder[T] {
	var coll *mongo.Collection
	var ctx context.Context = context.Background()

	switch src := source.(type) {
	case interface{ Native() *mongo.Collection }:
		coll = src.Native()
		// 尝试获取上下文
		if c, ok := source.(interface{ Context() context.Context }); ok {
			ctx = c.Context()
		}
	case *mongo.Collection:
		coll = src
	default:
		panic("agg: invalid source type")
	}

	return &Builder[T]{
		coll:     coll,
		ctx:      ctx,
		pipeline: mgo.Pipeline{},
	}
}

// WithContext 设置上下文
func (b *Builder[T]) WithContext(ctx context.Context) *Builder[T] {
	b.ctx = ctx
	return b
}

// Context 获取上下文
func (b *Builder[T]) Context() context.Context {
	if b.ctx == nil {
		return context.Background()
	}
	return b.ctx
}

// ==================== Pipeline 构建方法 ====================

// Match 添加匹配阶段
func (b *Builder[T]) Match(filter mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$match", Value: filter}})
	return b
}

// GroupBy 添加分组阶段
func (b *Builder[T]) GroupBy(field interface{}) *GroupStage[T] {
	return &GroupStage[T]{
		builder: b,
		groupID: field,
		fields:  mgo.D{},
	}
}

// Sort 添加排序阶段
func (b *Builder[T]) Sort(sort mgo.D) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$sort", Value: sort}})
	b.sort = sort // 记录排序信息
	return b
}

// SortAsc 升序排序
func (b *Builder[T]) SortAsc(fields ...string) *Builder[T] {
	sort := make(mgo.D, 0, len(fields))
	for _, field := range fields {
		sort = append(sort, mgo.E{Key: field, Value: 1})
	}
	b.sort = sort // 记录排序信息
	return b.Sort(sort)
}

// SortDesc 降序排序
func (b *Builder[T]) SortDesc(fields ...string) *Builder[T] {
	sort := make(mgo.D, 0, len(fields))
	for _, field := range fields {
		sort = append(sort, mgo.E{Key: field, Value: -1})
	}
	b.sort = sort // 记录排序信息
	return b.Sort(sort)
}

// Limit 限制数量
func (b *Builder[T]) Limit(n int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$limit", Value: n}})
	return b
}

// Skip 跳过数量
func (b *Builder[T]) Skip(n int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$skip", Value: n}})
	return b
}

// Project 添加投影阶段
func (b *Builder[T]) Project(projection mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$project", Value: projection}})
	return b
}

// Unwind 展开数组
func (b *Builder[T]) Unwind(path string) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$unwind", Value: path}})
	return b
}

// Lookup 关联查询
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
func (b *Builder[T]) AddFields(fields mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$addFields", Value: fields}})
	return b
}

// ReplaceRoot 替换根文档
func (b *Builder[T]) ReplaceRoot(newRoot string) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$replaceRoot", Value: mgo.M{
		"newRoot": newRoot,
	}}})
	return b
}

// Sample 随机抽样
func (b *Builder[T]) Sample(size int64) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$sample", Value: mgo.M{
		"size": size,
	}}})
	return b
}

// ==================== 执行方法 ====================

// All 执行聚合并返回所有结果
func (b *Builder[T]) All() ([]*T, error) {
	ctx := b.Context()
	cursor, err := b.coll.Aggregate(ctx, b.pipeline)
	if err != nil {
		return nil, mgo.WrapError(err, "failed to aggregate")
	}
	defer cursor.Close(ctx)

	var results []*T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, mgo.WrapError(err, "failed to decode aggregate results")
	}

	return results, nil
}

// One 执行聚合并返回第一条结果
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
func (b *Builder[T]) Count(ctx context.Context) (int64, error) {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$count", Value: "count"}})

	type countResult struct {
		Count int64 `bson:"count"`
	}

	cursor, err := b.coll.Aggregate(ctx, b.pipeline)
	if err != nil {
		return 0, mgo.WrapError(err, "failed to count aggregate")
	}
	defer cursor.Close(ctx)

	var result countResult
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, mgo.WrapError(err, "failed to decode count result")
		}
		return result.Count, nil
	}

	return 0, nil
}

// Pipeline 获取构建的 pipeline
func (b *Builder[T]) Pipeline() mgo.Pipeline {
	return b.pipeline
}

// CursorPage 使用游标的分页
func (b *Builder[T]) CursorPage(cursor string, perPage int) (*mgo.CursorPage[T], error) {
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}

	ctx := b.Context()

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
			// 游标解析失败，忽略并返回第一页
		} else {
			cursorDir = data.Direction

			// 根据游标方向反转排序
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
				insertPos := len(pipeline)
				for i := len(pipeline) - 1; i >= 0; i-- {
					if len(pipeline[i]) > 0 {
						if pipeline[i][0].Key == "$sort" {
							insertPos = i
						}
					}
				}

				matchStage := mgo.D{{Key: "$match", Value: filter}}
				pipeline = append(pipeline[:insertPos], append(mgo.Pipeline{matchStage}, pipeline[insertPos:]...)...)
			}

			// 确保排序在正确位置
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

	// 添加 limit
	limit := int64(perPage + 1)
	pipeline = append(pipeline, mgo.D{{Key: "$limit", Value: limit}})

	// 执行聚合
	aggCursor, err := b.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, mgo.WrapError(err, "failed to aggregate for cursor page")
	}
	defer aggCursor.Close(ctx)

	var items []*T
	if err := aggCursor.All(ctx, &items); err != nil {
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

	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = mgo.EncodeCursorFromItem(lastItem, sortDoc, "next")
	}

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
	groupID interface{} // 支持复合 ID
	fields  mgo.D       // 使用有序切片
}

// Count 统计数量
func (g *GroupStage[T]) Count(field string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$sum": 1}})
	return g
}

// CountIf 条件统计数量
//
// 示例：
//
//	g.CountIf("adults", agg.Gt("$age", 18))
func (g *GroupStage[T]) CountIf(field string, cond interface{}) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$sum": mgo.M{"$cond": []interface{}{cond, 1, 0}}}})
	return g
}

// Sum 求和
func (g *GroupStage[T]) Sum(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$sum": expr}})
	return g
}

// SumIf 条件求和
//
// 示例：
//
//	g.SumIf("vip_balance", "$balance", agg.Eq("$vip", true))
func (g *GroupStage[T]) SumIf(field string, expr interface{}, cond interface{}) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$sum": mgo.M{"$cond": []interface{}{cond, expr, 0}}}})
	return g
}

// Avg 平均值
func (g *GroupStage[T]) Avg(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$avg": expr}})
	return g
}

// Max 最大值
func (g *GroupStage[T]) Max(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$max": expr}})
	return g
}

// Min 最小值
func (g *GroupStage[T]) Min(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$min": expr}})
	return g
}

// First 第一个值
func (g *GroupStage[T]) First(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$first": expr}})
	return g
}

// Last 最后一个值
func (g *GroupStage[T]) Last(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$last": expr}})
	return g
}

// Push 收集到数组
func (g *GroupStage[T]) Push(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$push": expr}})
	return g
}

// AddToSet 收集到数组（去重）
func (g *GroupStage[T]) AddToSet(field, expr string) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: mgo.M{"$addToSet": expr}})
	return g
}

// Custom 自定义聚合操作
//
// 示例：
//
//	g.Custom("custom_field", mgo.M{"$sum": mgo.M{"$cond": ...}})
func (g *GroupStage[T]) Custom(field string, expr interface{}) *GroupStage[T] {
	g.fields = append(g.fields, mgo.E{Key: field, Value: expr})
	return g
}

// 完成分组并返回 Builder
func (g *GroupStage[T]) build() *Builder[T] {
	// 构建 group 文档，确保 _id 在最前面
	group := mgo.D{{Key: "_id", Value: g.groupID}}
	group = append(group, g.fields...)

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
