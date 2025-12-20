package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== Database 数据库封装 ====================

// Database MongoDB 数据库封装
//
// 提供便捷的集合访问和事务管理
//
// 示例：
//
//	db := mgo.MustOpen("mongodb://localhost/myapp")
//	defer db.Close()
//
//	users := db.Collection("users")
//	orders := db.Collection("orders")
type Database struct {
	client *mongo.Client
	db     *mongo.Database
	ctx    context.Context
	name   string
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
	return newCollection(d, d.db.Collection(name), opts...)
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
	return d.name
}

// Client 获取原生 mongo.Client
//
// 示例：
//
//	client := db.Client()
func (d *Database) Client() *mongo.Client {
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

// Context 获取默认上下文
//
// 示例：
//
//	ctx := db.Context()
func (d *Database) Context() context.Context {
	return getContext(d.ctx)
}

// WithContext 设置默认上下文并返回新的 Database
//
// 示例：
//
//	ctx := context.WithTimeout(context.Background(), 5*time.Second)
//	db = db.WithContext(ctx)
func (d *Database) WithContext(ctx context.Context) *Database {
	return &Database{
		client: d.client,
		db:     d.db,
		ctx:    ctx,
		name:   d.name,
	}
}

// Drop 删除数据库
//
// 示例：
//
//	err := db.Drop()
func (d *Database) Drop() error {
	return d.db.Drop(d.Context())
}

// ListCollectionNames 列出所有集合名称
//
// 示例：
//
//	names, err := db.ListCollectionNames()
//	for _, name := range names {
//	    fmt.Println(name)
//	}
func (d *Database) ListCollectionNames() ([]string, error) {
	return d.db.ListCollectionNames(d.Context(), M{})
}

// Ping 测试数据库连接
//
// 示例：
//
//	err := db.Ping()
//	if err != nil {
//	    log.Fatal("连接失败:", err)
//	}
func (d *Database) Ping() error {
	return d.client.Ping(d.Context(), nil)
}

// Close 关闭数据库连接
//
// 示例：
//
//	defer db.Close()
func (d *Database) Close() error {
	return d.client.Disconnect(d.Context())
}

// Disconnect Close 的别名
//
// 示例：
//
//	defer db.Disconnect()
func (d *Database) Disconnect() error {
	return d.Close()
}

// ==================== 事务相关方法 ====================

// Transaction 执行自动事务
//
// 示例：
//
//	err := db.Transaction(func(sess *mgo.Session) error {
//	    users := mgo.Model[User](sess)
//	    orders := mgo.Model[Order](sess)
//
//	    if err := users.Find().ID(userID).Inc("balance", -100).Update(); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    if _, err := orders.Insert(order); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    return nil  // 自动提交
//	})
func (d *Database) Transaction(fn func(*Session) error) error {
	session, err := d.client.StartSession()
	if err != nil {
		return WrapError(err, "failed to start session")
	}
	defer session.EndSession(d.Context())

	// 执行事务回调
	callback := func(ctx context.Context) (interface{}, error) {
		sess := &Session{
			session: session,
			db:      d,
			ctx:     ctx,
		}
		return nil, fn(sess)
	}

	_, err = session.WithTransaction(d.Context(), callback)
	return err
}

// Session 事务会话
type Session struct {
	session *mongo.Session
	db      *Database
	ctx     context.Context
}

// Context 获取会话上下文
func (s *Session) Context() context.Context {
	return s.ctx
}

// Database 返回数据库（在事务上下文中）
func (s *Session) Database() *Database {
	return &Database{
		client: s.db.client,
		db:     s.db.db,
		ctx:    s.ctx,
		name:   s.db.name,
	}
}

// Collection 获取集合（在事务上下文中）
func (s *Session) Collection(name string, opts ...CollectionOption) *Collection {
	return newCollection(s.Database(), s.db.db.Collection(name), opts...)
}

// Coll Collection 的简写形式
func (s *Session) Coll(name string, opts ...CollectionOption) *Collection {
	return s.Collection(name, opts...)
}

// ==================== ClientSession 跨库事务会话 ====================

// ClientSession 客户端级别的事务会话（支持跨库）
//
// 用于跨库事务，可以访问任意数据库
//
// 示例：
//
//	err := client.Transaction(func(sess *mgo.ClientSession) error {
//	    accountsDB := sess.Database("accounts")
//	    logsDB := sess.Database("logs")
//	    // 跨库操作
//	    return nil
//	})
type ClientSession struct {
	client  *mongo.Client
	session *mongo.Session
	ctx     context.Context
}

// NewClientSession 创建客户端会话（内部使用）
func NewClientSession(client *mongo.Client, session *mongo.Session, ctx context.Context) *ClientSession {
	return &ClientSession{
		client:  client,
		session: session,
		ctx:     ctx,
	}
}

// Database 获取指定数据库（在事务上下文中）
//
// 示例：
//
//	accountsDB := sess.Database("accounts")
//	logsDB := sess.Database("logs")
func (cs *ClientSession) Database(name string) *Database {
	return &Database{
		client: cs.client,
		db:     cs.client.Database(name),
		ctx:    cs.ctx,
		name:   name,
	}
}

// Context 获取会话上下文
//
// 示例：
//
//	ctx := sess.Context()
func (cs *ClientSession) Context() context.Context {
	return cs.ctx
}

// Commit 手动提交事务（用于手动事务）
//
// 示例：
//
//	if err := sess.Commit(); err != nil {
//	    return err
//	}
func (cs *ClientSession) Commit() error {
	if err := cs.session.CommitTransaction(cs.ctx); err != nil {
		return WrapError(err, "failed to commit transaction")
	}
	cs.session.EndSession(cs.ctx)
	return nil
}

// Rollback 手动回滚事务
//
// 示例：
//
//	defer sess.Rollback()
func (cs *ClientSession) Rollback() error {
	if err := cs.session.AbortTransaction(cs.ctx); err != nil {
		// 回滚失败不返回错误，因为可能已经提交
		_ = err
	}
	cs.session.EndSession(cs.ctx)
	return nil
}

// Abort Rollback 的别名
//
// 示例：
//
//	defer sess.Abort()
func (cs *ClientSession) Abort() error {
	return cs.Rollback()
}

// Databases 同时获取多个数据库
//
// 示例：
//
//	dbs := sess.Databases("accounts", "logs", "orders")
//	accountsDB := dbs[0]
//	logsDB := dbs[1]
//	ordersDB := dbs[2]
func (cs *ClientSession) Databases(names ...string) []*Database {
	dbs := make([]*Database, len(names))
	for i, name := range names {
		dbs[i] = cs.Database(name)
	}
	return dbs
}

// IsActive 检查事务是否活跃
//
// 示例：
//
//	if sess.IsActive() {
//	    // 事务正在进行
//	}
func (cs *ClientSession) IsActive() bool {
	return cs.session != nil && cs.ctx != nil
}
