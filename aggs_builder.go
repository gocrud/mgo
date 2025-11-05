package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AggsBuilder 统一的聚合管道构建器
// 提供链式 API 用于构建和执行 MongoDB 聚合管道
//
// 使用示例：
//
//	// 基础聚合
//	var results []Result
//	err := coll.Aggs().
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{
//	        "total": Sum(1),
//	        "avgAge": Avg("$age"),
//	    }).
//	    Sort("total", -1).
//	    All(ctx, &results)
//
//	// 关联查询
//	err := coll.Aggs().
//	    Lookup("orders", "user_id", "_id", "orders").
//	    Unwind("$orders").
//	    All(ctx, &results)
type AggsBuilder struct {
	coll   *Collection
	ctx    context.Context
	stages []bson.D
}

// newAggsBuilder 创建聚合构建器（内部使用）
func newAggsBuilder(coll *Collection, ctx context.Context) *AggsBuilder {
	return &AggsBuilder{
		coll:   coll,
		ctx:    ctx,
		stages: make([]bson.D, 0),
	}
}

// ==================== 通用方法 ====================

// AddStage 添加自定义阶段
//
// 示例：
//
//	ab.AddStage(bson.D{{Key: "$match", Value: M{"status": "active"}}})
func (ab *AggsBuilder) AddStage(stage bson.D) *AggsBuilder {
	ab.stages = append(ab.stages, stage)
	return ab
}

// AddStages 添加多个自定义阶段
//
// 示例：
//
//	ab.AddStages(
//	    bson.D{{Key: "$match", Value: M{"status": "active"}}},
//	    bson.D{{Key: "$sort", Value: M{"created_at": -1}}},
//	)
func (ab *AggsBuilder) AddStages(stages ...bson.D) *AggsBuilder {
	ab.stages = append(ab.stages, stages...)
	return ab
}

// ==================== $match ====================

// Match 匹配阶段（使用 Filter 构建器）
//
// 示例：
//
//	ab.Match(Filter().Eq("status", "active").Gt("age", 18))
func (ab *AggsBuilder) Match(filter *FilterBuilder) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$match", Value: filter.BuildM()}})
	return ab
}

// MatchDoc 匹配阶段（使用文档）
//
// 示例：
//
//	ab.MatchDoc(M{"status": "active", "age": M{"$gt": 18}})
func (ab *AggsBuilder) MatchDoc(doc M) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$match", Value: doc}})
	return ab
}

// ==================== $project ====================

// Project 投影阶段（使用 Projection 构建器）
//
// 示例：
//
//	ab.Project(Projection().Include("name", "age").Exclude("_id"))
func (ab *AggsBuilder) Project(projection *Projection) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$project", Value: projection.BuildM()}})
	return ab
}

// ProjectDoc 投影阶段（使用文档）
//
// 示例：
//
//	ab.ProjectDoc(M{"name": 1, "age": 1, "_id": 0})
func (ab *AggsBuilder) ProjectDoc(doc M) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$project", Value: doc}})
	return ab
}

// ==================== $group ====================

// Group 分组阶段
//
// 示例：
//
//	ab.Group("$city", M{
//	    "total": Sum(1),
//	    "avgAge": Avg("$age"),
//	    "maxSalary": Max("$salary"),
//	})
func (ab *AggsBuilder) Group(id any, accumulators M) *AggsBuilder {
	groupDoc := M{"_id": id}
	for k, v := range accumulators {
		groupDoc[k] = v
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$group", Value: groupDoc}})
	return ab
}

// GroupBy 按字段分组（简化版）
//
// 示例：
//
//	ab.GroupBy("city", M{"count": Sum(1)})
func (ab *AggsBuilder) GroupBy(field string, accumulators M) *AggsBuilder {
	return ab.Group("$"+field, accumulators)
}

// ==================== $sort ====================

// Sort 排序阶段（字段名，方向）
//
// 示例：
//
//	ab.Sort("created_at", -1)  // 降序
//	ab.Sort("name", 1)          // 升序
func (ab *AggsBuilder) Sort(field string, direction int) *AggsBuilder {
	sortDoc := M{field: direction}
	ab.stages = append(ab.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return ab
}

// SortBy 排序阶段（使用 Sort 构建器）
//
// 示例：
//
//	ab.SortBy(Sort().Desc("created_at").Asc("name"))
func (ab *AggsBuilder) SortBy(sort *Sort) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$sort", Value: sort.BuildM()}})
	return ab
}

// SortDoc 排序阶段（使用文档）
//
// 示例：
//
//	ab.SortDoc(M{"created_at": -1, "name": 1})
func (ab *AggsBuilder) SortDoc(doc M) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$sort", Value: doc}})
	return ab
}

