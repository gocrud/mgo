package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// QueryBuilder 统一的查询构建器
// 提供链式 API 用于构建和执行 MongoDB 查询
//
// 使用示例：
//
//	// 查询单条
//	var user User
//	err := coll.Query().
//	    Eq("status", "active").
//	    Gt("age", 18).
//	    One(ctx, &user)
//
//	// 查询多条
//	var users []User
//	err := coll.Query().
//	    In("city", "北京", "上海").
//	    Select("name", "email").
//	    OrderBy("created_at", -1).
//	    Limit(10).
//	    All(ctx, &users)
//
//	// 计数
//	count, err := coll.Query().
//	    Eq("status", "active").
//	    Count(ctx)
type QueryBuilder struct {
	coll           *Collection
	ctx            context.Context
	filter         *FilterBuilder
	projection     *Projection
	sort           *Sort
	limit          int64
	skip           int64
	hint           any
	batchSize      int32
	includeDeleted *bool // nil: 仅已删除, true: 包含已删除, false/unset: 排除已删除
	hardDelete     bool  // true: 强制硬删除, false: 根据配置决定
}

// newQueryBuilder 创建查询构建器（内部使用）
func newQueryBuilder(coll *Collection, ctx context.Context) *QueryBuilder {
	return &QueryBuilder{
		coll:   coll,
		ctx:    ctx,
		filter: Filter(),
	}
}

// ==================== 条件构建 ====================

// Filter 设置过滤条件（使用 Filter 构建器）
//
// 示例：
//
//	err := coll.Query().
//	    Filter(Filter().Eq("status", "active").Gt("age", 18)).
//	    All(ctx, &users)
func (qb *QueryBuilder) Filter(filter *FilterBuilder) *QueryBuilder {
	qb.filter = filter
	return qb
}

// Eq 等于条件
//
// 示例：
//
//	qb.Eq("status", "active")
func (qb *QueryBuilder) Eq(field string, value any) *QueryBuilder {
	qb.filter.Eq(field, value)
	return qb
}

// Ne 不等于条件
//
// 示例：
//
//	qb.Ne("status", "deleted")
func (qb *QueryBuilder) Ne(field string, value any) *QueryBuilder {
	qb.filter.Ne(field, value)
	return qb
}

// Gt 大于条件
//
// 示例：
//
//	qb.Gt("age", 18)
func (qb *QueryBuilder) Gt(field string, value any) *QueryBuilder {
	qb.filter.Gt(field, value)
	return qb
}

// Gte 大于等于条件
//
// 示例：
//
//	qb.Gte("score", 60)
func (qb *QueryBuilder) Gte(field string, value any) *QueryBuilder {
	qb.filter.Gte(field, value)
	return qb
}

// Lt 小于条件
//
// 示例：
//
//	qb.Lt("age", 65)
func (qb *QueryBuilder) Lt(field string, value any) *QueryBuilder {
	qb.filter.Lt(field, value)
	return qb
}

// Lte 小于等于条件
//
// 示例：
//
//	qb.Lte("price", 100)
func (qb *QueryBuilder) Lte(field string, value any) *QueryBuilder {
	qb.filter.Lte(field, value)
	return qb
}

// In IN 条件
//
// 示例：
//
//	qb.In("status", "active", "pending", "approved")
func (qb *QueryBuilder) In(field string, values ...any) *QueryBuilder {
	qb.filter.In(field, values...)
	return qb
}

// Nin NOT IN 条件
//
// 示例：
//
//	qb.Nin("status", "deleted", "archived")
func (qb *QueryBuilder) Nin(field string, values ...any) *QueryBuilder {
	qb.filter.NotIn(field, values...)
	return qb
}

// Between 范围条件（包含边界）
//
// 示例：
//
//	qb.Between("age", 18, 60)
func (qb *QueryBuilder) Between(field string, min, max any) *QueryBuilder {
	qb.filter.Between(field, min, max)
	return qb
}

// Contains 包含字符串（模糊匹配）
//
// 示例：
//
//	qb.Contains("name", "张")
func (qb *QueryBuilder) Contains(field string, substr string) *QueryBuilder {
	qb.filter.Contains(field, substr)
	return qb
}

