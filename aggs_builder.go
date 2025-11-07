package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// AggsBuilder 聚合执行器
// 专门负责执行聚合操作，stage 构建使用 StageBuilder
//
// 使用示例：
//
//	// 构建 pipeline
//	pipeline := Stage().
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{"total": Sum(1)}).
//	    Sort("total", -1)
//
//	// 执行聚合
//	var results []Result
//	err := coll.Aggs(ctx).Stage(pipeline).All(&results)
//
//	// 或者内联方式
//	err := coll.Aggs(ctx).
//	    Stage(
//	        Stage().
//	            Match(Filter().Eq("status", "active")).
//	            Limit(10),
//	    ).
//	    All(&results)
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

// ==================== Pipeline 设置方法 ====================

// Pipes 设置原始 pipeline stages
//
// 示例：
//
//	stages := []bson.D{
//	    {{Key: "$match", Value: M{"status": "active"}}},
//	    {{Key: "$sort", Value: M{"created_at": -1}}},
//	}
//	err := coll.Aggs(ctx).Pipes(stages).All(&results)
func (ab *AggsBuilder) Pipes(stages []bson.D) *AggsBuilder {
	ab.stages = stages
	return ab
}

// Stage 设置 StageBuilder 构建的 pipeline
//
// 示例：
//
//	pipeline := Stage().Match(Filter().Eq("status", "active")).Limit(10)
//	err := coll.Aggs(ctx).Stage(pipeline).All(&results)
func (ab *AggsBuilder) Stage(sb *StageBuilder) *AggsBuilder {
	ab.stages = sb.Build()
	return ab
}

// AppendPipes 追加原始 stages
//
// 示例：
//
//	err := coll.Aggs(ctx).
//	    Stage(Stage().Match(Filter().Eq("type", "user"))).
//	    AppendPipes([]bson.D{
//	        {{Key: "$sort", Value: M{"created_at": -1}}},
//	    }).
//	    All(&results)
func (ab *AggsBuilder) AppendPipes(stages []bson.D) *AggsBuilder {
	ab.stages = append(ab.stages, stages...)
	return ab
}

// AppendStage 追加 StageBuilder 构建的 stages
//
// 示例：
//
//	err := coll.Aggs(ctx).
//	    AddPipe(bson.D{{Key: "$match", Value: M{"type": "user"}}}).
//	    AppendStage(Stage().Limit(10).Skip(20)).
//	    All(&results)
func (ab *AggsBuilder) AppendStage(sb *StageBuilder) *AggsBuilder {
	ab.stages = append(ab.stages, sb.Build()...)
	return ab
}

// AddPipe 添加单个原始 stage
//
// 示例：
//
//	err := coll.Aggs(ctx).
//	    AddPipe(bson.D{{Key: "$match", Value: M{"status": "active"}}}).
//	    AddPipe(bson.D{{Key: "$sort", Value: M{"created_at": -1}}}).
//	    All(&results)
func (ab *AggsBuilder) AddPipe(stage bson.D) *AggsBuilder {
	ab.stages = append(ab.stages, stage)
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
//	    Stage(
//	        Stage().
//	            Match(Filter().Eq("status", "active")).
//	            Group("$city", M{"total": Sum(1)}),
//	    ).
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
//	    Stage(
//	        Stage().
//	            Match(Filter().Eq("status", "active")).
//	            Group("$city", M{"total": Sum(1)}),
//	    ).
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
//	cursor, err := coll.Aggs(ctx).
//	    Stage(Stage().Match(...)).
//	    Cursor()
//	defer cursor.Close(ctx)
//	for cursor.Next(ctx) {
//	    var doc Document
//	    cursor.Decode(&doc)
//	}
func (ab *AggsBuilder) Cursor() (*mongo.Cursor, error) {
	return ab.coll.coll.Aggregate(ab.ctx, ab.stages)
}
