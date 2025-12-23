package mgo_test

import (
	"context"
	"testing"

	"github.com/gocrud/mgo"
)

type UpdateTestUser struct {
	ID     mgo.ObjectID `bson:"_id,omitempty"`
	Name   string       `bson:"name"`
	Age    int          `bson:"age"`
	Scores []int        `bson:"scores"`
	Tags   []string     `bson:"tags"`
}

func (UpdateTestUser) CollName() string {
	return "update_methods_test"
}

func TestUpdateMethods(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	defer db.Close(ctx)

	users := mgo.Model[UpdateTestUser](db)
	defer users.Truncate()

	user := &UpdateTestUser{
		Name:   "Test",
		Age:    20,
		Scores: []int{10, 20},
		Tags:   []string{"a", "b"},
	}
	id, _ := users.WithContext(ctx).Insert(user)

	t.Run("NumericOperators", func(t *testing.T) {
		// Mul
		users.Find().WithContext(ctx).ID(id).Mul("age", 2).Update()
		u, _ := users.WithContext(ctx).FindByID(id)
		if u.Age != 40 {
			t.Errorf("Mul: expected 40, got %d", u.Age)
		}

		// SetMin
		users.Find().WithContext(ctx).ID(id).SetMin("age", 30).Update() // Should change to 30
		u, _ = users.WithContext(ctx).FindByID(id)
		if u.Age != 30 {
			t.Errorf("SetMin: expected 30, got %d", u.Age)
		}
		users.Find().WithContext(ctx).ID(id).SetMin("age", 50).Update() // Should not change
		u, _ = users.WithContext(ctx).FindByID(id)
		if u.Age != 30 {
			t.Errorf("SetMin: expected 30, got %d", u.Age)
		}

		// SetMax
		users.Find().WithContext(ctx).ID(id).SetMax("age", 50).Update() // Should change to 50
		u, _ = users.WithContext(ctx).FindByID(id)
		if u.Age != 50 {
			t.Errorf("SetMax: expected 50, got %d", u.Age)
		}
	})

	t.Run("FieldOperators", func(t *testing.T) {
		// Rename
		users.Find().WithContext(ctx).ID(id).Rename("age", "years").Update()
		// Check raw document or use map to verify
		// Since struct doesn't have "years", we can't decode it easily to struct unless we change struct.
		// But we can check if "age" is zero (missing)
		u, _ := users.WithContext(ctx).FindByID(id)
		if u.Age != 0 {
			t.Errorf("Rename: expected age to be missing (0), got %d", u.Age)
		}

		// Rename back
		users.Find().WithContext(ctx).ID(id).Rename("years", "age").Update()
		u, _ = users.WithContext(ctx).FindByID(id)
		if u.Age != 50 {
			t.Errorf("Rename back: expected 50, got %d", u.Age)
		}

		// Unset
		users.Find().WithContext(ctx).ID(id).Unset("age").Update()
		u, _ = users.WithContext(ctx).FindByID(id)
		if u.Age != 0 {
			t.Errorf("Unset: expected age to be missing (0), got %d", u.Age)
		}

		// Restore age
		users.Find().WithContext(ctx).ID(id).Set("age", 20).Update()
	})

	t.Run("ArrayOperators", func(t *testing.T) {
		// Push
		users.Find().WithContext(ctx).ID(id).Push("scores", 30).Update()
		u, _ := users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 3 || u.Scores[2] != 30 {
			t.Errorf("Push: expected [10 20 30], got %v", u.Scores)
		}

		// PushAll
		users.Find().WithContext(ctx).ID(id).PushAll("scores", []interface{}{40, 50}).Update()
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 5 {
			t.Errorf("PushAll: expected 5 items, got %d", len(u.Scores))
		}

		// AddToSet
		users.Find().WithContext(ctx).ID(id).AddToSet("scores", 30).Update() // Should not add
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 5 {
			t.Errorf("AddToSet: expected 5 items, got %d", len(u.Scores))
		}
		users.Find().WithContext(ctx).ID(id).AddToSet("scores", 60).Update() // Should add
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 6 {
			t.Errorf("AddToSet: expected 6 items, got %d", len(u.Scores))
		}

		// Pop
		users.Find().WithContext(ctx).ID(id).Pop("scores", 1).Update() // Remove last
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 5 || u.Scores[len(u.Scores)-1] == 60 {
			t.Errorf("Pop: expected last item removed")
		}

		// Pull
		users.Find().WithContext(ctx).ID(id).Pull("scores", 10).Update()
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 4 {
			t.Errorf("Pull: expected 4 items, got %d", len(u.Scores))
		}

		// PullAll
		users.Find().WithContext(ctx).ID(id).PullAll("scores", []interface{}{20, 30}).Update()
		u, _ = users.WithContext(ctx).FindByID(id)
		if len(u.Scores) != 2 {
			t.Errorf("PullAll: expected 2 items, got %d", len(u.Scores))
		}
	})
}
