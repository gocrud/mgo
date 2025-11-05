package mgo

import (
	"testing"
)

func TestDoc(t *testing.T) {
	doc := NewDoc().
		Set("name", "张三").
		Set("age", 25).
		Set("email", "zhangsan@example.com")

	result := doc.BuildM()

	if result["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", result["name"])
	}
	if result["age"] != 25 {
		t.Errorf("Expected age=25, got %v", result["age"])
	}
}

func TestDocSetIf(t *testing.T) {
	name := "张三"
	age := 0

	doc := NewDoc().
		SetIf(name != "", "name", name).
		SetIf(age > 0, "age", age)

	result := doc.BuildM()

	if result["name"] != "张三" {
		t.Errorf("Expected name=张三, got %v", result["name"])
	}
	if _, exists := result["age"]; exists {
		t.Error("age should not be set when age <= 0")
	}
}

func TestDocNested(t *testing.T) {
	doc := NewDoc().
		Set("user", NewDoc().
			Set("name", "张三").
			Set("email", "zhangsan@example.com").
			Build()).
		Set("tags", []string{"tag1", "tag2"})

	result := doc.Build()

	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
}

func TestProjectionInclude(t *testing.T) {
	proj := NewProjection().
		Include("name", "age", "email")

	result := proj.BuildM()

	if result["name"] != 1 {
		t.Errorf("Expected name=1, got %v", result["name"])
	}
	if result["age"] != 1 {
		t.Errorf("Expected age=1, got %v", result["age"])
	}
}

func TestProjectionExclude(t *testing.T) {
	proj := NewProjection().
		Exclude("password", "secret")

	result := proj.BuildM()

	if result["password"] != 0 {
		t.Errorf("Expected password=0, got %v", result["password"])
	}
	if result["secret"] != 0 {
		t.Errorf("Expected secret=0, got %v", result["secret"])
	}
}

func TestProjectionExcludeID(t *testing.T) {
	proj := NewProjection().
		Include("name", "age").
		ExcludeID()

	result := proj.BuildM()

	if result["_id"] != 0 {
		t.Errorf("Expected _id=0, got %v", result["_id"])
	}
}

func TestProjectionSlice(t *testing.T) {
	proj := NewProjection().
		Include("name").
		Slice("items", 10)

	result := proj.Build()

	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
}

func TestProjectionSliceWithSkip(t *testing.T) {
	proj := NewProjection().
		Include("name").
		SliceWithSkip("comments", 10, 5)

	result := proj.Build()

	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
}

func TestProjectionElemMatch(t *testing.T) {
	proj := NewProjection().
		Include("name").
		ElemMatch("comments", Filter().Gt("rating", 4))

	result := proj.Build()

	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
}

func TestProjectionSetExpr(t *testing.T) {
	proj := NewProjection().
		Include("firstName", "lastName").
		SetExpr("fullName", Exp.Concat(F("firstName"), " ", F("lastName")))

	result := proj.Build()

	if len(result) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(result))
	}
}

func TestProjectionMeta(t *testing.T) {
	proj := NewProjection().
		Include("title", "content").
		Meta("score", "textScore")

	result := proj.Build()

	if len(result) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(result))
	}
}

func TestSortAsc(t *testing.T) {
	sort := NewSort().Asc("name", "age")

	result := sort.BuildM()

	if result["name"] != 1 {
		t.Errorf("Expected name=1, got %v", result["name"])
	}
	if result["age"] != 1 {
		t.Errorf("Expected age=1, got %v", result["age"])
	}
}

func TestSortDesc(t *testing.T) {
	sort := NewSort().Desc("created_at", "priority")

	result := sort.BuildM()

	if result["created_at"] != -1 {
		t.Errorf("Expected created_at=-1, got %v", result["created_at"])
	}
	if result["priority"] != -1 {
		t.Errorf("Expected priority=-1, got %v", result["priority"])
	}
}

func TestSortMixed(t *testing.T) {
	sort := NewSort().
		Desc("priority").
		Asc("name").
		Desc("created_at")

	result := sort.BuildM()

	if result["priority"] != -1 {
		t.Errorf("Expected priority=-1, got %v", result["priority"])
	}
	if result["name"] != 1 {
		t.Errorf("Expected name=1, got %v", result["name"])
	}
}

func TestSortTextScore(t *testing.T) {
	sort := NewSort().TextScore("score")

	result := sort.Build()

	if len(result) != 1 {
		t.Errorf("Expected 1 field, got %d", len(result))
	}
}

func TestIndexAsc(t *testing.T) {
	idx := NewIndex().Asc("name", "age")

	result := idx.BuildM()

	if result["name"] != 1 {
		t.Errorf("Expected name=1, got %v", result["name"])
	}
	if result["age"] != 1 {
		t.Errorf("Expected age=1, got %v", result["age"])
	}
}

func TestIndexDesc(t *testing.T) {
	idx := NewIndex().Desc("created_at")

	result := idx.BuildM()

	if result["created_at"] != -1 {
		t.Errorf("Expected created_at=-1, got %v", result["created_at"])
	}
}

func TestIndexCompound(t *testing.T) {
	idx := NewIndex().
		Asc("category").
		Desc("created_at")

	result := idx.BuildM()

	if result["category"] != 1 {
		t.Errorf("Expected category=1, got %v", result["category"])
	}
	if result["created_at"] != -1 {
		t.Errorf("Expected created_at=-1, got %v", result["created_at"])
	}
}

func TestIndexText(t *testing.T) {
	idx := NewIndex().Text("title", "content")

	result := idx.BuildM()

	if result["title"] != "text" {
		t.Errorf("Expected title=text, got %v", result["title"])
	}
	if result["content"] != "text" {
		t.Errorf("Expected content=text, got %v", result["content"])
	}
}

func TestIndexGeo2DSphere(t *testing.T) {
	idx := NewIndex().Geo2DSphere("location")

	result := idx.BuildM()

	if result["location"] != "2dsphere" {
		t.Errorf("Expected location=2dsphere, got %v", result["location"])
	}
}

func TestIndexHashed(t *testing.T) {
	idx := NewIndex().Hashed("user_id")

	result := idx.BuildM()

	if result["user_id"] != "hashed" {
		t.Errorf("Expected user_id=hashed, got %v", result["user_id"])
	}
}
