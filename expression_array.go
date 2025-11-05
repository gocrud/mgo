package mgo

import "go.mongodb.org/mongo-driver/v2/bson"

// ===== 数组表达式操作符 =====

// Size 数组大小 ($size)
//
// 示例：
//
//	expr := Exp.Size(F("items"))
func (eb *ExprBuilder) Size(array any) Expr {
	return &expr{makeD("$size", unwrap(array))}
}

// ArrayElemAt 获取数组元素 ($arrayElemAt)
//
// 示例：
//
//	// 获取第一个元素
//	expr := Exp.ArrayElemAt(F("items"), 0)
//	// 获取最后一个元素
//	expr := Exp.ArrayElemAt(F("items"), -1)
func (eb *ExprBuilder) ArrayElemAt(array any, index int) Expr {
	return &expr{makeD("$arrayElemAt", []any{unwrap(array), index})}
}

// ConcatArrays 连接数组 ($concatArrays)
//
// 示例：
//
//	expr := Exp.ConcatArrays(F("tags1"), F("tags2"), F("tags3"))
func (eb *ExprBuilder) ConcatArrays(arrays ...any) Expr {
	return &expr{makeD("$concatArrays", unwrapExprs(arrays))}
}

// Filter 过滤数组 ($filter)
//
// 示例：
//
//	// 过滤出价格大于100的商品
//	expr := Exp.Filter(
//	    F("items"),
//	    Exp.Gt(F("this").Dot("price"), 100),
//	    "this",
//	)
func (eb *ExprBuilder) Filter(input any, cond Expr, as string) Expr {
	return &expr{makeD("$filter", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "cond", Value: cond.Build()},
		{Key: "as", Value: as},
	})}
}

// Map 映射数组 ($map)
//
// 示例：
//
//	// 提取所有商品的价格
//	expr := Exp.Map(
//	    F("items"),
//	    F("this").Dot("price"),
//	    "this",
//	)
func (eb *ExprBuilder) Map(input any, inExpr any, as string) Expr {
	return &expr{makeD("$map", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "in", Value: unwrap(inExpr)},
		{Key: "as", Value: as},
	})}
}

// Reduce 归约数组 ($reduce)
//
// 示例：
//
//	// 计算数组总和
//	expr := Exp.Reduce(
//	    F("numbers"),
//	    0,
//	    Exp.Add(F("value"), F("this")),
//	)
func (eb *ExprBuilder) Reduce(input, initialValue, inExpr any) Expr {
	return &expr{makeD("$reduce", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "initialValue", Value: unwrap(initialValue)},
		{Key: "in", Value: unwrap(inExpr)},
	})}
}

// Slice 数组切片 ($slice)
//
// 示例：
//
//	// 获取前5个元素
//	expr := Exp.Slice(F("items"), 5)
//	// 从位置2开始获取3个元素
//	expr := Exp.Slice2(F("items"), 2, 3)
func (eb *ExprBuilder) Slice(array any, n int) Expr {
	return &expr{makeD("$slice", []any{unwrap(array), n})}
}

// Slice2 数组切片（指定位置和数量）($slice)
//
// 示例：
//
//	// 从位置2开始获取3个元素
//	expr := Exp.Slice2(F("items"), 2, 3)
func (eb *ExprBuilder) Slice2(array any, position, n int) Expr {
	return &expr{makeD("$slice", []any{unwrap(array), position, n})}
}

// ReverseArray 反转数组 ($reverseArray)
//
// 示例：
//
//	expr := Exp.ReverseArray(F("items"))
func (eb *ExprBuilder) ReverseArray(array any) Expr {
	return &expr{makeD("$reverseArray", unwrap(array))}
}

// IndexOfArray 数组元素索引 ($indexOfArray)
//
// 示例：
//
//	expr := Exp.IndexOfArray(F("tags"), "important")
func (eb *ExprBuilder) IndexOfArray(array, search any) Expr {
	return &expr{makeD("$indexOfArray", []any{unwrap(array), unwrap(search)})}
}

