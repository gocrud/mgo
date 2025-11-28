package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// helpers.go - 辅助函数

// ==================== 连接辅助函数 ====================

// Connect 连接到 MongoDB 并返回 mgo.Client
//
// 推荐使用 NewClient 代替此函数
//
// 示例：
//
//	client, err := mgo.Connect(ctx, "mongodb://localhost:27017")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Disconnect(ctx)
func Connect(ctx context.Context, uri string, opts ...*options.ClientOptions) (*Client, error) {
	return NewClient(ctx, uri, opts...)
}

// MustConnect 连接到 MongoDB，失败时 panic
//
// 推荐使用 MustNewClient 代替此函数
//
// 示例：
//
//	client := mgo.MustConnect(ctx, "mongodb://localhost:27017")
//	defer client.Disconnect(ctx)
func MustConnect(ctx context.Context, uri string, opts ...*options.ClientOptions) *Client {
	return MustNewClient(ctx, uri, opts...)
}

// ==================== 内部辅助函数 ====================

// unwrap 解包表达式和字段引用
func unwrap(v any) any {
	if e, ok := v.(Expr); ok {
		return e.Build()
	}
	if f, ok := v.(Field); ok {
		return f.String()
	}
	return v
}

// unwrapExprs 批量解包表达式
func unwrapExprs(values []any) []any {
	result := make([]any, len(values))
	for i, v := range values {
		result[i] = unwrap(v)
	}
	return result
}

// D 简化 bson.D 创建的辅助函数
func makeD(key string, value any) bson.D {
	return bson.D{{Key: key, Value: value}}
}

// ptrInt 创建 int 指针
func ptrInt(v int) *int {
	return &v
}

// ptrInt64 创建 int64 指针
func ptrInt64(v int64) *int64 {
	return &v
}

// ptrFloat64 创建 float64 指针
func ptrFloat64(v float64) *float64 {
	return &v
}

// ptrBool 创建 bool 指针
func ptrBool(v bool) *bool {
	return &v
}

// ptrString 创建 string 指针
func ptrString(v string) *string {
	return &v
}

// ==================== 快捷构建辅助函数 ====================

// Set 创建 $set 更新文档的快捷函数
//
// 示例：
//
//	mgo.Set("name", "张三") // -> bson.M{"$set": bson.M{"name": "张三"}}
func Set(key string, value any) bson.M {
	return bson.M{"$set": bson.M{key: value}}
}
