package batch

import (
	"github.com/gocrud/mgo"
)

// ==================== 流式处理 ====================

// Each 遍历查询结果并对每条记录执行回调
//
// 示例：
//
//	err := batch.Each(users.Find().Where("status", "active"),
//	    func(user *User) error {
//	        return process(user)
//	    })
func Each[T any](query interface{}, fn func(*T) error) error {
	// 直接使用 query 的 Each 方法
	if q, ok := query.(interface{ Each(func(*T) error) error }); ok {
		return q.Each(fn)
	}

	panic("batch: query type does not support Each method")
}

// Stream 返回一个 channel 来流式处理查询结果
//
// 示例：
//
//	for user := range batch.Stream(users.Find(), 100) {
//	    process(user)
//	}
func Stream[T any](query interface{}, bufferSize int) <-chan *T {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	ch := make(chan *T, bufferSize)

	go func() {
		defer close(ch)

		// 使用 Each 实现
		_ = Each(query, func(doc *T) error {
			ch <- doc
			return nil
		})
	}()

	return ch
}

// StreamWithError 返回 channel 和 error channel
//
// 示例：
//
//	dataCh, errCh := batch.StreamWithError(users.Find(), 100)
//	for {
//	    select {
//	    case user, ok := <-dataCh:
//	        if !ok {
//	            return
//	        }
//	        process(user)
//	    case err := <-errCh:
//	        if err != nil {
//	            log.Error(err)
//	            return
//	        }
//	    }
//	}
func StreamWithError[T any](query interface{}, bufferSize int) (<-chan *T, <-chan error) {
	if bufferSize <= 0 {
		bufferSize = 100
	}

	dataCh := make(chan *T, bufferSize)
	errCh := make(chan error, 1)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		err := Each(query, func(doc *T) error {
			dataCh <- doc
			return nil
		})

		if err != nil {
			errCh <- err
		}
	}()

	return dataCh, errCh
}

// ==================== Chunk 分块处理 ====================

// Chunk 分块处理查询结果
//
// 示例：
//
//	err := batch.Chunk(users.Find(), 100, func(users []*User) error {
//	    for _, user := range users {
//	        process(user)
//	    }
//	    return nil
//	})
func Chunk[T any](query interface{}, size int, fn func([]*T) error) error {
	if size <= 0 {
		size = 100
	}

	var batch []*T

	err := Each(query, func(doc *T) error {
		batch = append(batch, doc)

		if len(batch) >= size {
			if err := fn(batch); err != nil {
				return err
			}
			batch = []*T{} // 清空批次
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 处理剩余的文档
	if len(batch) > 0 {
		if err := fn(batch); err != nil {
			return err
		}
	}

	return nil
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
//
// 示例：
//
//	stats, err := batch.InsertBatchWithStats(users, largeList)
//	fmt.Printf("成功: %d, 失败: %d\n", stats.Success, stats.Failed)
func InsertBatchWithStats[T any](coll interface{}, docs []*T, opts ...Option) (*BatchStats, error) {
	stats := &BatchStats{
		Total:  len(docs),
		Errors: []error{},
	}

	if len(docs) == 0 {
		return stats, nil
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: false, // 默认无序，继续处理错误
	}
	for _, opt := range opts {
		opt(options)
	}

	// 获取原生集合和上下文
	nativeColl, ctx := extractCollectionAndContext(coll)

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
		_, err := nativeColl.InsertMany(ctx, items)
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

// ==================== 并行批量插入 ====================

// InsertBatchParallel 并行批量插入
//
// 示例：
//
//	err := batch.InsertBatchParallel(users, largeList, 4)  // 4 个并发
func InsertBatchParallel[T any](coll interface{}, docs []*T, concurrency int, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}

	if concurrency <= 0 {
		concurrency = 4
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: false,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 获取原生集合和上下文
	nativeColl, ctx := extractCollectionAndContext(coll)

	// 分批
	chunks := mgo.ChunkSlice(docs, options.Size)

	// 创建任务 channel
	tasks := make(chan []interface{}, len(chunks))
	errors := make(chan error, len(chunks))
	done := make(chan bool, concurrency)

	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		go func() {
			for items := range tasks {
				_, err := nativeColl.InsertMany(ctx, items)
				if err != nil {
					errors <- err
				}
			}
			done <- true
		}()
	}

	// 发送任务
	for _, chunk := range chunks {
		items := make([]interface{}, len(chunk))
		for i, doc := range chunk {
			items[i] = doc
		}
		tasks <- items
	}
	close(tasks)

	// 等待完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
	close(errors)

	// 收集错误
	var firstErr error
	for err := range errors {
		if firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
