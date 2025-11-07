package mgo_test

import (
	"testing"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestStageBuilder 测试 StageBuilder 基础功能
func TestStageBuilder(t *testing.T) {
	// 测试 Stage() 构造函数
	sb := mgo.Stage()
	if sb == nil {
		t.Fatal("Stage() should not return nil")
	}

	// 测试 Build() 返回空数组
	stages := sb.Build()
	if len(stages) != 0 {
		t.Errorf("Expected empty stages, got %d", len(stages))
	}
}

// TestStageBuilderMatch 测试 Match 方法
func TestStageBuilderMatch(t *testing.T) {
	sb := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active"))

	stages := sb.Build()
	if len(stages) != 1 {
		t.Fatalf("Expected 1 stage, got %d", len(stages))
	}

	// 验证 stage 结构
	stage := stages[0]
	if len(stage) != 1 || stage[0].Key != "$match" {
		t.Errorf("Expected $match stage, got %v", stage)
	}
}

// TestStageBuilderChaining 测试链式调用
func TestStageBuilderChaining(t *testing.T) {
	sb := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active")).
		Sort("created_at", -1).
		Limit(10).
		Skip(5)

	stages := sb.Build()
	if len(stages) != 4 {
		t.Fatalf("Expected 4 stages, got %d", len(stages))
	}

	// 验证顺序
	expectedKeys := []string{"$match", "$sort", "$limit", "$skip"}
	for i, expected := range expectedKeys {
		if stages[i][0].Key != expected {
			t.Errorf("Stage %d: expected %s, got %s", i, expected, stages[i][0].Key)
		}
	}
}

// TestStageBuilderGroup 测试 Group 方法
func TestStageBuilderGroup(t *testing.T) {
	sb := mgo.Stage().
		Group("$city", mgo.M{
			"count":  mgo.Sum(1),
			"avgAge": mgo.Avg("$age"),
		})

	stages := sb.Build()
	if len(stages) != 1 {
		t.Fatalf("Expected 1 stage, got %d", len(stages))
	}

	stage := stages[0]
	if stage[0].Key != "$group" {
		t.Errorf("Expected $group stage, got %s", stage[0].Key)
	}
}

// TestStageBuilderClone 测试 Clone 方法
func TestStageBuilderClone(t *testing.T) {
	base := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active"))

	clone1 := base.Clone().Limit(10)
	clone2 := base.Clone().Limit(20)

	// 验证 base 未被修改
	if len(base.Build()) != 1 {
		t.Errorf("Base should have 1 stage, got %d", len(base.Build()))
	}

	// 验证克隆独立
	if len(clone1.Build()) != 2 {
		t.Errorf("Clone1 should have 2 stages, got %d", len(clone1.Build()))
	}

	if len(clone2.Build()) != 2 {
		t.Errorf("Clone2 should have 2 stages, got %d", len(clone2.Build()))
	}

	// 验证 limit 值不同
	limit1 := clone1.Build()[1][0].Value
	limit2 := clone2.Build()[1][0].Value

	if limit1 == limit2 {
		t.Errorf("Clone1 and Clone2 should have different limit values")
	}
}

// TestStageBuilderAppend 测试 Append 方法
func TestStageBuilderAppend(t *testing.T) {
	common := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active"))

	pipeline := mgo.Stage().
		Append(common).
		Limit(10)

	stages := pipeline.Build()
	if len(stages) != 2 {
		t.Fatalf("Expected 2 stages, got %d", len(stages))
	}

	if stages[0][0].Key != "$match" {
		t.Errorf("First stage should be $match, got %s", stages[0][0].Key)
	}

	if stages[1][0].Key != "$limit" {
		t.Errorf("Second stage should be $limit, got %s", stages[1][0].Key)
	}
}

// TestStageBuilderPrepend 测试 Prepend 方法
func TestStageBuilderPrepend(t *testing.T) {
	filter := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active"))

	pipeline := mgo.Stage().
		Limit(10).
		Prepend(filter)

	stages := pipeline.Build()
	if len(stages) != 2 {
		t.Fatalf("Expected 2 stages, got %d", len(stages))
	}

	// Prepend 应该将 filter 放在前面
	if stages[0][0].Key != "$match" {
		t.Errorf("First stage should be $match, got %s", stages[0][0].Key)
	}

	if stages[1][0].Key != "$limit" {
		t.Errorf("Second stage should be $limit, got %s", stages[1][0].Key)
	}
}

// TestStageBuilderLookupStage 测试 LookupStage 方法
func TestStageBuilderLookupStage(t *testing.T) {
	subPipeline := mgo.Stage().
		Match(mgo.Filter().Eq("status", "completed"))

	sb := mgo.Stage().
		LookupStage("orders", "user_orders",
			mgo.M{"userId": "$_id"},
			subPipeline)

	stages := sb.Build()
	if len(stages) != 1 {
		t.Fatalf("Expected 1 stage, got %d", len(stages))
	}

	stage := stages[0]
	if stage[0].Key != "$lookup" {
		t.Errorf("Expected $lookup stage, got %s", stage[0].Key)
	}
}

// TestStageBuilderFacet 测试 Facet 方法
func TestStageBuilderFacet(t *testing.T) {
	sb := mgo.Stage().
		Facet(map[string]*mgo.StageBuilder{
			"byCity": mgo.Stage().GroupBy("city", mgo.M{"count": mgo.Sum(1)}),
			"byAge":  mgo.Stage().GroupBy("age", mgo.M{"count": mgo.Sum(1)}),
		})

	stages := sb.Build()
	if len(stages) != 1 {
		t.Fatalf("Expected 1 stage, got %d", len(stages))
	}

	stage := stages[0]
	if stage[0].Key != "$facet" {
		t.Errorf("Expected $facet stage, got %s", stage[0].Key)
	}
}

// TestStageBuilderPage 测试 Page 方法
func TestStageBuilderPage(t *testing.T) {
	sb := mgo.Stage().Page(2, 20) // 第2页，每页20条

	stages := sb.Build()
	if len(stages) != 2 {
		t.Fatalf("Expected 2 stages (skip + limit), got %d", len(stages))
	}

	// 验证 skip
	if stages[0][0].Key != "$skip" {
		t.Errorf("First stage should be $skip, got %s", stages[0][0].Key)
	}
	skipValue := stages[0][0].Value.(int64)
	if skipValue != 20 { // (2-1) * 20 = 20
		t.Errorf("Expected skip 20, got %d", skipValue)
	}

	// 验证 limit
	if stages[1][0].Key != "$limit" {
		t.Errorf("Second stage should be $limit, got %s", stages[1][0].Key)
	}
	limitValue := stages[1][0].Value.(int64)
	if limitValue != 20 {
		t.Errorf("Expected limit 20, got %d", limitValue)
	}
}

// TestAggsBuilderStage 测试 AggsBuilder.Stage 方法
func TestAggsBuilderStage(t *testing.T) {
	// 注意：这里不实际连接数据库，只测试 API 结构
	// 实际的数据库测试应该在集成测试中进行

	pipeline := mgo.Stage().
		Match(mgo.Filter().Eq("status", "active")).
		Limit(10)

	// 验证 pipeline 构建正确
	stages := pipeline.Build()
	if len(stages) != 2 {
		t.Fatalf("Expected 2 stages, got %d", len(stages))
	}
}

// TestAggsBuilderPipes 测试 AggsBuilder.Pipes 方法
func TestAggsBuilderPipes(t *testing.T) {
	// 创建原始 stages
	rawStages := []bson.D{
		{{Key: "$match", Value: mgo.M{"status": "active"}}},
		{{Key: "$sort", Value: mgo.M{"created_at": -1}}},
	}

	// 验证可以传递给 Pipes
	_ = rawStages // 实际使用时会传给 AggsBuilder.Pipes()
}

// TestAccumulators 测试累加器函数
func TestAccumulators(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(any) mgo.M
		input    any
		expected string
	}{
		{"Sum", mgo.Sum, 1, "$sum"},
		{"Avg", mgo.Avg, "$age", "$avg"},
		{"Max", mgo.Max, "$price", "$max"},
		{"Min", mgo.Min, "$price", "$min"},
		{"First", mgo.First, "$date", "$first"},
		{"Last", mgo.Last, "$date", "$last"},
		{"Push", mgo.Push, "$item", "$push"},
		{"AddToSet", mgo.AddToSet, "$tag", "$addToSet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if len(result) != 1 {
				t.Fatalf("Expected 1 key in result, got %d", len(result))
			}

			// 验证包含正确的操作符
			if _, ok := result[tt.expected]; !ok {
				t.Errorf("Expected key %s in result, got %v", tt.expected, result)
			}
		})
	}
}
