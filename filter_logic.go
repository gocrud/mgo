package mgo

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
	conds := make([]interface{}, len(conditions))
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
		for k, v := range cond {
			result[k] = v
		}
	}
	return result
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
