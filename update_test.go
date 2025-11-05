package mgo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestUpdateSet(t *testing.T) {
	update := Update().
		Set("name", "张三").
		Set("age", 25).
		Set("user.email", "zhangsan@example.com")

	result := update.BuildM()

	if result["$set"] == nil {
		t.Fatal("$set operator not found")
	}

	setOp := result["$set"].(bson.M)
	if setOp["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", setOp["name"])
	}
	if setOp["age"] != 25 {
		t.Errorf("Expected age=25, got %v", setOp["age"])
	}
}

func TestUpdateSetM(t *testing.T) {
	update := Update().SetM(M{
		"name":   "张三",
		"age":    25,
		"status": "active",
	})

	result := update.BuildM()
	setOp := result["$set"].(bson.M)

	if len(setOp) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(setOp))
	}
}

func TestUpdateInc(t *testing.T) {
	update := Update().
		Inc("visits", 1).
		Inc("level", 2).
		Inc("score", -10)

	result := update.BuildM()
	incOp := result["$inc"].(bson.M)

	if incOp["visits"] != 1 {
		t.Errorf("Expected visits=1, got %v", incOp["visits"])
	}
	if incOp["score"] != -10 {
		t.Errorf("Expected score=-10, got %v", incOp["score"])
	}
}

func TestUpdateUnset(t *testing.T) {
	update := Update().Unset("temp_field", "old_field")

	result := update.BuildM()
	unsetOp := result["$unset"].(bson.M)

	if len(unsetOp) != 2 {
		t.Errorf("Expected 2 fields to unset, got %d", len(unsetOp))
	}
}

func TestUpdateMul(t *testing.T) {
	update := Update().
		Mul("price", 1.1).
		Mul("discount", 0.8)

	result := update.BuildM()
	mulOp := result["$mul"].(bson.M)

	if mulOp["price"] != 1.1 {
		t.Errorf("Expected price=1.1, got %v", mulOp["price"])
	}
}

func TestUpdateMinMax(t *testing.T) {
	update := Update().
		Min("price", 100).
		Max("stock", 1000)

	result := update.BuildM()

	minOp := result["$min"].(bson.M)
	maxOp := result["$max"].(bson.M)

	if minOp["price"] != 100 {
		t.Errorf("Expected min price=100, got %v", minOp["price"])
	}
	if maxOp["stock"] != 1000 {
		t.Errorf("Expected max stock=1000, got %v", maxOp["stock"])
	}
}

func TestUpdateRename(t *testing.T) {
	update := Update().Rename("old_name", "new_name")

	result := update.BuildM()
	renameOp := result["$rename"].(bson.M)

	if renameOp["old_name"] != "new_name" {
		t.Errorf("Expected rename old_name to new_name, got %v", renameOp["old_name"])
	}
}

func TestUpdateCurrentDate(t *testing.T) {
	update := Update().
		CurrentDate("updated_at", false).
		CurrentDate("last_seen", true)

	result := update.Build()

	// 检查 $currentDate 操作符存在
	found := false
	for _, elem := range result {
		if elem.Key == "$currentDate" {
			found = true
			break
		}
	}

	if !found {
		t.Error("$currentDate operator not found")
	}
}

func TestUpdateSetOnInsert(t *testing.T) {
	now := time.Now()
	update := Update().
		SetOnInsert("created_at", now).
		SetOnInsert("version", 1)

	result := update.BuildM()
	setOnInsertOp := result["$setOnInsert"].(bson.M)

	if setOnInsertOp["version"] != 1 {
		t.Errorf("Expected version=1, got %v", setOnInsertOp["version"])
	}
}

func TestUpdatePush(t *testing.T) {
	update := Update().
		Push("tags", "new-tag").
		Push("items", M{"id": 1, "name": "item1"})

	result := update.BuildM()
	pushOp := result["$push"].(bson.M)

	if pushOp["tags"] != "new-tag" {
		t.Errorf("Expected tag=new-tag, got %v", pushOp["tags"])
	}
}

func TestUpdatePushEach(t *testing.T) {
	update := Update().PushEach("tags", "tag1", "tag2", "tag3")

	result := update.Build()

	// 验证 $push 操作符存在
	found := false
	for _, elem := range result {
		if elem.Key == "$push" {
			found = true
			break
		}
	}

	if !found {
		t.Error("$push operator not found")
	}
}

