package mgo_test

import (
	"context"
	"testing"

	"github.com/gocrud/mgo"
)

func TestQueryWhen(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	defer db.Close(ctx)

	users := mgo.Model[TestUser](db, "query_when_test")
	defer users.Truncate()

	// 准备数据
	user1 := &TestUser{Name: "User1", Age: 20, Status: "active"}
	user2 := &TestUser{Name: "User2", Age: 30, Status: "inactive"}
	user3 := &TestUser{Name: "User3", Age: 40, Status: "active"}

	users.WithContext(ctx).Insert(user1)
	users.WithContext(ctx).Insert(user2)
	users.WithContext(ctx).Insert(user3)

	t.Run("When_True", func(t *testing.T) {
		// When(true) 应该执行闭包内的逻辑
		count, err := users.Find().WithContext(ctx).
			When(true, func(q *mgo.Query[TestUser]) {
				q.Eq("name", "User1")
			}).
			Count()

		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 user, got %d", count)
		}
	})

	t.Run("When_False", func(t *testing.T) {
		// When(false) 应该跳过闭包内的逻辑
		// 所以应该查出所有 3 个用户
		count, err := users.Find().WithContext(ctx).
			When(false, func(q *mgo.Query[TestUser]) {
				q.Eq("name", "User1")
			}).
			Count()

		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected 3 users, got %d", count)
		}
	})

	t.Run("When_Chained", func(t *testing.T) {
		// 组合测试
		// 1. When(true).Eq("status", "active") -> 执行，剩 User1, User3
		// 2. When(false).Gt("age", 30) -> 跳过 Gt(age > 30)，仍然是 User1, User3
		// 3. When(true).Lt("age", 50) -> 执行，User1(20), User3(40) 都满足
		count, err := users.Find().WithContext(ctx).
			When(true, func(q *mgo.Query[TestUser]) {
				q.Eq("status", "active")
			}).
			When(false, func(q *mgo.Query[TestUser]) {
				q.Gt("age", 30)
			}).
			When(true, func(q *mgo.Query[TestUser]) {
				q.Lt("age", 50)
			}).
			Count()

		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 users, got %d", count)
		}
	})

	t.Run("WhenFunc_True", func(t *testing.T) {
		// WhenFunc(true) 应该执行闭包内的逻辑
		count, err := users.Find().WithContext(ctx).
			When(true, func(q *mgo.Query[TestUser]) {
				q.Eq("status", "active").Gt("age", 30)
			}).
			Count()

		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		// active & age > 30 -> User3
		if count != 1 {
			t.Errorf("Expected 1 user, got %d", count)
		}
	})

	t.Run("WhenFunc_False", func(t *testing.T) {
		// WhenFunc(false) 应该跳过闭包内的逻辑
		count, err := users.Find().WithContext(ctx).
			When(false, func(q *mgo.Query[TestUser]) {
				q.Eq("status", "impossible")
			}).
			Count()

		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		// 没有任何条件 -> 3
		if count != 3 {
			t.Errorf("Expected 3 users, got %d", count)
		}
	})

	t.Run("When_Update", func(t *testing.T) {
		// 测试 Update 中的 When
		// 将 User1 的名字改为 "User1_Updated"，但跳过 Age 的更新
		err := users.Find().WithContext(ctx).
			Eq("name", "User1").
			When(true, func(q *mgo.Query[TestUser]) {
				q.Set("name", "User1_Updated")
			}).
			When(false, func(q *mgo.Query[TestUser]) {
				q.Set("age", 100)
			}).
			Update()

		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// 验证
		// 重新查询
		updatedUser, err := users.Find().WithContext(ctx).Eq("name", "User1_Updated").One()
		if err != nil {
			t.Fatalf("Find updated user failed: %v", err)
		}

		if updatedUser.Age == 100 {
			t.Error("Age should not be updated")
		}
		if updatedUser.Age != 20 {
			t.Errorf("Age should be 20, got %d", updatedUser.Age)
		}
	})
}
