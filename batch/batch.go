package batch

import (
	"context"

	"github.com/gocrud/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 批量插入 ====================

// InsertBatch 批量插入文档（自动分批）
//
// 示例：
//
//	largeUserList := make([]*User, 10000)
//	// ... 填充数据
//
//	// 自动分批（默认 1000 条/批）
//	err := batch.InsertBatch(users, largeUserList)
//
//	// 自定义批次大小
//	err := batch.InsertBatch(users, largeUserList, batch.Size(500))
func InsertBatch[T any](coll interface{}, docs []*T, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
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

		// 插入当前批次
		_, err := nativeColl.InsertMany(ctx, items)
		if err != nil {
			return mgo.WrapError(err, "failed to insert batch")
		}
	}

	return nil
}

// InsertBatchWithCallback 批量插入并在每批后执行回调
//
// 示例：
//
//	err := batch.InsertBatchWithCallback(users, largeList, func(inserted int) error {
//	    fmt.Printf("已插入 %d 条\n", inserted)
//	    return nil
//	})
func InsertBatchWithCallback[T any](coll interface{}, docs []*T, callback func(int) error, opts ...Option) error {
	if len(docs) == 0 {
		return nil
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 获取原生集合和上下文
	nativeColl, ctx := extractCollectionAndContext(coll)

	// 分批处理
	chunks := mgo.ChunkSlice(docs, options.Size)
	totalInserted := 0

	for _, chunk := range chunks {
		// 转换为 []interface{}
		items := make([]interface{}, len(chunk))
		for i, doc := range chunk {
			items[i] = doc
		}

		// 插入当前批次
		_, err := nativeColl.InsertMany(ctx, items)
		if err != nil {
			return mgo.WrapError(err, "failed to insert batch")
		}

		totalInserted += len(chunk)

		// 执行回调
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
//
// 示例：
//
//	updates := []batch.UpdateDoc{
//	    {Filter: mgo.M{"_id": id1}, Update: mgo.M{"$set": mgo.M{"status": "active"}}},
//	    {Filter: mgo.M{"_id": id2}, Update: mgo.M{"$set": mgo.M{"status": "inactive"}}},
//	}
//	err := batch.UpdateBatch(users, updates)
type UpdateDoc struct {
	Filter mgo.M
	Update mgo.M
}

func UpdateBatch(coll interface{}, updates []UpdateDoc, opts ...Option) error {
	if len(updates) == 0 {
		return nil
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 获取原生集合和上下文
	nativeColl, ctx := extractCollectionAndContext(coll)

	// 分批处理
	chunks := mgo.ChunkSlice(updates, options.Size)

	for _, chunk := range chunks {
		// 构建批量写操作
		models := make([]mongo.WriteModel, len(chunk))
		for i, update := range chunk {
			models[i] = mongo.NewUpdateOneModel().
				SetFilter(update.Filter).
				SetUpdate(update.Update)
		}

		// 执行批量写
		_, err := nativeColl.BulkWrite(ctx, models)
		if err != nil {
			return mgo.WrapError(err, "failed to update batch")
		}
	}

	return nil
}

// ==================== 批量删除 ====================

// DeleteBatch 批量删除文档
//
// 示例：
//
//	filters := []mgo.M{
//	    {"_id": id1},
//	    {"_id": id2},
//	}
//	n, err := batch.DeleteBatch(users, filters)
func DeleteBatch(coll interface{}, filters []mgo.M, opts ...Option) (int64, error) {
	if len(filters) == 0 {
		return 0, nil
	}

	// 解析选项
	options := &Options{
		Size:    1000,
		Ordered: true,
	}
	for _, opt := range opts {
		opt(options)
	}

	// 获取原生集合和上下文
	nativeColl, ctx := extractCollectionAndContext(coll)

	// 分批处理
	chunks := mgo.ChunkSlice(filters, options.Size)
	totalDeleted := int64(0)

	for _, chunk := range chunks {
		// 构建批量写操作
		models := make([]mongo.WriteModel, len(chunk))
		for i, filter := range chunk {
			models[i] = mongo.NewDeleteOneModel().SetFilter(filter)
		}

		// 执行批量写
		result, err := nativeColl.BulkWrite(ctx, models)
		if err != nil {
			return totalDeleted, mgo.WrapError(err, "failed to delete batch")
		}

		totalDeleted += result.DeletedCount
	}

	return totalDeleted, nil
}

// ==================== 辅助函数 ====================

// extractCollectionAndContext 提取集合和上下文
func extractCollectionAndContext(coll interface{}) (*mongo.Collection, context.Context) {
	var nativeColl *mongo.Collection
	var ctx context.Context = context.Background()

	// 尝试提取原生集合
	if nc, ok := coll.(interface{ Native() *mongo.Collection }); ok {
		nativeColl = nc.Native()
	} else if nc, ok := coll.(*mongo.Collection); ok {
		nativeColl = nc
	} else {
		panic("batch: invalid collection type")
	}

	// 尝试提取上下文
	if ctxGetter, ok := coll.(interface{ Context() context.Context }); ok {
		ctx = ctxGetter.Context()
	}

	return nativeColl, ctx
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
//
// 示例：
//
//	err := batch.InsertBatch(users, docs, batch.Size(500))
func Size(size int) Option {
	return func(opts *Options) {
		opts.Size = size
	}
}

// Ordered 设置是否有序执行
//
// 示例：
//
//	err := batch.InsertBatch(users, docs, batch.Ordered(false))
func Ordered(ordered bool) Option {
	return func(opts *Options) {
		opts.Ordered = ordered
	}
}
