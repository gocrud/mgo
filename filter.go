package mgo

import "maps"

// ==================== 条件构建函数 ====================

// Eq 等于条件
//
// 示例：
//
//	filter := mgo.Eq("status", "active")
//	// 生成: {"status": "active"}
func Eq(field string, value any) M {
	return M{field: value}
}

// Ne 不等于条件
//
// 示例：
//
//	filter := mgo.Ne("status", "deleted")
//	// 生成: {"status": {"$ne": "deleted"}}
func Ne(field string, value any) M {
	return M{field: M{"$ne": value}}
}

// Gt 大于条件
//
// 示例：
//
//	filter := mgo.Gt("age", 18)
//	// 生成: {"age": {"$gt": 18}}
func Gt(field string, value any) M {
	return M{field: M{"$gt": value}}
}

// Gte 大于等于条件
//
// 示例：
//
//	filter := mgo.Gte("age", 18)
//	// 生成: {"age": {"$gte": 18}}
func Gte(field string, value any) M {
	return M{field: M{"$gte": value}}
}

// Lt 小于条件
//
// 示例：
//
//	filter := mgo.Lt("age", 60)
//	// 生成: {"age": {"$lt": 60}}
func Lt(field string, value any) M {
	return M{field: M{"$lt": value}}
}

// Lte 小于等于条件
//
// 示例：
//
//	filter := mgo.Lte("age", 60)
//	// 生成: {"age": {"$lte": 60}}
func Lte(field string, value any) M {
	return M{field: M{"$lte": value}}
}

// In 包含于条件
//
// 示例：
//
//	filter := mgo.In("status", "active", "pending")
//	// 生成: {"status": {"$in": ["active", "pending"]}}
func In(field string, values ...any) M {
	return M{field: M{"$in": values}}
}

// Nin 不包含于条件
//
// 示例：
//
//	filter := mgo.Nin("status", "deleted", "expired")
//	// 生成: {"status": {"$nin": ["deleted", "expired"]}}
func Nin(field string, values ...any) M {
	return M{field: M{"$nin": values}}
}

// Exists 字段存在条件
//
// 示例：
//
//	filter := mgo.Exists("email", true)
//	// 生成: {"email": {"$exists": true}}
func Exists(field string, exists bool) M {
	return M{field: M{"$exists": exists}}
}

// Type 字段类型条件
//
// 示例：
//
//	filter := mgo.Type("age", "int")
//	// 生成: {"age": {"$type": "int"}}
func Type(field string, bsonType string) M {
	return M{field: M{"$type": bsonType}}
}

// RegexFilter 正则表达式条件
//
// 示例：
//
//	filter := mgo.RegexFilter("name", "^John", "i")
//	// 生成: {"name": {"$regex": "^John", "$options": "i"}}
func RegexFilter(field, pattern, options string) M {
	return M{field: M{
		"$regex":   pattern,
		"$options": options,
	}}
}

// Like 模糊匹配条件（简化版正则）
//
// 示例：
//
//	filter := mgo.Like("name", "John")
//	// 生成: {"name": {"$regex": "John", "$options": "i"}}
func Like(field, value string) M {
	return M{field: M{
		"$regex":   value,
		"$options": "i",
	}}
}

// StartsWith 以...开头条件
//
// 示例：
//
//	filter := mgo.StartsWith("name", "John")
//	// 生成: {"name": {"$regex": "^John", "$options": "i"}}
func StartsWith(field, value string) M {
	return M{field: M{
		"$regex":   "^" + value,
		"$options": "i",
	}}
}

// EndsWith 以...结尾条件
//
// 示例：
//
//	filter := mgo.EndsWith("email", "@example.com")
//	// 生成: {"email": {"$regex": "@example\\.com$", "$options": "i"}}
func EndsWith(field, value string) M {
	return M{field: M{
		"$regex":   value + "$",
		"$options": "i",
	}}
}

// ==================== 数组条件 ====================

// All 数组包含所有指定值
//
// 示例：
//
//	filter := mgo.All("tags", "go", "mongodb")
//	// 生成: {"tags": {"$all": ["go", "mongodb"]}}
func All(field string, values ...any) M {
	return M{field: M{"$all": values}}
}

