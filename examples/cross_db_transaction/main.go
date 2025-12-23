package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gocrud/mgo"
)

// User 账户模型
type User struct {
	ID      mgo.ObjectID `bson:"_id,omitempty"`
	Name    string       `bson:"name"`
	Balance float64      `bson:"balance"`
}

func (User) CollName() string {
	return "users"
}

// TransferLog 转账日志模型
type TransferLog struct {
	ID         mgo.ObjectID `bson:"_id,omitempty"`
	FromUserID mgo.ObjectID `bson:"from_user_id"`
	ToUserID   mgo.ObjectID `bson:"to_user_id"`
	Amount     float64      `bson:"amount"`
	Status     string       `bson:"status"`
}

func (TransferLog) CollName() string {
	return "transfer_logs"
}

func main() {
	ctx := context.Background()

	// 创建 Client 实例（支持跨库操作）
	client, err := mgo.OpenClient("mongodb://example:example@localhost:27017/admin?authSource=admin&directConnection=true")
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer client.Close()

	// 访问不同的数据库
	accountsDB := client.Database("accounts")
	logsDB := client.Database("logs")

	// 准备数据
	users := mgo.Model[User](accountsDB, "users")
	logs := mgo.Model[TransferLog](logsDB, "transfer_logs")

	// 清理测试数据
	users.Truncate()
	logs.Truncate()

	// 创建两个测试用户
	user1 := &User{Name: "张三", Balance: 1000}
	user2 := &User{Name: "李四", Balance: 500}

	user1ID, _ := users.WithContext(ctx).Insert(user1)
	user2ID, _ := users.WithContext(ctx).Insert(user2)

	fmt.Println("=== 跨库事务示例 ===")
	fmt.Printf("转账前 - 张三余额: %.2f, 李四余额: %.2f\n", user1.Balance, user2.Balance)

	// 跨库事务：从 user1 转账 200 到 user2，并在日志数据库记录
	amount := 200.0
	err = client.WithContext(ctx).Transaction(func(sess *mgo.ClientSession) error {
		// 在事务中访问多个数据库
		txAccountsDB := sess.Database("accounts")
		txLogsDB := sess.Database("logs")

		txUsers := mgo.Model[User](txAccountsDB, "users")
		txLogs := mgo.Model[TransferLog](txLogsDB, "transfer_logs")

		// 扣款
		if err := txUsers.Find().WithContext(sess.Context()).ID(user1ID).Inc("balance", -amount).Update(); err != nil {
			return fmt.Errorf("扣款失败: %w", err)
		}

		// 入账
		if err := txUsers.Find().WithContext(sess.Context()).ID(user2ID).Inc("balance", amount).Update(); err != nil {
			return fmt.Errorf("入账失败: %w", err)
		}

		// 记录日志（不同数据库）
		log := &TransferLog{
			FromUserID: user1ID,
			ToUserID:   user2ID,
			Amount:     amount,
			Status:     "completed",
		}
		if _, err := txLogs.WithContext(sess.Context()).Insert(log); err != nil {
			return fmt.Errorf("记录日志失败: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Fatal("跨库事务失败:", err)
	}
	fmt.Println("跨库事务成功")
}
