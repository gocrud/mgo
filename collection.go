package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// HookHandler 钩子函数类型
type HookHandler[T any] func(ctx context.Context, doc *T) error

// CollectionHooks 集合钩子
type CollectionHooks[T any] struct {
	BeforeInsert []HookHandler[T]
	AfterFind    []HookHandler[T]
}

// Collection 泛型集合封装
//
// 提供类型安全的集合操作
//
// 示例：
//
//	users := mgo.Model[User](db)
//	user, err := users.FindByID(ctx, id)
type Collection[T any] struct {
	coll  *mongo.Collection
	db    *Database
	ctx   context.Context
	opts  *CollectionOptions
	hooks *CollectionHooks[T]
}

// newCollection 创建新的集合实例
func newCollection[T any](db *Database, coll *mongo.Collection, opts ...CollectionOption) *Collection[T] {
	options := &CollectionOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &Collection[T]{
		coll:  coll,
		db:    db,
		ctx:   context.Background(),
		opts:  options,
		hooks: &CollectionHooks[T]{},
	}
}

// Hooks 获取钩子管理器
func (c *Collection[T]) Hooks() *CollectionHooks[T] {
	if c.hooks == nil {
		c.hooks = &CollectionHooks[T]{}
	}
	return c.hooks
}

// RegisterBeforeInsert 注册插入前钩子
func (h *CollectionHooks[T]) RegisterBeforeInsert(handler HookHandler[T]) {
	h.BeforeInsert = append(h.BeforeInsert, handler)
}

// RegisterAfterFind 注册查询后钩子
func (h *CollectionHooks[T]) RegisterAfterFind(handler HookHandler[T]) {
	h.AfterFind = append(h.AfterFind, handler)
}

// WithContext 设置上下文
func (c *Collection[T]) WithContext(ctx context.Context) *Collection[T] {
	// 浅拷贝
	newC := *c
	newC.ctx = ctx
	return &newC
}

// WithTx 使用事务上下文
func (c *Collection[T]) WithTx(tx *Tx) *Collection[T] {
	return c.WithContext(tx.Context())
}

// Context 获取上下文
func (c *Collection[T]) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// Name 获取集合名称
func (c *Collection[T]) Name() string {
	return c.coll.Name()
}

// Database 获取所属数据库
func (c *Collection[T]) Database() *Database {
	return c.db
}

// Native 返回原生 mongo.Collection
func (c *Collection[T]) Native() *mongo.Collection {
	return c.coll
}

// Options 获取集合选项
func (c *Collection[T]) Options() *CollectionOptions {
	return c.opts
}

// Find 创建查询构建器
func (c *Collection[T]) Find() *Query[T] {
	return newQuery(c).WithContext(c.ctx)
}

// Update 创建更新构建器
func (c *Collection[T]) Update() *UpdateBuilder[T] {
	return newUpdateBuilder(c)
}

// Delete 创建删除构建器
func (c *Collection[T]) Delete() *DeleteBuilder[T] {
	return newDeleteBuilder(c)
}

// FindOne 查询单条文档
func (c *Collection[T]) FindOne(filter any) (*T, error) {
	var result T
	err := c.coll.FindOne(c.ctx, filter).Decode(&result)
	if err != nil {
		return nil, WrapError(err, "failed to find one")
	}

	if c.hooks != nil {
		for _, hook := range c.hooks.AfterFind {
			if err := hook(c.ctx, &result); err != nil {
				return nil, WrapError(err, "failed to execute AfterFind hook")
			}
		}
	}

	return &result, nil
}

// FindByID 根据 ID 查询文档
func (c *Collection[T]) FindByID(id any) (*T, error) {
	return c.Find().ID(id).One()
}

