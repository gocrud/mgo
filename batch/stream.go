package batch

// ==================== 流式处理 ====================

// Each 遍历查询结果并对每条记录执行回调
//
// 示例：
//
//	err := batch.Each(users.Find().Eq("status", "active"),
//	    func(user *User) error {
//	        return process(user)
//	    })
func Each[T any](query interface{}, fn func(*T) error) error {
	// 直接使用 query 的 Each 方法
	if q, ok := query.(interface {
		Each(func(*T) error) error
	}); ok {
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
