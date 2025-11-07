package mgo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// StageBuilder 聚合 Stage 构建器
// 专门用于构建各种聚合管道阶段，不依赖 Collection 和 Context
//
// 使用示例：
//
//	// 构建 pipeline
//	pipeline := Stage().
//	    Match(Filter().Eq("status", "active")).
//	    Sort("created_at", -1).
//	    Limit(10)
//
//	// 执行聚合
//	var results []Result
//	err := coll.Aggs(ctx).Stage(pipeline).All(&results)
type StageBuilder struct {
	stages []bson.D
}

// Stage 创建新的 StageBuilder
func Stage() *StageBuilder {
	return &StageBuilder{
		stages: make([]bson.D, 0),
	}
}

// ==================== 基础方法 ====================

// Build 返回构建的 stages
func (sb *StageBuilder) Build() []bson.D {
	return sb.stages
}

// Clone 克隆一个新的 StageBuilder（用于复用）
//
// 示例：
//
//	base := Stage().Match(Filter().Eq("status", "active"))
//	pipeline1 := base.Clone().Limit(10)
//	pipeline2 := base.Clone().Limit(20)
func (sb *StageBuilder) Clone() *StageBuilder {
	newStages := make([]bson.D, len(sb.stages))
	copy(newStages, sb.stages)
	return &StageBuilder{stages: newStages}
}

// Append 追加另一个 StageBuilder 的 stages
//
// 示例：
//
//	common := Stage().Match(Filter().Eq("status", "active"))
//	pipeline := Stage().
//	    Append(common).
//	    Limit(10)
func (sb *StageBuilder) Append(other *StageBuilder) *StageBuilder {
	sb.stages = append(sb.stages, other.Build()...)
	return sb
}

// Prepend 在前面插入另一个 StageBuilder 的 stages
//
// 示例：
//
//	filter := Stage().Match(Filter().Eq("status", "active"))
//	pipeline := Stage().
//	    Limit(10).
//	    Prepend(filter)
func (sb *StageBuilder) Prepend(other *StageBuilder) *StageBuilder {
	sb.stages = append(other.Build(), sb.stages...)
	return sb
}

// AddStage 添加单个自定义 stage
//
// 示例：
//
//	sb.AddStage(bson.D{{Key: "$match", Value: M{"status": "active"}}})
func (sb *StageBuilder) AddStage(stage bson.D) *StageBuilder {
	sb.stages = append(sb.stages, stage)
	return sb
}

// AddStages 添加多个自定义 stages
//
// 示例：
//
//	sb.AddStages(
//	    bson.D{{Key: "$match", Value: M{"status": "active"}}},
//	    bson.D{{Key: "$sort", Value: M{"created_at": -1}}},
//	)
func (sb *StageBuilder) AddStages(stages ...bson.D) *StageBuilder {
	sb.stages = append(sb.stages, stages...)
	return sb
}

// ==================== $match ====================

// Match 匹配阶段（使用 Filter 构建器）
//
// 示例：
//
//	sb.Match(Filter().Eq("status", "active").Gt("age", 18))
func (sb *StageBuilder) Match(filter *FilterBuilder) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$match", Value: filter.BuildM()}})
	return sb
}

// MatchDoc 匹配阶段（使用文档）
//
// 示例：
//
//	sb.MatchDoc(M{"status": "active", "age": M{"$gt": 18}})
func (sb *StageBuilder) MatchDoc(doc M) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$match", Value: doc}})
	return sb
}

// ==================== $project ====================

// Project 投影阶段（使用 Projection 构建器）
//
// 示例：
//
//	sb.Project(Projection().Include("name", "age").Exclude("_id"))
func (sb *StageBuilder) Project(projection *Projection) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$project", Value: projection.BuildM()}})
	return sb
}

// ProjectDoc 投影阶段（使用文档）
//
// 示例：
//
//	sb.ProjectDoc(M{"name": 1, "age": 1, "_id": 0})
func (sb *StageBuilder) ProjectDoc(doc M) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$project", Value: doc}})
	return sb
}

// ==================== $group ====================