func TestUpdatePull(t *testing.T) {
	update := Update().
		Pull("tags", "old-tag").
		Pull("items", M{"status": "deleted"})

	result := update.BuildM()
	pullOp := result["$pull"].(bson.M)

	if pullOp["tags"] != "old-tag" {
		t.Errorf("Expected pull old-tag, got %v", pullOp["tags"])
	}
}

func TestUpdatePullAll(t *testing.T) {
	update := Update().PullAll("tags", "tag1", "tag2", "tag3")

	result := update.BuildM()
	pullAllOp := result["$pullAll"].(bson.M)

	tags := pullAllOp["tags"].([]any)
	if len(tags) != 3 {
		t.Errorf("Expected 3 tags to pull, got %d", len(tags))
	}
}

func TestUpdatePullFilter(t *testing.T) {
	update := Update().PullFilter("items", Filter().Gt("price", 1000))

	result := update.Build()

	// 验证 $pull 操作符存在
	found := false
	for _, elem := range result {
		if elem.Key == "$pull" {
			found = true
			break
		}
	}

	if !found {
		t.Error("$pull operator not found")
	}
}

func TestUpdatePop(t *testing.T) {
	update := Update().
		Pop("items", 1). // 删除最后一个
		Pop("queue", -1) // 删除第一个

	result := update.BuildM()
	popOp := result["$pop"].(bson.M)

	if popOp["items"] != 1 {
		t.Errorf("Expected pop items=1, got %v", popOp["items"])
	}
	if popOp["queue"] != -1 {
		t.Errorf("Expected pop queue=-1, got %v", popOp["queue"])
	}
}

func TestUpdateAddToSet(t *testing.T) {
	update := Update().
		AddToSet("tags", "unique-tag").
		AddToSet("user_ids", 12345)

	result := update.BuildM()
	addToSetOp := result["$addToSet"].(bson.M)

	if addToSetOp["tags"] != "unique-tag" {
		t.Errorf("Expected tag=unique-tag, got %v", addToSetOp["tags"])
	}
}

func TestUpdateAddToSetEach(t *testing.T) {
	update := Update().AddToSetEach("tags", "tag1", "tag2", "tag3")

	result := update.Build()

	// 验证 $addToSet 操作符存在
	found := false
	for _, elem := range result {
		if elem.Key == "$addToSet" {
			found = true
			break
		}
	}

	if !found {
		t.Error("$addToSet operator not found")
	}
}

func TestUpdateBit(t *testing.T) {
	update := Update().
		Bit("flags", "or", 4).
		Bit("perms", "and", 3)

	result := update.Build()

	// 验证 $bit 操作符存在
	found := false
	for _, elem := range result {
		if elem.Key == "$bit" {
			found = true
			break
		}
	}

	if !found {
		t.Error("$bit operator not found")
	}
}

func TestUpdateComplex(t *testing.T) {
	update := Update().
		Set("name", "张三").
		Set("status", "active").
		Inc("level", 1).
		Inc("exp", 100).
		Push("tags", "vip").
		CurrentDate("updated_at", false)

	result := update.BuildM()

	// 验证多个操作符都存在
	if result["$set"] == nil {
		t.Error("$set operator missing")
	}
	if result["$inc"] == nil {
		t.Error("$inc operator missing")
	}
	if result["$push"] == nil {
		t.Error("$push operator missing")
	}
	if result["$currentDate"] == nil {
		t.Error("$currentDate operator missing")
	}
}

func TestUpdateClone(t *testing.T) {
	update1 := Update().Set("name", "张三").Inc("age", 1)
	update2 := update1.Clone().Set("email", "test@example.com")

	result1 := update1.BuildM()
	result2 := update2.BuildM()

	// update1 应该只有 name 和 age
	set1 := result1["$set"].(bson.M)
	if len(set1) != 1 {
		t.Errorf("update1 $set should have 1 field, got %d", len(set1))
	}

	// update2 应该有 name, age 和 email
	set2 := result2["$set"].(bson.M)
	if len(set2) != 2 {
		t.Errorf("update2 $set should have 2 fields, got %d", len(set2))
	}
}
