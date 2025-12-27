package mgo

// ==================== 更新操作 ====================

// UpdateBuilder 更新构建器
type UpdateBuilder[T any] struct {
	coll    *Collection[T]
	filter  any
	updates M
}

// newUpdateBuilder 创建更新构建器
func newUpdateBuilder[T any](coll *Collection[T]) *UpdateBuilder[T] {
	return &UpdateBuilder[T]{
		coll:    coll,
		filter:  M{},
		updates: M{},
	}
}

// Where 设置更新条件
func (b *UpdateBuilder[T]) Where(filter any) *UpdateBuilder[T] {
	b.filter = filter
	return b
}

// Set 设置字段值
func (b *UpdateBuilder[T]) Set(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$set"]; !ok {
		b.updates["$set"] = M{}
	}
	b.updates["$set"].(M)[field] = NormalizeValue(value)
	return b
}

// Inc 增加字段值
func (b *UpdateBuilder[T]) Inc(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$inc"]; !ok {
		b.updates["$inc"] = M{}
	}
	b.updates["$inc"].(M)[field] = value
	return b
}

// Mul 乘以字段值
func (b *UpdateBuilder[T]) Mul(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$mul"]; !ok {
		b.updates["$mul"] = M{}
	}
	b.updates["$mul"].(M)[field] = value
	return b
}

// SetMin 设置字段最小值
func (b *UpdateBuilder[T]) SetMin(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$min"]; !ok {
		b.updates["$min"] = M{}
	}
	b.updates["$min"].(M)[field] = value
	return b
}

// SetMax 设置字段最大值
func (b *UpdateBuilder[T]) SetMax(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$max"]; !ok {
		b.updates["$max"] = M{}
	}
	b.updates["$max"].(M)[field] = value
	return b
}

// Unset 删除字段
func (b *UpdateBuilder[T]) Unset(fields ...string) *UpdateBuilder[T] {
	if _, ok := b.updates["$unset"]; !ok {
		b.updates["$unset"] = M{}
	}
	for _, field := range fields {
		b.updates["$unset"].(M)[field] = ""
	}
	return b
}

// Rename 重命名字段
func (b *UpdateBuilder[T]) Rename(oldField, newField string) *UpdateBuilder[T] {
	if _, ok := b.updates["$rename"]; !ok {
		b.updates["$rename"] = M{}
	}
	b.updates["$rename"].(M)[oldField] = newField
	return b
}

// Push 向数组添加元素
func (b *UpdateBuilder[T]) Push(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$push"]; !ok {
		b.updates["$push"] = M{}
	}
	b.updates["$push"].(M)[field] = value
	return b
}

// PushAll 向数组添加多个元素
func (b *UpdateBuilder[T]) PushAll(field string, values []interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$push"]; !ok {
		b.updates["$push"] = M{}
	}
	b.updates["$push"].(M)[field] = M{"$each": values}
	return b
}

// Pull 从数组删除元素
func (b *UpdateBuilder[T]) Pull(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$pull"]; !ok {
		b.updates["$pull"] = M{}
	}
	b.updates["$pull"].(M)[field] = value
	return b
}

// PullAll 从数组删除多个元素
func (b *UpdateBuilder[T]) PullAll(field string, values []interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$pullAll"]; !ok {
		b.updates["$pullAll"] = M{}
	}
	b.updates["$pullAll"].(M)[field] = values
	return b
}

// AddToSet 向数组添加元素（去重）
func (b *UpdateBuilder[T]) AddToSet(field string, value interface{}) *UpdateBuilder[T] {
	if _, ok := b.updates["$addToSet"]; !ok {
		b.updates["$addToSet"] = M{}
	}
	b.updates["$addToSet"].(M)[field] = value
	return b
}

// Pop 从数组移除第一个或最后一个元素
func (b *UpdateBuilder[T]) Pop(field string, position int) *UpdateBuilder[T] {
	if _, ok := b.updates["$pop"]; !ok {
		b.updates["$pop"] = M{}
	}
	b.updates["$pop"].(M)[field] = position
	return b
}

// Restore 恢复软删除的文档
func (b *UpdateBuilder[T]) Restore() error {
	if b.coll.opts.SoftDelete == nil || !b.coll.opts.SoftDelete.Enabled {
		return nil
	}
	return b.Set(b.coll.opts.SoftDelete.Field, nil).Exec()
}

// Exec 执行更新 (默认更新单条)
func (b *UpdateBuilder[T]) Exec() error {
	return b.coll.UpdateOne(b.filter, b.updates)
}

// UpdateMany 执行批量更新
func (b *UpdateBuilder[T]) UpdateMany() (int64, error) {
	return b.coll.UpdateMany(b.filter, b.updates)
}
