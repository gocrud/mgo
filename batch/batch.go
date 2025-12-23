package batch

import (
	"context"
	"sync"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 批量插入 ====================

// InsertBatch 批量插入文档（自动分批）
func InsertBatch[T any](coll *mgo.Collection[T], docs []*T, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}

	ctx := coll.Context()

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	nativeColl := coll.Native()
	batchSize := options.Size
	total := len(docs)

	// 预分配 buffer
	buffer := make([]interface{}, 0, batchSize)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		// 重置 buffer
		buffer = buffer[:0]
		for _, doc := range docs[i:end] {
			buffer = append(buffer, doc)
		}

		// 插入当前批次
		_, err := nativeColl.InsertMany(ctx, buffer)
		if err != nil {
			return mgo.WrapError(err, "failed to insert batch")
		}
	}

	return nil
}

// InsertBatchWithCallback 批量插入并在每批后执行回调
func InsertBatchWithCallback[T any](coll *mgo.Collection[T], docs []*T, callback func(int) error, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}

	ctx := coll.Context()

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	nativeColl := coll.Native()
	batchSize := options.Size
	total := len(docs)
	totalInserted := 0

	buffer := make([]interface{}, 0, batchSize)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		buffer = buffer[:0]
		for _, doc := range docs[i:end] {
			buffer = append(buffer, doc)
		}

		_, err := nativeColl.InsertMany(ctx, buffer)
		if err != nil {
			return mgo.WrapError(err, "failed to insert batch")
		}

		totalInserted += len(buffer)

		if callback != nil {
			if err := callback(totalInserted); err != nil {
				return err
			}
		}
	}

	return nil
}

// ==================== 批量更新 ====================

// UpdateBatch 批量更新文档
type UpdateDoc struct {
	Filter mgo.M
	Update mgo.M
}

func UpdateBatch[T any](ctx context.Context, coll *mgo.Collection[T], updates []UpdateDoc, opts ...Option) error {
	if len(updates) == 0 {
		return nil
	}

	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	nativeColl := coll.Native()
	batchSize := options.Size
	total := len(updates)

	models := make([]mongo.WriteModel, 0, batchSize)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		models = models[:0]
		for _, update := range updates[i:end] {
			models = append(models, mongo.NewUpdateOneModel().
				SetFilter(update.Filter).
				SetUpdate(update.Update))
		}

		_, err := nativeColl.BulkWrite(ctx, models)
		if err != nil {
			return mgo.WrapError(err, "failed to update batch")
		}
	}

	return nil
}

// ==================== 批量删除 ====================

// DeleteBatch 批量删除文档
func DeleteBatch[T any](ctx context.Context, coll *mgo.Collection[T], filters []mgo.M, opts ...Option) (int64, error) {
	if len(filters) == 0 {
		return 0, nil
	}

	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	nativeColl := coll.Native()
	batchSize := options.Size
	total := len(filters)
	totalDeleted := int64(0)

	models := make([]mongo.WriteModel, 0, batchSize)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		models = models[:0]
		for _, filter := range filters[i:end] {
			models = append(models, mongo.NewDeleteOneModel().SetFilter(filter))
		}

		result, err := nativeColl.BulkWrite(ctx, models)
		if err != nil {
			return totalDeleted, mgo.WrapError(err, "failed to delete batch")
		}

		totalDeleted += result.DeletedCount
	}

	return totalDeleted, nil
}

// InsertBatchParallel 并行批量插入
func InsertBatchParallel[T any](coll *mgo.Collection[T], docs []*T, workers int, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}
	ctx := coll.Context()
	if workers <= 0 {
		workers = 1
	}

	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	batchSize := options.Size
	total := len(docs)

	// 计算总批次数
	numBatches := (total + batchSize - 1) / batchSize

	// 任务通道
	type batchTask struct {
		start int
		end   int
	}
	tasks := make(chan batchTask, numBatches)

	// 错误通道
	errCh := make(chan error, 1)

	// 启动 worker
	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				// 检查是否有错误发生
				select {
				case <-errCh:
					return
				case <-ctx.Done():
					return
				default:
				}

				// 准备批次数据
				// 注意：这里需要将 []*T 转换为 []interface{}
				// 为了避免并发问题，每个 worker 创建自己的 buffer
				batchDocs := make([]interface{}, task.end-task.start)
				for i, doc := range docs[task.start:task.end] {
					batchDocs[i] = doc
				}

				_, err := coll.Native().InsertMany(ctx, batchDocs)
				if err != nil {
					select {
					case errCh <- mgo.WrapError(err, "failed to insert batch parallel"):
					default:
					}
					return
				}
			}
		}()
	}

	// 发送任务
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		tasks <- batchTask{start: i, end: end}
	}
	close(tasks)

	// 等待完成
	wg.Wait()

	// 检查错误
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// ==================== 选项 ====================

// Option 批量操作选项
type Option func(*Options)

// Options 批量操作配置
type Options struct {
	Size    int
	Ordered bool
}

// Size 设置批次大小
func Size(size int) Option {
	return func(opts *Options) {
		opts.Size = size
	}
}

// Ordered 设置是否有序执行
func Ordered(ordered bool) Option {
	return func(opts *Options) {
		opts.Ordered = ordered
	}
}

// ==================== 批量操作统计 ====================

// BatchStats 批量操作统计
type BatchStats struct {
	Total     int     // 总数
	Processed int     // 已处理
	Success   int     // 成功
	Failed    int     // 失败
	Errors    []error // 错误列表
}

// InsertBatchWithStats 批量插入并返回统计信息
func InsertBatchWithStats[T any](coll *mgo.Collection[T], docs []*T, opts ...Option) (*BatchStats, error) {
	stats := &BatchStats{
		Total:  len(docs),
		Errors: []error{},
	}

	if len(docs) == 0 {
		return stats, nil
	}
	ctx := coll.Context()

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: false, // 默认无序，继续处理错误
	}
	for _, opt := range opts {
		opt(options)
	}

	// 分批处理
	chunks := mgo.ChunkSlice(docs, options.Size)

	for _, chunk := range chunks {
		// 转换为 []interface{}
		items := make([]interface{}, len(chunk))
		for i, doc := range chunk {
			items[i] = doc
		}

		stats.Processed += len(chunk)

		// 插入当前批次
		_, err := coll.Native().InsertMany(ctx, items)
		if err != nil {
			stats.Failed += len(chunk)
			stats.Errors = append(stats.Errors, err)

			// 如果是有序执行，遇到错误就停止
			if options.Ordered {
				return stats, err
			}
		} else {
			stats.Success += len(chunk)
		}
	}

	return stats, nil
}
