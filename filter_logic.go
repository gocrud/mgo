package mgo

import "go.mongodb.org/mongo-driver/v2/bson"

// filter_logic.go - Filter 的逻辑操作符和辅助方法

// ===== 逻辑操作符 =====

// And 逻辑与（合并多个过滤器）($and)
//
// MongoDB: {$and: [{condition1}, {condition2}, ...]}
//
// 示例：
//
//	// age >= 18 AND status == "active"
//	filter := coll.Query().And(
//	    mgo.Filter().Gte("age", 18),
//	    mgo.Filter().Eq("status", "active"),
//	)
//
//	// 复杂组合
//	filter := coll.Query().
//	    Eq("status", "active").
//	    And(
//	        mgo.Filter().Gte("age", 18).Lt("age", 65),
//	        mgo.Filter().In("city", "北京", "上海"),
//	    )
func (f *FilterBuilder) And(filters ...*FilterBuilder) *FilterBuilder {
	conditions := make([]any, len(filters))
	for i, filter := range filters {
		conditions[i] = filter.Build()
	}
	f.conditions = append(f.conditions, bson.E{
		Key:   "$and",
		Value: conditions,
	})
	return f
}

// Or 逻辑或 ($or)
//
// MongoDB: {$or: [{condition1}, {condition2}, ...]}
//
// 示例：
//
//	// vip == true OR level >= 5
//	filter := coll.Query().Or(
//	    mgo.Filter().Eq("vip", true),
//	    mgo.Filter().Gte("level", 5),
//	)
//
//	// 多个或条件
//	filter := coll.Query().
//	    Eq("status", "active").
//	    Or(
//	        mgo.Filter().Eq("type", "premium"),
//	        mgo.Filter().Gte("total_spent", 10000),
//	        mgo.Filter().All("tags", "vip", "member"),
//	    )
func (f *FilterBuilder) Or(filters ...*FilterBuilder) *FilterBuilder {
	conditions := make([]any, len(filters))
	for i, filter := range filters {
		conditions[i] = filter.Build()
	}
	f.conditions = append(f.conditions, bson.E{
		Key:   "$or",
		Value: conditions,
	})
	return f
}

// Nor 逻辑或非 ($nor)
//
// MongoDB: {$nor: [{condition1}, {condition2}, ...]}
//
// 示例：
//
//	// NOT (status == "deleted" OR status == "archived")
//	filter := coll.Query().Nor(
//	    mgo.Filter().Eq("status", "deleted"),
//	    mgo.Filter().Eq("status", "archived"),
//	)
func (f *FilterBuilder) Nor(filters ...*FilterBuilder) *FilterBuilder {
	conditions := make([]any, len(filters))
	for i, filter := range filters {
		conditions[i] = filter.Build()
	}
	f.conditions = append(f.conditions, bson.E{
		Key:   "$nor",
		Value: conditions,
	})
	return f
}

// Not 逻辑非 ($not)
//
// MongoDB: {field: {$not: {condition}}}
//
// 示例：
//
//	// NOT (age > 65)
//	filter := mgo.Filter().Not("age", makeD("$gt", 65))
func (f *FilterBuilder) Not(field string, condition any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$not", condition),
	})
	return f
}

// ===== 地理位置查询 =====

// Near 地理位置附近 ($near)
//
// MongoDB: {field: {$near: {$geometry: {type: "Point", coordinates: [lng, lat]}, $maxDistance: distance}}}
//
// 示例：
//
//	// 查找距离指定坐标5公里内的地点
//	filter := mgo.Filter().Near("location", 116.404, 39.915, 5000)
func (f *FilterBuilder) Near(field string, lng, lat float64, maxDistance ...float64) *FilterBuilder {
	nearDoc := bson.D{
		{Key: "$geometry", Value: bson.D{
			{Key: "type", Value: "Point"},
			{Key: "coordinates", Value: []float64{lng, lat}},
		}},
	}
	if len(maxDistance) > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$maxDistance", Value: maxDistance[0]})
	}
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$near", nearDoc),
	})
	return f
}

// NearSphere 球面地理位置附近 ($nearSphere)
//
// 示例：
//
//	filter := mgo.Filter().NearSphere("location", 116.404, 39.915, 5000)
func (f *FilterBuilder) NearSphere(field string, lng, lat float64, maxDistance ...float64) *FilterBuilder {
	nearDoc := bson.D{
		{Key: "$geometry", Value: bson.D{
			{Key: "type", Value: "Point"},
			{Key: "coordinates", Value: []float64{lng, lat}},
		}},
	}
	if len(maxDistance) > 0 {
		nearDoc = append(nearDoc, bson.E{Key: "$maxDistance", Value: maxDistance[0]})
	}
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$nearSphere", nearDoc),
	})
	return f
}

// GeoWithin 在地理区域内 ($geoWithin)
//
// 示例：
//
//	// 在多边形内
//	geometry := bson.D{
//	    {"type", "Polygon"},
//	    {"coordinates", [][]float64{{...}}},
//	}
//	filter := mgo.Filter().GeoWithin("location", geometry)
func (f *FilterBuilder) GeoWithin(field string, geometry any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$geoWithin", makeD("$geometry", geometry)),
	})
	return f
}

// GeoIntersects 地理位置相交 ($geoIntersects)
//
// 示例：
//
//	filter := mgo.Filter().GeoIntersects("location", geometry)
func (f *FilterBuilder) GeoIntersects(field string, geometry any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$geoIntersects", makeD("$geometry", geometry)),
	})
	return f
}