// Insert 插入单条文档
func (c *Collection[T]) Insert(doc *T) (ObjectID, error) {
	if c.hooks != nil {
		for _, hook := range c.hooks.BeforeInsert {
			if err := hook(c.ctx, doc); err != nil {
				return NilObjectID, WrapError(err, "failed to execute BeforeInsert hook")
			}
		}
	}

	// 应用时间戳
	if c.opts.Timestamps != nil && c.opts.Timestamps.Enabled {
		applyTimestamps(doc, c.opts.Timestamps, true)
	}

	result, err := c.coll.InsertOne(c.ctx, doc)
	if err != nil {
		return NilObjectID, WrapError(err, "failed to insert")
	}

	if oid, ok := result.InsertedID.(ObjectID); ok {
		// 尝试回填 ID
		SetFieldValue(doc, "_id", oid)
		return oid, nil
	}

	return NilObjectID, nil
}

// InsertMany 插入多条文档
func (c *Collection[T]) InsertMany(docs []*T) ([]ObjectID, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	// 转换 interface{} 切片以适配驱动
	items := make([]interface{}, len(docs))
	for i, doc := range docs {
		if c.opts.Timestamps != nil && c.opts.Timestamps.Enabled {
			applyTimestamps(doc, c.opts.Timestamps, true)
		}
		items[i] = doc
	}

	result, err := c.coll.InsertMany(c.ctx, items)
	if err != nil {
		return nil, WrapError(err, "failed to insert many")
	}

	ids := make([]ObjectID, 0, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(ObjectID); ok {
			ids = append(ids, oid)
			if i < len(docs) {
				SetFieldValue(docs[i], "_id", oid)
			}
		}
	}

	return ids, nil
}

// CountDocuments 统计文档数量
func (c *Collection[T]) CountDocuments(filter any) (int64, error) {
	count, err := c.coll.CountDocuments(c.ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count documents")
	}
	return count, nil
}
	count, err := c.coll.CountDocuments(c.ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count documents")
	}
	return count, nil
}

// Distinct 获取字段的不重复值
func (c *Collection[T]) Distinct(fieldName string, filter any) ([]any, error) {
	distinctResult := c.coll.Distinct(c.ctx, fieldName, filter)
	if distinctResult.Err() != nil {
		return nil, WrapError(distinctResult.Err(), "failed to get distinct values")
	}

	var values []interface{}
	if err := distinctResult.Decode(&values); err != nil {
		return nil, WrapError(err, "failed to decode distinct values")
	}

	return values, nil
}

// Aggregate 执行聚合查询
func (c *Collection[T]) Aggregate(pipeline any, results any) error {
	cursor, err := c.coll.Aggregate(c.ctx, pipeline)
	if err != nil {
		return WrapError(err, "failed to aggregate")
	}
	defer cursor.Close(c.ctx)

	if err := cursor.All(c.ctx, results); err != nil {
		return WrapError(err, "failed to decode aggregate results")
	}
	return nil
}

// Drop 删除集合
func (c *Collection[T]) Drop() error {
	if err := c.coll.Drop(c.ctx); err != nil {
		return WrapError(err, "failed to drop collection")
	}
	return nil
}

// Truncate 清空集合
func (c *Collection[T]) Truncate() error {
	_, err := c.coll.DeleteMany(c.ctx, M{})
	if err != nil {
		return WrapError(err, "failed to truncate collection")
	}
	return nil
}

// EstimatedDocumentCount 估算文档数量
func (c *Collection[T]) EstimatedDocumentCount() (int64, error) {
	count, err := c.coll.EstimatedDocumentCount(c.ctx)
	if err != nil {
		return 0, WrapError(err, "failed to estimate document count")
	}
	return count, nil
}

// WithTimestamps 启用自动时间戳
func (c *Collection[T]) WithTimestamps(fields ...string) *Collection[T] {
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
func (c *Collection[T]) WithSoftDelete(fields ...string) *Collection[T] {
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

// Count 统计符合条件的文档数量
func (c *Collection[T]) Count(filter any) (int64, error) {
	return c.coll.CountDocuments(c.ctx, filter)
}

// CountAll 统计所有文档数量
func (c *Collection[T]) CountAll() (int64, error) {
	return c.coll.CountDocuments(c.ctx, M{})
}
