package mgo

// ==================== 更新操作 ====================

// UpdateBuilder 更新构建器
type UpdateBuilder struct {
	updates M
}

// newUpdateBuilder 创建更新构建器
func newUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		updates: M{},
	}
}

// ==================== 更新方法（Query）====================

// Set 设置字段值
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Set("status", "inactive").
//	    Update()
func (q *Query[T]) Set(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	// 将更新操作存储在 filter 的特殊键中
	if _, ok := q.filter["$set"]; !ok {
		q.filter["$set"] = M{}
	}
	q.filter["$set"].(M)[field] = NormalizeValue(value)
	return q
}

// Inc 增加字段值
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Inc("login_count", 1).
//	    Update()
func (q *Query[T]) Inc(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$inc"]; !ok {
		q.filter["$inc"] = M{}
	}
	q.filter["$inc"].(M)[field] = value
	return q
}

// Mul 乘以字段值
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Mul("score", 2).
//	    Update()
func (q *Query[T]) Mul(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$mul"]; !ok {
		q.filter["$mul"] = M{}
	}
	q.filter["$mul"].(M)[field] = value
	return q
}

// SetMin 设置字段最小值
//
// 示例：
//
//	err := users.Find().ID(id).
//	    SetMin("age", 18).
//	    Update()
func (q *Query[T]) SetMin(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$min"]; !ok {
		q.filter["$min"] = M{}
	}
	q.filter["$min"].(M)[field] = value
	return q
}

// SetMax 设置字段最大值
//
// 示例：
//
//	err := users.Find().ID(id).
//	    SetMax("age", 60).
//	    Update()
func (q *Query[T]) SetMax(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$max"]; !ok {
		q.filter["$max"] = M{}
	}
	q.filter["$max"].(M)[field] = value
	return q
}

// Unset 删除字段
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Unset("temp_field").
//	    Update()
func (q *Query[T]) Unset(fields ...string) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$unset"]; !ok {
		q.filter["$unset"] = M{}
	}
	for _, field := range fields {
		q.filter["$unset"].(M)[field] = ""
	}
	return q
}

// Rename 重命名字段
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Rename("old_name", "new_name").
//	    Update()
func (q *Query[T]) Rename(oldField, newField string) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$rename"]; !ok {
		q.filter["$rename"] = M{}
	}
	q.filter["$rename"].(M)[oldField] = newField
	return q
}

// ==================== 数组更新操作 ====================

// Push 向数组添加元素
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Push("tags", "new_tag").
//	    Update()
func (q *Query[T]) Push(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$push"]; !ok {
		q.filter["$push"] = M{}
	}
	q.filter["$push"].(M)[field] = value
	return q
}

// PushAll 向数组添加多个元素
//
// 示例：
//
//	err := users.Find().ID(id).
//	    PushAll("tags", []string{"tag1", "tag2"}).
//	    Update()
func (q *Query[T]) PushAll(field string, values []interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$push"]; !ok {
		q.filter["$push"] = M{}
	}
	q.filter["$push"].(M)[field] = M{"$each": values}
	return q
}

// Pull 从数组删除元素
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Pull("tags", "old_tag").
//	    Update()
func (q *Query[T]) Pull(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$pull"]; !ok {
		q.filter["$pull"] = M{}
	}
	q.filter["$pull"].(M)[field] = value
	return q
}

// PullAll 从数组删除多个元素
//
// 示例：
//
//	err := users.Find().ID(id).
//	    PullAll("tags", []string{"tag1", "tag2"}).
//	    Update()
func (q *Query[T]) PullAll(field string, values []interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$pullAll"]; !ok {
		q.filter["$pullAll"] = M{}
	}
	q.filter["$pullAll"].(M)[field] = values
	return q
}

// AddToSet 向数组添加元素（去重）
//
// 示例：
//
//	err := users.Find().ID(id).
//	    AddToSet("roles", "admin").
//	    Update()
func (q *Query[T]) AddToSet(field string, value interface{}) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$addToSet"]; !ok {
		q.filter["$addToSet"] = M{}
	}
	q.filter["$addToSet"].(M)[field] = value
	return q
}

