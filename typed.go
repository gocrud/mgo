package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== TypedCollection 泛型集合 ====================

// TypedCollection 泛型集合封装
//
// 提供类型安全的集合操作
//
// 示例：
//
//	users := mgo.Model[User](db)
//	user, err := users.FindByID(id)  // 返回 *User
//	results, err := users.Find().All()  // 返回 []*User
type TypedCollection[T any] struct {
	coll *mongo.Collection
	db   *Database
	opts *CollectionOptions
}

// newTypedCollection 创建新的泛型集合实例
func newTypedCollection[T any](db *Database, coll *mongo.Collection) *TypedCollection[T] {
	return &TypedCollection[T]{
		coll: coll,
		db:   db,
		opts: &CollectionOptions{
			Context: db.Context(),
		},
	}
}

// ==================== 配置方法 ====================

// WithTimestamps 启用自动时间戳
func (c *TypedCollection[T]) WithTimestamps(fields ...string) *TypedCollection[T] {
	createdField := "created_at"
	updatedField := "updated_at"

	if len(fields) > 0 {
		createdField = fields[0]
	}
	if len(fields) > 1 {
		updatedField = fields[1]
	}

	c.opts.Timestamps = &TimestampConfig{
		CreatedField: createdField,
		UpdatedField: updatedField,
		Enabled:      true,
	}
	return c
}

// WithSoftDelete 启用软删除
func (c *TypedCollection[T]) WithSoftDelete(fields ...string) *TypedCollection[T] {
	deletedField := "deleted_at"

	if len(fields) > 0 {
		deletedField = fields[0]
	}

	c.opts.SoftDelete = &SoftDeleteConfig{
		Field:   deletedField,
		Enabled: true,
	}
	return c
}

// WithContext 设置默认上下文
func (c *TypedCollection[T]) WithContext(ctx context.Context) *TypedCollection[T] {
	c.opts.Context = ctx
	return c
}

// ==================== 基本信息方法 ====================

// Name 获取集合名称
func (c *TypedCollection[T]) Name() string {
	return c.coll.Name()
}

// Database 获取所属数据库
func (c *TypedCollection[T]) Database() *Database {
	return c.db
}

// Native 返回原生 mongo.Collection
func (c *TypedCollection[T]) Native() *mongo.Collection {
	return c.coll
}

// Context 获取默认上下文
func (c *TypedCollection[T]) Context() context.Context {
	return getContext(c.opts.Context)
}

// Options 获取集合选项
func (c *TypedCollection[T]) Options() *CollectionOptions {
	return c.opts
}

// ==================== 查询方法 ====================

// Find 创建查询构建器
func (c *TypedCollection[T]) Find() *Query[T] {
	return newQuery[T](c)
}

// FindByID 根据 ID 查询文档
func (c *TypedCollection[T]) FindByID(id interface{}) (*T, error) {
	return c.Find().ID(id).One()
}

// FindOne 查询单条文档
func (c *TypedCollection[T]) FindOne(filter M) (*T, error) {
	return c.Find().Filter(filter).One()
}

// FindAll 查询所有文档
func (c *TypedCollection[T]) FindAll(filter M) ([]*T, error) {
	return c.Find().Filter(filter).All()
}

// ==================== 插入方法 ====================

// Insert 插入单条文档
func (c *TypedCollection[T]) Insert(doc *T) (ObjectID, error) {
	ctx := c.Context()

	// 应用时间戳
	if c.opts.Timestamps != nil && c.opts.Timestamps.Enabled {
		applyTimestamps(doc, c.opts.Timestamps, true)
	}

	result, err := c.coll.InsertOne(ctx, doc)
	if err != nil {
		return NilObjectID, WrapError(err, "failed to insert")
	}

	// 回填 ID 到原始文档
	if oid, ok := result.InsertedID.(ObjectID); ok {
		if err := SetFieldValue(doc, "_id", oid); err == nil {
			return oid, nil
		}
		return oid, nil
	}

	return NilObjectID, nil
}

// InsertMany 插入多条文档
func (c *TypedCollection[T]) InsertMany(docs ...*T) ([]ObjectID, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	ctx := c.Context()

	// 应用时间戳
	if c.opts.Timestamps != nil && c.opts.Timestamps.Enabled {
		for _, doc := range docs {
			applyTimestamps(doc, c.opts.Timestamps, true)
		}
	}

	// 转换为 []interface{}
	items := make([]interface{}, len(docs))
	for i, doc := range docs {
		items[i] = doc
	}

	result, err := c.coll.InsertMany(ctx, items)
	if err != nil {
		return nil, WrapError(err, "failed to insert many")
	}

	ids := make([]ObjectID, 0, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(ObjectID); ok {
			ids = append(ids, oid)
			// 回填 ID 到原始文档
			if i < len(docs) {
				SetFieldValue(docs[i], "_id", oid)
			}
		}
	}

	return ids, nil
}

// ==================== 聚合方法 ====================

// Count 统计文档数量
func (c *TypedCollection[T]) Count(filter M) (int64, error) {
	return c.Find().Filter(filter).Count()
}

// CountAll 统计所有文档数量
func (c *TypedCollection[T]) CountAll() (int64, error) {
	return c.Find().Count()
}

// Exists 检查文档是否存在
func (c *TypedCollection[T]) Exists(filter M) (bool, error) {
	return c.Find().Filter(filter).Exists()
}

// ==================== 其他方法 ====================

// Drop 删除集合
func (c *TypedCollection[T]) Drop() error {
	ctx := c.Context()
	if err := c.coll.Drop(ctx); err != nil {
		return WrapError(err, "failed to drop collection")
	}
	return nil
}
