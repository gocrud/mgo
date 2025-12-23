package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Database MongoDB 数据库封装
type Database struct {
	client *mongo.Client
	db     *mongo.Database
	name   string
	ctx    context.Context
}

// WithContext 返回带有新上下文的 Database 副本
func (d *Database) WithContext(ctx context.Context) *Database {
	return &Database{
		client: d.client,
		db:     d.db,
		name:   d.name,
		ctx:    ctx,
	}
}

// Context 返回当前上下文
func (d *Database) Context() context.Context {
	if d.ctx == nil {
		return context.Background()
	}
	return d.ctx
}

// Collection 获取集合 (默认返回 bson.M 类型)
func (d *Database) Collection(name string, opts ...CollectionOption) *Collection[M] {
	return newCollection[M](d, d.db.Collection(name), opts...).WithContext(d.Context())
}

// Name 获取数据库名称
func (d *Database) Name() string {
	return d.name
}

// Client 获取原生 mongo.Client
func (d *Database) Client() *mongo.Client {
	return d.client
}

// Native 返回原生 mongo.Database
func (d *Database) Native() *mongo.Database {
	return d.db
}

// Drop 删除数据库
func (d *Database) Drop() error {
	return d.db.Drop(d.Context())
}

// ListCollectionNames 列出所有集合名称
func (d *Database) ListCollectionNames() ([]string, error) {
	return d.db.ListCollectionNames(d.Context(), M{})
}

// Ping 测试数据库连接
func (d *Database) Ping() error {
	return d.client.Ping(d.Context(), nil)
}

// Close 关闭数据库连接
func (d *Database) Close(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

// Disconnect Close 的别名
func (d *Database) Disconnect(ctx context.Context) error {
	return d.Close(ctx)
}

// ==================== 事务相关方法 ====================

// Transaction 执行自动事务
func (d *Database) Transaction(fn func(*Session) error) error {
	ctx := d.Context()
	session, err := d.client.StartSession()
	if err != nil {
		return WrapError(err, "failed to start session")
	}
	defer session.EndSession(ctx)

	// 执行事务回调
	callback := func(sessCtx context.Context) (interface{}, error) {
		sess := &Session{
			session: session,
			db:      d,
			ctx:     sessCtx,
		}
		return nil, fn(sess)
	}

	_, err = session.WithTransaction(ctx, callback)
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
	return s.db
}

// Collection 获取集合（在事务上下文中）
func (s *Session) Collection(name string, opts ...CollectionOption) *Collection[M] {
	return newCollection[M](s.db, s.db.db.Collection(name), opts...)
}

// Coll Collection 的简写形式
func (s *Session) Coll(name string, opts ...CollectionOption) *Collection[M] {
	return s.Collection(name, opts...)
}

// ==================== ClientSession 跨库事务会话 ====================

// ClientSession 客户端级别的事务会话（支持跨库）
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
func (cs *ClientSession) Database(name string) *Database {
	return &Database{
		client: cs.client,
		db:     cs.client.Database(name),
		name:   name,
	}
}

// Context 获取会话上下文
func (cs *ClientSession) Context() context.Context {
	return cs.ctx
}

// Commit 手动提交事务
func (cs *ClientSession) Commit() error {
	if err := cs.session.CommitTransaction(cs.ctx); err != nil {
		return WrapError(err, "failed to commit transaction")
	}
	cs.session.EndSession(cs.ctx)
	return nil
}

// Rollback 手动回滚事务
func (cs *ClientSession) Rollback() error {
	if err := cs.session.AbortTransaction(cs.ctx); err != nil {
		_ = err
	}
	cs.session.EndSession(cs.ctx)
	return nil
}

// Abort Rollback 的别名
func (cs *ClientSession) Abort() error {
	return cs.Rollback()
}

// Databases 同时获取多个数据库
func (cs *ClientSession) Databases(names ...string) []*Database {
	dbs := make([]*Database, len(names))
	for i, name := range names {
		dbs[i] = cs.Database(name)
	}
	return dbs
}

// IsActive 检查事务是否活跃
func (cs *ClientSession) IsActive() bool {
	return cs.session != nil && cs.ctx != nil
}