// SortAsc 升序排序
//
// 示例：
//
//	ab.SortAsc("name", "age")
func (ab *AggsBuilder) SortAsc(fields ...string) *AggsBuilder {
	sortDoc := M{}
	for _, field := range fields {
		sortDoc[field] = 1
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return ab
}

// SortDesc 降序排序
//
// 示例：
//
//	ab.SortDesc("created_at", "updated_at")
func (ab *AggsBuilder) SortDesc(fields ...string) *AggsBuilder {
	sortDoc := M{}
	for _, field := range fields {
		sortDoc[field] = -1
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return ab
}

// ==================== $limit / $skip ====================

// Limit 限制返回数量
//
// 示例：
//
//	ab.Limit(10)
func (ab *AggsBuilder) Limit(limit int64) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$limit", Value: limit}})
	return ab
}

// Skip 跳过记录数
//
// 示例：
//
//	ab.Skip(20)
func (ab *AggsBuilder) Skip(skip int64) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$skip", Value: skip}})
	return ab
}

// Page 分页（页码从 1 开始）
//
// 示例：
//
//	ab.Page(2, 20)  // 第2页，每页20条
func (ab *AggsBuilder) Page(page, pageSize int64) *AggsBuilder {
	if page < 1 {
		page = 1
	}
	skip := (page - 1) * pageSize
	return ab.Skip(skip).Limit(pageSize)
}

// ==================== $unwind ====================

// Unwind 展开数组字段
//
// 示例：
//
//	ab.Unwind("$tags")
func (ab *AggsBuilder) Unwind(path string) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$unwind", Value: path}})
	return ab
}

// UnwindPreserveEmpty 展开数组字段（保留空数组）
//
// 示例：
//
//	ab.UnwindPreserveEmpty("$tags")
func (ab *AggsBuilder) UnwindPreserveEmpty(path string) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{
		Key: "$unwind",
		Value: M{
			"path":                       path,
			"preserveNullAndEmptyArrays": true,
		},
	}})
	return ab
}

// ==================== $lookup ====================

// Lookup 关联查询（简单版）
//
// 示例：
//
//	ab.Lookup("orders", "user_id", "_id", "orders")
func (ab *AggsBuilder) Lookup(from, localField, foreignField, as string) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{
		Key: "$lookup",
		Value: M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	}})
	return ab
}

// LookupPipeline 关联查询（使用管道）
//
// 示例：
//
//	ab.LookupPipeline("orders", "user_orders", M{"user_id": "$_id"},
//	    Aggs().Match(Filter().Eq("status", "completed")))
func (ab *AggsBuilder) LookupPipeline(from, as string, let M, pipeline *AggsBuilder) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{
		Key: "$lookup",
		Value: M{
			"from":     from,
			"let":      let,
			"pipeline": pipeline.Build(),
			"as":       as,
		},
	}})
	return ab
}

// ==================== $addFields ====================

// AddFields 添加字段
//
// 示例：
//
//	ab.AddFields(M{
//	    "fullName": M{"$concat": []any{"$firstName", " ", "$lastName"}},
//	    "totalPrice": M{"$multiply": []any{"$price", "$quantity"}},
//	})
func (ab *AggsBuilder) AddFields(fields M) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$addFields", Value: fields}})
	return ab
}

