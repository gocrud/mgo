package mgo

import "errors"

// 预定义的错误类型
var (
	// ErrNoDocuments 未找到文档
	ErrNoDocuments = errors.New("mgo: no documents found")

	// ErrDuplicateKey 重复键错误
	ErrDuplicateKey = errors.New("mgo: duplicate key error")

	// ErrInvalidID 无效的对象 ID
	ErrInvalidID = errors.New("mgo: invalid object id")

	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("mgo: connection failed")

	// ErrTimeout 操作超时
	ErrTimeout = errors.New("mgo: operation timeout")

	// ErrSoftDeleteNotEnabled 软删除未启用
	ErrSoftDeleteNotEnabled = errors.New("mgo: soft delete not enabled")

	// ErrInvalidOperation 无效操作
	ErrInvalidOperation = errors.New("mgo: invalid operation")

	// ErrNilFilter 过滤器为空
	ErrNilFilter = errors.New("mgo: filter cannot be nil")

	// ErrNilUpdate 更新文档为空
	ErrNilUpdate = errors.New("mgo: update cannot be nil")
)

// IsNoDocuments 检查是否是未找到文档错误
//
// 示例：
//
//	err := coll.FindOne(...)
//	if mgo.IsNoDocuments(err) {
//	    // 处理未找到文档的情况
//	}
func IsNoDocuments(err error) bool {
	return errors.Is(err, ErrNoDocuments)
}

// IsDuplicateKey 检查是否是重复键错误
//
// 示例：
//
//	err := coll.InsertOne(...)
//	if mgo.IsDuplicateKey(err) {
//	    // 处理重复键错误
//	}
func IsDuplicateKey(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// IsTimeout 检查是否是超时错误
//
// 示例：
//
//	err := coll.Find(...)
//	if mgo.IsTimeout(err) {
//	    // 处理超时错误
//	}
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}
