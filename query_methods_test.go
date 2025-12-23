package mgo_test

import (
	"context"
	"testing"

	"github.com/gocrud/mgo"
)

func TestQueryMethods(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	defer db.Close(ctx)

	users := mgo.Model[TestUser](db, "query_methods_test")
	defer users.Truncate()

	// 准备数据
	user1 := &TestUser{Name: "User1", Age: 20, Status: "active"}
	user2 := &TestUser{Name: "User2", Age: 30, Status: "inactive"}
	user3 := &TestUser{Name: "User3", Age: 40, Status: "active"}

	id1, _ := users.WithContext(ctx).Insert(user1)
	id2, _ := users.WithContext(ctx).Insert(user2)
	_, _ = users.WithContext(ctx).Insert(user3)

	t.Run("ComparisonOperators", func(t *testing.T) {
		// Lt
		count, _ := users.Find().WithContext(ctx).Lt("age", 30).Count()
		if count != 1 {
			t.Errorf("Lt: expected 1, got %d", count)
		}

		// Lte
		count, _ = users.Find().WithContext(ctx).Lte("age", 30).Count()
		if count != 2 {
			t.Errorf("Lte: expected 2, got %d", count)
		}

		// Gt
		count, _ = users.Find().WithContext(ctx).Gt("age", 30).Count()
		if count != 1 {
			t.Errorf("Gt: expected 1, got %d", count)
		}

		// Gte
		count, _ = users.Find().WithContext(ctx).Gte("age", 30).Count()
		if count != 2 {
			t.Errorf("Gte: expected 2, got %d", count)
		}

		// Ne
		count, _ = users.Find().WithContext(ctx).Ne("age", 30).Count()
		if count != 2 {
			t.Errorf("Ne: expected 2, got %d", count)
		}
	})

	t.Run("InAndNin", func(t *testing.T) {
		// In
		count, _ := users.Find().WithContext(ctx).In("age", 20, 40).Count()
		if count != 2 {
			t.Errorf("In: expected 2, got %d", count)
		}

		// Nin
		count, _ = users.Find().WithContext(ctx).Nin("age", 20, 40).Count()
		if count != 1 {
			t.Errorf("Nin: expected 1, got %d", count)
		}
	})

	t.Run("Regex", func(t *testing.T) {
		count, _ := users.Find().WithContext(ctx).Regex("name", "^User").Count()
		if count != 3 {
			t.Errorf("Regex: expected 3, got %d", count)
		}

		count, _ = users.Find().WithContext(ctx).Regex("name", "1$").Count()
		if count != 1 {
			t.Errorf("Regex: expected 1, got %d", count)
		}
	})

	t.Run("IDs", func(t *testing.T) {
		count, _ := users.Find().WithContext(ctx).IDs(id1, id2).Count()
		if count != 2 {
			t.Errorf("IDs: expected 2, got %d", count)
		}
	})

	t.Run("SelectAndOmit", func(t *testing.T) {
		// Select
		u, _ := users.Find().WithContext(ctx).ID(id1).Select("name").One()
		if u.Name == "" || u.Age != 0 {
			t.Errorf("Select: expected only name, got %+v", u)
		}

		// Omit
		u, _ = users.Find().WithContext(ctx).ID(id1).Omit("name").One()
		if u.Name != "" || u.Age == 0 {
			t.Errorf("Omit: expected no name, got %+v", u)
		}
	})
}