// StartsWith 以字符串开头
//
// 示例：
//
//	qb.StartsWith("email", "admin")
func (qb *QueryBuilder) StartsWith(field string, prefix string) *QueryBuilder {
	qb.filter.StartsWith(field, prefix)
	return qb
}

// EndsWith 以字符串结尾
//
// 示例：
//
//	qb.EndsWith("email", "@example.com")
func (qb *QueryBuilder) EndsWith(field string, suffix string) *QueryBuilder {
	qb.filter.EndsWith(field, suffix)
	return qb
}

// Regex 正则表达式匹配
//
// 示例：
//
//	qb.Regex("name", "^张.*", "i")
func (qb *QueryBuilder) Regex(field string, pattern string, options ...string) *QueryBuilder {
	qb.filter.Regex(field, pattern, options...)
	return qb
}

// FieldExists 字段存在性
//
// 示例：
//
//	qb.FieldExists("email")
func (qb *QueryBuilder) FieldExists(field string) *QueryBuilder {
	qb.filter.Exists(field)
	return qb
}

// FieldNotExists 字段不存在
//
// 示例：
//
//	qb.FieldNotExists("deleted_at")
func (qb *QueryBuilder) FieldNotExists(field string) *QueryBuilder {
	qb.filter.NotExists(field)
	return qb
}

// Type 字段类型判断
//
// 示例：
//
//	qb.Type("age", "int")
func (qb *QueryBuilder) Type(field string, bsonType string) *QueryBuilder {
	qb.filter.Type(field, bsonType)
	return qb
}

// ==================== 逻辑操作 ====================

// And AND 逻辑（多个条件）
//
// 示例：
//
//	qb.And(
//	    Filter().Eq("status", "active"),
//	    Filter().Gt("age", 18),
//	)
func (qb *QueryBuilder) And(filters ...*FilterBuilder) *QueryBuilder {
	qb.filter.And(filters...)
	return qb
}

// Or OR 逻辑
//
// 示例：
//
//	qb.Or(
//	    Filter().Eq("vip", true),
//	    Filter().Gte("level", 5),
//	)
func (qb *QueryBuilder) Or(filters ...*FilterBuilder) *QueryBuilder {
	qb.filter.Or(filters...)
	return qb
}

// Nor NOR 逻辑
//
// 示例：
//
//	qb.Nor(
//	    Filter().Eq("status", "deleted"),
//	    Filter().Eq("status", "archived"),
//	)
func (qb *QueryBuilder) Nor(filters ...*FilterBuilder) *QueryBuilder {
	qb.filter.Nor(filters...)
	return qb
}

// Not NOT 逻辑（单个字段）
//
// 示例：
//
//	qb.Not("age", bson.M{"$lt": 18})
func (qb *QueryBuilder) Not(field string, condition any) *QueryBuilder {
	qb.filter.Not(field, condition)
	return qb
}

// ==================== 投影 ====================

// Select 选择要返回的字段
//
// 示例：
//
//	qb.Select("name", "email", "age")
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	if qb.projection == nil {
		qb.projection = NewProjection()
	}
	qb.projection.Include(fields...)
	return qb
}

// Omit 排除不返回的字段
//
// 示例：
//
//	qb.Omit("password", "secret")
func (qb *QueryBuilder) Omit(fields ...string) *QueryBuilder {
	if qb.projection == nil {
		qb.projection = NewProjection()
	}
	qb.projection.Exclude(fields...)
	return qb
}

// Project 使用 Projection 构建器设置投影
//
// 示例：
//
//	qb.Project(NewProjection().Include("name", "email"))
func (qb *QueryBuilder) Project(projection *Projection) *QueryBuilder {
	qb.projection = projection
	return qb
}

// ==================== 排序 ====================

// OrderBy 设置排序（使用 Sort 构建器）
//
// 示例：
//
//	qb.OrderBy(Sort().Desc("created_at").Asc("name"))
func (qb *QueryBuilder) OrderBy(sort *Sort) *QueryBuilder {
	qb.sort = sort
	return qb
}

