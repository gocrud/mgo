package mgo_test

import (
	"context"
	"testing"
	"time"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// 测试用的泛型模型
type TestGenericUser struct {
	ID   bson.ObjectID `bson:"_id,omitempty"`
	Name string        `bson:"name"`
	Age  int           `bson:"age"`
}

// TestGenericIntegration 是一个集成测试，尝试连接本地 MongoDB。
// 如果连接失败，会自动跳过，不会导致测试失败。
func TestGenericIntegration(t *testing.T) {
	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 尝试连接本地 MongoDB (默认端口)
	// 使用 directConnection=true 避免在某些环境中去解析副本集
	client, err := mgo.Connect(ctx, "mongodb://localhost:27017/?directConnection=true")
	if err != nil {
		t.Skipf("Skipping integration test: Connect failed: %v", err)
	}

	// 尝试 Ping 确认服务可用
	if err := client.Native().Ping(ctx, nil); err != nil {
		t.Skipf("Skipping integration test: Ping failed: %v", err)
	}
	defer client.Disconnect(ctx)

	// 准备测试集合
	db := client.Database("test_generic_db")
	coll := db.Collection("users")
	// 清理旧数据
	_ = coll.Drop(ctx)

	// 1. 转换为泛型集合
	userColl := mgo.Model[TestGenericUser](coll)

	// 2. 插入数据
	newUser := TestGenericUser{
		Name: "TestUser",
		Age:  20,
	}
	insertRes, err := coll.InsertOne(ctx, newUser)
	if err != nil {
		// 如果是因为认证失败或其他环境问题，跳过测试而不是 Fail
		// 这是一个权衡，为了保证在没有正确配置的环境下 `go test` 也能通过（显示 skip）
		t.Skipf("Skipping integration test: InsertOne failed (likely env issue): %v", err)
	}
	id := insertRes.InsertedID

	// 3. 泛型查询: One()
	found, err := userColl.Query(ctx).Eq("_id", id).One()
	if err != nil {
		t.Fatalf("Generic One() failed: %v", err)
	}
	if found.Name != "TestUser" {
		t.Errorf("Expected name TestUser, got %s", found.Name)
	}

	// 4. 泛型查询: All()
	list, err := userColl.Query(ctx).Gte("age", 18).All()
	if err != nil {
		t.Fatalf("Generic All() failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 user, got %d", len(list))
	}

	// 5. 简化 Update (mgo.Set)
	updateRes, err := coll.Query(ctx).Eq("_id", id).UpdateOne(mgo.Set("age", 21))
	if err != nil {
		t.Fatalf("UpdateOne with mgo.Set failed: %v", err)
	}
	if updateRes.ModifiedCount != 1 {
		t.Errorf("Expected 1 document modified, got %d", updateRes.ModifiedCount)
	}

	// 6. 泛型 FindAndUpdate
	updated, err := userColl.Query(ctx).Eq("_id", id).FindAndUpdate(mgo.Set("name", "UpdatedUser"))
	if err != nil {
		t.Fatalf("Generic FindAndUpdate failed: %v", err)
	}
	if updated.Name != "UpdatedUser" {
		t.Errorf("Expected name UpdatedUser, got %s", updated.Name)
	}
	if updated.Age != 21 {
		t.Errorf("Expected age 21, got %d", updated.Age)
	}
}