// Group 分组阶段
//
// 示例：
//
//	sb.Group("$city", M{
//	    "total": Sum(1),
//	    "avgAge": Avg("$age"),
//	    "maxSalary": Max("$salary"),
//	})
func (sb *StageBuilder) Group(id any, accumulators M) *StageBuilder {
	groupDoc := M{"_id": id}
	for k, v := range accumulators {
		groupDoc[k] = v
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$group", Value: groupDoc}})
	return sb
}

// GroupBy 按字段分组（简化版）
//
// 示例：
//
//	sb.GroupBy("city", M{"count": Sum(1)})
func (sb *StageBuilder) GroupBy(field string, accumulators M) *StageBuilder {
	return sb.Group("$"+field, accumulators)
}

// ==================== $sort ====================

// Sort 排序阶段（字段名，方向）
//
// 示例：
//
//	sb.Sort("created_at", -1)  // 降序
//	sb.Sort("name", 1)          // 升序
func (sb *StageBuilder) Sort(field string, direction int) *StageBuilder {
	sortDoc := M{field: direction}
	sb.stages = append(sb.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return sb
}

// SortBy 排序阶段（使用 Sort 构建器）
//
// 示例：
//
//	sb.SortBy(Sort().Desc("created_at").Asc("name"))
func (sb *StageBuilder) SortBy(sort *Sort) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$sort", Value: sort.BuildM()}})
	return sb
}

// SortDoc 排序阶段（使用文档）
//
// 示例：
//
//	sb.SortDoc(M{"created_at": -1, "name": 1})
func (sb *StageBuilder) SortDoc(doc M) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$sort", Value: doc}})
	return sb
}

// SortAsc 升序排序
//
// 示例：
//
//	sb.SortAsc("name", "age")
func (sb *StageBuilder) SortAsc(fields ...string) *StageBuilder {
	sortDoc := M{}
	for _, field := range fields {
		sortDoc[field] = 1
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return sb
}

// SortDesc 降序排序
//
// 示例：
//
//	sb.SortDesc("created_at", "updated_at")
func (sb *StageBuilder) SortDesc(fields ...string) *StageBuilder {
	sortDoc := M{}
	for _, field := range fields {
		sortDoc[field] = -1
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$sort", Value: sortDoc}})
	return sb
}

// ==================== $limit / $skip ====================

// Limit 限制返回数量
//
// 示例：
//
//	sb.Limit(10)
func (sb *StageBuilder) Limit(limit int64) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$limit", Value: limit}})
	return sb
}

// Skip 跳过记录数
//
// 示例：
//
//	sb.Skip(20)
func (sb *StageBuilder) Skip(skip int64) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$skip", Value: skip}})
	return sb
}

// Page 分页（页码从 1 开始）
//
// 示例：
//
//	sb.Page(2, 20)  // 第2页，每页20条
func (sb *StageBuilder) Page(page, pageSize int64) *StageBuilder {
	if page < 1 {
		page = 1
	}
	skip := (page - 1) * pageSize
	return sb.Skip(skip).Limit(pageSize)
}

// ==================== $unwind ====================

// Unwind 展开数组字段
//
// 示例：
//
//	sb.Unwind("$tags")
func (sb *StageBuilder) Unwind(path string) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$unwind", Value: path}})
	return sb
}

// UnwindPreserveEmpty 展开数组字段（保留空数组）
//
// 示例：
//
//	sb.UnwindPreserveEmpty("$tags")
func (sb *StageBuilder) UnwindPreserveEmpty(path string) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{
		Key: "$unwind",
		Value: M{
			"path":                       path,
			"preserveNullAndEmptyArrays": true,
		},
	}})
	return sb
}

// ==================== $lookup ====================

// Lookup 关联查询（简单版）
//
// 示例：
//
//	sb.Lookup("orders", "user_id", "_id", "orders")
func (sb *StageBuilder) Lookup(from, localField, foreignField, as string) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{
		Key: "$lookup",
		Value: M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	}})
	return sb
}

// LookupStage 关联查询（使用 StageBuilder pipeline）
//
// 示例：
//
//	sb.LookupStage("orders", "user_orders", M{"user_id": "$_id"},
//	    Stage().Match(Filter().Eq("status", "completed")))
func (sb *StageBuilder) LookupStage(from, as string, let M, pipeline *StageBuilder) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{
		Key: "$lookup",
		Value: M{
			"from":     from,
			"let":      let,
			"pipeline": pipeline.Build(),
			"as":       as,
		},
	}})
	return sb
}

