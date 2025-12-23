package batch

import (
	"context"
	"sync"
	"time"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 缓冲区批量插入 ====================

// Buffer 批量插入缓冲区
//
// 自动在达到大小或超时时刷新
//
// 示例：
//
//	buffer := batch.NewBuffer(ctx, users, 100, 5*time.Second)
//	defer buffer.Close()
//
//	for _, user := range largeList {
//	    buffer.Add(user)
//	}
type Buffer[T any] struct {
	coll      *mongo.Collection
	ctx       context.Context
	size      int
	flushTime time.Duration
	buffer    []*T
	mu        sync.Mutex
	timer     *time.Timer
	closed    bool
	errors    []error
}

// NewBuffer 创建新的缓冲区
//
// 参数：
//   - coll: 集合
//   - size: 缓冲区大小
//   - flushTime: 刷新超时时间
//
// 示例：
//
//	buffer := batch.NewBuffer(ctx, users, 100, 5*time.Second)
func NewBuffer[T any](ctx context.Context, coll *mgo.Collection[T], size int, flushTime time.Duration) *Buffer[T] {
	if size <= 0 {
		size = 100
	}
	if flushTime <= 0 {
		flushTime = 5 * time.Second
	}

	buffer := &Buffer[T]{
		coll:      coll.Native(),
		ctx:       ctx,
		size:      size,
		flushTime: flushTime,
		buffer:    make([]*T, 0, size),
		errors:    make([]error, 0),
	}

	// 启动定时器
	buffer.timer = time.AfterFunc(flushTime, func() {
		buffer.Flush()
	})

	return buffer
}

// Add 添加文档到缓冲区
//
// 示例：
//
//	buffer.Add(&user)
func (b *Buffer[T]) Add(doc *T) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return mgo.ErrInvalidOperation
	}

	b.buffer = append(b.buffer, doc)

	// 检查是否需要刷新
	if len(b.buffer) >= b.size {
		return b.flush()
	}

	return nil
}

// Flush 手动刷新缓冲区
//
// 示例：
//
//	buffer.Flush()
func (b *Buffer[T]) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.flush()
}

func (b *Buffer[T]) flush() error {
	if len(b.buffer) == 0 {
		return nil
	}

	// 批量插入
	// 注意：这里需要将 []*T 转换为 []interface{}
	docs := make([]interface{}, len(b.buffer))
	for i, doc := range b.buffer {
		docs[i] = doc
	}

	_, err := b.coll.InsertMany(b.ctx, docs)
	if err != nil {
		b.errors = append(b.errors, err)
		return err
	}

	// 清空缓冲区
	b.buffer = b.buffer[:0]

	// 重置定时器
	if b.timer != nil {
		b.timer.Reset(b.flushTime)
	}

	return nil
}

// Close 关闭缓冲区（并刷新剩余数据）
//
// 示例：
//
//	defer buffer.Close()
func (b *Buffer[T]) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	if b.timer != nil {
		b.timer.Stop()
	}

	return b.flush()
}

// Errors 获取发生的错误
func (b *Buffer[T]) Errors() []error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.errors
}

// Size 获取当前缓冲区大小
func (b *Buffer[T]) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buffer)
}

// ==================== 缓冲区批量更新 ====================

// UpdateBuffer 批量更新缓冲区
type UpdateBuffer struct {
	coll      *mongo.Collection
	ctx       context.Context
	size      int
	flushTime time.Duration
	buffer    []UpdateDoc
	mu        sync.Mutex
	timer     *time.Timer
	closed    bool
	errors    []error
}

// NewUpdateBuffer 创建更新缓冲区
//
// 示例：
//
//	buffer := batch.NewUpdateBuffer(ctx, users, 100, 5*time.Second)
func NewUpdateBuffer[T any](ctx context.Context, coll *mgo.Collection[T], size int, flushTime time.Duration) *UpdateBuffer {
	if size <= 0 {
		size = 100
	}
	if flushTime <= 0 {
		flushTime = 5 * time.Second
	}

	buffer := &UpdateBuffer{
		coll:      coll.Native(),
		ctx:       ctx,
		size:      size,
		flushTime: flushTime,
		buffer:    make([]UpdateDoc, 0, size),
		errors:    make([]error, 0),
	}

	buffer.timer = time.AfterFunc(flushTime, func() {
		buffer.Flush()
	})

	return buffer
}

// Add 添加更新操作
func (b *UpdateBuffer) Add(filter, update mgo.M) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return mgo.ErrInvalidOperation
	}

	b.buffer = append(b.buffer, UpdateDoc{
		Filter: filter,
		Update: update,
	})

	if len(b.buffer) >= b.size {
		return b.flush()
	}

	return nil
}

// Flush 刷新缓冲区
func (b *UpdateBuffer) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.flush()
}

func (b *UpdateBuffer) flush() error {
	if len(b.buffer) == 0 {
		return nil
	}

	// 构建批量写操作
	models := make([]mongo.WriteModel, len(b.buffer))
	for i, update := range b.buffer {
		models[i] = mongo.NewUpdateOneModel().
			SetFilter(update.Filter).
			SetUpdate(update.Update)
	}

	// 执行批量写
	_, err := b.coll.BulkWrite(b.ctx, models)
	if err != nil {
		b.errors = append(b.errors, err)
		return err
	}

	// 清空缓冲区
	b.buffer = b.buffer[:0]

	// 重置定时器
	if b.timer != nil {
		b.timer.Reset(b.flushTime)
	}

	return nil
}

// Close 关闭缓冲区
func (b *UpdateBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true

	if b.timer != nil {
		b.timer.Stop()
	}

	return b.flush()
}