// ElemMatch 数组元素匹配条件
//
// 示例：
//
//	filter := mgo.ElemMatch("orders", mgo.M{"status": "pending", "amount": mgo.M{"$gt": 100}})
//	// 生成: {"orders": {"$elemMatch": {"status": "pending", "amount": {"$gt": 100}}}}
func ElemMatch(field string, condition M) M {
	return M{field: M{"$elemMatch": condition}}
}

// Size 数组大小条件
//
// 示例：
//
//	filter := mgo.Size("tags", 3)
//	// 生成: {"tags": {"$size": 3}}
func Size(field string, size int) M {
	return M{field: M{"$size": size}}
}

// ==================== 地理位置条件 ====================

// Near 附近位置条件
//
// 示例：
//
//	filter := mgo.Near("location", 116.4, 39.9, 5000) // 5km 范围内
//	// 生成: {"location": {"$near": {"$geometry": {"type": "Point", "coordinates": [116.4, 39.9]}, "$maxDistance": 5000}}}
func Near(field string, longitude, latitude float64, maxDistance int) M {
	return M{field: M{
		"$near": M{
			"$geometry": M{
				"type":        "Point",
				"coordinates": []float64{longitude, latitude},
			},
			"$maxDistance": maxDistance,
		},
	}}
}

// GeoWithin 地理范围内条件
//
// 示例：
//
//	// 圆形范围
//	filter := mgo.GeoWithin("location", mgo.M{
//	    "$centerSphere": []interface{}{[]float64{116.4, 39.9}, 0.001},
//	})
func GeoWithin(field string, geometry M) M {
	return M{field: M{"$geoWithin": geometry}}
}

// ==================== 文本搜索 ====================

// Text 全文搜索条件
//
// 示例：
//
//	filter := mgo.Text("search term")
//	// 生成: {"$text": {"$search": "search term"}}
func Text(search string) M {
	return M{"$text": M{"$search": search}}
}

// TextWithLanguage 指定语言的全文搜索
//
// 示例：
//
//	filter := mgo.TextWithLanguage("搜索词", "chinese")
//	// 生成: {"$text": {"$search": "搜索词", "$language": "chinese"}}
func TextWithLanguage(search, language string) M {
	return M{"$text": M{
		"$search":   search,
		"$language": language,
	}}
}

// ==================== 高级条件 ====================

// Mod 取模条件
//
// 示例：
//
//	filter := mgo.Mod("age", 5, 0) // age % 5 == 0
//	// 生成: {"age": {"$mod": [5, 0]}}
func Mod(field string, divisor, remainder int) M {
	return M{field: M{"$mod": []int{divisor, remainder}}}
}

// Where JavaScript 表达式条件
//
// 示例：
//
//	filter := mgo.Where("this.age > 18 && this.status === 'active'")
//	// 生成: {"$where": "this.age > 18 && this.status === 'active'"}
func Where(js string) M {
	return M{"$where": js}
}

// ExprFilter 聚合表达式条件
//
// 示例：
//
//	filter := mgo.ExprFilter(mgo.M{"$gt": []string{"$spent", "$budget"}})
//	// 生成: {"$expr": {"$gt": ["$spent", "$budget"]}}
func ExprFilter(expression M) M {
	return M{"$expr": expression}
}

// ==================== 位操作条件 ====================

// BitsAllSet 位全部为 1
//
// 示例：
//
//	filter := mgo.BitsAllSet("permissions", 5) // 二进制 101
//	// 生成: {"permissions": {"$bitsAllSet": 5}}
func BitsAllSet(field string, bitmask int) M {
	return M{field: M{"$bitsAllSet": bitmask}}
}

// BitsAnySet 位任意为 1
//
// 示例：
//
//	filter := mgo.BitsAnySet("permissions", 5)
//	// 生成: {"permissions": {"$bitsAnySet": 5}}
func BitsAnySet(field string, bitmask int) M {
	return M{field: M{"$bitsAnySet": bitmask}}
}

// BitsAllClear 位全部为 0
//
// 示例：
//
//	filter := mgo.BitsAllClear("permissions", 5)
//	// 生成: {"permissions": {"$bitsAllClear": 5}}
func BitsAllClear(field string, bitmask int) M {
	return M{field: M{"$bitsAllClear": bitmask}}
}

