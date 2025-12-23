package mgo

import (
	"time"
)

// ==================== 删除操作 ====================

// Delete 删除单条文档
func (q *Query[T]) Delete() error {
	// 如果启用了软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		return q.softDelete()
	}

	// 物理删除
	return q.forceDelete()
}

// DeleteMany 批量删除文档
func (q *Query[T]) DeleteMany() (int64, error) {
	// 如果启用了软删除
	if q.coll.opts.SoftDelete != nil && q.coll.opts.SoftDelete.Enabled {
		return q.softDeleteMany()
	}

	// 物理删除
	return q.forceDeleteMany()
}

// ForceDelete 物理删除（忽略软删除设置）
func (q *Query[T]) ForceDelete() error {
	return q.forceDelete()
}

// ForceDeleteMany 批量物理删除
func (q *Query[T]) ForceDeleteMany() (int64, error) {
	return q.forceDeleteMany()
}

// ==================== 软删除实现 ====================

// softDelete 软删除单条
func (q *Query[T]) softDelete() error {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return q.forceDelete()
	}

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

	_, err := q.coll.coll.UpdateOne(q.ctx, filter, update)
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

	result, err := q.coll.coll.UpdateMany(q.ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to soft delete many")
	}

	return result.ModifiedCount, nil
}

// ==================== 物理删除实现 ====================

// forceDelete 物理删除单条
func (q *Query[T]) forceDelete() error {
	filter := q.buildFilter()

	// 移除软删除过滤（允许删除已软删除的记录）
	if q.coll.opts.SoftDelete != nil {
		delete(filter, q.coll.opts.SoftDelete.Field)
	}

	_, err := q.coll.coll.DeleteOne(q.ctx, filter)
	if err != nil {
		return WrapError(err, "failed to delete")
	}

	return nil
}

// forceDeleteMany 批量物理删除
func (q *Query[T]) forceDeleteMany() (int64, error) {
	filter := q.buildFilter()

	// 移除软删除过滤（允许删除已软删除的记录）
	if q.coll.opts.SoftDelete != nil {
		delete(filter, q.coll.opts.SoftDelete.Field)
	}

	result, err := q.coll.coll.DeleteMany(q.ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to delete many")
	}

	return result.DeletedCount, nil
}

// ==================== 恢复软删除 ====================

// Restore 恢复软删除的记录
func (q *Query[T]) Restore() error {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return ErrInvalidOperation
	}

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

	_, err := q.coll.coll.UpdateOne(q.ctx, filter, update)
	if err != nil {
		return WrapError(err, "failed to restore")
	}

	return nil
}

// RestoreMany 批量恢复软删除的记录
func (q *Query[T]) RestoreMany() (int64, error) {
	if q.coll.opts.SoftDelete == nil || !q.coll.opts.SoftDelete.Enabled {
		return 0, ErrInvalidOperation
	}

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

	result, err := q.coll.coll.UpdateMany(q.ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to restore many")
	}

	return result.ModifiedCount, nil
}

// ==================== DeleteAndReturn 原子操作 ====================

// DeleteAndReturn 删除并返回被删除的文档
func (q *Query[T]) DeleteAndReturn() (*T, error) {
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

		err := q.coll.coll.FindOneAndUpdate(q.ctx, filter, update).Decode(&result)
		if err != nil {
			return nil, WrapError(err, "failed to soft delete and return")
		}
	} else {
		// 物理删除并返回
		err := q.coll.coll.FindOneAndDelete(q.ctx, filter).Decode(&result)
		if err != nil {
			return nil, WrapError(err, "failed to delete and return")
		}
	}

	return &result, nil
}