// Sort 简单排序（字段名，方向）
//
// 示例：
//
//	qb.Sort("created_at", -1)  // 降序
//	qb.Sort("name", 1)          // 升序
func (qb *QueryBuilder) Sort(field string, direction int) *QueryBuilder {
	if qb.sort == nil {
		qb.sort = NewSort()
	}
	if direction >= 0 {
		qb.sort.Asc(field)
	} else {
		qb.sort.Desc(field)
	}
	return qb
}

// Asc 升序排序
//
// 示例：
//
//	qb.Asc("name", "age")
func (qb *QueryBuilder) Asc(fields ...string) *QueryBuilder {
	if qb.sort == nil {
		qb.sort = NewSort()
	}
	qb.sort.Asc(fields...)
	return qb
}

// Desc 降序排序
//
// 示例：
//
//	qb.Desc("created_at", "updated_at")
func (qb *QueryBuilder) Desc(fields ...string) *QueryBuilder {
	if qb.sort == nil {
		qb.sort = NewSort()
	}
	qb.sort.Desc(fields...)
	return qb
}

// ==================== 分页 ====================

// Limit 限制返回数量
//
// 示例：
//
//	qb.Limit(10)
func (qb *QueryBuilder) Limit(limit int64) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Skip 跳过记录数
//
// 示例：
//
//	qb.Skip(20)
func (qb *QueryBuilder) Skip(skip int64) *QueryBuilder {
	qb.skip = skip
	return qb
}

// Page 分页查询（页码从 1 开始）
//
// 示例：
//
//	qb.Page(2, 20)  // 第2页，每页20条
func (qb *QueryBuilder) Page(page, pageSize int64) *QueryBuilder {
	if page < 1 {
		page = 1
	}
	qb.limit = pageSize
	qb.skip = (page - 1) * pageSize
	return qb
}

// ==================== 其他选项 ====================

// Hint 设置索引提示
//
// 示例：
//
//	qb.Hint("status_1_age_1")
//	qb.Hint(bson.D{{Key: "status", Value: 1}})
func (qb *QueryBuilder) Hint(hint any) *QueryBuilder {
	qb.hint = hint
	return qb
}

// BatchSize 设置批量大小
//
// 示例：
//
//	qb.BatchSize(100)
func (qb *QueryBuilder) BatchSize(size int32) *QueryBuilder {
	qb.batchSize = size
	return qb
}

// ==================== 执行方法 ====================

