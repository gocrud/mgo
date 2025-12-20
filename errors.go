package mgo

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 标准错误 ====================

var (
	// ErrNoDocuments 未找到文档
	ErrNoDocuments = mongo.ErrNoDocuments

	// ErrNilDocument 文档为 nil
	ErrNilDocument = errors.New("mgo: document is nil")

	// ErrInvalidID ID 无效
	ErrInvalidID = errors.New("mgo: invalid id")

	// ErrEmptyFilter 过滤条件为空
	ErrEmptyFilter = errors.New("mgo: empty filter")

	// ErrEmptyUpdate 更新内容为空
	ErrEmptyUpdate = errors.New("mgo: empty update")

	// ErrInvalidOperation 无效操作
	ErrInvalidOperation = errors.New("mgo: invalid operation")

	// ErrAlreadyDeleted 文档已被删除
	ErrAlreadyDeleted = errors.New("mgo: document already deleted")

	// ErrNotFound 未找到
	ErrNotFound = errors.New("mgo: not found")

	// ErrDuplicateKey 重复键错误
	ErrDuplicateKey = errors.New("mgo: duplicate key")
)

// ==================== 错误检查函数 ====================

// IsNoDocuments 检查是否为未找到文档错误
//
// 示例：
//
//	user, err := users.FindByID(id)
//	if mgo.IsNoDocuments(err) {
//	    // 处理未找到的情况
//	}
func IsNoDocuments(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}

// IsDuplicateKey 检查是否为重复键错误
//
// 示例：
//
//	_, err := users.Insert(user)
//	if mgo.IsDuplicateKey(err) {
//	    // 处理重复键错误
//	}
func IsDuplicateKey(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// IsNetworkError 检查是否为网络错误
//
// 示例：
//
//	err := db.Ping()
//	if mgo.IsNetworkError(err) {
//	    // 处理网络错误
//	}
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// 检查是否包含常见的网络错误关键词
	errStr := err.Error()
	return contains(errStr, "connection", "timeout", "network", "refused")
}

// IsTimeout 检查是否为超时错误
//
// 示例：
//
//	_, err := users.Find().All()
//	if mgo.IsTimeout(err) {
//	    // 处理超时错误
//	}
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "timeout", "deadline exceeded")
}

// ==================== 错误包装函数 ====================

// WrapError 包装错误并添加上下文信息
//
// 示例：
//
//	if err != nil {
//	    return mgo.WrapError(err, "failed to insert user")
//	}
func WrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErrorf 包装错误并格式化上下文信息
//
// 示例：
//
//	if err != nil {
//	    return mgo.WrapErrorf(err, "failed to find user with id %s", id)
//	}
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, err)
}

// ==================== 自定义错误类型 ====================

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError 创建验证错误
//
// 示例：
//
//	if user.Email == "" {
//	    return mgo.NewValidationError("email", "email is required")
//	}
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// QueryError 查询错误
type QueryError struct {
	Operation string
	Err       error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("query error during %s: %v", e.Operation, e.Err)
}

func (e *QueryError) Unwrap() error {
	return e.Err
}

// NewQueryError 创建查询错误
func NewQueryError(operation string, err error) error {
	return &QueryError{
		Operation: operation,
		Err:       err,
	}
}

// ==================== 辅助函数 ====================

// contains 检查字符串是否包含任意一个子串
func contains(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