// Range 生成数字数组 ($range)
//
// 示例：
//
//	// 生成 [0, 1, 2, 3, 4]
//	expr := Exp.Range(0, 5, 1)
func (eb *ExprBuilder) Range(start, end, step int) Expr {
	return &expr{makeD("$range", []any{start, end, step})}
}

// Zip 合并数组 ($zip)
//
// 示例：
//
//	expr := Exp.Zip(
//	    []any{F("names"), F("ages")},
//	    false,
//	    nil,
//	)
func (eb *ExprBuilder) Zip(inputs any, useLongestLength bool, defaults any) Expr {
	doc := bson.D{
		{Key: "inputs", Value: unwrap(inputs)},
		{Key: "useLongestLength", Value: useLongestLength},
	}
	if defaults != nil {
		doc = append(doc, bson.E{Key: "defaults", Value: unwrap(defaults)})
	}
	return &expr{makeD("$zip", doc)}
}

// In 值是否在数组中 ($in)
//
// 示例：
//
//	expr := Exp.In("admin", F("roles"))
func (eb *ExprBuilder) In(value, array any) Expr {
	return &expr{makeD("$in", []any{unwrap(value), unwrap(array)})}
}

// IsArray 是否是数组 ($isArray)
//
// 示例：
//
//	expr := Exp.IsArray(F("tags"))
func (eb *ExprBuilder) IsArray(value any) Expr {
	return &expr{makeD("$isArray", unwrap(value))}
}

// ArrayToObject 数组转对象 ($arrayToObject)
//
// 示例：
//
//	expr := Exp.ArrayToObject(F("keyValuePairs"))
func (eb *ExprBuilder) ArrayToObject(array any) Expr {
	return &expr{makeD("$arrayToObject", unwrap(array))}
}

// ObjectToArray 对象转数组 ($objectToArray)
//
// 示例：
//
//	expr := Exp.ObjectToArray(F("document"))
func (eb *ExprBuilder) ObjectToArray(object any) Expr {
	return &expr{makeD("$objectToArray", unwrap(object))}
}

// ===== 日期表达式操作符 =====

// Year 获取年份 ($year)
//
// 示例：
//
//	expr := Exp.Year(F("created_at"))
func (eb *ExprBuilder) Year(date any) Expr {
	return &expr{makeD("$year", unwrap(date))}
}

// Month 获取月份 ($month)
//
// 示例：
//
//	expr := Exp.Month(F("created_at"))
func (eb *ExprBuilder) Month(date any) Expr {
	return &expr{makeD("$month", unwrap(date))}
}

// DayOfMonth 获取日 ($dayOfMonth)
//
// 示例：
//
//	expr := Exp.DayOfMonth(F("created_at"))
func (eb *ExprBuilder) DayOfMonth(date any) Expr {
	return &expr{makeD("$dayOfMonth", unwrap(date))}
}

// DayOfWeek 获取星期几 ($dayOfWeek)
//
// 示例：
//
//	expr := Exp.DayOfWeek(F("created_at"))
func (eb *ExprBuilder) DayOfWeek(date any) Expr {
	return &expr{makeD("$dayOfWeek", unwrap(date))}
}

// DayOfYear 获取一年中的第几天 ($dayOfYear)
//
// 示例：
//
//	expr := Exp.DayOfYear(F("created_at"))
func (eb *ExprBuilder) DayOfYear(date any) Expr {
	return &expr{makeD("$dayOfYear", unwrap(date))}
}

// Hour 获取小时 ($hour)
//
// 示例：
//
//	expr := Exp.Hour(F("timestamp"))
func (eb *ExprBuilder) Hour(date any) Expr {
	return &expr{makeD("$hour", unwrap(date))}
}

