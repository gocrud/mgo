package mgo_test

import (
	"context"
	"testing"

	"github.com/gocrud/mgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	ID             string   `bson:"_id,omitempty"`
	Name           string   `bson:"name"`
	Age            int      `bson:"age"`
	Role           string   `bson:"role"`
	LoginCount     int      `bson:"login_count"`
	Tags           []string `bson:"tags"`
	mgo.TimeFields `bson:",inline"`
	mgo.SoftDelete `bson:",inline"`
}

type Order struct {
	ID             string `bson:"_id,omitempty"`
	UserID         string `bson:"user_id"`
	Amount         int    `bson:"amount"`
	Status         string `bson:"status"`
	mgo.TimeFields `bson:",inline"`
}

type OrderWithUser struct {
	Order `bson:",inline"`
	User  *User `bson:"user_info"`
}

const mongoURI = "mongodb://example:example@localhost:27017/?directConnection=true"

func TestMgoFullCoverage(t *testing.T) {
	// 1. Connect
	cli, err := mgo.Connect(mongoURI)
	require.NoError(t, err)
	defer cli.Disconnect(context.Background())

	db := cli.Database("test_mgo_full_db")

	// Clean up
	_ = db.Collection("users").Drop(context.Background())
	_ = db.Collection("orders").Drop(context.Background())

	users := db.Collection("users").AutoTime().SoftDelete()
	orders := db.Collection("orders").AutoTime()

	// 2. Insert & Basic Types
	t.Run("Insert & Types", func(t *testing.T) {
		// One
		_, err := users.Insert().Doc(&User{ID: "user_alice", Name: "Alice", Age: 20, Role: "admin", Tags: []string{"a"}}).One()
		assert.NoError(t, err)

		// Many
		_, err = users.Insert().Docs(
			&User{ID: "user_bob", Name: "Bob", Age: 22, Role: "user", Tags: []string{"b"}},
			&User{ID: "user_charlie", Name: "Charlie", Age: 25, Role: "user", Tags: []string{"c"}},
			&User{ID: "user_david", Name: "David", Age: 30, Role: "admin", Tags: []string{"d"}},
			&User{ID: "user_eve", Name: "Eve", Age: 18, Role: "user", Tags: []string{"e"}},
		).Many()
		assert.NoError(t, err)
	})

	// 3. Find Advanced
	t.Run("Find Advanced", func(t *testing.T) {
		// Where, Sort, Limit, Select
		var list []User
		err := users.Find().
			Where(mgo.Gt("age", 20)).
			SortDesc("age").
			Limit(2).
			Select("name", "age"). // Projection
			All(&list)

		assert.NoError(t, err)
		assert.Len(t, list, 2)
		if len(list) == 2 {
			assert.Equal(t, "David", list[0].Name)
			assert.Equal(t, "Charlie", list[1].Name)
			assert.Empty(t, list[0].Role) // Should be empty due to Select
		}
	})

	// 4. Cursor Pagination
	t.Run("Cursor Pagination", func(t *testing.T) {
		// Sort by Age (Unique: 18, 20, 22, 25, 30)
		// Get first page
		var page1 []User
		query := users.Find().SortAsc("age").Limit(2)
		err := query.All(&page1)
		require.NoError(t, err)
		require.Len(t, page1, 2)
		assert.Equal(t, "Eve", page1[0].Name)   // 18
		assert.Equal(t, "Alice", page1[1].Name) // 20

		lastUser := page1[1]

		// Get second page using Seek
		var page2 []User
		err = users.Find().
			SortAsc("age").
			Limit(2).
			Seek(lastUser).
			All(&page2)

		require.NoError(t, err)
		require.Len(t, page2, 2)
		assert.Equal(t, "Bob", page2[0].Name)     // 22
		assert.Equal(t, "Charlie", page2[1].Name) // 25
	})

	// 5. Update Advanced
	t.Run("Update Advanced", func(t *testing.T) {
		// Inc, Push, Pull
		_, err := users.Update().
			Where(mgo.Eq("name", "Alice")).
			Inc("login_count", 1).
			Push("tags", "new_tag").
			One()
		assert.NoError(t, err)

		var u User
		err = users.Find().Where(mgo.Eq("name", "Alice")).One(&u)
		assert.NoError(t, err)
		assert.Equal(t, 1, u.LoginCount)
		assert.Contains(t, u.Tags, "new_tag")

		// Pull
		_, err = users.Update().
			Where(mgo.Eq("name", "Alice")).
			Pull("tags", "a").
			One()
		assert.NoError(t, err)

		err = users.Find().Where(mgo.Eq("name", "Alice")).One(&u)
		assert.NotContains(t, u.Tags, "a")
	})

	// 6. Soft Delete & Restore
	t.Run("Soft Delete", func(t *testing.T) {
		// Soft Delete
		_, err := users.Delete().Where(mgo.Eq("name", "Bob")).Many()
		assert.NoError(t, err)

		// Find should not return Bob
		var list []User
		err = users.Find().Where(mgo.Eq("name", "Bob")).All(&list)
		assert.NoError(t, err)
		assert.Len(t, list, 0)

		// Unscoped should return Bob
		err = users.Find().Unscoped().Where(mgo.Eq("name", "Bob")).All(&list)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		if len(list) > 0 {
			assert.NotNil(t, list[0].DeletedAt)
		}

		// Restore
		_, err = users.Update().Where(mgo.Eq("name", "Bob")).Restore().One()
		assert.NoError(t, err)

		// Find should return Bob again
		err = users.Find().Where(mgo.Eq("name", "Bob")).All(&list)
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		if len(list) > 0 {
			assert.Nil(t, list[0].DeletedAt)
		}

		// Hard Delete
		_, err = users.Delete().Where(mgo.Eq("name", "Bob")).Hard().Many()
		assert.NoError(t, err)

		// Unscoped should NOT return Bob
		err = users.Find().Unscoped().Where(mgo.Eq("name", "Bob")).All(&list)
		assert.NoError(t, err)
		assert.Len(t, list, 0)
	})

	// 7. Aggregate & Join
	t.Run("Aggregate Join", func(t *testing.T) {
		// Setup Orders
		// Alice ID is "user_alice"
		_, err = orders.Insert().Docs(
			&Order{UserID: "user_alice", Amount: 100, Status: "paid"},
			&Order{UserID: "user_alice", Amount: 200, Status: "pending"},
		).Many()
		require.NoError(t, err)

		var results []OrderWithUser
		err = orders.Aggregate().
			Match(mgo.Eq("status", "paid")).
			Join("users", "user_id", "user_info").
			All(&results)

		require.NoError(t, err)
		require.Len(t, results, 1)
		if len(results) > 0 {
			assert.Equal(t, 100, results[0].Amount)
			assert.NotNil(t, results[0].User)
			assert.Equal(t, "Alice", results[0].User.Name)
		}
	})

	// 8. Aggregate Pagination
	t.Run("Aggregate Pagination", func(t *testing.T) {
		var list []Order
		res, err := orders.Aggregate().
			SortDesc("amount").
			Paginate(1, 1).
			All(&list)

		require.NoError(t, err)
		assert.Equal(t, int64(2), res.Total)
		assert.Len(t, list, 1)
		if len(list) > 0 {
			assert.Equal(t, 200, list[0].Amount) // SortDesc
		}
	})

	// 9. Transactions
	t.Run("Transactions", func(t *testing.T) {
		// Success Case
		err := cli.Tx(context.Background(), func(ctx context.Context) error {
			_, err := users.Insert().Ctx(ctx).Doc(&User{ID: "tx_user_1", Name: "TxUser1"}).One()
			if err != nil {
				return err
			}
			_, err = users.Insert().Ctx(ctx).Doc(&User{ID: "tx_user_2", Name: "TxUser2"}).One()
			return err
		})
		assert.NoError(t, err)

		count, _ := users.Find().Where(mgo.In("name", []string{"TxUser1", "TxUser2"})).Count()
		assert.Equal(t, int64(2), count)

		// Rollback Case
		err = cli.Tx(context.Background(), func(ctx context.Context) error {
			_, err := users.Insert().Ctx(ctx).Doc(&User{ID: "tx_user_3", Name: "TxUser3"}).One()
			if err != nil {
				return err
			}
			return assert.AnError // Force rollback
		})
		assert.Error(t, err)

		count, _ = users.Find().Where(mgo.Eq("name", "TxUser3")).Count()
		assert.Equal(t, int64(0), count)
	})
}