// ==================== $addFields ====================

// AddFields 添加字段
//
// 示例：
//
//	sb.AddFields(M{
//	    "fullName": M{"$concat": []any{"$firstName", " ", "$lastName"}},
//	    "totalPrice": M{"$multiply": []any{"$price", "$quantity"}},
//	})
func (sb *StageBuilder) AddFields(fields M) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$addFields", Value: fields}})
	return sb
}

// ==================== $replaceRoot ====================

// ReplaceRoot 替换根文档
//
// 示例：
//
//	sb.ReplaceRoot("$user")
//	sb.ReplaceRoot(M{"newRoot": "$embedded"})
func (sb *StageBuilder) ReplaceRoot(newRoot any) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$replaceRoot", Value: M{"newRoot": newRoot}}})
	return sb
}

// ==================== $count ====================

// Count 计数阶段
//
// 示例：
//
//	sb.Count("total")
func (sb *StageBuilder) Count(field string) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$count", Value: field}})
	return sb
}

// ==================== $facet ====================

// Facet 多面查询
//
// 示例：
//
//	sb.Facet(map[string]*StageBuilder{
//	    "byCity": Stage().GroupBy("city", M{"count": Sum(1)}),
//	    "byAge": Stage().GroupBy("age", M{"count": Sum(1)}),
//	})
func (sb *StageBuilder) Facet(facets map[string]*StageBuilder) *StageBuilder {
	facetDoc := M{}
	for name, pipeline := range facets {
		facetDoc[name] = pipeline.Build()
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$facet", Value: facetDoc}})
	return sb
}

// ==================== $bucket ====================

// Bucket 分桶
//
// 示例：
//
//	sb.Bucket("$age", []any{0, 18, 30, 50, 100}, "其他", M{
//	    "count": Sum(1),
//	    "avgSalary": Avg("$salary"),
//	})
func (sb *StageBuilder) Bucket(groupBy any, boundaries []any, defaultBucket any, output M) *StageBuilder {
	bucketDoc := M{
		"groupBy":    groupBy,
		"boundaries": boundaries,
		"output":     output,
	}
	if defaultBucket != nil {
		bucketDoc["default"] = defaultBucket
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$bucket", Value: bucketDoc}})
	return sb
}

// ==================== $sample ====================

// Sample 随机抽样
//
// 示例：
//
//	sb.Sample(10)  // 随机抽取10条
func (sb *StageBuilder) Sample(size int64) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$sample", Value: M{"size": size}}})
	return sb
}

// ==================== $out / $merge ====================

// Out 输出到集合
//
// 示例：
//
//	sb.Out("result_collection")
func (sb *StageBuilder) Out(collection string) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$out", Value: collection}})
	return sb
}

// Merge 合并到集合
//
// 示例：
//
//	sb.Merge("target_collection", "_id", "replace", "insert")
func (sb *StageBuilder) Merge(collection string, on any, whenMatched, whenNotMatched string) *StageBuilder {
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
	sb.stages = append(sb.stages, bson.D{{Key: "$merge", Value: mergeDoc}})
	return sb
}

// ==================== $geoNear ====================

// GeoNear 地理位置查询
//
// 示例：
//
//	sb.GeoNear(
//	    M{"type": "Point", "coordinates": []float64{116.4074, 39.9042}},
//	    "distance",
//	    M{"maxDistance": 5000, "spherical": true},
//	)
func (sb *StageBuilder) GeoNear(near any, distanceField string, options M) *StageBuilder {
	geoNearDoc := M{
		"near":          near,
		"distanceField": distanceField,
	}
	for k, v := range options {
		geoNearDoc[k] = v
	}
	sb.stages = append(sb.stages, bson.D{{Key: "$geoNear", Value: geoNearDoc}})
	return sb
}

// ==================== $redact ====================

// Redact 条件过滤
//
// 示例：
//
//	sb.Redact(M{
//	    "$cond": M{
//	        "if": M{"$eq": []any{"$level", 5}},
//	        "then": "$$DESCEND",
//	        "else": "$$PRUNE",
//	    },
//	})
func (sb *StageBuilder) Redact(expression any) *StageBuilder {
	sb.stages = append(sb.stages, bson.D{{Key: "$redact", Value: expression}})
	return sb
}
