package agg

import "github.com/gocrud/mgo"

// ==================== 聚合操作符 ====================

// Add 加法操作符
//
// 示例：
//
//	expr := agg.Add("$price", "$tax")
func Add(exprs ...string) mgo.M {
	return mgo.M{"$add": exprs}
}

// Subtract 减法操作符
//
// 示例：
//
//	expr := agg.Subtract("$price", "$discount")
func Subtract(expr1, expr2 string) mgo.M {
	return mgo.M{"$subtract": []string{expr1, expr2}}
}

// Multiply 乘法操作符
//
// 示例：
//
//	expr := agg.Multiply("$price", "$quantity")
func Multiply(exprs ...string) mgo.M {
	return mgo.M{"$multiply": exprs}
}

// Divide 除法操作符
//
// 示例：
//
//	expr := agg.Divide("$total", "$count")
func Divide(expr1, expr2 string) mgo.M {
	return mgo.M{"$divide": []string{expr1, expr2}}
}

// Mod 取模操作符
//
// 示例：
//
//	expr := agg.Mod("$age", 10)
func Mod(expr string, divisor int) mgo.M {
	return mgo.M{"$mod": []interface{}{expr, divisor}}
}

// ==================== 字符串操作符 ====================

// Concat 字符串连接
//
// 示例：
//
//	expr := agg.Concat("$firstName", " ", "$lastName")
func Concat(exprs ...string) mgo.M {
	return mgo.M{"$concat": exprs}
}

// Substr 字符串截取
//
// 示例：
//
//	expr := agg.Substr("$name", 0, 5)
func Substr(expr string, start, length int) mgo.M {
	return mgo.M{"$substr": []interface{}{expr, start, length}}
}

// ToLower 转小写
//
// 示例：
//
//	expr := agg.ToLower("$email")
func ToLower(expr string) mgo.M {
	return mgo.M{"$toLower": expr}
}

// ToUpper 转大写
//
// 示例：
//
//	expr := agg.ToUpper("$name")
func ToUpper(expr string) mgo.M {
	return mgo.M{"$toUpper": expr}
}

// ==================== 数组操作符 ====================

// Size 数组大小
//
// 示例：
//
//	expr := agg.Size("$tags")
func Size(expr string) mgo.M {
	return mgo.M{"$size": expr}
}

// ArrayElemAt 获取数组元素
//
// 示例：
//
//	expr := agg.ArrayElemAt("$tags", 0)
func ArrayElemAt(expr string, index int) mgo.M {
	return mgo.M{"$arrayElemAt": []interface{}{expr, index}}
}

// Slice 数组切片
//
// 示例：
//
//	expr := agg.Slice("$tags", 0, 5)
func Slice(expr string, start, length int) mgo.M {
	return mgo.M{"$slice": []interface{}{expr, start, length}}
}

// Filter 过滤数组
//
// 示例：
//
//	expr := agg.Filter("$items", "item", mgo.M{"$gte": []string{"$$item.price", 100}})
func Filter(input, as string, cond mgo.M) mgo.M {
	return mgo.M{"$filter": mgo.M{
		"input": input,
		"as":    as,
		"cond":  cond,
	}}
}

// Map 映射数组
//
// 示例：
//
//	expr := agg.Map("$items", "item", "$$item.price")
func Map(input, as, in string) mgo.M {
	return mgo.M{"$map": mgo.M{
		"input": input,
		"as":    as,
		"in":    in,
	}}
}

// Reduce 归约数组
//
// 示例：
//
//	expr := agg.Reduce("$items", 0, mgo.M{"$add": []string{"$$value", "$$this.price"}})
func Reduce(input string, initialValue interface{}, in mgo.M) mgo.M {
	return mgo.M{"$reduce": mgo.M{
		"input":        input,
		"initialValue": initialValue,
		"in":           in,
	}}
}

// ==================== 条件操作符 ====================

// Cond 条件表达式
//
// 示例：
//
//	expr := agg.Cond(mgo.M{"$gt": []string{"$age", 18}}, "adult", "minor")
func Cond(condition mgo.M, ifTrue, ifFalse interface{}) mgo.M {
	return mgo.M{"$cond": []interface{}{condition, ifTrue, ifFalse}}
}

// IfNull 如果为 null 则返回默认值
//
// 示例：
//
//	expr := agg.IfNull("$email", "no-email")
func IfNull(expr string, replacement interface{}) mgo.M {
	return mgo.M{"$ifNull": []interface{}{expr, replacement}}
}

