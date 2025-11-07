package mgo

// ==================== 聚合累加器辅助函数 ====================
// 用于 $group 阶段的累加器操作

// Sum 求和累加器
//
// 示例：
//
//	M{"total": Sum(1)}          // 计数
//	M{"totalSales": Sum("$amount")}  // 求和
func Sum(expression any) M {
	return M{"$sum": expression}
}

// Avg 平均值累加器
//
// 示例：
//
//	M{"avgAge": Avg("$age")}
func Avg(expression any) M {
	return M{"$avg": expression}
}

// Max 最大值累加器
//
// 示例：
//
//	M{"maxPrice": Max("$price")}
func Max(expression any) M {
	return M{"$max": expression}
}

// Min 最小值累加器
//
// 示例：
//
//	M{"minPrice": Min("$price")}
func Min(expression any) M {
	return M{"$min": expression}
}

// First 第一个值累加器
//
// 示例：
//
//	M{"firstOrder": First("$order_date")}
func First(expression any) M {
	return M{"$first": expression}
}

// Last 最后一个值累加器
//
// 示例：
//
//	M{"lastOrder": Last("$order_date")}
func Last(expression any) M {
	return M{"$last": expression}
}

// Push 数组追加累加器
//
// 示例：
//
//	M{"orders": Push("$order_id")}
func Push(expression any) M {
	return M{"$push": expression}
}

// AddToSet 数组去重追加累加器
//
// 示例：
//
//	M{"uniqueTags": AddToSet("$tag")}
func AddToSet(expression any) M {
	return M{"$addToSet": expression}
}

// StdDevPop 总体标准差累加器
//
// 示例：
//
//	M{"stdDev": StdDevPop("$score")}
func StdDevPop(expression any) M {
	return M{"$stdDevPop": expression}
}

// StdDevSamp 样本标准差累加器
//
// 示例：
//
//	M{"stdDev": StdDevSamp("$score")}
func StdDevSamp(expression any) M {
	return M{"$stdDevSamp": expression}
}