// Pop 从数组移除第一个或最后一个元素
//
// position: 1 移除最后一个, -1 移除第一个
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Pop("tags", 1).  // 移除最后一个
//	    Update()
func (q *Query[T]) Pop(field string, position int) *Query[T] {
	if q.coll == nil {
		return q
	}

	if _, ok := q.filter["$pop"]; !ok {
		q.filter["$pop"] = M{}
	}
	q.filter["$pop"].(M)[field] = position
	return q
}

// ==================== 执行更新操作 ====================

// buildUpdateDoc 构建更新文档
func (q *Query[T]) buildUpdateDoc() M {
	update := M{}

	// 提取更新操作（不删除，因为 buildFilter 会自动排除）
	updateOps := []string{"$set", "$inc", "$mul", "$min", "$max", "$unset", "$rename", "$push", "$pull", "$pullAll", "$addToSet", "$pop"}
	for _, op := range updateOps {
		if val, ok := q.filter[op]; ok {
			update[op] = val
		}
	}

	// 应用时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		if _, ok := update["$set"]; !ok {
			update["$set"] = M{}
		}
		update["$set"].(M)[q.coll.opts.Timestamps.UpdatedField] = NormalizeValue(Now())
	}

	return update
}

// Update 执行更新（单条）
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Set("status", "inactive").
//	    Update()
func (q *Query[T]) Update() error {
	ctx := q.Context()
	filter := q.buildFilter()
	update := q.buildUpdateDoc()

	if len(update) == 0 {
		return ErrEmptyUpdate
	}

	_, err := q.coll.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return WrapError(err, "failed to update")
	}

	return nil
}

// UpdateMany 执行批量更新
//
// 示例：
//
//	n, err := users.Find().
//	    Where("status", "pending").
//	    Set("status", "active").
//	    UpdateMany()
func (q *Query[T]) UpdateMany() (int64, error) {
	ctx := q.Context()
	filter := q.buildFilter()
	update := q.buildUpdateDoc()

	if len(update) == 0 {
		return 0, ErrEmptyUpdate
	}

	result, err := q.coll.coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to update many")
	}

	return result.ModifiedCount, nil
}

// Patch 部分更新（从结构体）
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Patch(&User{Status: "inactive", Age: 30})
func (q *Query[T]) Patch(doc *T) error {
	// TODO: 使用反射提取非零值字段
	// 然后调用 Set 方法
	return q.Update()
}

// Replace 完整替换文档
//
// 示例：
//
//	err := users.Find().ID(id).
//	    Replace(newUser)
func (q *Query[T]) Replace(doc *T) error {
	ctx := q.Context()
	filter := q.buildFilter()

	// 应用时间戳
	if q.coll.opts.Timestamps != nil && q.coll.opts.Timestamps.Enabled {
		applyTimestamps(doc, q.coll.opts.Timestamps, false)
	}

	_, err := q.coll.coll.ReplaceOne(ctx, filter, doc)
	if err != nil {
		return WrapError(err, "failed to replace")
	}

	return nil
}

// ==================== FindAndModify 原子操作 ====================

// UpdateAndReturn 更新并返回更新后的文档
//
// 示例：
//
//	user, err := users.Find().ID(id).
//	    Set("status", "processing").
//	    UpdateAndReturn()
func (q *Query[T]) UpdateAndReturn() (*T, error) {
	// TODO: 使用 FindOneAndUpdate 实现
	return nil, nil
}

// UpdateAndReturnOld 更新并返回更新前的文档
//
// 示例：
//
//	oldUser, err := users.Find().ID(id).
//	    Set("status", "processing").
//	    UpdateAndReturnOld()
func (q *Query[T]) UpdateAndReturnOld() (*T, error) {
	// TODO: 使用 FindOneAndUpdate 实现
	return nil, nil
}

// ==================== Upsert 插入或更新 ====================

// Upsert 存在则更新，不存在则插入
//
// 示例：
//
//	err := users.Find().
//	    Where("email", email).
//	    Upsert(&user)
func (q *Query[T]) Upsert(doc *T) error {
	// TODO: 使用 UpdateOne 的 upsert 选项实现
	return nil
}
