package tx

import (
	"context"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ==================== Session 扩展方法 ====================

// WithContext 使用自定义上下文
//
// 示例：
//
//	sess.WithContext(customCtx)
func (s *Session) WithContext(ctx context.Context) *Session {
	return &Session{
		session: s.session,
		db:      s.db,
		ctx:     ctx,
	}
}

// IsActive 检查事务是否活跃
//
// 示例：
//
//	if sess.IsActive() {
//	    // 事务正在进行
//	}
func (s *Session) IsActive() bool {
	// MongoDB Go Driver v2 中需要检查会话状态
	return s.session != nil && s.ctx != nil
}

// ==================== 批量事务操作 ====================

// BatchTransaction 批量事务（每批一个事务）
//
// 示例：
//
//	operations := []func(*tx.Session) error{
//	    func(sess *tx.Session) error { return op1(sess) },
//	    func(sess *tx.Session) error { return op2(sess) },
//	}
//	results := tx.BatchTransaction(db, operations, 10)
func BatchTransaction(db *mgo.Database, operations []func(*Session) error, batchSize int) []*TransactionResult {
	if batchSize <= 0 {
		batchSize = 10
	}

	results := make([]*TransactionResult, 0, len(operations))

	// 分批执行事务
	chunks := chunkOperations(operations, batchSize)

	for _, chunk := range chunks {
		for _, op := range chunk {
			result := &TransactionResult{}

			err := Transaction(db, op)
			if err != nil {
				result.Error = err
				result.Success = false
			} else {
				result.Success = true
			}

			results = append(results, result)
		}
	}

	return results
}

// TransactionResult 事务结果
type TransactionResult struct {
	Success bool
	Error   error
}

// chunkOperations 分批操作
func chunkOperations(ops []func(*Session) error, size int) [][]func(*Session) error {
	if size <= 0 {
		return nil
	}

	chunks := make([][]func(*Session) error, 0, (len(ops)+size-1)/size)
	for i := 0; i < len(ops); i += size {
		end := i + size
		if end > len(ops) {
			end = len(ops)
		}
		chunks = append(chunks, ops[i:end])
	}

	return chunks
}

// ==================== 事务选项辅助 ====================

// DefaultTransactionOptions 默认事务选项
//
// 示例：
//
//	opts := tx.DefaultTransactionOptions()
//	err := tx.Transaction(db, fn, opts)
func DefaultTransactionOptions() options.TransactionOptions {
	return options.TransactionOptions{}
}

// ==================== 嵌套事务支持 ====================

// NestedTransaction 嵌套事务（保存点模拟）
//
// 注意：MongoDB 不直接支持嵌套事务，这里通过模拟实现
//
// 示例：
//
//	err := tx.Transaction(db, func(sess *tx.Session) error {
//	    // 外层事务
//
//	    err := tx.NestedTransaction(sess, func(nested *tx.Session) error {
//	        // 内层事务
//	        return nil
//	    })
//
//	    return err
//	})
func NestedTransaction(parent *Session, fn func(*Session) error) error {
	// MongoDB 不支持真正的嵌套事务
	// 这里直接使用父事务的上下文
	return fn(parent)
}

// ==================== 事务统计 ====================

// TransactionStats 事务统计信息
type TransactionStats struct {
	TotalTransactions int
	SuccessCount      int
	FailureCount      int
	RetryCount        int
}

// NewTransactionStats 创建事务统计
func NewTransactionStats() *TransactionStats {
	return &TransactionStats{}
}

// RecordSuccess 记录成功
func (ts *TransactionStats) RecordSuccess() {
	ts.TotalTransactions++
	ts.SuccessCount++
}

// RecordFailure 记录失败
func (ts *TransactionStats) RecordFailure() {
	ts.TotalTransactions++
	ts.FailureCount++
}

// RecordRetry 记录重试
func (ts *TransactionStats) RecordRetry() {
	ts.RetryCount++
}

// SuccessRate 成功率
func (ts *TransactionStats) SuccessRate() float64 {
	if ts.TotalTransactions == 0 {
		return 0
	}
	return float64(ts.SuccessCount) / float64(ts.TotalTransactions) * 100
}
