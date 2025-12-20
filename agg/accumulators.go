package agg

import "github.com/gocrud/mgo"

// ==================== 累加器函数 ====================

// Sum 求和累加器
//
// 示例：
//
//	expr := agg.Sum("$amount")
func Sum(expr string) mgo.M {
	return mgo.M{"$sum": expr}
}

// Avg 平均值累加器
//
// 示例：
//
//	expr := agg.Avg("$age")
func Avg(expr string) mgo.M {
	return mgo.M{"$avg": expr}
}

// Max 最大值累加器
//
// 示例：
//
//	expr := agg.Max("$price")
func Max(expr string) mgo.M {
	return mgo.M{"$max": expr}
}

// Min 最小值累加器
//
// 示例：
//
//	expr := agg.Min("$price")
func Min(expr string) mgo.M {
	return mgo.M{"$min": expr}
}

// First 第一个值累加器
//
// 示例：
//
//	expr := agg.First("$name")
func First(expr string) mgo.M {
	return mgo.M{"$first": expr}
}

// Last 最后一个值累加器
//
// 示例：
//
//	expr := agg.Last("$name")
func Last(expr string) mgo.M {
	return mgo.M{"$last": expr}
}

// Push 收集到数组累加器
//
// 示例：
//
//	expr := agg.Push("$item")
func Push(expr string) mgo.M {
	return mgo.M{"$push": expr}
}

// AddToSet 收集到数组累加器（去重）
//
// 示例：
//
//	expr := agg.AddToSet("$tag")
func AddToSet(expr string) mgo.M {
	return mgo.M{"$addToSet": expr}
}

// StdDevPop 总体标准差
//
// 示例：
//
//	expr := agg.StdDevPop("$score")
func StdDevPop(expr string) mgo.M {
	return mgo.M{"$stdDevPop": expr}
}

// StdDevSamp 样本标准差
//
// 示例：
//
//	expr := agg.StdDevSamp("$score")
func StdDevSamp(expr string) mgo.M {
	return mgo.M{"$stdDevSamp": expr}
}

// MergeObjects 合并对象
//
// 示例：
//
//	expr := agg.MergeObjects("$profile")
func MergeObjects(expr string) mgo.M {
	return mgo.M{"$mergeObjects": expr}
}

// ==================== 条件累加器 ====================

// CountIf 条件计数
//
// 示例：
//
//	expr := agg.CountIf(mgo.M{"$gt": []string{"$age", 18}})
func CountIf(condition mgo.M) mgo.M {
	return mgo.M{"$sum": mgo.M{"$cond": []interface{}{condition, 1, 0}}}
}

// SumIf 条件求和
//
// 示例：
//
//	expr := agg.SumIf(condition, "$amount")
func SumIf(condition mgo.M, expr string) mgo.M {
	return mgo.M{"$sum": mgo.M{"$cond": []interface{}{condition, expr, 0}}}
}
