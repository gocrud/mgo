package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== Collection 传统集合封装 ====================

// Collection MongoDB 集合封装（传统方式）
//
// 提供非泛型的集合操作，适用于动态类型场景
//
// 示例：
//
//	users := db.Collection("users")
//	var user User
//	err := users.FindOne(mgo.M{"email": email}, &user)
type Collection struct {
	coll *mongo.Collection
	db   *Database
	opts *CollectionOptions
}

// newCollection 创建新的集合实例
func newCollection(db *Database, coll *mongo.Collection, opts ...CollectionOption) *Collection {
	options := &CollectionOptions{
		Context: db.Context(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Collection{
		coll: coll,
		db:   db,
		opts: options,
	}
}

// Name 获取集合名称
//
// 示例：
//
//	name := coll.Name()
func (c *Collection) Name() string {
	return c.coll.Name()
}

// Database 获取所属数据库
//
// 示例：
//
//	db := coll.Database()
func (c *Collection) Database() *Database {
	return c.db
}

// Native 返回原生 mongo.Collection
//
// 示例：
//
//	nativeColl := coll.Native()
func (c *Collection) Native() *mongo.Collection {
	return c.coll
}

// Context 获取默认上下文
func (c *Collection) Context() context.Context {
	return getContext(c.opts.Context)
}

// Options 获取集合选项
func (c *Collection) Options() *CollectionOptions {
	return c.opts
}

// ==================== 基础查询方法 ====================

// FindOne 查询单条文档
//
// 示例：
//
//	var user User
//	err := coll.FindOne(mgo.M{"email": email}, &user)
func (c *Collection) FindOne(filter interface{}, result interface{}) error {
	ctx := c.Context()
	err := c.coll.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return WrapError(err, "failed to find one")
	}
	return nil
}

// FindByID 根据 ID 查询文档
//
// 示例：
//
//	var user User
//	err := coll.FindByID(id, &user)
func (c *Collection) FindByID(id interface{}, result interface{}) error {
	return c.FindOne(M{"_id": id}, result)
}

// Find 查询多条文档
//
// 示例：
//
//	var users []User
//	err := coll.Find(mgo.M{"status": "active"}, &users)
func (c *Collection) Find(filter interface{}, results interface{}) error {
	ctx := c.Context()
	cursor, err := c.coll.Find(ctx, filter)
	if err != nil {
		return WrapError(err, "failed to find")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return WrapError(err, "failed to decode results")
	}
	return nil
}

// ==================== 插入方法 ====================

// InsertOne 插入单条文档
//
// 示例：
//
//	id, err := coll.InsertOne(&user)
func (c *Collection) InsertOne(doc interface{}) (ObjectID, error) {
	ctx := c.Context()
	result, err := c.coll.InsertOne(ctx, doc)
	if err != nil {
		return NilObjectID, WrapError(err, "failed to insert one")
	}

	if oid, ok := result.InsertedID.(ObjectID); ok {
		return oid, nil
	}

	return NilObjectID, nil
}

// InsertMany 插入多条文档
//
// 示例：
//
//	ids, err := coll.InsertMany([]interface{}{user1, user2, user3})
func (c *Collection) InsertMany(docs []interface{}) ([]ObjectID, error) {
	ctx := c.Context()
	result, err := c.coll.InsertMany(ctx, docs)
	if err != nil {
		return nil, WrapError(err, "failed to insert many")
	}

	ids := make([]ObjectID, 0, len(result.InsertedIDs))
	for _, id := range result.InsertedIDs {
		if oid, ok := id.(ObjectID); ok {
			ids = append(ids, oid)
		}
	}

	return ids, nil
}

// ==================== 更新方法 ====================

// UpdateOne 更新单条文档
//
// 示例：
//
//	err := coll.UpdateOne(
//	    mgo.M{"_id": id},
//	    mgo.M{"$set": mgo.M{"status": "inactive"}},
//	)
func (c *Collection) UpdateOne(filter, update interface{}) error {
	ctx := c.Context()
	_, err := c.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return WrapError(err, "failed to update one")
	}
	return nil
}

// UpdateByID 根据 ID 更新文档
//
// 示例：
//
//	err := coll.UpdateByID(id, mgo.M{"$set": mgo.M{"status": "inactive"}})
func (c *Collection) UpdateByID(id, update interface{}) error {
	return c.UpdateOne(M{"_id": id}, update)
}

