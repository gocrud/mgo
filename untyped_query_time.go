package mgo

import (
	"time"
)

// ==================== 时间查询辅助方法 (UntypedQuery) ====================

// WhereDateBetween 时间范围查询
//
// 示例:
//
//	coll.Query().WhereDateBetween("created_at", "2024-01-01", "2024-12-31").All(&users)
func (q *UntypedQuery) WhereDateBetween(field, start, end string) *UntypedQuery {
	startTime, err := ParseTime(start)
	if err != nil {
		return q
	}

	endTime, err := ParseTime(end)
	if err != nil {
		return q
	}

	// 转换为 UTC
	startTime = startTime.In(time.Local).UTC()
	endTime = endTime.In(time.Local).UTC()

	// 设置为当天结束时间
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

	q.filter[field] = M{
		"$gte": startTime,
		"$lte": endTime,
	}

	return q
}

// WhereDateBetweenIf 条件性时间范围查询
//
// 示例:
//
//	coll.Query().WhereDateBetweenIf(hasDateRange, "created_at", startDate, endDate).All(&users)
func (q *UntypedQuery) WhereDateBetweenIf(condition bool, field, start, end string) *UntypedQuery {
	if condition {
		return q.WhereDateBetween(field, start, end)
	}
	return q
}

// WhereDateAfter 大于指定日期
//
// 示例:
//
//	coll.Query().WhereDateAfter("created_at", "2024-01-01").All(&users)
func (q *UntypedQuery) WhereDateAfter(field, date string) *UntypedQuery {
	t, err := ParseTime(date)
	if err != nil {
		return q
	}

	t = t.In(time.Local).UTC()
	q.filter[field] = M{"$gt": t}
	return q
}

// WhereDateAfterIf 条件性大于指定日期
func (q *UntypedQuery) WhereDateAfterIf(condition bool, field, date string) *UntypedQuery {
	if condition {
		return q.WhereDateAfter(field, date)
	}
	return q
}

// WhereDateBefore 小于指定日期
//
// 示例:
//
//	coll.Query().WhereDateBefore("created_at", "2024-12-31").All(&users)
func (q *UntypedQuery) WhereDateBefore(field, date string) *UntypedQuery {
	t, err := ParseTime(date)
	if err != nil {
		return q
	}

	t = t.In(time.Local).UTC()
	// 设置为当天结束时间
	t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
	q.filter[field] = M{"$lt": t}
	return q
}

// WhereDateBeforeIf 条件性小于指定日期
func (q *UntypedQuery) WhereDateBeforeIf(condition bool, field, date string) *UntypedQuery {
	if condition {
		return q.WhereDateBefore(field, date)
	}
	return q
}

// ==================== 相对时间查询 ====================

// WhereToday 查询今天
//
// 示例:
//
//	coll.Query().WhereToday("created_at").All(&users)
func (q *UntypedQuery) WhereToday(field string) *UntypedQuery {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereTodayIf 条件性查询今天
func (q *UntypedQuery) WhereTodayIf(condition bool, field string) *UntypedQuery {
	if condition {
		return q.WhereToday(field)
	}
	return q
}

// WhereYesterday 查询昨天
//
// 示例:
//
//	coll.Query().WhereYesterday("created_at").All(&users)
func (q *UntypedQuery) WhereYesterday(field string) *UntypedQuery {
	now := time.Now().In(time.Local)
	yesterday := now.AddDate(0, 0, -1)
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereYesterdayIf 条件性查询昨天
func (q *UntypedQuery) WhereYesterdayIf(condition bool, field string) *UntypedQuery {
	if condition {
		return q.WhereYesterday(field)
	}
	return q
}

// WhereThisWeek 查询本周
//
// 示例:
//
//	coll.Query().WhereThisWeek("created_at").All(&users)
func (q *UntypedQuery) WhereThisWeek(field string) *UntypedQuery {
	now := time.Now().In(time.Local)

	// 计算本周开始（周一）
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := now.AddDate(0, 0, -(weekday - 1))
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local).UTC()

	// 计算本周结束（周日）
	end := start.AddDate(0, 0, 6)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, time.UTC)

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereThisWeekIf 条件性查询本周
func (q *UntypedQuery) WhereThisWeekIf(condition bool, field string) *UntypedQuery {
	if condition {
		return q.WhereThisWeek(field)
	}
	return q
}

// WhereThisMonth 查询本月
//
// 示例:
//
//	coll.Query().WhereThisMonth("created_at").All(&users)
func (q *UntypedQuery) WhereThisMonth(field string) *UntypedQuery {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).UTC()

	// 下个月的第一天
	nextMonth := start.AddDate(0, 1, 0)
	// 本月最后一天
	end := nextMonth.Add(-time.Second)

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereThisMonthIf 条件性查询本月
func (q *UntypedQuery) WhereThisMonthIf(condition bool, field string) *UntypedQuery {
	if condition {
		return q.WhereThisMonth(field)
	}
	return q
}

// WhereThisYear 查询本年
//
// 示例:
//
//	coll.Query().WhereThisYear("created_at").All(&users)
func (q *UntypedQuery) WhereThisYear(field string) *UntypedQuery {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.Local).UTC()

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereThisYearIf 条件性查询本年
func (q *UntypedQuery) WhereThisYearIf(condition bool, field string) *UntypedQuery {
	if condition {
		return q.WhereThisYear(field)
	}
	return q
}

// WhereLastDays 查询最近 N 天
//
// 示例:
//
//	coll.Query().WhereLastDays("created_at", 7).All(&users)  // 最近 7 天
func (q *UntypedQuery) WhereLastDays(field string, days int) *UntypedQuery {
	now := time.Now().In(time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	start := now.AddDate(0, 0, -days)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local).UTC()

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereLastDaysIf 条件性查询最近 N 天
func (q *UntypedQuery) WhereLastDaysIf(condition bool, field string, days int) *UntypedQuery {
	if condition {
		return q.WhereLastDays(field, days)
	}
	return q
}

// WhereLastHours 查询最近 N 小时
//
// 示例:
//
//	coll.Query().WhereLastHours("created_at", 24).All(&users)  // 最近 24 小时
func (q *UntypedQuery) WhereLastHours(field string, hours int) *UntypedQuery {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(hours) * time.Hour)

	q.filter[field] = M{
		"$gte": start,
		"$lte": now,
	}
	return q
}

// WhereLastHoursIf 条件性查询最近 N 小时
func (q *UntypedQuery) WhereLastHoursIf(condition bool, field string, hours int) *UntypedQuery {
	if condition {
		return q.WhereLastHours(field, hours)
	}
	return q
}

// WhereLastMinutes 查询最近 N 分钟
//
// 示例:
//
//	coll.Query().WhereLastMinutes("created_at", 30).All(&users)  // 最近 30 分钟
func (q *UntypedQuery) WhereLastMinutes(field string, minutes int) *UntypedQuery {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(minutes) * time.Minute)

	q.filter[field] = M{
		"$gte": start,
		"$lte": now,
	}
	return q
}

// WhereLastMinutesIf 条件性查询最近 N 分钟
func (q *UntypedQuery) WhereLastMinutesIf(condition bool, field string, minutes int) *UntypedQuery {
	if condition {
		return q.WhereLastMinutes(field, minutes)
	}
	return q
}
