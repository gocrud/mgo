package mgo

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestBuildUpdate 测试 QueryBuilder.buildUpdate 方法
// 由于该方法是私有的，且我们在 package mgo 中，所以可以直接访问
func TestBuildUpdate(t *testing.T) {
	qb := &QueryBuilder{}

	// Case 1: *UpdateBuilder
	ub := Update().Set("name", "test")
	res1 := qb.buildUpdate(ub)

	// 验证结果是否为 bson.M 且包含 $set
	if m, ok := res1.(bson.M); ok {
		if _, exists := m["$set"]; !exists {
			t.Error("Expected *UpdateBuilder to be built into map with $set")
		}
	} else if d, ok := res1.(bson.D); ok {
		// UpdateBuilder.Build() 有可能返回 bson.D (虽然目前实现是返回 map 或 D)
		// 检查 UpdateBuilder.Build() 实现：
		// 它是返回 map 还是 D？mgo 代码里 UpdateBuilder.Build() 返回的是 map[string]any (即 M) 或者 []bson.E
		// 让我们假设它返回的是 map 或 D
		found := false
		for _, e := range d {
			if e.Key == "$set" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected *UpdateBuilder to be built into D with $set")
		}
	} else {
		// 甚至可能是一个 []bson.E
		t.Errorf("Unexpected type from UpdateBuilder: %T", res1)
	}

	// Case 2: map (mgo.Set 返回的就是 map)
	mapUpdate := Set("age", 18)
	res2 := qb.buildUpdate(mapUpdate)
	// 无法直接比较 map，比较其内容或指针（如果是同一个 map 对象，指针应该相同吗？Go map 是引用类型）
	// 但在 Go 中，res2 == mapUpdate 会报编译错误
	// 我们只需要检查 res2 是否是我们传入的 map
	if m, ok := res2.(bson.M); !ok {
		t.Errorf("Expected bson.M, got %T", res2)
	} else {
		// 检查内容
		val := m["$set"].(bson.M)["age"]
		if val != 18 {
			t.Errorf("Expected 18, got %v", val)
		}
	}

	// Case 3: struct (直接替换文档)
	type UserUpdate struct {
		Name string `bson:"name"`
	}
	structUpdate := UserUpdate{Name: "new"}
	res3 := qb.buildUpdate(structUpdate)
	if res3 != structUpdate {
		t.Error("Expected struct to be passed through directly")
	}
}