// ===== 位操作符 =====

// BitsAllSet 所有位都设置 ($bitsAllSet)
//
// 示例：
//
//	filter := mgo.Filter().BitsAllSet("permissions", []int{1, 5})
func (f *FilterBuilder) BitsAllSet(field string, positions []int) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$bitsAllSet", positions),
	})
	return f
}

// BitsAnySet 任意位设置 ($bitsAnySet)
//
// 示例：
//
//	filter := mgo.Filter().BitsAnySet("permissions", []int{1, 5})
func (f *FilterBuilder) BitsAnySet(field string, positions []int) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$bitsAnySet", positions),
	})
	return f
}

// BitsAllClear 所有位都清除 ($bitsAllClear)
//
// 示例：
//
//	filter := mgo.Filter().BitsAllClear("permissions", []int{1, 5})
func (f *FilterBuilder) BitsAllClear(field string, positions []int) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$bitsAllClear", positions),
	})
	return f
}

// BitsAnyClear 任意位清除 ($bitsAnyClear)
//
// 示例：
//
//	filter := mgo.Filter().BitsAnyClear("permissions", []int{1, 5})
func (f *FilterBuilder) BitsAnyClear(field string, positions []int) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD("$bitsAnyClear", positions),
	})
	return f
}

// ===== 评估查询操作符 =====

// Expr 使用聚合表达式 ($expr)
//
// 示例：
//
//	// spent_amount > budget
//	filter := mgo.Filter().Expr(
//	    mgo.Exp.Gt(mgo.F("spent_amount"), mgo.F("budget")),
//	)
func (f *FilterBuilder) Expr(expression Expr) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   "$expr",
		Value: expression.Build(),
	})
	return f
}

// JsonSchema JSON Schema 验证 ($jsonSchema)
//
// 示例：
//
//	schema := bson.D{{"bsonType", "object"}, ...}
//	filter := mgo.Filter().JsonSchema(schema)
func (f *FilterBuilder) JsonSchema(schema any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   "$jsonSchema",
		Value: schema,
	})
	return f
}

// Where JavaScript 表达式 ($where)
//
// 示例：
//
//	filter := mgo.Filter().Where("this.age > 18")
func (f *FilterBuilder) Where(jsExpression string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   "$where",
		Value: jsExpression,
	})
	return f
}

// ===== 注释 =====

// Comment 添加查询注释 ($comment)
//
// 示例：
//
//	filter := mgo.Filter().
//	    Eq("status", "active").
//	    Comment("查询活跃用户")
func (f *FilterBuilder) Comment(comment string) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{
		Key:   "$comment",
		Value: comment,
	})
	return f
}

// ===== 辅助方法 =====

// Raw 添加原始条件（用于特殊情况）
//
// 示例：
//
//	// 使用原始 bson.D
//	filter := mgo.Filter().Raw("custom_field", bson.D{{"$custom", "value"}})
func (f *FilterBuilder) Raw(key string, value any) *FilterBuilder {
	f.conditions = append(f.conditions, bson.E{Key: key, Value: value})
	return f
}

// Merge 合并其他过滤器
//
// 示例：
//
//	baseFilter := mgo.Filter().Eq("status", "active")
//	additionalFilter := mgo.Filter().Gte("age", 18)
//	combined := baseFilter.Merge(additionalFilter)
func (f *FilterBuilder) Merge(other *FilterBuilder) *FilterBuilder {
	f.conditions = append(f.conditions, other.conditions...)
	return f
}

// Build 构建最终查询条件
//
// 示例：
//
//	filter := mgo.Filter().Eq("status", "active")
//	bsonD := filter.Build() // => bson.D{{"status", "active"}}
func (f *FilterBuilder) Build() bson.D {
	return f.conditions
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	filter := mgo.Filter().Eq("status", "active")
//	bsonM := filter.BuildM() // => bson.M{"status": "active"}
func (f *FilterBuilder) BuildM() bson.M {
	result := make(bson.M, len(f.conditions))
	for _, e := range f.conditions {
		result[e.Key] = e.Value
	}
	return result
}

// Empty 检查是否为空过滤器
//
// 示例：
//
//	filter := mgo.Filter()
//	if filter.Empty() {
//	    // 没有任何条件
//	}
func (f *FilterBuilder) Empty() bool {
	return len(f.conditions) == 0
}

// Clone 克隆过滤器
//
// 示例：
//
//	original := mgo.Filter().Eq("status", "active")
//	cloned := original.Clone().Gte("age", 18)
func (f *FilterBuilder) Clone() *FilterBuilder {
	clone := Filter()
	clone.conditions = make(bson.D, len(f.conditions))
	copy(clone.conditions, f.conditions)
	return clone
}

// addOperator 添加操作符（合并同字段多个操作符）
// 这是一个内部方法，用于处理同一字段有多个比较操作符的情况
//
// 示例：
//
//	// age >= 18 AND age < 65
//	// 会被合并为: {age: {$gte: 18, $lt: 65}}
func (f *FilterBuilder) addOperator(field string, operator string, value any) {
	// 查找是否已存在该字段
	for i, e := range f.conditions {
		if e.Key == field {
			// 如果值已经是 bson.D，追加操作符
			if existingDoc, ok := e.Value.(bson.D); ok {
				f.conditions[i].Value = append(existingDoc, bson.E{Key: operator, Value: value})
				return
			}
		}
	}
	// 否则新增
	f.conditions = append(f.conditions, bson.E{
		Key:   field,
		Value: makeD(operator, value),
	})
}
