package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collection MongoDB 集合封装
//
// 提供两个核心入口:
// - Query() 用于查询操作
// - Aggs() 用于聚合操作
//
// 使用示例：
//
//	// 查询
//	var users []User
//	err := coll.Query().
//	    Eq("status", "active").
//	    Gt("age", 18).
//	    All(ctx, &users)
//
//	// 聚合
//	var results []Result
//	err := coll.Aggs().
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{"count": Sum(1)}).
//	    All(ctx, &results)
//
//	// 简单插入
//	id, err := coll.InsertOne(ctx, user)
type Collection struct {
	coll       *mongo.Collection
	softDelete *SoftDeleteConfig
}

// NewCollection 创建集合封装（内部使用）
//
// 对外使用 Client.Collection() 或 Database.Collection() 方法
//
// 示例：
//
//	// 不启用软删除
//	coll := mgo.newCollection(mongoCollection)
//
//	// 启用软删除
//	coll := mgo.newCollection(mongoCollection, mgo.WithSoftDelete())
//
//	// 自定义软删除字段
//	coll := mgo.newCollection(mongoCollection, mgo.WithSoftDelete("removed_at"))
func newCollection(coll *mongo.Collection, opts ...CollectionOption) *Collection {
	c := &Collection{
		coll:       coll,
		softDelete: defaultSoftDeleteConfig(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Name 获取集合名称
func (c *Collection) Name() string {
	return c.coll.Name()
}

// Database 获取数据库
func (c *Collection) Database() *mongo.Database {
	return c.coll.Database()
}

// Native 返回原生 mongo.Collection（用于高级场景）
func (c *Collection) Native() *mongo.Collection {
	return c.coll
}

// ==================== 核心入口 ====================

// Query 创建查询构建器
//
// 这是查询操作的唯一入口，支持所有查询相关功能
//
// 示例：
//
//	// 查询单条
//	var user User
//	err := coll.Query().Eq("_id", id).One(ctx, &user)
//
//	// 查询多条
//	var users []User
//	err := coll.Query().
//	    Eq("status", "active").
//	    Select("name", "email").
//	    Desc("created_at").
//	    Limit(10).
//	    All(ctx, &users)
//
//	// 计数
//	count, err := coll.Query().Eq("status", "active").Count(ctx)
//
//	// 更新
//	result, err := coll.Query().
//	    Eq("_id", id).
//	    UpdateOne(ctx, Update().Set("status", "inactive"))
//
//	// 删除
//	result, err := coll.Query().Eq("status", "expired").DeleteMany(ctx)
func (c *Collection) Query(ctx context.Context) *QueryBuilder {
	return newQueryBuilder(c, ctx)
}

// Aggs 创建聚合构建器
//
// 这是聚合操作的唯一入口
//
// 示例：
//
//	// 分组统计
//	var results []Result
//	err := coll.Aggs().
//	    Match(Filter().Eq("status", "active")).
//	    Group("$city", M{
//	        "total": Sum(1),
//	        "avgAge": Avg("$age"),
//	    }).
//	    SortDesc("total").
//	    All(ctx, &results)
//
//	// 关联查询
//	err := coll.Aggs().
//	    Lookup("orders", "user_id", "_id", "orders").
//	    Unwind("$orders").
//	    All(ctx, &results)
func (c *Collection) Aggs(ctx context.Context) *AggsBuilder {
	return newAggsBuilder(c, ctx)
}

// ==================== 简单 CRUD 操作 ====================
// 这些方法用于最基本的操作，不需要构建器的场景

// InsertOne 插入单条文档
//
// 示例：
//
//	result, err := coll.InsertOne(ctx, user)
//	insertedID := result.InsertedID
func (c *Collection) InsertOne(ctx context.Context, document any) (*mongo.InsertOneResult, error) {
	return c.coll.InsertOne(ctx, document)
}

// InsertMany 插入多条文档
//
// 示例：
//
//	result, err := coll.InsertMany(ctx, []any{user1, user2, user3})
//	insertedIDs := result.InsertedIDs
func (c *Collection) InsertMany(ctx context.Context, documents []any) (*mongo.InsertManyResult, error) {
	return c.coll.InsertMany(ctx, documents)
}

// UpdateByID 通过 ID 更新文档（快捷方法）
//
// 示例：
//
//	result, err := coll.UpdateByID(ctx, id,
//	    Update().Set("status", "inactive"))
func (c *Collection) UpdateByID(ctx context.Context, id any, update any) (*mongo.UpdateResult, error) {
	// 这里没有 QueryBuilder 实例，所以我们需要手动处理 update
	var updateDoc any
	if ub, ok := update.(*UpdateBuilder); ok {
		updateDoc = ub.Build()
	} else {
		updateDoc = update
	}
	return c.coll.UpdateOne(ctx, bson.D{{Key: "_id", Value: id}}, updateDoc)
}

// DeleteByID 通过 ID 删除文档（快捷方法）
//
// 示例：
//
//	result, err := coll.DeleteByID(ctx, id)
func (c *Collection) DeleteByID(ctx context.Context, id any) (*mongo.DeleteResult, error) {
	return c.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
}

// ReplaceByID 通过 ID 替换文档（快捷方法）
//
// 示例：
//
//	result, err := coll.ReplaceByID(ctx, id, newUser)
func (c *Collection) ReplaceByID(ctx context.Context, id any, replacement any) (*mongo.UpdateResult, error) {
	return c.coll.ReplaceOne(ctx, bson.D{{Key: "_id", Value: id}}, replacement)
}

// Count 统计所有文档数量（快捷方法）
//
// 示例：
//
//	count, err := coll.Count(ctx)
//
// 注意：如果需要条件计数，请使用 Query().Count()
func (c *Collection) Count(ctx context.Context) (int64, error) {
	return c.coll.EstimatedDocumentCount(ctx)
}

// Drop 删除集合
//
// 示例：
//
//	err := coll.Drop(ctx)
func (c *Collection) Drop(ctx context.Context) error {
	return c.coll.Drop(ctx)
}

// ==================== 索引操作 ====================

// CreateIndex 创建索引
//
// 示例：
//
//	// 单字段索引
//	err := coll.CreateIndex(ctx, "email", true)  // 唯一索引
//
//	// 组合索引（使用原生方式）
//	indexModel := mongo.IndexModel{
//	    Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
//	}
//	_, err := coll.Native().Indexes().CreateOne(ctx, indexModel)
func (c *Collection) CreateIndex(ctx context.Context, field string, unique bool) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: field, Value: 1}},
	}
	if unique {
		indexModel.Options = options.Index().SetUnique(true)
	}
	_, err := c.coll.Indexes().CreateOne(ctx, indexModel)
	return err
}

// DropIndex 删除索引
//
// 示例：
//
//	err := coll.DropIndex(ctx, "email_1")
func (c *Collection) DropIndex(ctx context.Context, name string) error {
	return c.coll.Indexes().DropOne(ctx, name)
}

// ListIndexes 列出所有索引
//
// 示例：
//
//	cursor, err := coll.ListIndexes(ctx)
func (c *Collection) ListIndexes(ctx context.Context) (*mongo.Cursor, error) {
	return c.coll.Indexes().List(ctx)
}

// ==================== 批量操作 ====================

// BulkWrite 批量写入操作
//
// 示例：
//
//	models := []mongo.WriteModel{
//	    mongo.NewInsertOneModel().SetDocument(user1),
//	    mongo.NewUpdateOneModel().SetFilter(...).SetUpdate(...),
//	    mongo.NewDeleteOneModel().SetFilter(...),
//	}
//	result, err := coll.BulkWrite(ctx, models)
func (c *Collection) BulkWrite(ctx context.Context, models []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	return c.coll.BulkWrite(ctx, models)
}