// Minute 获取分钟 ($minute)
//
// 示例：
//
//	expr := Exp.Minute(F("timestamp"))
func (eb *ExprBuilder) Minute(date any) Expr {
	return &expr{makeD("$minute", unwrap(date))}
}

// Second 获取秒 ($second)
//
// 示例：
//
//	expr := Exp.Second(F("timestamp"))
func (eb *ExprBuilder) Second(date any) Expr {
	return &expr{makeD("$second", unwrap(date))}
}

// Millisecond 获取毫秒 ($millisecond)
//
// 示例：
//
//	expr := Exp.Millisecond(F("timestamp"))
func (eb *ExprBuilder) Millisecond(date any) Expr {
	return &expr{makeD("$millisecond", unwrap(date))}
}

// Week 获取周数 ($week)
//
// 示例：
//
//	expr := Exp.Week(F("created_at"))
func (eb *ExprBuilder) Week(date any) Expr {
	return &expr{makeD("$week", unwrap(date))}
}

// IsoWeek ISO周数 ($isoWeek)
//
// 示例：
//
//	expr := Exp.IsoWeek(F("created_at"))
func (eb *ExprBuilder) IsoWeek(date any) Expr {
	return &expr{makeD("$isoWeek", unwrap(date))}
}

// IsoWeekYear ISO周年份 ($isoWeekYear)
//
// 示例：
//
//	expr := Exp.IsoWeekYear(F("created_at"))
func (eb *ExprBuilder) IsoWeekYear(date any) Expr {
	return &expr{makeD("$isoWeekYear", unwrap(date))}
}

// IsoDayOfWeek ISO星期几 ($isoDayOfWeek)
//
// 示例：
//
//	expr := Exp.IsoDayOfWeek(F("created_at"))
func (eb *ExprBuilder) IsoDayOfWeek(date any) Expr {
	return &expr{makeD("$isoDayOfWeek", unwrap(date))}
}

// DateToString 日期转字符串 ($dateToString)
//
// 示例：
//
//	// 格式化日期
//	expr := Exp.DateToString("%Y-%m-%d", F("created_at"))
//	// 带时区
//	expr := Exp.DateToStringTZ("%Y-%m-%d %H:%M:%S", F("created_at"), "Asia/Shanghai")
func (eb *ExprBuilder) DateToString(format string, date any) Expr {
	return &expr{makeD("$dateToString", bson.D{
		{Key: "format", Value: format},
		{Key: "date", Value: unwrap(date)},
	})}
}

// DateToStringTZ 日期转字符串（带时区）($dateToString)
//
// 示例：
//
//	expr := Exp.DateToStringTZ("%Y-%m-%d %H:%M:%S", F("created_at"), "Asia/Shanghai")
func (eb *ExprBuilder) DateToStringTZ(format string, date any, timezone string) Expr {
	return &expr{makeD("$dateToString", bson.D{
		{Key: "format", Value: format},
		{Key: "date", Value: unwrap(date)},
		{Key: "timezone", Value: timezone},
	})}
}

// DateFromString 字符串转日期 ($dateFromString)
//
// 示例：
//
//	expr := Exp.DateFromString("2024-01-01")
func (eb *ExprBuilder) DateFromString(dateString any) Expr {
	return &expr{makeD("$dateFromString", bson.D{
		{Key: "dateString", Value: unwrap(dateString)},
	})}
}

// DateAdd 日期加法 ($dateAdd)
//
// 示例：
//
//	// 加7天
//	expr := Exp.DateAdd(F("created_at"), "day", 7)
func (eb *ExprBuilder) DateAdd(startDate any, unit string, amount int) Expr {
	return &expr{makeD("$dateAdd", bson.D{
		{Key: "startDate", Value: unwrap(startDate)},
		{Key: "unit", Value: unit},
		{Key: "amount", Value: amount},
	})}
}

