package mgo

import (
	"context"
	"testing"
)

func TestWithTransaction(t *testing.T) {
	// 注意：这只是示例代码结构，需要实际的 MongoDB 连接来运行
	t.Skip("需要实际的 MongoDB 连接")

	_ = context.Background()
	// client, _ := Connect(ctx, "mongodb://localhost:27017")

	// 示例：使用 client.WithTransaction 执行事务
	// err := client.WithTransaction(ctx, func(ctx context.Context) error {
	//     db := client.Database("test")
	//     coll := db.Collection("orders")
	//
	//     // 插入订单
	//     _, err := coll.InsertOne(ctx, M{
	//         "order_id": "ORD001",
	//         "amount":   100.00,
	//     })
	//     if err != nil {
	//         return err
	//     }
	//
	//     // 更新库存
	//     invColl := db.Collection("inventory")
	//     _, err = invColl.Query(ctx).
	//         Eq("product_id", "PROD001").
	//         UpdateOne(Update().Inc("stock", -1))
	//
	//     return err
	// })
	//
	// if err != nil {
	//     t.Fatal("事务执行失败:", err)
	// }
}

func TestSessionTransaction(t *testing.T) {
	// 注意：这只是示例代码结构，需要实际的 MongoDB 连接来运行
	t.Skip("需要实际的 MongoDB 连接")

	_ = context.Background()
	// client, _ := Connect(ctx, "mongodb://localhost:27017")

	// 示例：使用 SessionTransaction 手动控制事务
	// txn, err := client.StartTransaction(ctx)
	// if err != nil {
	//     t.Fatal(err)
	// }
	// defer txn.EndSession()
	//
	// db := client.Database("test")
	// coll := db.Collection("accounts")
	//
	// // 从账户A扣款
	// _, err = coll.Query(txn.Context()).
	//     Eq("account_id", "A").
	//     UpdateOne(Update().Inc("balance", -100))
	// if err != nil {
	//     txn.Abort()
	//     t.Fatal(err)
	// }
	//
	// // 向账户B加款
	// _, err = coll.Query(txn.Context()).
	//     Eq("account_id", "B").
	//     UpdateOne(Update().Inc("balance", 100))
	// if err != nil {
	//     txn.Abort()
	//     t.Fatal(err)
	// }
	//
	// // 提交事务
	// if err := txn.Commit(); err != nil {
	//     t.Fatal("提交事务失败:", err)
	// }
}

func TestTransactionWithOptions(t *testing.T) {
	// 注意：这只是示例代码结构，需要实际的 MongoDB 连接来运行
	t.Skip("需要实际的 MongoDB 连接")

	_ = context.Background()
	// client, _ := Connect(ctx, "mongodb://localhost:27017")

	// 示例：使用事务选项
	// opts := TransactionOptions().
	//     SetReadConcern(readconcern.Majority()).
	//     SetWriteConcern(writeconcern.Majority())
	//
	// err := client.WithTransaction(ctx, func(ctx context.Context) error {
	//     // 执行需要强一致性的操作
	//     db := client.Database("test")
	//     coll := db.Collection("critical_data")
	//     _, err := coll.InsertOne(ctx, M{"data": "important"})
	//     return err
	// }, opts)
	//
	// if err != nil {
	//     t.Fatal(err)
	// }
}
