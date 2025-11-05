package mgo

import (
	"regexp"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// FilterBuilder 查询条件构建器，用于构建 MongoDB 查询过滤器
// 完全消除 "$gte", "$in" 等操作符字符串，提供流畅的链式 API
//
// 示例：
//
//	// 基础查询
//	filter := mgo.Filter().
//	    Eq("status", "active").
//	    Gte("age", 18).
//	    In("city", "北京", "上海", "深圳")
//
//	// 复杂逻辑
//	filter := mgo.Filter().
//	    Eq("status", "active").
//	    Or(
//	        mgo.Filter().Eq("vip", true),
//	        mgo.Filter().Gte("level", 5),
//	    )
type FilterBuilder struct {
	conditions bson.D
}

// Filter 创建新的过滤器构建器
//
// 使用简洁的语法：
//
//	filter := mgo.Filter().Eq("status", "active")
//
//	// 在聚合中使用
//	coll.Aggs(ctx).Match(mgo.Filter().Eq("status", "active"))
//
//	// 在逻辑操作中使用
//	coll.Query(ctx).Or(
//	    mgo.Filter().Eq("vip", true),
//	    mgo.Filter().Gte("level", 5),
//	)
func Filter() *FilterBuilder {
	return &FilterBuilder{conditions: bson.D{}}
}

// ===== 基础条件方法 =====

// Eq 等于条件
//
// MongoDB: {field: value}
//
// 示例：
//
//	filter := mgo.Filter().Eq("status", "active")
//	filter := mgo.Filter().Eq("age", 25)
func (f *FilterBuilder) Eq(field string, value any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{Key: field, Value: value})
	return f
}

// Ne 不等于条件 ($ne)
//
// MongoDB: {field: {$ne: value}}
//
// 示例：
//
//	filter := mgo.Filter().Ne("status", "deleted")
func (f *FilterBuilder) Ne(field string, value any) *FilterBuilder {
	f.addOperator(field, "$ne", value)
	return f
}

// Gt 大于条件 ($gt)
//
// MongoDB: {field: {$gt: value}}
//
// 示例：
//
//	filter := mgo.Filter().Gt("age", 18)
//	filter := mgo.Filter().Gt("price", 100.0)
func (f *FilterBuilder) Gt(field string, value any) *FilterBuilder {
	f.addOperator(field, "$gt", value)
	return f
}

// Gte 大于等于条件 ($gte)
//
// MongoDB: {field: {$gte: value}}
//
// 示例：
//
//	filter := mgo.Filter().Gte("age", 18)
//	filter := mgo.Filter().Gte("created_at", startDate)
func (f *FilterBuilder) Gte(field string, value any) *FilterBuilder {
	f.addOperator(field, "$gte", value)
	return f
}

// Lt 小于条件 ($lt)
//
// MongoDB: {field: {$lt: value}}
//
// 示例：
//
//	filter := mgo.Filter().Lt("age", 65)
//	filter := mgo.Filter().Lt("stock", 10)
func (f *FilterBuilder) Lt(field string, value any) *FilterBuilder {
	f.addOperator(field, "$lt", value)
	return f
}

// Lte 小于等于条件 ($lte)
//
// MongoDB: {field: {$lte: value}}
//
// 示例：
//
//	filter := mgo.Filter().Lte("age", 65)
//	filter := mgo.Filter().Lte("price", 1000.0)
func (f *FilterBuilder) Lte(field string, value any) *FilterBuilder {
	f.addOperator(field, "$lte", value)
	return f
}

// Between 范围查询（包含边界）
//
// MongoDB: {field: {$gte: min, $lte: max}}
//
// 示例：
//
//	// 年龄在 18-65 之间
//	filter := mgo.Filter().Between("age", 18, 65)
//	// 价格在 100-1000 之间
//	filter := mgo.Filter().Between("price", 100.0, 1000.0)
func (f *FilterBuilder) Between(field string, min, max any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key: field,
		Value: bson.D{
			{Key: "$gte", Value: min},
			{Key: "$lte", Value: max},
		},
	})
	return f
}

// BetweenExclusive 范围查询（不包含边界）
//
// MongoDB: {field: {$gt: min, $lt: max}}
//
// 示例：
//
//	// 年龄在 18-65 之间（不包含18和65）
//	filter := mgo.Filter().BetweenExclusive("age", 18, 65)
func (f *FilterBuilder) BetweenExclusive(field string, min, max any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key: field,
		Value: bson.D{
			{Key: "$gt", Value: min},
			{Key: "$lt", Value: max},
		},
	})
	return f
}

// ===== 数组和集合操作 =====

// In 在列表中 ($in)
//
// MongoDB: {field: {$in: [values...]}}
//
// 示例：
//
//	// 城市在指定列表中
//	filter := mgo.Filter().In("city", "北京", "上海", "深圳")
//	// 状态在指定列表中
//	filter := mgo.Filter().In("status", "pending", "approved")
func (f *FilterBuilder) In(field string, values ...any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$in", values),
	})
	return f
}

// NotIn 不在列表中 ($nin)
//
// MongoDB: {field: {$nin: [values...]}}
//
// 示例：
//
//	filter := mgo.Filter().NotIn("status", "deleted", "archived")
func (f *FilterBuilder) NotIn(field string, values ...any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$nin", values),
	})
	return f
}

// All 数组包含所有元素 ($all)
//
// MongoDB: {field: {$all: [values...]}}
//
// 示例：
//
//	// tags 必须包含所有指定标签
//	filter := mgo.Filter().All("tags", "active", "verified", "premium")
func (f *FilterBuilder) All(field string, values ...any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$all", values),
	})
	return f
}