// BitsAnyClear 位任意为 0
//
// 示例：
//
//	filter := mgo.BitsAnyClear("permissions", 5)
//	// 生成: {"permissions": {"$bitsAnyClear": 5}}
func BitsAnyClear(field string, bitmask int) M {
	return M{field: M{"$bitsAnyClear": bitmask}}
}

// ==================== 逻辑条件组合 ====================

// And 逻辑与条件
//
// 示例：
//
//	filter := mgo.And(
//	    mgo.Eq("status", "active"),
//	    mgo.Gt("age", 18),
//	    mgo.In("city", "北京", "上海"),
//	)
//	// 生成: {"$and": [{"status": "active"}, {"age": {"$gt": 18}}, {"city": {"$in": ["北京", "上海"]}}]}
func And(conditions ...M) M {
	if len(conditions) == 0 {
		return M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}

	// 转换为 []interface{}
	conds := make([]interface{}, len(conditions))
	for i, c := range conditions {
		conds[i] = c
	}

	return M{"$and": conds}
}

// Or 逻辑或条件
//
// 示例：
//
//	filter := mgo.Or(
//	    mgo.Eq("status", "active"),
//	    mgo.Eq("status", "pending"),
//	)
//	// 生成: {"$or": [{"status": "active"}, {"status": "pending"}]}
func Or(conditions ...M) M {
	if len(conditions) == 0 {
		return M{}
	}
	if len(conditions) == 1 {
		return conditions[0]
	}

	// 转换为 []interface{}
	conds := make([]interface{}, len(conditions))
	for i, c := range conditions {
		conds[i] = c
	}

	return M{"$or": conds}
}

// Not 逻辑非条件
//
// 示例：
//
//	filter := mgo.Not("status", mgo.M{"$eq": "deleted"})
//	// 生成: {"status": {"$not": {"$eq": "deleted"}}}
func Not(field string, condition M) M {
	return M{field: M{"$not": condition}}
}

// Nor 逻辑非或条件
//
// 示例：
//
//	filter := mgo.Nor(
//	    mgo.Eq("status", "deleted"),
//	    mgo.Eq("status", "expired"),
//	)
//	// 生成: {"$nor": [{"status": "deleted"}, {"status": "expired"}]}
func Nor(conditions ...M) M {
	if len(conditions) == 0 {
		return M{}
	}

	// 转换为 []interface{}
	conds := make([]any, len(conditions))
	for i, c := range conditions {
		conds[i] = c
	}

	return M{"$nor": conds}
}

// ==================== 条件合并辅助函数 ====================

// Merge 合并多个条件
//
// 示例：
//
//	filter := mgo.Merge(
//	    mgo.Eq("status", "active"),
//	    mgo.Gt("age", 18),
//	)
//	// 生成: {"status": "active", "age": {"$gt": 18}}
func Merge(conditions ...M) M {
	result := M{}
	for _, cond := range conditions {
		maps.Copy(result, cond)
	}
	return result
}

// ==================== 条件逻辑组合函数 ====================

// AndIf 条件逻辑与（仅当 condition 为 true 时才包含该条件）
//
// 用于根据业务逻辑动态构建 AND 条件组合
//
// 示例：
//
//	filter := mgo.And(
//	    mgo.Eq("status", "active"),
//	    mgo.AndIf(minAge > 0, mgo.Gt("age", minAge)),
//	    mgo.AndIf(hasCity, mgo.Eq("city", city)),
//	)
func AndIf(condition bool, m M) M {
	if !condition {
		return M{}
	}
	return m
}

// OrIf 条件逻辑或（仅当 condition 为 true 时才包含该条件）
//
// 用于根据业务逻辑动态构建 OR 条件组合
//
// 示例：
//
//	filter := mgo.Or(
//	    mgo.OrIf(searchName, mgo.Like("name", keyword)),
//	    mgo.OrIf(searchEmail, mgo.Like("email", keyword)),
//	)
func OrIf(condition bool, m M) M {
	if !condition {
		return M{}
	}
	return m
}