// ==================== $replaceRoot ====================

// ReplaceRoot 替换根文档
//
// 示例：
//
//	ab.ReplaceRoot("$user")
//	ab.ReplaceRoot(M{"newRoot": "$embedded"})
func (ab *AggsBuilder) ReplaceRoot(newRoot any) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$replaceRoot", Value: M{"newRoot": newRoot}}})
	return ab
}

// ==================== $count ====================

// Count 计数阶段
//
// 示例：
//
//	ab.Count("total")
func (ab *AggsBuilder) Count(field string) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$count", Value: field}})
	return ab
}

// ==================== $facet ====================

// Facet 多面查询
//
// 示例：
//
//	ab.Facet(map[string]*AggsBuilder{
//	    "byCity": Aggs().GroupBy("city", M{"count": Sum(1)}),
//	    "byAge": Aggs().GroupBy("age", M{"count": Sum(1)}),
//	})
func (ab *AggsBuilder) Facet(facets map[string]*AggsBuilder) *AggsBuilder {
	facetDoc := M{}
	for name, pipeline := range facets {
		facetDoc[name] = pipeline.Build()
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$facet", Value: facetDoc}})
	return ab
}

// ==================== $bucket ====================

// Bucket 分桶
//
// 示例：
//
//	ab.Bucket("$age", []any{0, 18, 30, 50, 100}, "其他", M{
//	    "count": Sum(1),
//	    "avgSalary": Avg("$salary"),
//	})
func (ab *AggsBuilder) Bucket(groupBy any, boundaries []any, defaultBucket any, output M) *AggsBuilder {
	bucketDoc := M{
		"groupBy":    groupBy,
		"boundaries": boundaries,
		"output":     output,
	}
	if defaultBucket != nil {
		bucketDoc["default"] = defaultBucket
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$bucket", Value: bucketDoc}})
	return ab
}

// ==================== $sample ====================

// Sample 随机抽样
//
// 示例：
//
//	ab.Sample(10)  // 随机抽取10条
func (ab *AggsBuilder) Sample(size int64) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$sample", Value: M{"size": size}}})
	return ab
}

// ==================== $out / $merge ====================

// Out 输出到集合
//
// 示例：
//
//	ab.Out("result_collection")
func (ab *AggsBuilder) Out(collection string) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$out", Value: collection}})
	return ab
}

// Merge 合并到集合
//
// 示例：
//
//	ab.Merge("target_collection", "_id", "replace", "insert")
func (ab *AggsBuilder) Merge(collection string, on any, whenMatched, whenNotMatched string) *AggsBuilder {
	mergeDoc := M{
		"into": collection,
		"on":   on,
	}
	if whenMatched != "" {
		mergeDoc["whenMatched"] = whenMatched
	}
	if whenNotMatched != "" {
		mergeDoc["whenNotMatched"] = whenNotMatched
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$merge", Value: mergeDoc}})
	return ab
}

// ==================== $geoNear ====================

// GeoNear 地理位置查询
//
// 示例：
//
//	ab.GeoNear(
//	    M{"type": "Point", "coordinates": []float64{116.4074, 39.9042}},
//	    "distance",
//	    M{"maxDistance": 5000, "spherical": true},
//	)
func (ab *AggsBuilder) GeoNear(near any, distanceField string, options M) *AggsBuilder {
	geoNearDoc := M{
		"near":          near,
		"distanceField": distanceField,
	}
	for k, v := range options {
		geoNearDoc[k] = v
	}
	ab.stages = append(ab.stages, bson.D{{Key: "$geoNear", Value: geoNearDoc}})
	return ab
}

// ==================== $redact ====================

// Redact 条件过滤
//
// 示例：
//
//	ab.Redact(M{
//	    "$cond": M{
//	        "if": M{"$eq": []any{"$level", 5}},
//	        "then": "$$DESCEND",
//	        "else": "$$PRUNE",
//	    },
//	})
func (ab *AggsBuilder) Redact(expression any) *AggsBuilder {
	ab.stages = append(ab.stages, bson.D{{Key: "$redact", Value: expression}})
	return ab
}

// ==================== 执行方法 ====================

// Build 构建管道（返回原始 pipeline 数组）
func (ab *AggsBuilder) Build() []bson.D {
	return ab.stages
}

// One 执行聚合并返回单条结果
//
// 示例：
//
//	var result Result
//	err := coll.Aggs(ctx).
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{"total": Sum(1)}).
//	    One(&result)
func (ab *AggsBuilder) One(result any) error {
	cursor, err := ab.coll.coll.Aggregate(ab.ctx, ab.stages)
	if err != nil {
		return err
	}
	defer cursor.Close(ab.ctx)

	if !cursor.Next(ab.ctx) {
		if err := cursor.Err(); err != nil {
			return err
		}
		return mongo.ErrNoDocuments
	}

	return cursor.Decode(result)
}

