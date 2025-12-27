package mgo

import (
	"time"
)

// ==================== 删除构建器 ====================

// DeleteBuilder 删除构建器
type DeleteBuilder[T any] struct {
	coll    *Collection[T]
	filter  any
	isForce bool
}

// newDeleteBuilder 创建删除构建器
func newDeleteBuilder[T any](coll *Collection[T]) *DeleteBuilder[T] {
	return &DeleteBuilder[T]{
		coll:   coll,
		filter: M{},
	}
}

// Where 设置删除条件
func (b *DeleteBuilder[T]) Where(filter any) *DeleteBuilder[T] {
	b.filter = filter
	return b
}

// Force 强制物理删除
func (b *DeleteBuilder[T]) Force() *DeleteBuilder[T] {
	b.isForce = true
	return b
}

// Exec 执行删除 (默认删除单条)
func (b *DeleteBuilder[T]) Exec() error {
	// 软删除逻辑
	if !b.isForce && b.coll.opts.SoftDelete != nil && b.coll.opts.SoftDelete.Enabled {
		update := M{
			"$set": M{
				b.coll.opts.SoftDelete.Field: time.Now().UTC(),
			},
		}
		// 应用更新时间戳
		if b.coll.opts.Timestamps != nil && b.coll.opts.Timestamps.Enabled {
			update["$set"].(M)[b.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
		}
		_, err := b.coll.coll.UpdateOne(b.coll.ctx, b.filter, update)
		return err
	}

	// 物理删除
	_, err := b.coll.coll.DeleteOne(b.coll.ctx, b.filter)
	return err
}

// DeleteMany 执行批量删除
func (b *DeleteBuilder[T]) DeleteMany() (int64, error) {
	// 软删除逻辑
	if !b.isForce && b.coll.opts.SoftDelete != nil && b.coll.opts.SoftDelete.Enabled {
		update := M{
			"$set": M{
				b.coll.opts.SoftDelete.Field: time.Now().UTC(),
			},
		}
		if b.coll.opts.Timestamps != nil && b.coll.opts.Timestamps.Enabled {
			update["$set"].(M)[b.coll.opts.Timestamps.UpdatedField] = time.Now().UTC()
		}
		res, err := b.coll.coll.UpdateMany(b.coll.ctx, b.filter, update)
		if err != nil {
			return 0, err
		}
		return res.ModifiedCount, nil
	}

	// 物理删除
	res, err := b.coll.coll.DeleteMany(b.coll.ctx, b.filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
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
