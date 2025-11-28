package mgo

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHelperSet(t *testing.T) {
	key := "age"
	val := 25
	m := Set(key, val)

	if len(m) != 1 {
		t.Fatalf("Expected map length 1, got %d", len(m))
	}

	setVal, ok := m["$set"]
	if !ok {
		t.Fatal("Expected $set key in map")
	}

	setMap, ok := setVal.(bson.M)
	if !ok {
		t.Fatal("Expected $set value to be bson.M")
	}

	if setMap[key] != val {
		t.Errorf("Expected %v, got %v", val, setMap[key])
	}
}