// buildFindOptions 构建查询选项
func (qb *QueryBuilder) buildFindOptions() *options.FindOptionsBuilder {
	opts := options.Find()

	if qb.projection != nil {
		opts.SetProjection(qb.projection.BuildM())
	}

	if qb.sort != nil {
		opts.SetSort(qb.sort.BuildM())
	}

	if qb.limit > 0 {
		opts.SetLimit(qb.limit)
	}

	if qb.skip > 0 {
		opts.SetSkip(qb.skip)
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	if qb.batchSize > 0 {
		opts.SetBatchSize(qb.batchSize)
	}

	return opts
}

// One 查询单条记录并解码到结果
//
// 示例：
//
//	var user User
//	err := coll.Query(ctx).Eq("_id", id).One(&user)
//	if err == mongo.ErrNoDocuments {
//	    // 未找到
//	}
func (qb *QueryBuilder) One(result any) error {
	opts := options.FindOne()

	if qb.projection != nil {
		opts.SetProjection(qb.projection.BuildM())
	}

	if qb.sort != nil {
		opts.SetSort(qb.sort.BuildM())
	}

	if qb.skip > 0 {
		opts.SetSkip(qb.skip)
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.FindOne(qb.ctx, qb.buildFilterWithSoftDelete(), opts).Decode(result)
}

// All 查询多条记录并解码到结果切片
//
// 示例：
//
//	var users []User
//	err := coll.Query(ctx).Eq("status", "active").All(&users)
func (qb *QueryBuilder) All(results any) error {
	cursor, err := qb.coll.coll.Find(qb.ctx, qb.buildFilterWithSoftDelete(), qb.buildFindOptions())
	if err != nil {
		return err
	}
	defer cursor.Close(qb.ctx)

	return cursor.All(qb.ctx, results)
}

// Count 计数查询
//
// 示例：
//
//	count, err := coll.Query(ctx).Eq("status", "active").Count()
func (qb *QueryBuilder) Count() (int64, error) {
	opts := options.Count()

	if qb.limit > 0 {
		opts.SetLimit(qb.limit)
	}

	if qb.skip > 0 {
		opts.SetSkip(qb.skip)
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.CountDocuments(qb.ctx, qb.buildFilterWithSoftDelete(), opts)
}

// Exists 判断是否存在满足条件的文档
//
// 示例：
//
//	exists, err := coll.Query(ctx).Eq("email", "test@example.com").Exists()
func (qb *QueryBuilder) Exists() (bool, error) {
	count, err := qb.Limit(1).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Cursor 获取游标（用于高级场景，如流式处理）
//
// 示例：
//
//	cursor, err := coll.Query(ctx).Eq("status", "active").Cursor()
//	if err != nil {
//	    return err
//	}
//	defer cursor.Close(qb.ctx)
//
//	for cursor.Next(qb.ctx) {
//	    var user User
//	    if err := cursor.Decode(&user); err != nil {
//	        return err
//	    }
//	    // 处理 user
//	}
func (qb *QueryBuilder) Cursor() (*mongo.Cursor, error) {
	return qb.coll.coll.Find(qb.ctx, qb.buildFilterWithSoftDelete(), qb.buildFindOptions())
}

// ==================== 查询并修改 ====================

// FindAndUpdate 查找并更新文档
//
// 示例：
//
//	var updatedUser User
//	err := coll.Query(ctx).
//	    Eq("_id", id).
//	    FindAndUpdate(Update().Set("status", "inactive"), &updatedUser)
func (qb *QueryBuilder) FindAndUpdate(update *UpdateBuilder, result any) error {
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	if qb.projection != nil {
		opts.SetProjection(qb.projection.BuildM())
	}

	if qb.sort != nil {
		opts.SetSort(qb.sort.BuildM())
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.FindOneAndUpdate(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		update.Build(),
		opts,
	).Decode(result)
}

// FindAndReplace 查找并替换文档
//
// 示例：
//
//	var replacedUser User
//	err := coll.Query(ctx).
//	    Eq("_id", id).
//	    FindAndReplace(newUser, &replacedUser)
func (qb *QueryBuilder) FindAndReplace(replacement any, result any) error {
	opts := options.FindOneAndReplace().
		SetReturnDocument(options.After)

	if qb.projection != nil {
		opts.SetProjection(qb.projection.BuildM())
	}

	if qb.sort != nil {
		opts.SetSort(qb.sort.BuildM())
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.FindOneAndReplace(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		replacement,
		opts,
	).Decode(result)
}

// FindAndDelete 查找并删除文档
//
// 行为取决于是否启用软删除：
//   - 未启用软删除：执行硬删除（物理删除）
//   - 启用软删除：执行软删除（设置 deleted_at 字段）
//   - 启用软删除 + WithHardDelete()：强制硬删除
//
// 示例：
//
//	var deletedUser User
//	err := coll.Query(ctx).
//	    Eq("_id", id).
//	    FindAndDelete(&deletedUser)
//
//	// 强制硬删除
//	err := coll.Query(ctx).
//	    Eq("_id", id).
//	    WithHardDelete().
//	    FindAndDelete(&deletedUser)
func (qb *QueryBuilder) FindAndDelete(result any) error {
	// 如果未启用软删除或强制硬删除，执行硬删除
	if !qb.coll.softDelete.Enabled || qb.hardDelete {
		opts := options.FindOneAndDelete()

		if qb.projection != nil {
			opts.SetProjection(qb.projection.BuildM())
		}

		if qb.sort != nil {
			opts.SetSort(qb.sort.BuildM())
		}

		if qb.hint != nil {
			opts.SetHint(qb.hint)
		}

		return qb.coll.coll.FindOneAndDelete(
			qb.ctx,
			qb.buildFilterWithSoftDelete(),
			opts,
		).Decode(result)
	}

	// 启用软删除时，使用 FindOneAndUpdate 来软删除
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.Before) // 返回更新前的文档

	if qb.projection != nil {
		opts.SetProjection(qb.projection.BuildM())
	}

	if qb.sort != nil {
		opts.SetSort(qb.sort.BuildM())
	}

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.FindOneAndUpdate(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		qb.coll.buildSoftDeleteUpdate(),
		opts,
	).Decode(result)
}

// ==================== 更新和删除 ====================

// UpdateOne 更新单条文档
//
// 示例：
//
//	result, err := coll.Query(ctx).
//	    Eq("_id", id).
//	    UpdateOne(Update().Set("status", "inactive"))
func (qb *QueryBuilder) UpdateOne(update *UpdateBuilder) (*mongo.UpdateResult, error) {
	opts := options.UpdateOne()

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.UpdateOne(qb.ctx, qb.buildFilterWithSoftDelete(), update.Build(), opts)
}

// UpdateMany 更新多条文档
//
// 示例：
//
//	result, err := coll.Query(ctx).
//	    Eq("status", "pending").
//	    UpdateMany(Update().Set("status", "processed"))
func (qb *QueryBuilder) UpdateMany(update *UpdateBuilder) (*mongo.UpdateResult, error) {
	opts := options.UpdateMany()

	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.UpdateMany(qb.ctx, qb.buildFilterWithSoftDelete(), update.Build(), opts)
}

// DeleteOne 删除单条文档
//
// 行为取决于是否启用软删除：
//   - 未启用软删除：执行硬删除（物理删除）
//   - 启用软删除：执行软删除（设置 deleted_at 字段）
//   - 启用软删除 + WithHardDelete()：强制硬删除
//
// 示例：
//
//	// 未启用软删除时
//	result, err := coll.Query(ctx).Eq("_id", id).DeleteOne()  // 硬删除
//
//	// 启用软删除时
//	result, err := coll.Query(ctx).Eq("_id", id).DeleteOne()  // 软删除
//
//	// 强制硬删除
//	result, err := coll.Query(ctx).Eq("_id", id).WithHardDelete().DeleteOne()
func (qb *QueryBuilder) DeleteOne() (*mongo.DeleteResult, error) {
	// 如果未启用软删除或强制硬删除，执行硬删除
	if !qb.coll.softDelete.Enabled || qb.hardDelete {
		opts := options.DeleteOne()
		if qb.hint != nil {
			opts.SetHint(qb.hint)
		}
		return qb.coll.coll.DeleteOne(qb.ctx, qb.buildFilterWithSoftDelete(), opts)
	}

	// 启用软删除时，执行软删除（更新 deleted_at 字段）
	opts := options.UpdateOne()
	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	updateResult, err := qb.coll.coll.UpdateOne(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		qb.coll.buildSoftDeleteUpdate(),
		opts,
	)

	// 将 UpdateResult 转换为 DeleteResult
	deleteResult := &mongo.DeleteResult{
		DeletedCount: updateResult.ModifiedCount,
	}
	return deleteResult, err
}

// DeleteMany 删除多条文档
//
// 行为取决于是否启用软删除：
//   - 未启用软删除：执行硬删除（物理删除）
//   - 启用软删除：执行软删除（设置 deleted_at 字段）
//   - 启用软删除 + WithHardDelete()：强制硬删除
//
// 示例：
//
//	// 未启用软删除时
//	result, err := coll.Query(ctx).Eq("status", "expired").DeleteMany()  // 硬删除
//
//	// 启用软删除时
//	result, err := coll.Query(ctx).Eq("status", "expired").DeleteMany()  // 软删除
//
//	// 强制硬删除
//	result, err := coll.Query(ctx).OnlyDeleted().WithHardDelete().DeleteMany()
func (qb *QueryBuilder) DeleteMany() (*mongo.DeleteResult, error) {
	// 如果未启用软删除或强制硬删除，执行硬删除
	if !qb.coll.softDelete.Enabled || qb.hardDelete {
		opts := options.DeleteMany()
		if qb.hint != nil {
			opts.SetHint(qb.hint)
		}
		return qb.coll.coll.DeleteMany(qb.ctx, qb.buildFilterWithSoftDelete(), opts)
	}

	// 启用软删除时，执行软删除（更新 deleted_at 字段）
	opts := options.UpdateMany()
	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	updateResult, err := qb.coll.coll.UpdateMany(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		qb.coll.buildSoftDeleteUpdate(),
		opts,
	)

	// 将 UpdateResult 转换为 DeleteResult
	deleteResult := &mongo.DeleteResult{
		DeletedCount: updateResult.ModifiedCount,
	}
	return deleteResult, err
}

// ==================== 软删除相关方法 ====================

// WithDeleted 包含已软删除的文档
//
// 仅在启用软删除时有效
//
// 示例：
//
//	// 查询包含已删除的文档
//	err := coll.Query(ctx).Eq("status", "active").WithDeleted().All(&users)
func (qb *QueryBuilder) WithDeleted() *QueryBuilder {
	includeDeleted := true
	qb.includeDeleted = &includeDeleted
	return qb
}

// OnlyDeleted 仅查询已软删除的文档
//
// 仅在启用软删除时有效
//
// 示例：
//
//	// 仅查询已删除的文档
//	err := coll.Query(ctx).Eq("status", "active").OnlyDeleted().All(&users)
//
//	// 清理已删除的数据
//	result, err := coll.Query(ctx).OnlyDeleted().WithHardDelete().DeleteMany()
func (qb *QueryBuilder) OnlyDeleted() *QueryBuilder {
	qb.includeDeleted = nil
	return qb
}

// WithHardDelete 强制执行硬删除（永久删除）
//
// 在启用软删除时，使用此方法可以强制执行真正的删除操作
//
// 示例：
//
//	// 永久删除文档
//	result, err := coll.Query(ctx).Eq("_id", id).WithHardDelete().DeleteOne()
//
//	// 清理已软删除的数据
//	result, err := coll.Query(ctx).OnlyDeleted().WithHardDelete().DeleteMany()
func (qb *QueryBuilder) WithHardDelete() *QueryBuilder {
	qb.hardDelete = true
	return qb
}

// Restore 恢复已软删除的文档（移除 deleted_at 字段）
//
// 仅在启用软删除时有效，否则返回错误
//
// 示例：
//
//	// 恢复单条
//	result, err := coll.Query(ctx).Eq("_id", id).Restore()
//
//	// 批量恢复
//	result, err := coll.Query(ctx).In("_id", ids...).Restore()
func (qb *QueryBuilder) Restore() (*mongo.UpdateResult, error) {
	if !qb.coll.softDelete.Enabled {
		return nil, ErrSoftDeleteNotEnabled
	}

	// 恢复操作默认只针对已删除的文档
	if qb.includeDeleted == nil || *qb.includeDeleted {
		qb.OnlyDeleted()
	}

	opts := options.UpdateMany()
	if qb.hint != nil {
		opts.SetHint(qb.hint)
	}

	return qb.coll.coll.UpdateMany(
		qb.ctx,
		qb.buildFilterWithSoftDelete(),
		qb.coll.buildRestoreUpdate(),
		opts,
	)
}

// buildFilterWithSoftDelete 构建包含软删除过滤的完整过滤器
func (qb *QueryBuilder) buildFilterWithSoftDelete() bson.M {
	baseFilter := qb.filter.BuildM()
	softDeleteFilter := qb.coll.buildSoftDeleteFilter(qb.includeDeleted)

	if len(softDeleteFilter) == 0 {
		return baseFilter
	}

	// 合并过滤条件
	result := bson.M{}
	for k, v := range baseFilter {
		result[k] = v
	}
	for _, elem := range softDeleteFilter {
		result[elem.Key] = elem.Value
	}

	return result
}