// Switch 多条件分支
//
// 示例：
//
//	expr := agg.Switch(
//	    []mgo.M{
//	        {"case": mgo.M{"$eq": []string{"$status", "active"}}, "then": 1},
//	        {"case": mgo.M{"$eq": []string{"$status", "pending"}}, "then": 2},
//	    },
//	    0,  // default
//	)
func Switch(branches []mgo.M, defaultValue interface{}) mgo.M {
	return mgo.M{"$switch": mgo.M{
		"branches": branches,
		"default":  defaultValue,
	}}
}

// ==================== 比较操作符 ====================

// Eq 等于比较
//
// 示例：
//
//	expr := agg.Eq("$status", "active")
func Eq(expr1, expr2 string) mgo.M {
	return mgo.M{"$eq": []string{expr1, expr2}}
}

// Ne 不等于比较
//
// 示例：
//
//	expr := agg.Ne("$status", "deleted")
func Ne(expr1, expr2 string) mgo.M {
	return mgo.M{"$ne": []string{expr1, expr2}}
}

// Gt 大于比较
//
// 示例：
//
//	expr := agg.Gt("$age", "18")
func Gt(expr1, expr2 string) mgo.M {
	return mgo.M{"$gt": []string{expr1, expr2}}
}

// Gte 大于等于比较
//
// 示例：
//
//	expr := agg.Gte("$age", "18")
func Gte(expr1, expr2 string) mgo.M {
	return mgo.M{"$gte": []string{expr1, expr2}}
}

// Lt 小于比较
//
// 示例：
//
//	expr := agg.Lt("$age", "60")
func Lt(expr1, expr2 string) mgo.M {
	return mgo.M{"$lt": []string{expr1, expr2}}
}

// Lte 小于等于比较
//
// 示例：
//
//	expr := agg.Lte("$age", "60")
func Lte(expr1, expr2 string) mgo.M {
	return mgo.M{"$lte": []string{expr1, expr2}}
}

// ==================== 逻辑操作符 ====================

// And 逻辑与
//
// 示例：
//
//	expr := agg.And(condition1, condition2)
func And(exprs ...mgo.M) mgo.M {
	return mgo.M{"$and": exprs}
}

// Or 逻辑或
//
// 示例：
//
//	expr := agg.Or(condition1, condition2)
func Or(exprs ...mgo.M) mgo.M {
	return mgo.M{"$or": exprs}
}

// Not 逻辑非
//
// 示例：
//
//	expr := agg.Not(condition)
func Not(expr mgo.M) mgo.M {
	return mgo.M{"$not": expr}
}

// ==================== 日期操作符 ====================

// Year 获取年份
//
// 示例：
//
//	expr := agg.Year("$created_at")
func Year(expr string) mgo.M {
	return mgo.M{"$year": expr}
}

// Month 获取月份
//
// 示例：
//
//	expr := agg.Month("$created_at")
func Month(expr string) mgo.M {
	return mgo.M{"$month": expr}
}

// DayOfMonth 获取日期
//
// 示例：
//
//	expr := agg.DayOfMonth("$created_at")
func DayOfMonth(expr string) mgo.M {
	return mgo.M{"$dayOfMonth": expr}
}

// DayOfWeek 获取星期
//
// 示例：
//
//	expr := agg.DayOfWeek("$created_at")
func DayOfWeek(expr string) mgo.M {
	return mgo.M{"$dayOfWeek": expr}
}

// Hour 获取小时
//
// 示例：
//
//	expr := agg.Hour("$created_at")
func Hour(expr string) mgo.M {
	return mgo.M{"$hour": expr}
}

// Minute 获取分钟
//
// 示例：
//
//	expr := agg.Minute("$created_at")
func Minute(expr string) mgo.M {
	return mgo.M{"$minute": expr}
}

// Second 获取秒
//
// 示例：
//
//	expr := agg.Second("$created_at")
func Second(expr string) mgo.M {
	return mgo.M{"$second": expr}
}

// DateToString 日期转字符串
//
// 示例：
//
//	expr := agg.DateToString("$created_at", "%Y-%m-%d")
func DateToString(expr, format string) mgo.M {
	return mgo.M{"$dateToString": mgo.M{
		"format": format,
		"date":   expr,
	}}
}
