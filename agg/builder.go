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
//	builder.Sort(mgo.M{"count": -1})
func (b *Builder[T]) Sort(sort mgo.M) *Builder[T] {
	b.pipeline = append(b.pipeline, mgo.D{{Key: "$sort", Value: sort}})
	return b
}

// SortAsc 升序排序
//
// 示例：
//
//	builder.SortAsc("age")
func (b *Builder[T]) SortAsc(fields ...string) *Builder[T] {
	sort := mgo.M{}
	for _, field := range fields {
		sort[field] = 1
	}
	return b.Sort(sort)
}

// SortDesc 降序排序
//
// 示例：
//
//	builder.SortDesc("count")
func (b *Builder[T]) SortDesc(fields ...string) *Builder[T] {
	sort := mgo.M{}
	for _, field := range fields {
		sort[field] = -1
	}
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
func (g *GroupStage[T]) Sort(sort mgo.M) *Builder[T] {
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
