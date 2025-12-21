package agg

import "github.com/gocrud/mgo"

// ==================== Pipeline 辅助函数 ====================

// NewPipeline 创建新的 Pipeline
//
// 示例：
//
//	pipeline := agg.NewPipeline().
//	    Match(mgo.Eq("status", "active")).
//	    Group("$city", mgo.M{"count": agg.Sum(1)}).
//	    Build()
func NewPipeline() *PipelineBuilder {
	return &PipelineBuilder{
		stages: mgo.Pipeline{},
	}
}

// PipelineBuilder Pipeline 构建器
type PipelineBuilder struct {
	stages mgo.Pipeline
}

// Match 添加 match 阶段
func (p *PipelineBuilder) Match(filter mgo.M) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$match", Value: filter}})
	return p
}

// Group 添加 group 阶段
func (p *PipelineBuilder) Group(id string, fields mgo.M) *PipelineBuilder {
	group := mgo.M{"_id": id}
	for k, v := range fields {
		group[k] = v
	}
	p.stages = append(p.stages, mgo.D{{Key: "$group", Value: group}})
	return p
}

// Sort 添加 sort 阶段
func (p *PipelineBuilder) Sort(sort mgo.D) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$sort", Value: sort}})
	return p
}

// Limit 添加 limit 阶段
func (p *PipelineBuilder) Limit(n int64) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$limit", Value: n}})
	return p
}

// Skip 添加 skip 阶段
func (p *PipelineBuilder) Skip(n int64) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$skip", Value: n}})
	return p
}

// Project 添加 project 阶段
func (p *PipelineBuilder) Project(projection mgo.M) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$project", Value: projection}})
	return p
}

// Unwind 添加 unwind 阶段
func (p *PipelineBuilder) Unwind(path string) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$unwind", Value: path}})
	return p
}

// Lookup 添加 lookup 阶段
func (p *PipelineBuilder) Lookup(from, localField, foreignField, as string) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$lookup", Value: mgo.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}}})
	return p
}

// AddFields 添加 addFields 阶段
func (p *PipelineBuilder) AddFields(fields mgo.M) *PipelineBuilder {
	p.stages = append(p.stages, mgo.D{{Key: "$addFields", Value: fields}})
	return p
}

// Build 构建最终的 Pipeline
func (p *PipelineBuilder) Build() mgo.Pipeline {
	return p.stages
}

// ==================== 快捷 Pipeline 函数 ====================

// MatchStage 创建 match 阶段
func MatchStage(filter mgo.M) mgo.D {
	return mgo.D{{Key: "$match", Value: filter}}
}

// GroupStageDoc 创建 group 阶段文档
func GroupStageDoc(id string, fields mgo.M) mgo.D {
	group := mgo.M{"_id": id}
	for k, v := range fields {
		group[k] = v
	}
	return mgo.D{{Key: "$group", Value: group}}
}

// SortStage 创建 sort 阶段
func SortStage(sort mgo.D) mgo.D {
	return mgo.D{{Key: "$sort", Value: sort}}
}

// LimitStage 创建 limit 阶段
func LimitStage(n int64) mgo.D {
	return mgo.D{{Key: "$limit", Value: n}}
}

// SkipStage 创建 skip 阶段
func SkipStage(n int64) mgo.D {
	return mgo.D{{Key: "$skip", Value: n}}
}

// ProjectStage 创建 project 阶段
func ProjectStage(projection mgo.M) mgo.D {
	return mgo.D{{Key: "$project", Value: projection}}
}

// UnwindStage 创建 unwind 阶段
func UnwindStage(path string) mgo.D {
	return mgo.D{{Key: "$unwind", Value: path}}
}

// LookupStage 创建 lookup 阶段
func LookupStage(from, localField, foreignField, as string) mgo.D {
	return mgo.D{{Key: "$lookup", Value: mgo.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}}}
}
