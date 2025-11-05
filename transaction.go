package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// TransactionFunc 事务回调函数类型
//
// 在此函数中执行需要事务保护的数据库操作
//
// 示例：
//
//	fn := func(ctx context.Context) error {
//	    _, err := coll.InsertOne(ctx, doc1)
//	    if err != nil {
//	        return err
//	    }
//	    _, err = coll.UpdateOne(ctx, filter, update)
//	    return err
//	}
type TransactionFunc func(ctx context.Context) error

// SessionTransaction 事务封装器
//
// 提供更灵活的事务控制，支持手动提交和回滚
//
// 示例：
//
//	txn, err := client.StartTransaction(ctx)
//	if err != nil {
//	    return err
//	}
//	defer txn.EndSession()
//
//	// 执行操作1
//	_, err = coll1.InsertOne(txn.Context(), doc1)
//	if err != nil {
//	    txn.Abort()
//	    return err
//	}
//
//	// 执行操作2
//	_, err = coll2.UpdateOne(txn.Context(), filter, update)
//	if err != nil {
//	    txn.Abort()
//	    return err
//	}
//
//	// 提交事务
//	return txn.Commit()
type SessionTransaction struct {
	session *mongo.Session
	ctx     context.Context
}

// Context 获取带会话的上下文
//
// 在事务中执行数据库操作时，必须使用此上下文
//
// 示例：
//
//	_, err := coll.InsertOne(txn.Context(), doc)
func (st *SessionTransaction) Context() context.Context {
	return st.ctx
}

// Commit 提交事务
//
// 示例：
//
//	err := txn.Commit()
//	if err != nil {
//	    return err
//	}
func (st *SessionTransaction) Commit() error {
	return st.session.CommitTransaction(st.ctx)
}

// Abort 回滚事务
//
// 示例：
//
//	if err != nil {
//	    txn.Abort()
//	    return err
//	}
func (st *SessionTransaction) Abort() error {
	return st.session.AbortTransaction(st.ctx)
}

// EndSession 结束会话
//
// 应该在 defer 中调用
//
// 示例：
//
//	txn, err := client.StartTransaction(ctx)
//	if err != nil {
//	    return err
//	}
//	defer txn.EndSession()
func (st *SessionTransaction) EndSession() {
	st.session.EndSession(st.ctx)
}

// TransactionOptions 事务选项构建器
//
// 用于配置事务的读关注、写关注等选项
//
// 示例：
//
//	opts := mgo.TransactionOptions().
//	    SetReadConcern(readconcern.Majority()).
//	    SetWriteConcern(writeconcern.Majority())
//
//	err := client.WithTransaction(ctx, fn, opts)
func TransactionOptions() *options.TransactionOptionsBuilder {
	return options.Transaction()
}
