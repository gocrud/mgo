package mgo

// ==================== 查询执行方法 ====================

// One 查询单条文档
func (q *Query[T]) One() (*T, error) {
	filter := q.buildFilter()
	opts := q.buildFindOneOptions()

	var result T
	err := q.coll.coll.FindOne(q.ctx, filter, opts).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return nil, ErrNoDocuments
		}
		return nil, WrapError(err, "failed to find one")
	}

	return &result, nil
}

// All 查询所有匹配的文档
func (q *Query[T]) All() ([]*T, error) {
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(q.ctx, filter, opts)
	if err != nil {
		return nil, WrapError(err, "failed to find")
	}
	defer cursor.Close(q.ctx)

	var results []*T
	if err := cursor.All(q.ctx, &results); err != nil {
		return nil, WrapError(err, "failed to decode results")
	}

	return results, nil
}

// Count 统计匹配的文档数量
func (q *Query[T]) Count() (int64, error) {
	filter := q.buildFilter()

	count, err := q.coll.coll.CountDocuments(q.ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count")
	}

	return count, nil
}

// Exists 检查是否存在匹配的文档
func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Limit(1).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Must 系列（panic on error）====================

// MustOne 查询单条，失败时 panic
func (q *Query[T]) MustOne() *T {
	result, err := q.One()
	if err != nil {
		panic(err)
	}
	return result
}

// MustAll 查询所有，失败时 panic
func (q *Query[T]) MustAll() []*T {
	results, err := q.All()
	if err != nil {
		panic(err)
	}
	return results
}

// MustCount 统计数量，失败时 panic
func (q *Query[T]) MustCount() int64 {
	count, err := q.Count()
	if err != nil {
		panic(err)
	}
	return count
}

// ==================== Or 系列（返回默认值）====================

// OneOrNil 查询单条，未找到返回 nil
func (q *Query[T]) OneOrNil() *T {
	result, err := q.One()
	if err != nil {
		return nil
	}
	return result
}

// Each 遍历查询结果
func (q *Query[T]) Each(fn func(*T) error) error {
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(q.ctx, filter, opts)
	if err != nil {
		return WrapError(err, "failed to find")
	}
	defer cursor.Close(q.ctx)

	for cursor.Next(q.ctx) {
		var result T
		if err := cursor.Decode(&result); err != nil {
			return WrapError(err, "failed to decode")
		}
		if err := fn(&result); err != nil {
			return err
		}
	}

	return cursor.Err()
}

// AllOrEmpty 查询所有，失败返回空切片
func (q *Query[T]) AllOrEmpty() []*T {
	results, err := q.All()
	if err != nil || results == nil {
		return []*T{}
	}
	return results
}

// OneOr 查询单条，未找到返回默认值
func (q *Query[T]) OneOr(defaultValue *T) *T {
	result, err := q.One()
	if err != nil {
		return defaultValue
	}
	return result
}

// ==================== 高级查询方法 ====================

// First 查询第一条记录
func (q *Query[T]) First() (*T, error) {
	return q.Limit(1).One()
}

// Last 查询最后一条记录
func (q *Query[T]) Last() (*T, error) {
	// 反转排序
	reversed := q.Clone()
	newSort := make(D, len(reversed.sort))
	for i, elem := range reversed.sort {
		if order, ok := elem.Value.(int); ok {
			newSort[i] = E{Key: elem.Key, Value: -order}
		} else {
			newSort[i] = elem
		}
	}
	reversed.sort = newSort
	return reversed.Limit(1).One()
}

// FirstOrCreate 查询第一条，不存在则创建
func (q *Query[T]) FirstOrCreate(doc *T) (*T, bool, error) {
	result, err := q.One()
	if err == nil {
		return result, false, nil
	}

	if !IsNoDocuments(err) {
		return nil, false, err
	}

	// 不存在，创建新记录
	id, err := q.coll.WithContext(q.ctx).Insert(doc)
	if err != nil {
		return nil, false, err
	}

	// 重新查询以获取完整数据
	result, err = q.coll.WithContext(q.ctx).FindByID(id)
	if err != nil {
		return nil, true, err
	}

	return result, true, nil
}

// ==================== Distinct 去重 ====================

// Distinct 获取字段的不重复值
func (q *Query[T]) Distinct(field string) ([]interface{}, error) {
	filter := q.buildFilter()

	distinctResult := q.coll.coll.Distinct(q.ctx, field, filter)
	if distinctResult.Err() != nil {
		return nil, WrapError(distinctResult.Err(), "failed to get distinct values")
	}

	var values []interface{}
	if err := distinctResult.Decode(&values); err != nil {
		return nil, WrapError(err, "failed to decode distinct values")
	}

	return values, nil
}

// ==================== Chunk 批量处理 ====================

// Chunk 分块处理查询结果
func (q *Query[T]) Chunk(size int, fn func([]*T) error) error {
	if size <= 0 {
		size = 100
	}

	skip := int64(0)
	limit := int64(size)

	for {
		results, err := q.Clone().Skip(skip).Limit(limit).All()
		if err != nil {
			return err
		}

		if len(results) == 0 {
			break
		}

		if err := fn(results); err != nil {
			return err
		}

		if len(results) < size {
			break
		}

		skip += int64(size)
	}

	return nil
}

// ==================== 聚合辅助方法 ====================

// Sum 求和
func (q *Query[T]) Sum(field string) (float64, error) {
	// TODO: 使用聚合实现
	return 0, nil
}

// Avg 平均值
func (q *Query[T]) Avg(field string) (float64, error) {
	// TODO: 使用聚合实现
	return 0, nil
}

// Max 最大值
func (q *Query[T]) Max(field string) (interface{}, error) {
	// TODO: 使用聚合实现
	return nil, nil
}

// Min 最小值
func (q *Query[T]) Min(field string) (interface{}, error) {
	// TODO: 使用聚合实现
	return nil, nil
}
