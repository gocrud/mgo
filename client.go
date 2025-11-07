package mgo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Client MongoDB 客户端封装
//
// 提供优雅的数据库和集合访问接口，以及事务管理
//
// 示例：
//
//	// 连接数据库
//	client, err := mgo.NewClient(ctx, "mongodb://localhost:27017")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Disconnect(ctx)
//
//	// 访问数据库
//	db := client.Database("myapp")
//
//	// 访问集合
//	users := db.Collection("users")
//
//	// 快捷访问
//	users := client.DB("myapp").Coll("users")
type Client struct {
	client *mongo.Client
}

// NewClient 创建新的 MongoDB 客户端并连接
//
// 自动执行 ping 验证连接
//
// 示例：
//
//	client, err := mgo.NewClient(ctx, "mongodb://localhost:27017")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Disconnect(ctx)
func NewClient(ctx context.Context, uri string, opts ...*options.ClientOptions) (*Client, error) {
	clientOpts := options.Client().ApplyURI(uri)
	for _, opt := range opts {
		clientOpts = options.MergeClientOptions(clientOpts, opt)
	}

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Ping 验证连接
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return &Client{client: client}, nil
}

// MustNewClient 连接到 MongoDB，失败时 panic
//
// 示例：
//
//	client := mgo.MustNewClient(ctx, "mongodb://localhost:27017")
//	defer client.Disconnect(ctx)
func MustNewClient(ctx context.Context, uri string, opts ...*options.ClientOptions) *Client {
	client, err := NewClient(ctx, uri, opts...)
	if err != nil {
		panic(err)
	}
	return client
}

// WrapClient 包装原生 mongo.Client
//
// 用于将现有的 mongo.Client 包装为 mgo.Client
//
// 示例：
//
//	nativeClient, _ := mongo.Connect(opts)
//	client := mgo.WrapClient(nativeClient)
func WrapClient(client *mongo.Client) *Client {
	return &Client{client: client}
}

// Database 获取数据库封装
//
// 示例：
//
//	db := client.Database("myapp")
//	users := db.Collection("users")
func (c *Client) Database(name string) *Database {
	return &Database{
		db:     c.client.Database(name),
		client: c,
	}
}

// DB Database 的简写形式
//
// 示例：
//
//	users := client.DB("myapp").Coll("users")
func (c *Client) DB(name string) *Database {
	return c.Database(name)
}

// WithTransaction 执行事务
//
// 自动处理事务的开始、提交和回滚
//
// 示例：
//
//	err := client.WithTransaction(ctx, func(ctx context.Context) error {
//	    db := client.Database("myapp")
//	    users := db.Collection("users")
//	    orders := db.Collection("orders")
//
//	    _, err := users.InsertOne(ctx, user)
//	    if err != nil {
//	        return err
//	    }
//
//	    _, err = orders.InsertOne(ctx, order)
//	    return err
//	})
func (c *Client) WithTransaction(ctx context.Context, fn TransactionFunc, opts ...*options.TransactionOptionsBuilder) error {
	session, err := c.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx context.Context) (any, error) {
		return nil, fn(sessCtx)
	}

	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	_, err = session.WithTransaction(ctx, callback, txOpts)
	return err
}

// StartTransaction 开始一个新事务
//
// 返回 SessionTransaction 用于手动控制事务
//
// 示例：
//
//	txn, err := client.StartTransaction(ctx)
//	if err != nil {
//	    return err
//	}
//	defer txn.EndSession()
//
//	// 执行操作
//	_, err = users.InsertOne(txn.Context(), user)
//	if err != nil {
//	    txn.Abort()
//	    return err
//	}
//
//	return txn.Commit()
func (c *Client) StartTransaction(ctx context.Context, opts ...*options.TransactionOptionsBuilder) (*SessionTransaction, error) {
	session, err := c.client.StartSession()
	if err != nil {
		return nil, err
	}

	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	if err := session.StartTransaction(txOpts); err != nil {
		session.EndSession(ctx)
		return nil, err
	}

	return &SessionTransaction{
		session: session,
		ctx:     mongo.NewSessionContext(ctx, session),
	}, nil
}

// Ping 测试连接
//
// 示例：
//
//	err := client.Ping(ctx)
//	if err != nil {
//	    log.Fatal("连接失败:", err)
//	}
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

// Disconnect 断开连接
//
// 示例：
//
//	defer client.Disconnect(ctx)
func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Native 返回原生 mongo.Client
//
// 用于需要直接访问底层 API 的场景
//
// 示例：
//
//	nativeClient := client.Native()
func (c *Client) Native() *mongo.Client {
	return c.client
}

// Database 数据库封装
//
// 提供便捷的集合访问
//
// 示例：
//
//	db := client.Database("myapp")
//	users := db.Collection("users")
//	orders := db.Collection("orders")
type Database struct {
	db     *mongo.Database
	client *Client
}

// Collection 获取集合
//
// 示例：
//
//	users := db.Collection("users")
//
//	// 启用软删除
//	users := db.Collection("users", mgo.WithSoftDelete())
//
//	// 自定义软删除字段
//	users := db.Collection("users", mgo.WithSoftDelete("removed_at"))
func (d *Database) Collection(name string, opts ...CollectionOption) *Collection {
	return newCollection(d.db.Collection(name), opts...)
}

// Coll Collection 的简写形式
//
// 示例：
//
//	users := db.Coll("users")
//
//	// 启用软删除
//	users := db.Coll("users", mgo.WithSoftDelete())
func (d *Database) Coll(name string, opts ...CollectionOption) *Collection {
	return d.Collection(name, opts...)
}

// Name 获取数据库名称
//
// 示例：
//
//	name := db.Name()
func (d *Database) Name() string {
	return d.db.Name()
}

// Drop 删除数据库
//
// 示例：
//
//	err := db.Drop(ctx)
func (d *Database) Drop(ctx context.Context) error {
	return d.db.Drop(ctx)
}

// ListCollectionNames 列出所有集合名称
//
// 示例：
//
//	names, err := db.ListCollectionNames(ctx)
//	for _, name := range names {
//	    fmt.Println(name)
//	}
func (d *Database) ListCollectionNames(ctx context.Context) ([]string, error) {
	return d.db.ListCollectionNames(ctx, nil)
}

// Client 获取客户端
//
// 示例：
//
//	client := db.Client()
func (d *Database) Client() *Client {
	return d.client
}

// Native 返回原生 mongo.Database
//
// 示例：
//
//	nativeDB := db.Native()
func (d *Database) Native() *mongo.Database {
	return d.db
}
