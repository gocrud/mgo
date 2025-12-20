package mgo

import "time"

// ==================== 删除操作 ====================

// Delete 删除单条文档
//
// 如果启用了软删除，会设置 deleted_at 字段
//
// 示例：
//
//	err := users.Find().ID(id).Delete()
func (q *Query[T]) Delete() error {
	// 如果启用了软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		return q.softDelete()
	}

	// 物理删除
	return q.forceDelete()
}

// DeleteMany 批量删除文档
//
// 如果启用了软删除，会设置 deleted_at 字段
//
// 示例：
//
//	n, err := users.Find().
//	    Where("status", "expired").
//	    DeleteMany()
func (q *Query[T]) DeleteMany() (int64, error) {
	// 如果启用了软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		return q.softDeleteMany()
	}

	// 物理删除
	return q.forceDeleteMany()
}

// ForceDelete 物理删除（忽略软删除设置）
//
// 示例：
//
//	err := users.Find().ID(id).ForceDelete()
func (q *Query[T]) ForceDelete() error {
	return q.forceDelete()
}

// ForceDeleteMany 批量物理删除
//
// 示例：
//
//	n, err := users.Find().
//	    Where("status", "expired").
//	    ForceDeleteMany()
func (q *Query[T]) ForceDeleteMany() (int64, error) {
	return q.forceDeleteMany()
}

// ==================== 软删除实现 ====================

// softDelete 软删除单条
func (q *Query[T]) softDelete() error {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return q.forceDelete()
	}

	ctx := q.Context()
	filter := q.buildFilter()

	// 移除软删除过滤（不能包含 deleted_at 条件，否则已删除记录无法被软删除）
	// 但这里我们要删除的是未删除的记录，所以保留软删除过滤是正确的

	// 构建更新文档
	update := M{
		"$set": M{
			q.coll.opts.SoftDelete.Field: time.Now().UTC(),
		},
	}

	// 应用更新时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
	}

	_, err := q.coll.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return WrapError(err, "failed to soft delete")
	}

	return nil
}

// softDeleteMany 批量软删除
func (q *Query[T]) softDeleteMany() (int64, error) {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return q.forceDeleteMany()
	}

	ctx := q.Context()
	filter := q.buildFilter()

	// 构建更新文档
	update := M{
		"$set": M{
			q.coll.opts.SoftDelete.Field: time.Now().UTC(),
		},
	}

	// 应用更新时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
	}

	result, err := q.coll.coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to soft delete many")
	}

	return result.ModifiedCount, nil
}

// ==================== 物理删除实现 ====================

// forceDelete 物理删除单条
func (q *Query[T]) forceDelete() error {
	ctx := q.Context()
	filter := q.buildFilter()

	// 移除软删除过滤（允许删除已软删除的记录）
	if q.coll.opts.SoftDelete != nil {
		delete(filter, q.coll.opts.SoftDelete.Field)
	}

	_, err := q.coll.coll.DeleteOne(ctx, filter)
	if err != nil {
		return WrapError(err, "failed to delete")
	}

	return nil
}

// forceDeleteMany 批量物理删除
func (q *Query[T]) forceDeleteMany() (int64, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	// 移除软删除过滤（允许删除已软删除的记录）
	if q.coll.opts.SoftDelete != nil {
		delete(filter, q.coll.opts.SoftDelete.Field)
	}

	result, err := q.coll.coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to delete many")
	}

	return result.DeletedCount, nil
}

// ==================== 恢复软删除 ====================

// Restore 恢复软删除的记录
//
// 示例：
//
//	err := users.Find().ID(id).
//	    WithTrashed().
//	    Restore()
func (q *Query[T]) Restore() error {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return ErrInvalidOperation
	}

	ctx := q.Context()

	// 构建过滤条件：只恢复已删除的记录
	filter := q.buildFilter()
	filter[q.coll.opts.SoftDelete.Field] = M{"$ne": nil}

	// 构建更新文档
	update := M{
		"$set": M{
			q.coll.opts.SoftDelete.Field: nil,
		},
	}

	// 应用更新时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
	}

	_, err := q.coll.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return WrapError(err, "failed to restore")
	}

	return nil
}

// RestoreMany 批量恢复软删除的记录
//
// 示例：
//
//	n, err := users.Find().
//	    Where("status", "expired").
//	    WithTrashed().
//	    RestoreMany()
func (q *Query[T]) RestoreMany() (int64, error) {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return 0, ErrInvalidOperation
	}

	ctx := q.Context()

	// 构建过滤条件：只恢复已删除的记录
	filter := q.buildFilter()
	filter[q.coll.opts.SoftDelete.Field] = M{"$ne": nil}

	// 构建更新文档
	update := M{
		"$set": M{
			q.coll.opts.SoftDelete.Field: nil,
		},
	}

	// 应用更新时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
	}

	result, err := q.coll.coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to restore many")
	}

	return result.ModifiedCount, nil
}

// ==================== DeleteAndReturn 原子操作 ====================

// DeleteAndReturn 删除并返回被删除的文档
//
// 示例：
//
//	user, err := users.Find().ID(id).DeleteAndReturn()
func (q *Query[T]) DeleteAndReturn() (*T, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	var result T

	// 如果启用了软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		// 软删除并返回
		update := M{
			"$set": M{
				q.coll.opts.SoftDelete.Field: time.Now().UTC(),
			},
		}

		if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
			update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
		}

		err := q.coll.coll.FindOneAndUpdate(ctx, filter, update).Decode(&result)
		if err != nil {
			return nil, WrapError(err, "failed to soft delete and return")
		}
	} else {
		// 物理删除并返回
		err := q.coll.coll.FindOneAndDelete(ctx, filter).Decode(&result)
		if err != nil {
			return nil, WrapError(err, "failed to delete and return")
		}
	}

	return &result, nil
}

// ==================== Truncate 清空集合 ====================

// Truncate 清空集合（删除所有文档）
//
// 注意：这是物理删除，不受软删除影响
//
// 示例：
//
//	err := users.Truncate()
func (c *TypedCollection[T]) Truncate() error {
	ctx := c.Context()
	_, err := c.coll.DeleteMany(ctx, M{})
	if err != nil {
		return WrapError(err, "failed to truncate collection")
	}
	return nil
}
