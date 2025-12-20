package mgo

// ==================== 条件构建函数 ====================

// Eq 等于条件
//
// 示例：
//
//	filter := mgo.Eq("status", "active")
//	// 生成: {"status": "active"}
func Eq(field string, value interface{}) M {
	return M{field: value}
}

// Ne 不等于条件
//
// 示例：
//
//	filter := mgo.Ne("status", "deleted")
//	// 生成: {"status": {"$ne": "deleted"}}
func Ne(field string, value interface{}) M {
	return M{field: M{"$ne": value}}
}

// Gt 大于条件
//
// 示例：
//
//	filter := mgo.Gt("age", 18)
//	// 生成: {"age": {"$gt": 18}}
func Gt(field string, value interface{}) M {
	return M{field: M{"$gt": value}}
}

// Gte 大于等于条件
//
// 示例：
//
//	filter := mgo.Gte("age", 18)
//	// 生成: {"age": {"$gte": 18}}
func Gte(field string, value interface{}) M {
	return M{field: M{"$gte": value}}
}

// Lt 小于条件
//
// 示例：
//
//	filter := mgo.Lt("age", 60)
//	// 生成: {"age": {"$lt": 60}}
func Lt(field string, value interface{}) M {
	return M{field: M{"$lt": value}}
}

// Lte 小于等于条件
//
// 示例：
//
//	filter := mgo.Lte("age", 60)
//	// 生成: {"age": {"$lte": 60}}
func Lte(field string, value interface{}) M {
	return M{field: M{"$lte": value}}
}

// In 包含于条件
//
// 示例：
//
//	filter := mgo.In("status", "active", "pending")
//	// 生成: {"status": {"$in": ["active", "pending"]}}
func In(field string, values ...interface{}) M {
	return M{field: M{"$in": values}}
}

// Nin 不包含于条件
//
// 示例：
//
//	filter := mgo.Nin("status", "deleted", "expired")
//	// 生成: {"status": {"$nin": ["deleted", "expired"]}}
func Nin(field string, values ...interface{}) M {
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
func All(field string, values ...interface{}) M {
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
