package mgo

import (
	"time"
)

// ==================== 时间查询辅助方法 ====================

// WhereDateBetween 时间范围查询
//
// 示例：
//
//	users.Find().WhereDateBetween("created_at", "2024-01-01", "2024-12-31").All()
func (q *Query[T]) WhereDateBetween(field, start, end string) *Query[T] {
	if q.err != nil {
		return q
	}

	startTime, err := ParseTime(start)
	if err != nil {
		q.err = err
		return q
	}

	endTime, err := ParseTime(end)
	if err != nil {
		q.err = err
		return q
	}

	// 转换为 UTC
	startTime = startTime.In(time.Local).UTC()
	endTime = endTime.In(time.Local).UTC()

	// 设置为当天结束时间
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.UTC)

	return q.Gte(field, startTime).Lte(field, endTime)
}

// WhereDateAfter 大于指定日期
//
// 示例：
//
//	users.Find().WhereDateAfter("created_at", "2024-01-01").All()
func (q *Query[T]) WhereDateAfter(field, date string) *Query[T] {
	if q.err != nil {
		return q
	}

	t, err := ParseTime(date)
	if err != nil {
		q.err = err
		return q
	}

	t = t.In(time.Local).UTC()
	return q.Gt(field, t)
}

// WhereDateBefore 小于指定日期
//
// 示例：
//
//	users.Find().WhereDateBefore("created_at", "2024-12-31").All()
func (q *Query[T]) WhereDateBefore(field, date string) *Query[T] {
	if q.err != nil {
		return q
	}

	t, err := ParseTime(date)
	if err != nil {
		q.err = err
		return q
	}

	t = t.In(time.Local).UTC()
	// 设置为当天结束时间
	t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
	return q.Lt(field, t)
}

// ==================== 相对时间查询 ====================

// WhereToday 查询今天
//
// 示例：
//
//	users.Find().WhereToday("created_at").All()
func (q *Query[T]) WhereToday(field string) *Query[T] {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	return q.Gte(field, start).Lte(field, end)
}

// WhereYesterday 查询昨天
//
// 示例：
//
//	users.Find().WhereYesterday("created_at").All()
func (q *Query[T]) WhereYesterday(field string) *Query[T] {
	now := time.Now().In(time.Local)
	yesterday := now.AddDate(0, 0, -1)
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	return q.Gte(field, start).Lte(field, end)
}

// WhereThisWeek 查询本周
//
// 示例：
//
//	users.Find().WhereThisWeek("created_at").All()
func (q *Query[T]) WhereThisWeek(field string) *Query[T] {
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

	return q.Gte(field, start).Lte(field, end)
}

// WhereThisMonth 查询本月
//
// 示例：
//
//	users.Find().WhereThisMonth("created_at").All()
func (q *Query[T]) WhereThisMonth(field string) *Query[T] {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).UTC()

	// 下个月的第一天
	nextMonth := start.AddDate(0, 1, 0)
	// 本月最后一天
	end := nextMonth.Add(-time.Second)

	return q.Gte(field, start).Lte(field, end)
}

// WhereThisYear 查询本年
//
// 示例：
//
//	users.Find().WhereThisYear("created_at").All()
func (q *Query[T]) WhereThisYear(field string) *Query[T] {
	now := time.Now().In(time.Local)
	start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local).UTC()
	end := time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.Local).UTC()

	return q.Gte(field, start).Lte(field, end)
}

// WhereLastDays 查询最近 N 天
//
// 示例：
//
//	users.Find().WhereLastDays("created_at", 7).All()  // 最近 7 天
func (q *Query[T]) WhereLastDays(field string, days int) *Query[T] {
	now := time.Now().In(time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.Local).UTC()

	start := now.AddDate(0, 0, -days)
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local).UTC()

	return q.Gte(field, start).Lte(field, end)
}

// WhereLastHours 查询最近 N 小时
//
// 示例：
//
//	users.Find().WhereLastHours("created_at", 24).All()  // 最近 24 小时
func (q *Query[T]) WhereLastHours(field string, hours int) *Query[T] {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(hours) * time.Hour)

	return q.Gte(field, start).Lte(field, now)
}

// WhereLastMinutes 查询最近 N 分钟
//
// 示例：
//
//	users.Find().WhereLastMinutes("created_at", 30).All()  // 最近 30 分钟
func (q *Query[T]) WhereLastMinutes(field string, minutes int) *Query[T] {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(minutes) * time.Minute)

	return q.Gte(field, start).Lte(field, now)
}
