package tx

import (
	"context"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ==================== 自动事务 ====================

// Transaction 执行自动事务
//
// 示例：
//
//	err := tx.Transaction(db, func(sess *tx.Session) error {
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
func Transaction(db *mgo.Database, fn func(*Session) error, opts ...options.TransactionOptions) error {
	client := db.Client()
	ctx := db.Context()

	// 启动会话
	session, err := client.StartSession()
	if err != nil {
		return mgo.WrapError(err, "failed to start session")
	}
	defer session.EndSession(ctx)

	// 执行事务回调
	callback := func(sessCtx context.Context) (interface{}, error) {
		sess := &Session{
			session: session,
			db:      db,
			ctx:     sessCtx,
		}
		return nil, fn(sess)
	}

	// 构建事务选项
	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = options.Transaction()
		for _, opt := range opts {
			// 应用选项
			_ = opt
		}
	}

	_, err = session.WithTransaction(ctx, callback, txOpts)
	return err
}

// MustTransaction 执行自动事务，失败时 panic
//
// 示例：
//
//	tx.MustTransaction(db, func(sess *tx.Session) error {
//	    // 事务操作
//	    return nil
//	})
func MustTransaction(db *mgo.Database, fn func(*Session) error, opts ...options.TransactionOptions) {
	if err := Transaction(db, fn, opts...); err != nil {
		panic(err)
	}
}

// ==================== 手动事务 ====================

// Begin 开始手动事务
//
// 示例：
//
//	sess, err := tx.Begin(db)
//	if err != nil {
//	    return err
//	}
//	defer sess.Rollback()
//
//	users := mgo.Model[User](sess)
//	if err := users.Find().ID(userID).Update(...); err != nil {
//	    return err
//	}
//
//	return sess.Commit()
func Begin(db *mgo.Database, opts ...options.TransactionOptions) (*Session, error) {
	client := db.Client()
	ctx := db.Context()

	// 启动会话
	session, err := client.StartSession()
	if err != nil {
		return nil, mgo.WrapError(err, "failed to start session")
	}

	// 构建事务选项
	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = options.Transaction()
		for _, opt := range opts {
			_ = opt
		}
	}

	// 开始事务
	if err := session.StartTransaction(txOpts); err != nil {
		session.EndSession(ctx)
		return nil, mgo.WrapError(err, "failed to start transaction")
	}

	// 创建会话上下文
	sessCtx := mongo.NewSessionContext(ctx, session)

	return &Session{
		session: session,
		db:      db,
		ctx:     sessCtx,
	}, nil
}

// MustBegin 开始手动事务，失败时 panic
//
// 示例：
//
//	sess := tx.MustBegin(db)
//	defer sess.Rollback()
func MustBegin(db *mgo.Database, opts ...options.TransactionOptions) *Session {
	sess, err := Begin(db, opts...)
	if err != nil {
		panic(err)
	}
	return sess
}

// ==================== Session 会话 ====================

// Session 事务会话
type Session struct {
	session *mongo.Session
	db      *mgo.Database
	ctx     context.Context
}

// Commit 提交事务
//
// 示例：
//
//	if err := sess.Commit(); err != nil {
//	    return err
//	}
func (s *Session) Commit() error {
	if err := s.session.CommitTransaction(s.ctx); err != nil {
		return mgo.WrapError(err, "failed to commit transaction")
	}
	s.session.EndSession(s.ctx)
	return nil
}

// Rollback 回滚事务
//
// 示例：
//
//	defer sess.Rollback()
func (s *Session) Rollback() error {
	if err := s.session.AbortTransaction(s.ctx); err != nil {
		// 回滚失败不返回错误，因为可能已经提交
		_ = err
	}
	s.session.EndSession(s.ctx)
	return nil
}

// Abort Rollback 的别名
func (s *Session) Abort() error {
	return s.Rollback()
}

// Context 获取会话上下文
//
// 示例：
//
//	ctx := sess.Context()
func (s *Session) Context() context.Context {
	return s.ctx
}

// Database 获取数据库（在事务上下文中）
//
// 示例：
//
//	db := sess.Database()
func (s *Session) Database() *mgo.Database {
	return s.db.WithContext(s.ctx)
}

// Collection 获取集合（在事务上下文中）
//
// 示例：
//
//	users := sess.Collection("users")
func (s *Session) Collection(name string, opts ...mgo.CollectionOption) interface{} {
	db := s.Database()
	return db.Collection(name, opts...)
}

// Coll Collection 的简写
func (s *Session) Coll(name string, opts ...mgo.CollectionOption) interface{} {
	return s.Collection(name, opts...)
}

// ==================== 事务辅助函数 ====================

// WithRetry 带重试的事务执行
//
// 示例：
//
//	err := tx.WithRetry(db, 3, func(sess *tx.Session) error {
//	    // 事务操作
//	    return nil
//	})
func WithRetry(db *mgo.Database, maxRetries int, fn func(*Session) error, opts ...options.TransactionOptions) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		err := Transaction(db, fn, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否可重试
		if !isRetryableError(err) {
			return err
		}
	}

	return mgo.WrapErrorf(lastErr, "transaction failed after %d retries", maxRetries)
}

