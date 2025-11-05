package mgo

import (
	"testing"
)

func TestFilterBasic(t *testing.T) {
	// 测试基础条件
	filter := Filter().
		Eq("status", "active").
		Gte("age", 18).
		Lt("age", 65)

	result := filter.Build()
	// age 字段的两个条件会合并，所以是 2 个条件：status 和 age
	if len(result) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(result))
	}
}

func TestFilterIn(t *testing.T) {
	// 测试 In 操作
	filter := Filter().
		In("city", "北京", "上海", "深圳")

	result := filter.Build()
	if len(result) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(result))
	}
}

func TestFilterOr(t *testing.T) {
	// 测试 Or 逻辑
	filter := Filter().
		Eq("status", "active").
		Or(
			Filter().Eq("vip", true),
			Filter().Gte("level", 5),
		)

	result := filter.Build()
	if len(result) != 2 {
		t.Errorf("Expected 2 conditions, got %d", len(result))
	}
}

func TestFilterBetween(t *testing.T) {
	// 测试 Between
	filter := Filter().Between("age", 18, 65)

	result := filter.Build()
	if len(result) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(result))
	}
}

func TestFilterAll(t *testing.T) {
	// 测试 All
	filter := Filter().All("tags", "active", "verified")

	result := filter.Build()
	if len(result) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(result))
	}
}

func TestExpressionBuilder(t *testing.T) {
	// 测试表达式构建
	expr := Exp.Add(F("price"), F("tax"))
	result := expr.Build()

	if result == nil {
		t.Error("Expression build failed")
	}
}

func TestFieldReference(t *testing.T) {
	// 测试字段引用
	field := F("user").Dot("email")
	if field.String() != "$user.email" {
		t.Errorf("Expected $user.email, got %s", field.String())
	}

	// 测试数组索引
	field2 := F("items").Index(0).Dot("price")
	if field2.String() != "$items.0.price" {
		t.Errorf("Expected $items.0.price, got %s", field2.String())
	}
}

func TestComplexFilter(t *testing.T) {
	// 测试复杂过滤器
	filter := Filter().
		Eq("status", "active").
		Between("age", 18, 65).
		In("city", "北京", "上海", "深圳").
		All("tags", "verified", "premium").
		Or(
			Filter().Eq("vip", true),
			Filter().Gte("total_spent", 10000),
		)

	result := filter.Build()
	if len(result) < 5 {
		t.Errorf("Expected at least 5 conditions, got %d", len(result))
	}
}