// DateSubtract 日期减法 ($dateSubtract)
//
// 示例：
//
//	// 减7天
//	expr := Exp.DateSubtract(F("created_at"), "day", 7)
func (eb *ExprBuilder) DateSubtract(startDate any, unit string, amount int) Expr {
	return &expr{makeD("$dateSubtract", bson.D{
		{Key: "startDate", Value: unwrap(startDate)},
		{Key: "unit", Value: unit},
		{Key: "amount", Value: amount},
	})}
}

// DateDiff 日期差 ($dateDiff)
//
// 示例：
//
//	// 计算两个日期相差的天数
//	expr := Exp.DateDiff(F("start_date"), F("end_date"), "day")
func (eb *ExprBuilder) DateDiff(startDate, endDate any, unit string) Expr {
	return &expr{makeD("$dateDiff", bson.D{
		{Key: "startDate", Value: unwrap(startDate)},
		{Key: "endDate", Value: unwrap(endDate)},
		{Key: "unit", Value: unit},
	})}
}

// DateTrunc 日期截断 ($dateTrunc)
//
// 示例：
//
//	// 截断到小时
//	expr := Exp.DateTrunc(F("timestamp"), "hour")
func (eb *ExprBuilder) DateTrunc(date any, unit string) Expr {
	return &expr{makeD("$dateTrunc", bson.D{
		{Key: "date", Value: unwrap(date)},
		{Key: "unit", Value: unit},
	})}
}

// ToDate 转日期 ($toDate)
//
// 示例：
//
//	expr := Exp.ToDate(F("timestamp_string"))
func (eb *ExprBuilder) ToDate(value any) Expr {
	return &expr{makeD("$toDate", unwrap(value))}
}

// ===== 类型转换操作符 =====

// ToInt 转整数 ($toInt)
//
// 示例：
//
//	expr := Exp.ToInt(F("score_string"))
func (eb *ExprBuilder) ToInt(value any) Expr {
	return &expr{makeD("$toInt", unwrap(value))}
}

// ToLong 转长整数 ($toLong)
//
// 示例：
//
//	expr := Exp.ToLong(F("id_string"))
func (eb *ExprBuilder) ToLong(value any) Expr {
	return &expr{makeD("$toLong", unwrap(value))}
}

// ToDouble 转浮点数 ($toDouble)
//
// 示例：
//
//	expr := Exp.ToDouble(F("price_string"))
func (eb *ExprBuilder) ToDouble(value any) Expr {
	return &expr{makeD("$toDouble", unwrap(value))}
}

// ToDecimal 转Decimal ($toDecimal)
//
// 示例：
//
//	expr := Exp.ToDecimal(F("amount_string"))
func (eb *ExprBuilder) ToDecimal(value any) Expr {
	return &expr{makeD("$toDecimal", unwrap(value))}
}

// ToBool 转布尔值 ($toBool)
//
// 示例：
//
//	expr := Exp.ToBool(F("flag_string"))
func (eb *ExprBuilder) ToBool(value any) Expr {
	return &expr{makeD("$toBool", unwrap(value))}
}

// ToObjectId 转ObjectId ($toObjectId)
//
// 示例：
//
//	expr := Exp.ToObjectId(F("id_string"))
func (eb *ExprBuilder) ToObjectId(value any) Expr {
	return &expr{makeD("$toObjectId", unwrap(value))}
}

// TypeExpr 获取类型 ($type)
//
// 示例：
//
//	expr := Exp.TypeExpr(F("value"))
func (eb *ExprBuilder) TypeExpr(value any) Expr {
	return &expr{makeD("$type", unwrap(value))}
}

// IsNumber 是否是数字 ($isNumber)
//
// 示例：
//
//	expr := Exp.IsNumber(F("value"))
func (eb *ExprBuilder) IsNumber(value any) Expr {
	return &expr{makeD("$isNumber", unwrap(value))}
}