// All 执行聚合并返回所有结果
//
// 示例：
//
//	var results []Result
//	err := coll.Aggs(ctx).
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{"total": Sum(1)}).
//	    All(&results)
func (ab *AggsBuilder) All(results any) error {
	cursor, err := ab.coll.coll.Aggregate(ab.ctx, ab.stages)
	if err != nil {
		return err
	}
	defer cursor.Close(ab.ctx)

	return cursor.All(ab.ctx, results)
}

// Cursor 获取游标（用于高级场景）
//
// 示例：
//
//	cursor, err := coll.Aggs(ctx).Match(...).Cursor()
//	defer cursor.Close(ab.ctx)
//	for cursor.Next(ab.ctx) {
//	    var doc Document
//	    cursor.Decode(&doc)
//	}
func (ab *AggsBuilder) Cursor() (*mongo.Cursor, error) {
	return ab.coll.coll.Aggregate(ab.ctx, ab.stages)
}

// ==================== 聚合累加器辅助函数 ====================

// Sum 求和累加器
//
// 示例：
//
//	M{"total": Sum(1)}          // 计数
//	M{"totalSales": Sum("$amount")}  // 求和
func Sum(expression any) M {
	return M{"$sum": expression}
}

// Avg 平均值累加器
//
// 示例：
//
//	M{"avgAge": Avg("$age")}
func Avg(expression any) M {
	return M{"$avg": expression}
}

// Max 最大值累加器
//
// 示例：
//
//	M{"maxPrice": Max("$price")}
func Max(expression any) M {
	return M{"$max": expression}
}

// Min 最小值累加器
//
// 示例：
//
//	M{"minPrice": Min("$price")}
func Min(expression any) M {
	return M{"$min": expression}
}

// First 第一个值累加器
//
// 示例：
//
//	M{"firstOrder": First("$order_date")}
func First(expression any) M {
	return M{"$first": expression}
}

// Last 最后一个值累加器
//
// 示例：
//
//	M{"lastOrder": Last("$order_date")}
func Last(expression any) M {
	return M{"$last": expression}
}

// Push 数组追加累加器
//
// 示例：
//
//	M{"orders": Push("$order_id")}
func Push(expression any) M {
	return M{"$push": expression}
}

// AddToSet 数组去重追加累加器
//
// 示例：
//
//	M{"uniqueTags": AddToSet("$tag")}
func AddToSet(expression any) M {
	return M{"$addToSet": expression}
}

// StdDevPop 总体标准差累加器
//
// 示例：
//
//	M{"stdDev": StdDevPop("$score")}
func StdDevPop(expression any) M {
	return M{"$stdDevPop": expression}
}

// StdDevSamp 样本标准差累加器
//
// 示例：
//
//	M{"stdDev": StdDevSamp("$score")}
func StdDevSamp(expression any) M {
	return M{"$stdDevSamp": expression}
}
