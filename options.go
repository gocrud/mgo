package mgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ==================== 客户端选项 ====================

// ClientOption 客户端配置选项
type ClientOption func(*clientOptions)

type clientOptions struct {
	uri         string
	timeout     time.Duration
	maxPoolSize uint64
	minPoolSize uint64
	retryWrites bool
	retryReads  bool
	ctx         context.Context
}

// URI 设置连接 URI
//
// 示例：
//
//	db, err := mgo.Connect(mgo.URI("mongodb://localhost:27017/myapp"))
func URI(uri string) ClientOption {
	return func(opts *clientOptions) {
		opts.uri = uri
	}
}

// Timeout 设置连接超时时间
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.Timeout(10*time.Second),
//	)
func Timeout(d time.Duration) ClientOption {
	return func(opts *clientOptions) {
		opts.timeout = d
	}
}

// MaxPoolSize 设置最大连接池大小
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.MaxPoolSize(100),
//	)
func MaxPoolSize(size uint64) ClientOption {
	return func(opts *clientOptions) {
		opts.maxPoolSize = size
	}
}

// MinPoolSize 设置最小连接池大小
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.MinPoolSize(10),
//	)
func MinPoolSize(size uint64) ClientOption {
	return func(opts *clientOptions) {
		opts.minPoolSize = size
	}
}

// RetryWrites 设置是否重试写操作
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.RetryWrites(true),
//	)
func RetryWrites(retry bool) ClientOption {
	return func(opts *clientOptions) {
		opts.retryWrites = retry
	}
}

// RetryReads 设置是否重试读操作
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.RetryReads(true),
//	)
func RetryReads(retry bool) ClientOption {
	return func(opts *clientOptions) {
		opts.retryReads = retry
	}
}

// WithContext 设置默认上下文
//
// 示例：
//
//	ctx := context.WithTimeout(context.Background(), 30*time.Second)
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.WithContext(ctx),
//	)
func WithContext(ctx context.Context) ClientOption {
	return func(opts *clientOptions) {
		opts.ctx = ctx
	}
}

// buildClientOptions 构建 mongo.ClientOptions
func buildClientOptions(opts *clientOptions) *options.ClientOptions {
	clientOpts := options.Client().ApplyURI(opts.uri)

	if opts.timeout > 0 {
		clientOpts.SetConnectTimeout(opts.timeout)
	}

	if opts.maxPoolSize > 0 {
		clientOpts.SetMaxPoolSize(opts.maxPoolSize)
	}

	if opts.minPoolSize > 0 {
		clientOpts.SetMinPoolSize(opts.minPoolSize)
	}

	if opts.retryWrites {
		clientOpts.SetRetryWrites(true)
	}

	if opts.retryReads {
		clientOpts.SetRetryReads(true)
	}

	return clientOpts
}

// ==================== 集合选项 ====================

// CollectionOption 集合配置选项
type CollectionOption func(*CollectionOptions)

// CollectionOptions 集合配置
type CollectionOptions struct {
	Timestamps *TimestampConfig
	SoftDelete *SoftDeleteConfig
	Hooks      interface{} // 将在 hooks.go 中定义具体类型
	Context    context.Context
}

// WithTimestamps 启用自动时间戳
//
// 示例：
//
//	// 使用默认字段名
//	users := mgo.Model[User](db).WithTimestamps()
//
//	// 自定义字段名
//	users := mgo.Model[User](db).WithTimestamps("created_at", "updated_at")
func WithTimestamps(fields ...string) CollectionOption {
	return func(opts *CollectionOptions) {
		createdField := "created_at"
		updatedField := "updated_at"

		if len(fields) > 0 {
			createdField = fields[0]
		}
		if len(fields) > 1 {
			updatedField = fields[1]
		}

		opts.Timestamps = &TimestampConfig{
			CreatedField: createdField,
			UpdatedField: updatedField,
			Enabled:      true,
		}
	}
}

// WithSoftDelete 启用软删除
//
// 示例：
//
//	// 使用默认字段名
//	users := mgo.Model[User](db).WithSoftDelete()
//
//	// 自定义字段名
//	users := mgo.Model[User](db).WithSoftDelete("deleted_at")
func WithSoftDelete(fields ...string) CollectionOption {
	return func(opts *CollectionOptions) {
		deletedField := "deleted_at"

		if len(fields) > 0 {
			deletedField = fields[0]
		}

		opts.SoftDelete = &SoftDeleteConfig{
			Field:   deletedField,
			Enabled: true,
		}
	}
}

// WithCollectionContext 设置集合默认上下文
//
// 示例：
//
//	ctx := context.WithTimeout(context.Background(), 5*time.Second)
//	users := mgo.Model[User](db).WithCollectionContext(ctx)
func WithCollectionContext(ctx context.Context) CollectionOption {
	return func(opts *CollectionOptions) {
		opts.Context = ctx
	}
}

// ==================== 时间戳配置 ====================

// TimestampConfig 时间戳配置
type TimestampConfig struct {
	CreatedField string
	UpdatedField string
	Enabled      bool
}

// ==================== 软删除配置 ====================

// SoftDeleteConfig 软删除配置
type SoftDeleteConfig struct {
	Field   string
	Enabled bool
}

// ==================== 查询选项 ====================

// QueryOption 查询选项
type QueryOption func(*QueryOptions)

// QueryOptions 查询配置
type QueryOptions struct {
	Skip       *int64
	Limit      *int64
	Sort       M
	Projection M
	Hint       interface{}
	Context    context.Context
}

// ==================== 批量操作选项 ====================

// BatchOption 批量操作选项
type BatchOption func(*BatchOptions)

// BatchOptions 批量操作配置
type BatchOptions struct {
	Size    int
	Ordered bool
}

// BatchSize 设置批量操作的批次大小
//
// 示例：
//
//	err := batch.InsertBatch(users, largeList, mgo.BatchSize(500))
func BatchSize(size int) BatchOption {
	return func(opts *BatchOptions) {
		opts.Size = size
	}
}

// Ordered 设置批量操作是否有序
//
// 示例：
//
//	err := batch.InsertBatch(users, largeList, mgo.Ordered(false))
func Ordered(ordered bool) BatchOption {
	return func(opts *BatchOptions) {
		opts.Ordered = ordered
	}
}

// ==================== 分页选项 ====================

// PageOption 分页选项
type PageOption func(*PageOptions)

// PageOptions 分页配置
type PageOptions struct {
	DisableCount bool // 禁用总数统计（提升性能）
}

// DisableCount 禁用分页时的总数统计
//
// 示例：
//
//	page, err := users.Find().PageList(1, 20, mgo.DisableCount())
func DisableCount() PageOption {
	return func(opts *PageOptions) {
		opts.DisableCount = true
	}
}