// UpdateMany 更新多条文档
//
// 示例：
//
//	n, err := coll.UpdateMany(
//	    mgo.M{"status": "pending"},
//	    mgo.M{"$set": mgo.M{"status": "active"}},
//	)
func (c *Collection) UpdateMany(filter, update interface{}) (int64, error) {
	ctx := c.Context()
	result, err := c.coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, WrapError(err, "failed to update many")
	}
	return result.ModifiedCount, nil
}

// ReplaceOne 替换单条文档
//
// 示例：
//
//	err := coll.ReplaceOne(mgo.M{"_id": id}, newUser)
func (c *Collection) ReplaceOne(filter, replacement interface{}) error {
	ctx := c.Context()
	_, err := c.coll.ReplaceOne(ctx, filter, replacement)
	if err != nil {
		return WrapError(err, "failed to replace one")
	}
	return nil
}

// ==================== 删除方法 ====================

// DeleteOne 删除单条文档
//
// 示例：
//
//	err := coll.DeleteOne(mgo.M{"_id": id})
func (c *Collection) DeleteOne(filter interface{}) error {
	ctx := c.Context()
	_, err := c.coll.DeleteOne(ctx, filter)
	if err != nil {
		return WrapError(err, "failed to delete one")
	}
	return nil
}

// DeleteByID 根据 ID 删除文档
//
// 示例：
//
//	err := coll.DeleteByID(id)
func (c *Collection) DeleteByID(id interface{}) error {
	return c.DeleteOne(M{"_id": id})
}

// DeleteMany 删除多条文档
//
// 示例：
//
//	n, err := coll.DeleteMany(mgo.M{"status": "expired"})
func (c *Collection) DeleteMany(filter interface{}) (int64, error) {
	ctx := c.Context()
	result, err := c.coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to delete many")
	}
	return result.DeletedCount, nil
}

// ==================== 聚合方法 ====================

// CountDocuments 统计文档数量
//
// 示例：
//
//	count, err := coll.CountDocuments(mgo.M{"status": "active"})
func (c *Collection) CountDocuments(filter interface{}) (int64, error) {
	ctx := c.Context()
	count, err := c.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, WrapError(err, "failed to count documents")
	}
	return count, nil
}

// Distinct 获取字段的不重复值
//
// 示例：
//
//	values, err := coll.Distinct("city", mgo.M{"status": "active"})
func (c *Collection) Distinct(fieldName string, filter interface{}) ([]interface{}, error) {
	ctx := c.Context()
	distinctResult := c.coll.Distinct(ctx, fieldName, filter)
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
//
// 示例：
//
//	pipeline := mgo.Pipeline{
//	    {{"$match", mgo.M{"status": "active"}}},
//	    {{"$group", mgo.M{"_id": "$city", "count": mgo.M{"$sum": 1}}}},
//	}
//	var results []CityStats
//	err := coll.Aggregate(pipeline, &results)
func (c *Collection) Aggregate(pipeline interface{}, results interface{}) error {
	ctx := c.Context()
	cursor, err := c.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return WrapError(err, "failed to aggregate")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, results); err != nil {
		return WrapError(err, "failed to decode aggregate results")
	}
	return nil
}

// ==================== 索引方法 ====================

// CreateIndex 创建索引
//
// 示例：
//
//	err := coll.CreateIndex("email", true) // unique index
func (c *Collection) CreateIndex(field string, unique bool) error {
	// TODO: 实现索引创建
	return nil
}

// DropIndex 删除索引
//
// 示例：
//
//	err := coll.DropIndex("email_1")
func (c *Collection) DropIndex(name string) error {
	// TODO: 实现索引删除
	return nil
}

// ==================== 其他方法 ====================

// Drop 删除集合
//
// 示例：
//
//	err := coll.Drop()
func (c *Collection) Drop() error {
	ctx := c.Context()
	if err := c.coll.Drop(ctx); err != nil {
		return WrapError(err, "failed to drop collection")
	}
	return nil
}

// EstimatedDocumentCount 估算文档数量（快速）
//
// 示例：
//
//	count, err := coll.EstimatedDocumentCount()
func (c *Collection) EstimatedDocumentCount() (int64, error) {
	ctx := c.Context()
	count, err := c.coll.EstimatedDocumentCount(ctx)
	if err != nil {
		return 0, WrapError(err, "failed to estimate document count")
	}
	return count, nil
}
