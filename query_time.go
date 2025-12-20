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

// WhereDateAfter 大于指定日期
//
// 示例：
//
//	users.Find().WhereDateAfter("created_at", "2024-01-01").All()
func (q *Query[T]) WhereDateAfter(field, date string) *Query[T] {
	t, err := ParseTime(date)
	if err != nil {
		return q
	}

	t = t.In(time.Local).UTC()
	q.filter[field] = M{"$gt": t}
	return q
}

// WhereDateBefore 小于指定日期
//
// 示例：
//
//	users.Find().WhereDateBefore("created_at", "2024-12-31").All()
func (q *Query[T]) WhereDateBefore(field, date string) *Query[T] {
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
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

	q.filter[field] = M{
		"$gte": start,
		"$lte": end,
	}
	return q
}

// WhereLastHours 查询最近 N 小时
//
// 示例：
//
//	users.Find().WhereLastHours("created_at", 24).All()  // 最近 24 小时
func (q *Query[T]) WhereLastHours(field string, hours int) *Query[T] {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(hours) * time.Hour)

	q.filter[field] = M{
		"$gte": start,
		"$lte": now,
	}
	return q
}

// WhereLastMinutes 查询最近 N 分钟
//
// 示例：
//
//	users.Find().WhereLastMinutes("created_at", 30).All()  // 最近 30 分钟
func (q *Query[T]) WhereLastMinutes(field string, minutes int) *Query[T] {
	now := time.Now().UTC()
	start := now.Add(-time.Duration(minutes) * time.Minute)

	q.filter[field] = M{
		"$gte": start,
		"$lte": now,
	}
	return q
}

// ==================== 时间范围构建器 ====================

// DateRange 时间范围构建器
type DateRange[T any] struct {
	query *Query[T]
	field string
	start *time.Time
	end   *time.Time
}

// WhereDate 创建时间范围构建器
//
// 示例：
//
//	users.Find().WhereDate("created_at").
//	    After("2024-01-01").
//	    Before("2024-12-31")
func (q *Query[T]) WhereDate(field string) *DateRange[T] {
	return &DateRange[T]{
		query: q,
		field: field,
	}
}

// After 设置开始时间
func (dr *DateRange[T]) After(date string) *DateRange[T] {
	t, err := ParseTime(date)
	if err == nil {
		t = t.In(time.Local).UTC()
		dr.start = &t
	}
	return dr
}

// Before 设置结束时间
func (dr *DateRange[T]) Before(date string) *DateRange[T] {
	t, err := ParseTime(date)
	if err == nil {
		t = t.In(time.Local).UTC()
		// 设置为当天结束时间
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
		dr.end = &t
	}
	return dr
}

// Since 最近一段时间
func (dr *DateRange[T]) Since(duration time.Duration) *DateRange[T] {
	now := time.Now().UTC()
	start := now.Add(-duration)
	dr.start = &start
	dr.end = &now
	return dr
}

// Build 构建查询（内部使用）
func (dr *DateRange[T]) build() {
	filter := M{}
	if dr.start != nil {
		filter["$gte"] = *dr.start
	}
	if dr.end != nil {
		filter["$lte"] = *dr.end
	}

	if len(filter) > 0 {
		dr.query.filter[dr.field] = filter
	}
}