// MergeIf 条件合并（仅当 condition 为 true 时才合并条件）
//
// 用于根据业务逻辑动态合并过滤条件
//
// 示例：
//
//	filter := mgo.Merge(
//	    mgo.Eq("status", "active"),
//	    mgo.MergeIf(hasAgeFilter, mgo.Gt("age", minAge)),
//	    mgo.MergeIf(hasCityFilter, mgo.Eq("city", city)),
//	)
func MergeIf(condition bool, m M) M {
	if !condition {
		return M{}
	}
	return m
}

// ==================== 高级逻辑组合 ====================

// Filter 构建复杂过滤条件的辅助结构
type Filter struct {
	conditions []M
	operator   string
}

// NewFilter 创建新的过滤器
//
// 示例：
//
//	f := mgo.NewFilter()
//	f.Add(mgo.Eq("status", "active"))
//	f.Add(mgo.Gt("age", 18))
//	filter := f.Build()
func NewFilter() *Filter {
	return &Filter{
		conditions: make([]M, 0),
		operator:   "$and",
	}
}

// Add 添加条件
func (f *Filter) Add(condition M) *Filter {
	f.conditions = append(f.conditions, condition)
	return f
}

// SetOperator 设置操作符（$and, $or, $nor）
func (f *Filter) SetOperator(op string) *Filter {
	f.operator = op
	return f
}

// Build 构建最终的过滤条件
func (f *Filter) Build() M {
	if len(f.conditions) == 0 {
		return M{}
	}
	if len(f.conditions) == 1 {
		return f.conditions[0]
	}

	switch f.operator {
	case "$and":
		return And(f.conditions...)
	case "$or":
		return Or(f.conditions...)
	case "$nor":
		return Nor(f.conditions...)
	default:
		return And(f.conditions...)
	}
}

// ==================== 便捷组合函数 ====================

// Between 范围条件（包含边界）
//
// 示例：
//
//	filter := mgo.Between("age", 18, 60)
//	// 生成: {"age": {"$gte": 18, "$lte": 60}}
func Between(field string, min, max interface{}) M {
	return M{field: M{
		"$gte": min,
		"$lte": max,
	}}
}

// NotBetween 不在范围条件
//
// 示例：
//
//	filter := mgo.NotBetween("age", 18, 60)
//	// 生成: {"$or": [{"age": {"$lt": 18}}, {"age": {"$gt": 60}}]}
func NotBetween(field string, min, max interface{}) M {
	return Or(
		Lt(field, min),
		Gt(field, max),
	)
}

// IsNull 字段为空条件
//
// 示例：
//
//	filter := mgo.IsNull("deleted_at")
//	// 生成: {"deleted_at": null}
func IsNull(field string) M {
	return M{field: nil}
}

// IsNotNull 字段不为空条件
//
// 示例：
//
//	filter := mgo.IsNotNull("deleted_at")
//	// 生成: {"deleted_at": {"$ne": null}}
func IsNotNull(field string) M {
	return M{field: M{"$ne": nil}}
}

// IsEmpty 数组为空条件
//
// 示例：
//
//	filter := mgo.IsEmpty("tags")
//	// 生成: {"$or": [{"tags": {"$exists": false}}, {"tags": {"$size": 0}}]}
func IsEmpty(field string) M {
	return Or(
		Exists(field, false),
		Size(field, 0),
	)
}

// IsNotEmpty 数组不为空条件
//
// 示例：
//
//	filter := mgo.IsNotEmpty("tags")
//	// 生成: {"$and": [{"tags": {"$exists": true}}, {"tags": {"$not": {"$size": 0}}}]}
func IsNotEmpty(field string) M {
	return And(
		Exists(field, true),
		Not(field, M{"$size": 0}),
	)
}

// ==================== 条件判断辅助 ====================

// If 条件判断辅助函数
//
// 示例：
//
//	filter := mgo.Merge(
//	    mgo.Eq("status", "active"),
//	    mgo.If(hasCity, mgo.Eq("city", city)),
//	)
func If(condition bool, m M) M {
	if condition {
		return m
	}
	return M{}
}

// IfElse 条件判断辅助函数（带 else）
//
// 示例：
//
//	filter := mgo.Merge(
//	    mgo.IfElse(isVIP, mgo.Eq("vip", true), mgo.Eq("score", mgo.M{"$gte": 90})),
//	)
func IfElse(condition bool, ifTrue, ifFalse M) M {
	if condition {
		return ifTrue
	}
	return ifFalse
}