// ElemMatch 数组元素匹配 ($elemMatch)
//
// MongoDB: {field: {$elemMatch: {condition...}}}
//
// 示例：
//
//	// items 数组中至少有一个元素的 price > 1000
//	filter := mgo.Filter().ElemMatch("items",
//	    mgo.Filter().Gt("price", 1000),
//	)
//
//	// 复杂匹配
//	filter := mgo.Filter().ElemMatch("orders",
//	    mgo.Filter().
//	        Eq("status", "completed").
//	        Gt("amount", 100),
//	)
func (f *FilterBuilder) ElemMatch(field string, subFilter *FilterBuilder) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$elemMatch", subFilter.Build()),
	})
	return f
}

// Size 数组大小 ($size)
//
// MongoDB: {field: {$size: n}}
//
// 示例：
//
//	// tags 数组必须包含恰好 3 个元素
//	filter := mgo.Filter().Size("tags", 3)
func (f *FilterBuilder) Size(field string, size int) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$size", size),
	})
	return f
}

// ===== 字段存在性 =====

// Exists 字段存在 ($exists: true)
//
// MongoDB: {field: {$exists: true}}
//
// 示例：
//
//	// email 字段必须存在
//	filter := mgo.Filter().Exists("email")
func (f *FilterBuilder) Exists(field string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$exists", true),
	})
	return f
}

// NotExists 字段不存在 ($exists: false)
//
// MongoDB: {field: {$exists: false}}
//
// 示例：
//
//	// deleted_at 字段不存在
//	filter := mgo.Filter().NotExists("deleted_at")
func (f *FilterBuilder) NotExists(field string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$exists", false),
	})
	return f
}

// IsNull 字段为 null
//
// MongoDB: {field: null}
//
// 示例：
//
//	// deleted_at 为 null
//	filter := mgo.Filter().IsNull("deleted_at")
func (f *FilterBuilder) IsNull(field string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{Key: field, Value: nil})
	return f
}

// IsNotNull 字段不为 null ($ne: null)
//
// MongoDB: {field: {$ne: null}}
//
// 示例：
//
//	// email 不为 null
//	filter := mgo.Filter().IsNotNull("email")
func (f *FilterBuilder) IsNotNull(field string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$ne", nil),
	})
	return f
}

// ===== 字符串匹配 =====

// Regex 正则匹配 ($regex)
//
// MongoDB: {field: {$regex: pattern, $options: options}}
//
// 示例：
//
//	// 名字以 "张" 开头（不区分大小写）
//	filter := mgo.Filter().Regex("name", "^张", "i")
//	// 邮箱格式验证
//	filter := mgo.Filter().Regex("email", "^[a-z]+@[a-z]+\\.[a-z]+$", "i")
func (f *FilterBuilder) Regex(field string, pattern string, options ...string) *FilterBuilder {
	regexDoc := makeD("$regex", pattern)
	if len(options) > 0 && options[0] != "" {
		regexDoc = append(regexDoc, bson.E{Key: "$options", Value: options[0]})
	}
	f.conditions = append(f.conditions, bson.E{Key: field, Value: regexDoc})
	return f
}

// StartsWith 以...开头（不区分大小写）
//
// 示例：
//
//	// 名字以 "张" 开头
//	filter := mgo.Filter().StartsWith("name", "张")
func (f *FilterBuilder) StartsWith(field string, prefix string) *FilterBuilder {
	return f.Regex(field, "^"+regexp.QuoteMeta(prefix), "i")
}

// EndsWith 以...结尾（不区分大小写）
//
// 示例：
//
//	// 文件名以 ".pdf" 结尾
//	filter := mgo.Filter().EndsWith("filename", ".pdf")
func (f *FilterBuilder) EndsWith(field string, suffix string) *FilterBuilder {
	return f.Regex(field, regexp.QuoteMeta(suffix)+"$", "i")
}

// Contains 包含子字符串（不区分大小写）
//
// 示例：
//
//	// 描述中包含 "优惠"
//	filter := mgo.Filter().Contains("description", "优惠")
func (f *FilterBuilder) Contains(field string, substring string) *FilterBuilder {
	return f.Regex(field, regexp.QuoteMeta(substring), "i")
}

// Text 全文搜索 ($text)
//
// MongoDB: {$text: {$search: search, $language: language}}
//
// 示例：
//
//	// 搜索包含 "机器学习" 的文档（中文）
//	filter := mgo.Filter().Text("机器学习", "zh")
//	// 英文搜索
//	filter := mgo.Filter().Text("machine learning", "en")
func (f *FilterBuilder) Text(search string, language ...string) *FilterBuilder {
	textDoc := makeD("$search", search)
	if len(language) > 0 && language[0] != "" {
		textDoc = append(textDoc, bson.E{Key: "$language", Value: language[0]})
	}
	f.conditions = append(f.conditions, bson.E{Key: "$text", Value: textDoc})
	return f
}

// ===== 类型判断 =====

// Type 类型匹配 ($type)
//
// MongoDB: {field: {$type: bsonType}}
//
// 示例：
//
//	// 字段类型是字符串
//	filter := mgo.Filter().Type("name", "string")
//	// 字段类型是数字
//	filter := mgo.Filter().Type("age", "number")
func (f *FilterBuilder) Type(field string, bsonType string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$type", bsonType),
	})
	return f
}

// ===== 数值运算 =====

// Mod 取模运算 ($mod)
//
// MongoDB: {field: {$mod: [divisor, remainder]}}
//
// 示例：
//
//	// 偶数
//	filter := mgo.Filter().Mod("value", 2, 0)
//	// 奇数
//	filter := mgo.Filter().Mod("value", 2, 1)
func (f *FilterBuilder) Mod(field string, divisor, remainder int64) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$mod", []int64{divisor, remainder}),
	})
	return f
}