// Convert 类型转换 ($convert)
//
// 示例：
//
//	// 转换为整数
//	expr := Exp.Convert(F("value"), "int")
func (eb *ExprBuilder) Convert(input any, to string) Expr {
	return &expr{makeD("$convert", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "to", Value: to},
	})}
}

// ===== 对象操作符 =====

// MergeObjects 合并对象 ($mergeObjects)
//
// 示例：
//
//	expr := Exp.MergeObjects(F("defaults"), F("user_settings"))
func (eb *ExprBuilder) MergeObjects(objects ...any) Expr {
	return &expr{makeD("$mergeObjects", unwrapExprs(objects))}
}

// ===== 集合操作符 =====

// SetDifference 集合差 ($setDifference)
//
// 示例：
//
//	expr := Exp.SetDifference(F("all_tags"), F("excluded_tags"))
func (eb *ExprBuilder) SetDifference(set1, set2 any) Expr {
	return &expr{makeD("$setDifference", []any{unwrap(set1), unwrap(set2)})}
}

// SetEquals 集合相等 ($setEquals)
//
// 示例：
//
//	expr := Exp.SetEquals(F("tags1"), F("tags2"))
func (eb *ExprBuilder) SetEquals(sets ...any) Expr {
	return &expr{makeD("$setEquals", unwrapExprs(sets))}
}

// SetIntersection 集合交 ($setIntersection)
//
// 示例：
//
//	expr := Exp.SetIntersection(F("user_tags"), F("required_tags"))
func (eb *ExprBuilder) SetIntersection(sets ...any) Expr {
	return &expr{makeD("$setIntersection", unwrapExprs(sets))}
}

// SetUnion 集合并 ($setUnion)
//
// 示例：
//
//	expr := Exp.SetUnion(F("tags1"), F("tags2"), F("tags3"))
func (eb *ExprBuilder) SetUnion(sets ...any) Expr {
	return &expr{makeD("$setUnion", unwrapExprs(sets))}
}

// SetIsSubset 是否是子集 ($setIsSubset)
//
// 示例：
//
//	expr := Exp.SetIsSubset(F("user_permissions"), F("required_permissions"))
func (eb *ExprBuilder) SetIsSubset(set1, set2 any) Expr {
	return &expr{makeD("$setIsSubset", []any{unwrap(set1), unwrap(set2)})}
}

// AllElementsTrue 所有元素为真 ($allElementsTrue)
//
// 示例：
//
//	expr := Exp.AllElementsTrue(F("flags"))
func (eb *ExprBuilder) AllElementsTrue(set any) Expr {
	return &expr{makeD("$allElementsTrue", unwrap(set))}
}

// AnyElementTrue 任一元素为真 ($anyElementTrue)
//
// 示例：
//
//	expr := Exp.AnyElementTrue(F("flags"))
func (eb *ExprBuilder) AnyElementTrue(set any) Expr {
	return &expr{makeD("$anyElementTrue", unwrap(set))}
}

// ===== 辅助函数 =====

// Literal 字面量 ($literal)
//
// 示例：
//
//	// 将 "$price" 作为字面量字符串，而不是字段引用
//	expr := Exp.Literal("$price")
func (eb *ExprBuilder) Literal(value any) Expr {
	return &expr{makeD("$literal", value)}
}

// Let 定义变量 ($let)
//
// 示例：
//
//	expr := Exp.Let(
//	    bson.D{{"total", Exp.Add(F("price"), F("tax"))}},
//	    Exp.Mul(F("total"), 0.9),
//	)
func (eb *ExprBuilder) Let(vars, inExpr any) Expr {
	return &expr{makeD("$let", bson.D{
		{Key: "vars", Value: unwrap(vars)},
		{Key: "in", Value: unwrap(inExpr)},
	})}
}

// TextScore 文本搜索分数 ($meta: "textScore")
//
// 示例：
//
//	expr := Exp.TextScore()
func (eb *ExprBuilder) TextScore() Expr {
	return &expr{makeD("$meta", "textScore")}
}
