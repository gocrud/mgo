package mgo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

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

// Pipeline MongoDB 聚合管道类型
type Pipeline = mongo.Pipeline

// M bson.M 的别名，用于快速创建文档
type M = bson.M

// D bson.D 的别名，用于有序文档
type D = bson.D

// A bson.A 的别名，用于数组
type A = bson.A

// E bson.E 的别名，用于文档元素
type E = bson.E
