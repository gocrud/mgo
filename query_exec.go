package mgo

// ==================== 查询执行方法 ====================

// One 查询单条文档
//
// 示例：
//
//	user, err := users.Find().Where("email", email).One()
func (q *Query[T]) One() (*T, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	var result T
	err := q.coll.coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if IsNoDocuments(err) {
			return nil, ErrNoDocuments
		}
		return nil, WrapError(err, "failed to find one")
	}

	return &result, nil
}

// All 查询所有匹配的文档
//
// 示例：
//
//	users, err := users.Find().Where("status", "active").All()
func (q *Query[T]) All() ([]*T, error) {
	ctx := q.Context()
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, WrapError(err, "failed to find")
	}
	defer cursor.Close(ctx)

	var results []*T
	if err := cursor.All(ctx, &results); err != nil {
		return nil, WrapError(err, "failed to decode results")
	}

	return results, nil
}

// Count 统计匹配的文档数量
//
// 示例：
//
//	count, err := users.Find().Where("status", "active").Count()
func (q *Query[T]) Count() (int64, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	count, err := q.coll.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count")
	}

	return count, nil
}

// Exists 检查是否存在匹配的文档
//
// 示例：
//
//	exists, err := users.Find().Where("email", email).Exists()
func (q *Query[T]) Exists() (bool, error) {
	count, err := q.Limit(1).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Must 系列（panic on error）====================

// MustOne 查询单条，失败时 panic
//
// 示例：
//
//	user := users.Find().Where("email", email).MustOne()
func (q *Query[T]) MustOne() *T {
	result, err := q.One()
	if err != nil {
		panic(err)
	}
	return result
}

// MustAll 查询所有，失败时 panic
//
// 示例：
//
//	users := users.Find().Where("status", "active").MustAll()
func (q *Query[T]) MustAll() []*T {
	results, err := q.All()
	if err != nil {
		panic(err)
	}
	return results
}

// MustCount 统计数量，失败时 panic
//
// 示例：
//
//	count := users.Find().Where("status", "active").MustCount()
func (q *Query[T]) MustCount() int64 {
	count, err := q.Count()
	if err != nil {
		panic(err)
	}
	return count
}

// ==================== Or 系列（返回默认值）====================

// OneOrNil 查询单条，未找到返回 nil
//
// 示例：
//
//	user := users.Find().Where("email", email).OneOrNil()
//	if user != nil {
//	    // 找到用户
//	}
func (q *Query[T]) OneOrNil() *T {
	result, err := q.One()
	if err != nil {
		return nil
	}
	return result
}

// AllOrEmpty 查询所有，失败返回空切片
//
// 示例：
//
//	users := users.Find().Where("status", "active").AllOrEmpty()
func (q *Query[T]) AllOrEmpty() []*T {
	results, err := q.All()
	if err != nil || results == nil {
		return []*T{}
	}
	return results
}

// OneOr 查询单条，未找到返回默认值
//
// 示例：
//
//	user := users.Find().Where("email", email).OneOr(&User{
//	    Status: "guest",
//	})
func (q *Query[T]) OneOr(defaultValue *T) *T {
	result, err := q.One()
	if err != nil {
		return defaultValue
	}
	return result
}

// ==================== 高级查询方法 ====================

// First 查询第一条记录
//
// 示例：
//
//	user, err := users.Find().OrderBy("created_at").First()
func (q *Query[T]) First() (*T, error) {
	return q.Limit(1).One()
}

// Last 查询最后一条记录
//
// 示例：
//
//	user, err := users.Find().OrderBy("created_at").Last()
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
//
// 示例：
//
//	user, created, err := users.Find().
//	    Where("email", email).
//	    FirstOrCreate(&User{Email: email, Name: "新用户"})
func (q *Query[T]) FirstOrCreate(doc *T) (*T, bool, error) {
	result, err := q.One()
	if err == nil {
		return result, false, nil
	}

	if !IsNoDocuments(err) {
		return nil, false, err
	}

	// 不存在，创建新记录
	id, err := q.coll.Insert(doc)
	if err != nil {
		return nil, false, err
	}

	// 重新查询以获取完整数据
	result, err = q.coll.FindByID(id)
	if err != nil {
		return nil, true, err
	}

	return result, true, nil
}

// ==================== Distinct 去重 ====================

// Distinct 获取字段的不重复值
//
// 示例：
//
//	cities, err := users.Find().
//	    Where("status", "active").
//	    Distinct("city")
func (q *Query[T]) Distinct(field string) ([]interface{}, error) {
	ctx := q.Context()
	filter := q.buildFilter()

	distinctResult := q.coll.coll.Distinct(ctx, field, filter)
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
//
// 示例：
//
//	err := users.Find().Chunk(100, func(users []*User) error {
//	    for _, user := range users {
//	        process(user)
//	    }
//	    return nil
//	})
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

// Each 遍历每一条记录
//
// 示例：
//
//	err := users.Find().Each(func(user *User) error {
//	    return process(user)
//	})
func (q *Query[T]) Each(fn func(*T) error) error {
	ctx := q.Context()
	filter := q.buildFilter()
	opts := q.buildOptions()

	cursor, err := q.coll.coll.Find(ctx, filter, opts)
	if err != nil {
		return WrapError(err, "failed to find")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc T
		if err := cursor.Decode(&doc); err != nil {
			return WrapError(err, "failed to decode document")
		}

		if err := fn(&doc); err != nil {
			return err
		}
	}

	if err := cursor.Err(); err != nil {
		return WrapError(err, "cursor error")
	}

	return nil
}

// ==================== 聚合辅助方法 ====================

// Sum 求和
//
// 示例：
//
//	total, err := users.Find().Sum("balance")
func (q *Query[T]) Sum(field string) (float64, error) {
	// TODO: 使用聚合实现
	return 0, nil
}

// Avg 平均值
//
// 示例：
//
//	avg, err := users.Find().Avg("age")
func (q *Query[T]) Avg(field string) (float64, error) {
	// TODO: 使用聚合实现
	return 0, nil
}

// Max 最大值
//
// 示例：
//
//	max, err := users.Find().Max("age")
func (q *Query[T]) Max(field string) (interface{}, error) {
	// TODO: 使用聚合实现
	return nil, nil
}

// Min 最小值
//
// 示例：
//
//	min, err := users.Find().Min("age")
func (q *Query[T]) Min(field string) (interface{}, error) {
	// TODO: 使用聚合实现
	return nil, nil
}
