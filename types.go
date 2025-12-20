package mgo

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 核心类型别名 ====================

// ObjectID MongoDB ObjectID 类型
type ObjectID = bson.ObjectID

// Decimal128 MongoDB Decimal128 类型
type Decimal128 = bson.Decimal128

// DateTime MongoDB DateTime 类型
type DateTime = bson.DateTime

// Regex MongoDB 正则表达式类型
type Regex = bson.Regex

// JavaScript MongoDB JavaScript 代码类型
type JavaScript = bson.JavaScript

// Binary MongoDB 二进制数据类型
type Binary = bson.Binary

// Timestamp MongoDB 时间戳类型
type Timestamp = bson.Timestamp

// Symbol MongoDB Symbol 类型
type Symbol = bson.Symbol

// MinKey MongoDB MinKey 类型
type MinKey = bson.MinKey

// MaxKey MongoDB MaxKey 类型
type MaxKey = bson.MaxKey

// Undefined MongoDB Undefined 类型
type Undefined = bson.Undefined

// Null MongoDB Null 类型
type Null = bson.Null

// DBPointer MongoDB DBPointer 类型
type DBPointer = bson.DBPointer

// CodeWithScope MongoDB CodeWithScope 类型
type CodeWithScope = bson.CodeWithScope

// ==================== 文档类型别名 ====================

// M bson.M 的别名，用于快速创建文档
type M = bson.M

// D bson.D 的别名，用于有序文档
type D = bson.D

// A bson.A 的别名，用于数组
type A = bson.A

// E bson.E 的别名，用于文档元素
type E = bson.E

// Raw bson.Raw 的别名，用于原始 BSON 数据
type Raw = bson.Raw

// ==================== 上下文和会话类型 ====================

// SessionContext MongoDB 会话上下文类型（MongoDB v2 中已移除）
// type SessionContext = mongo.SessionContext

// ==================== Pipeline 类型 ====================

// Pipeline MongoDB 聚合管道类型
type Pipeline = mongo.Pipeline

// ==================== ObjectID 辅助函数 ====================

// NewObjectID 生成新的 ObjectID
//
// 示例：
//
//	id := mgo.NewObjectID()
func NewObjectID() ObjectID {
	return bson.NewObjectID()
}

// ObjectIDFromHex 从十六进制字符串创建 ObjectID
//
// 示例：
//
//	id, err := mgo.ObjectIDFromHex("507f1f77bcf86cd799439011")
//	if err != nil {
//	    return err
//	}
func ObjectIDFromHex(hex string) (ObjectID, error) {
	return bson.ObjectIDFromHex(hex)
}

// MustObjectIDFromHex 从十六进制字符串创建 ObjectID，失败时 panic
//
// 示例：
//
//	id := mgo.MustObjectIDFromHex("507f1f77bcf86cd799439011")
func MustObjectIDFromHex(hex string) ObjectID {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return id
}

// IsValidObjectID 检查字符串是否为有效的 ObjectID
//
// 示例：
//
//	if mgo.IsValidObjectID(idStr) {
//	    id, _ := mgo.ObjectIDFromHex(idStr)
//	}
func IsValidObjectID(hex string) bool {
	// 检查长度和字符
	if len(hex) != 24 {
		return false
	}
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ==================== DateTime 辅助函数 ====================

// NewDateTime 从 time.Time 创建 DateTime
//
// 示例：
//
//	dt := mgo.NewDateTime(time.Now())
func NewDateTime(t time.Time) DateTime {
	return DateTime(t.UnixMilli())
}

// Now 返回当前时间的 DateTime
//
// 示例：
//
//	dt := mgo.Now()
func Now() DateTime {
	return DateTime(time.Now().UnixMilli())
}

// ==================== Regex 辅助函数 ====================

// NewRegex 创建正则表达式
//
// 示例：
//
//	regex := mgo.NewRegex("^test", "i")
func NewRegex(pattern, options string) Regex {
	return bson.Regex{
		Pattern: pattern,
		Options: options,
	}
}

// ==================== JavaScript 辅助函数 ====================

// NewJavaScript 创建 JavaScript 代码
//
// 示例：
//
//	js := mgo.NewJavaScript("function() { return this.value * 2; }")
func NewJavaScript(code string) JavaScript {
	return JavaScript(code)
}

// NewCodeWithScope 创建带作用域的 JavaScript 代码
//
// 示例：
//
//	code := mgo.NewCodeWithScope("function() { return x * 2; }", mgo.M{"x": 10})
func NewCodeWithScope(code string, scope interface{}) CodeWithScope {
	return bson.CodeWithScope{
		Code:  bson.JavaScript(code),
		Scope: scope,
	}
}

// ==================== Decimal128 辅助函数 ====================

// NewDecimal128 从字符串创建 Decimal128
//
// 示例：
//
//	dec, err := mgo.NewDecimal128("123.45")
//	if err != nil {
//	    return err
//	}
func NewDecimal128(s string) (Decimal128, error) {
	return bson.ParseDecimal128(s)
}

// MustDecimal128 从字符串创建 Decimal128，失败时 panic
//
// 示例：
//
//	dec := mgo.MustDecimal128("123.45")
func MustDecimal128(s string) Decimal128 {
	dec, err := bson.ParseDecimal128(s)
	if err != nil {
		panic(err)
	}
	return dec
}

// NewDecimal128FromFloat64 从 float64 创建 Decimal128
//
// 示例：
//
//	dec := mgo.NewDecimal128FromFloat64(123.45)
func NewDecimal128FromFloat64(f float64) Decimal128 {
	// 简单实现，将 float64 转换为字符串再解析
	dec, _ := bson.ParseDecimal128(fmt.Sprintf("%.2f", f))
	return dec
}

// ==================== Binary 辅助函数 ====================

// NewBinary 创建二进制数据
//
// 示例：
//
//	bin := mgo.NewBinary(0x00, []byte("data"))
func NewBinary(subtype byte, data []byte) Binary {
	return bson.Binary{
		Subtype: subtype,
		Data:    data,
	}
}

// ==================== Timestamp 辅助函数 ====================

// NewTimestamp 创建时间戳
//
// 示例：
//
//	ts := mgo.NewTimestamp(uint32(time.Now().Unix()), 0)
func NewTimestamp(t, i uint32) Timestamp {
	return bson.Timestamp{
		T: t,
		I: i,
	}
}

// ==================== 特殊类型实例 ====================

var (
	// NilObjectID 零值 ObjectID
	NilObjectID = bson.NilObjectID

	// MinKeyVal MinKey 实例
	MinKeyVal = bson.MinKey{}

	// MaxKeyVal MaxKey 实例
	MaxKeyVal = bson.MaxKey{}

	// UndefinedVal Undefined 实例
	UndefinedVal = bson.Undefined{}

	// NullVal Null 实例
	NullVal = bson.Null{}
)

// ==================== Expr 表达式接口 ====================

// Expr 表达式接口，所有表达式构建器都实现此接口
//
// 示例：
//
//	expr := Exp.Add(F("price"), F("tax"))
//	value := expr.Build() // 构建为 bson 值
type Expr interface {
	Build() any
}

// expr 基础表达式实现
type expr struct {
	value any
}

func (e *expr) Build() any {
	return e.value
}
