package mgo

import (
	"fmt"
	"strings"
)

// Field 字段引用类型，用于在表达式中引用 MongoDB 字段
//
// 示例：
//
//	// 创建字段引用
//	userField := F("user")
//	emailField := F("user").Dot("email")
//	firstItemField := F("items").Index(0)
type Field string

// F 创建字段引用
//
// 示例：
//
//	field := mgo.F("name")          // => "$name"
//	field := mgo.F("user.email")    // => "$user.email"
func F(name string) Field {
	return Field(name)
}

// String 转换为 MongoDB 字段引用格式（添加 $ 前缀）
//
// 示例：
//
//	field := F("name")
//	str := field.String() // => "$name"
func (f Field) String() string {
	if strings.HasPrefix(string(f), "$") {
		return string(f)
	}
	if strings.HasPrefix(string(f), "$$") {
		return string(f)
	}
	return "$" + string(f)
}

// Dot 访问嵌套字段
//
// 示例：
//
//	field := F("user").Dot("address").Dot("city")
//	// => "$user.address.city"
func (f Field) Dot(subField string) Field {
	return Field(string(f) + "." + subField)
}

// Index 访问数组元素
//
// 示例：
//
//	field := F("items").Index(0)           // => "$items.0"
//	field := F("items").Index(0).Dot("price") // => "$items.0.price"
func (f Field) Index(i int) Field {
	return Field(fmt.Sprintf("%s.%d", f, i))
}

// Raw 返回原始字段名（不添加 $ 前缀）
//
// 示例：
//
//	field := F("name")
//	raw := field.Raw() // => "name"
func (f Field) Raw() string {
	return strings.TrimPrefix(string(f), "$")
}