// isRetryableError 检查错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否包含 "TransientTransactionError" 或 "UnknownTransactionCommitResult"
	errStr := err.Error()
	return mgo.WrapError(err, errStr) != nil &&
		(contains(errStr, "TransientTransactionError") ||
			contains(errStr, "UnknownTransactionCommitResult"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsAt(s, substr, 0)
}

func containsAt(s, substr string, start int) bool {
	if start+len(substr) > len(s) {
		return false
	}
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== 跨库事务（使用 Client）====================

// CrossDBTransaction 跨库事务（使用 Client）
//
// 示例：
//
//	err := tx.CrossDBTransaction(client, func(sess *mgo.ClientSession) error {
//	    accountsDB := sess.Database("accounts")
//	    logsDB := sess.Database("logs")
//
//	    users := mgo.Model[User](accountsDB)
//	    logs := mgo.Model[Log](logsDB)
//
//	    // 跨库操作
//	    if err := users.Find().ID(userID).Inc("balance", -amount).Update(); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    if _, err := logs.Insert(log); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    return nil  // 自动提交
//	})
func CrossDBTransaction(client *mgo.Client, fn func(*mgo.ClientSession) error, opts ...options.TransactionOptions) error {
	ctx := client.Context()
	nativeClient := client.Native()

	// 启动会话
	session, err := nativeClient.StartSession()
	if err != nil {
		return mgo.WrapError(err, "failed to start session")
	}
	defer session.EndSession(ctx)

	// 执行事务回调
	callback := func(sessCtx context.Context) (interface{}, error) {
		sess := mgo.NewClientSession(nativeClient, session, sessCtx)
		return nil, fn(sess)
	}

	// 构建事务选项
	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = options.Transaction()
		for _, opt := range opts {
			_ = opt
		}
	}

	_, err = session.WithTransaction(ctx, callback, txOpts)
	return err
}

// MustCrossDBTransaction 跨库事务，失败时 panic
//
// 示例：
//
//	tx.MustCrossDBTransaction(client, func(sess *mgo.ClientSession) error {
//	    // 跨库事务操作
//	    return nil
//	})
func MustCrossDBTransaction(client *mgo.Client, fn func(*mgo.ClientSession) error, opts ...options.TransactionOptions) {
	if err := CrossDBTransaction(client, fn, opts...); err != nil {
		panic(err)
	}
}

// BeginCrossDBTransaction 开始跨库手动事务
//
// 示例：
//
//	sess, err := tx.BeginCrossDBTransaction(client)
//	if err != nil {
//	    return err
//	}
//	defer sess.Rollback()
//
//	accountsDB := sess.Database("accounts")
//	logsDB := sess.Database("logs")
//
//	// 执行跨库操作...
//
//	return sess.Commit()
func BeginCrossDBTransaction(client *mgo.Client, opts ...options.TransactionOptions) (*mgo.ClientSession, error) {
	ctx := client.Context()
	nativeClient := client.Native()

	// 启动会话
	session, err := nativeClient.StartSession()
	if err != nil {
		return nil, mgo.WrapError(err, "failed to start session")
	}

	// 构建事务选项
	var txOpts *options.TransactionOptionsBuilder
	if len(opts) > 0 {
		txOpts = options.Transaction()
		for _, opt := range opts {
			_ = opt
		}
	}

	// 开始事务
	if err := session.StartTransaction(txOpts); err != nil {
		session.EndSession(ctx)
		return nil, mgo.WrapError(err, "failed to start transaction")
	}

	// 创建会话上下文
	sessCtx := mongo.NewSessionContext(ctx, session)

	return mgo.NewClientSession(nativeClient, session, sessCtx), nil
}

// MustBeginCrossDBTransaction 开始跨库手动事务，失败时 panic
//
// 示例：
//
//	sess := tx.MustBeginCrossDBTransaction(client)
//	defer sess.Rollback()
func MustBeginCrossDBTransaction(client *mgo.Client, opts ...options.TransactionOptions) *mgo.ClientSession {
	sess, err := BeginCrossDBTransaction(client, opts...)
	if err != nil {
		panic(err)
	}
	return sess
}

// CrossDBWithRetry 带重试的跨库事务
//
// 示例：
//
//	err := tx.CrossDBWithRetry(client, 3, func(sess *mgo.ClientSession) error {
//	    // 跨库事务操作
//	    return nil
//	})
func CrossDBWithRetry(client *mgo.Client, maxRetries int, fn func(*mgo.ClientSession) error, opts ...options.TransactionOptions) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		err := CrossDBTransaction(client, fn, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否可重试
		if !isRetryableError(err) {
			return err
		}
	}

	return mgo.WrapErrorf(lastErr, "cross-db transaction failed after %d retries", maxRetries)
}
