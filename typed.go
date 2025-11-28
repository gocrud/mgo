package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 泛型集合封装 ====================

// TypedCollection 泛型集合封装
//
// 提供类型安全的查询和聚合操作
// T 是文档对应的结构体类型
type TypedCollection[T any] struct {
	*Collection
}

// Model 将普通 Collection 转换为泛型 TypedCollection
//
// 示例：
//
//	userColl := mgo.Model[User](coll)
//	user, err := userColl.Query(ctx).Eq("name", "张三").First()
func Model[T any](c *Collection) *TypedCollection[T] {
	return &TypedCollection[T]{
		Collection: c,
	}
}

// Query 创建泛型查询构建器
func (tc *TypedCollection[T]) Query(ctx context.Context) *TypedQueryBuilder[T] {
	return &TypedQueryBuilder[T]{
		QueryBuilder: tc.Collection.Query(ctx),
	}
}

// ==================== 泛型查询构建器 ====================

// TypedQueryBuilder 泛型查询构建器
//
// 继承自 QueryBuilder，重写了返回结果的方法以支持泛型
type TypedQueryBuilder[T any] struct {
	*QueryBuilder
}

// Filter 设置过滤条件
func (tqb *TypedQueryBuilder[T]) Filter(filter *FilterBuilder) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Filter(filter)
	return tqb
}

// Eq 等于条件
func (tqb *TypedQueryBuilder[T]) Eq(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Eq(field, value)
	return tqb
}

// Ne 不等于条件
func (tqb *TypedQueryBuilder[T]) Ne(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Ne(field, value)
	return tqb
}

// Gt 大于条件
func (tqb *TypedQueryBuilder[T]) Gt(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Gt(field, value)
	return tqb
}

// Gte 大于等于条件
func (tqb *TypedQueryBuilder[T]) Gte(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Gte(field, value)
	return tqb
}

// Lt 小于条件
func (tqb *TypedQueryBuilder[T]) Lt(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Lt(field, value)
	return tqb
}

// Lte 小于等于条件
func (tqb *TypedQueryBuilder[T]) Lte(field string, value any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Lte(field, value)
	return tqb
}

// In IN 条件
func (tqb *TypedQueryBuilder[T]) In(field string, values ...any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.In(field, values...)
	return tqb
}

// Nin NOT IN 条件
func (tqb *TypedQueryBuilder[T]) Nin(field string, values ...any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Nin(field, values...)
	return tqb
}

// Between 范围条件
func (tqb *TypedQueryBuilder[T]) Between(field string, min, max any) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Between(field, min, max)
	return tqb
}

// Contains 包含字符串
func (tqb *TypedQueryBuilder[T]) Contains(field string, substr string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Contains(field, substr)
	return tqb
}

// StartsWith 以字符串开头
func (tqb *TypedQueryBuilder[T]) StartsWith(field string, prefix string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.StartsWith(field, prefix)
	return tqb
}

// EndsWith 以字符串结尾
func (tqb *TypedQueryBuilder[T]) EndsWith(field string, suffix string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.EndsWith(field, suffix)
	return tqb
}

// Regex 正则匹配
func (tqb *TypedQueryBuilder[T]) Regex(field string, pattern string, options ...string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Regex(field, pattern, options...)
	return tqb
}

// And AND 逻辑
func (tqb *TypedQueryBuilder[T]) And(filters ...*FilterBuilder) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.And(filters...)
	return tqb
}

// Or OR 逻辑
func (tqb *TypedQueryBuilder[T]) Or(filters ...*FilterBuilder) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Or(filters...)
	return tqb
}

// Select 选择字段
func (tqb *TypedQueryBuilder[T]) Select(fields ...string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Select(fields...)
	return tqb
}

// Omit 排除字段
func (tqb *TypedQueryBuilder[T]) Omit(fields ...string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Omit(fields...)
	return tqb
}

// OrderBy 设置排序
func (tqb *TypedQueryBuilder[T]) OrderBy(sort *Sort) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.OrderBy(sort)
	return tqb
}

// Sort 简单排序
func (tqb *TypedQueryBuilder[T]) Sort(field string, direction int) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Sort(field, direction)
	return tqb
}

// Asc 升序
func (tqb *TypedQueryBuilder[T]) Asc(fields ...string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Asc(fields...)
	return tqb
}

// Desc 降序
func (tqb *TypedQueryBuilder[T]) Desc(fields ...string) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Desc(fields...)
	return tqb
}

// Limit 限制数量
func (tqb *TypedQueryBuilder[T]) Limit(limit int64) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Limit(limit)
	return tqb
}

// Skip 跳过数量
func (tqb *TypedQueryBuilder[T]) Skip(skip int64) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Skip(skip)
	return tqb
}

// Page 分页
func (tqb *TypedQueryBuilder[T]) Page(page, pageSize int64) *TypedQueryBuilder[T] {
	tqb.QueryBuilder.Page(page, pageSize)
	return tqb
}

// WithDeleted 包含已删除
func (tqb *TypedQueryBuilder[T]) WithDeleted() *TypedQueryBuilder[T] {
	tqb.QueryBuilder.WithDeleted()
	return tqb
}

// OnlyDeleted 仅包含已删除
func (tqb *TypedQueryBuilder[T]) OnlyDeleted() *TypedQueryBuilder[T] {
	tqb.QueryBuilder.OnlyDeleted()
	return tqb
}

// ==================== 结果获取方法 (泛型增强) ====================

// One 查询单条记录并返回对象
//
// 示例：
//
//	user, err := coll.Query(ctx).Eq("_id", id).One()
func (tqb *TypedQueryBuilder[T]) One() (T, error) {
	var result T
	err := tqb.QueryBuilder.One(&result)
	return result, err
}

// All 查询多条记录并返回切片
//
// 示例：
//
//	users, err := coll.Query(ctx).Eq("status", "active").All()
func (tqb *TypedQueryBuilder[T]) All() ([]T, error) {
	var results []T
	err := tqb.QueryBuilder.All(&results)
	return results, err
}

// FindAndUpdate 查找并更新，返回更新后的对象
func (tqb *TypedQueryBuilder[T]) FindAndUpdate(update any) (T, error) {
	// UpdateBuilder 转换逻辑在 QueryBuilder.FindAndUpdate 中处理
	// 这里我们需要先调用 QueryBuilder 的通用方法，这里稍微麻烦，因为 update 参数类型可能是 *UpdateBuilder 或 map
	// 由于我们下面会优化 QueryBuilder.FindAndUpdate 接受 any，这里直接传即可

	var result T
	// 注意：此时 QueryBuilder.FindAndUpdate 签名还没改，可能会报错。
	// 我们假设接下来会修改 QueryBuilder 的签名接受 any
	err := tqb.QueryBuilder.FindAndUpdate(update, &result)
	return result, err
}

// FindAndReplace 查找并替换，返回替换后的对象
func (tqb *TypedQueryBuilder[T]) FindAndReplace(replacement any) (T, error) {
	var result T
	err := tqb.QueryBuilder.FindAndReplace(replacement, &result)
	return result, err
}

// FindAndDelete 查找并删除，返回被删除的对象
func (tqb *TypedQueryBuilder[T]) FindAndDelete() (T, error) {
	var result T
	err := tqb.QueryBuilder.FindAndDelete(&result)
	return result, err
}

// Cursor 获取游标
func (tqb *TypedQueryBuilder[T]) Cursor() (*mongo.Cursor, error) {
	return tqb.QueryBuilder.Cursor()
}
