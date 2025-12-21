package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocrud/mgo"
	"github.com/gocrud/mgo/agg"
	"github.com/gocrud/mgo/batch"
	"github.com/gocrud/mgo/tx"
)

// User 用户模型
type User struct {
	ID        mgo.ObjectID `bson:"_id,omitempty"`
	Name      string       `bson:"name"`
	Email     string       `bson:"email"`
	City      string       `bson:"city"`
	Age       int          `bson:"age"`
	Status    string       `bson:"status"`
	Balance   float64      `bson:"balance"`
	CreatedAt time.Time    `bson:"created_at"`
	UpdatedAt time.Time    `bson:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

// Order 订单模型
type Order struct {
	ID        mgo.ObjectID `bson:"_id,omitempty"`
	UserID    mgo.ObjectID `bson:"user_id"`
	Amount    float64      `bson:"amount"`
	Status    string       `bson:"status"`
	CreatedAt time.Time    `bson:"created_at"`
}

func (Order) TableName() string {
	return "orders"
}

// CityStats 城市统计
type CityStats struct {
	City   string  `bson:"_id"`
	Count  int     `bson:"count"`
	AvgAge float64 `bson:"avg_age"`
	MaxAge int     `bson:"max_age"`
	MinAge int     `bson:"min_age"`
}

func main() {
	fmt.Println("🚀 MGO 高级功能示例")
	fmt.Println("======================")

	// 连接数据库
	db := mgo.MustOpen("mongodb://localhost/mgo_advanced")
	defer db.Close()
	fmt.Println("✅ 已连接到数据库")

	// 获取集合
	users := mgo.Model[User](db).WithTimestamps()
	orders := mgo.Model[Order](db).WithTimestamps()

	// 清空集合
	users.Truncate()
	orders.Truncate()

	// ==================== 1. 批量插入示例 ====================
	fmt.Println("\n📦 批量插入示例...")

	// 创建大量测试数据
	largeUserList := make([]*User, 1000)
	cities := []string{"北京", "上海", "广州", "深圳", "杭州"}
	for i := 0; i < 1000; i++ {
		largeUserList[i] = &User{
			Name:    fmt.Sprintf("User%d", i),
			Email:   fmt.Sprintf("user%d@example.com", i),
			City:    cities[i%len(cities)],
			Age:     20 + (i % 40),
			Status:  "active",
			Balance: float64(i * 100),
		}
	}

	// 批量插入（自动分批）
	err := batch.InsertBatch(users, largeUserList, batch.Size(200))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ 批量插入 1000 条用户记录")

	// 带统计的批量插入
	moreUsers := make([]*User, 500)
	for i := 0; i < 500; i++ {
		moreUsers[i] = &User{
			Name:   fmt.Sprintf("User%d", i+1000),
			Email:  fmt.Sprintf("user%d@example.com", i+1000),
			City:   cities[i%len(cities)],
			Age:    20 + (i % 40),
			Status: "pending",
		}
	}

	stats, _ := batch.InsertBatchWithStats(users, moreUsers)
	fmt.Printf("✅ 批量插入统计 - 总数: %d, 成功: %d, 失败: %d\n",
		stats.Total, stats.Success, stats.Failed)

	// ==================== 2. 流式处理示例 ====================
	fmt.Println("\n🌊 流式处理示例...")

	// Each 遍历
	processedCount := 0
	err = batch.Each(users.Find().Where("status", "active"),
		func(user *User) error {
			processedCount++
			// 模拟处理
			return nil
		})
	fmt.Printf("✅ Each 遍历处理了 %d 条记录\n", processedCount)

	// Chunk 分块处理
	chunkCount := 0
	err = batch.Chunk(users.Find(), 100,
		func(userList []*User) error {
			chunkCount++
			fmt.Printf("   处理第 %d 批，包含 %d 条记录\n", chunkCount, len(userList))
			return nil
		})
	fmt.Printf("✅ Chunk 处理完成，共 %d 批\n", chunkCount)

	// Stream Channel
	streamCount := 0
	for user := range batch.Stream[User](users.Find().Limit(100), 20) {
		streamCount++
		_ = user
	}
	fmt.Printf("✅ Stream 处理了 %d 条记录\n", streamCount)

	// ==================== 3. 聚合查询示例 ====================
	fmt.Println("\n📊 聚合查询示例...")

	// 按城市统计
	cityStats, err := agg.Aggregate[CityStats](users).
		Match(mgo.Eq("status", "active")).
		GroupBy("$city").
		Count("count").
		Avg("avg_age", "$age").
		Max("max_age", "$age").
		Min("min_age", "$age").
		SortDesc("count").
		All()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ 城市统计结果:")
	for _, stat := range cityStats {
		fmt.Printf("   %s: %d 人，平均年龄 %.1f 岁 (最小 %d, 最大 %d)\n",
			stat.City, stat.Count, stat.AvgAge, stat.MinAge, stat.MaxAge)
	}

	// ==================== 4. 事务示例 ====================
	fmt.Println("\n💼 事务示例...")

	// 创建测试订单
	testUser, _ := users.Find().Where("status", "active").First()
	if testUser != nil {
		// 自动事务
		err := tx.Transaction(db, func(sess *tx.Session) error {
			txUsers := mgo.Model[User](sess)
			txOrders := mgo.Model[Order](sess)

			// 扣减余额
			amount := 50.0
			if err := txUsers.Find().ID(testUser.ID).
				Inc("balance", -amount).
				Update(); err != nil {
				return err // 自动回滚
			}

			// 创建订单
			order := &Order{
				UserID: testUser.ID,
				Amount: amount,
				Status: "completed",
			}
			if _, err := txOrders.Insert(order); err != nil {
				return err // 自动回滚
			}

			return nil // 自动提交
		})

		if err != nil {
			fmt.Printf("❌ 事务失败: %v\n", err)
		} else {
			fmt.Println("✅ 事务成功（扣减余额并创建订单）")
		}

		// 带重试的事务
		err = tx.WithRetry(db, 3, func(sess *tx.Session) error {
			txUsers := mgo.Model[User](sess)

			return txUsers.Find().ID(testUser.ID).
				Inc("balance", 10).
				Update()
		})

		if err == nil {
			fmt.Println("✅ 带重试的事务成功")
		}
	}

	// ==================== 5. 缓冲区示例 ====================
	fmt.Println("\n🔄 缓冲区示例...")

	// 创建插入缓冲区
	insertBuffer := batch.NewBuffer[User](users, 50, 2*time.Second)
	defer insertBuffer.Close()

	// 添加文档
	for i := 0; i < 120; i++ {
		insertBuffer.Add(&User{
			Name:   fmt.Sprintf("BufferedUser%d", i),
			Email:  fmt.Sprintf("buffered%d@example.com", i),
			City:   cities[i%len(cities)],
			Age:    25,
			Status: "active",
		})
	}

	fmt.Printf("✅ 缓冲区当前大小: %d\n", insertBuffer.Size())
	insertBuffer.Flush()
	fmt.Println("✅ 缓冲区已刷新")

	// ==================== 6. 复杂聚合示例 ====================
	fmt.Println("\n🎯 复杂聚合示例...")

	type UserWithOrders struct {
		User
		Orders      []Order `bson:"orders"`
		OrderCount  int     `bson:"order_count"`
		TotalAmount float64 `bson:"total_amount"`
	}

	// 关联查询（模拟，实际需要有订单数据）
	// usersWithOrders, err := agg.Aggregate[UserWithOrders](users).
	// 	Lookup("orders", "_id", "user_id", "orders").
	// 	Match(mgo.Eq("status", "active")).
	// 	AddFields(mgo.M{
	// 		"order_count":  agg.Size("$orders"),
	// 		"total_amount": agg.Sum("$orders.amount"),
	// 	}).
	// 	Limit(10).
	// 	All()

	// ==================== 7. 并行批量操作 ====================
	fmt.Println("\n⚡ 并行批量操作...")

	parallelUsers := make([]*User, 500)
	for i := 0; i < 500; i++ {
		parallelUsers[i] = &User{
			Name:   fmt.Sprintf("ParallelUser%d", i),
			Email:  fmt.Sprintf("parallel%d@example.com", i),
			Status: "active",
		}
	}

	// 4 个并发
	start := time.Now()
	err = batch.InsertBatchParallel(users, parallelUsers, 4, batch.Size(100))
	elapsed := time.Since(start)

	if err == nil {
		fmt.Printf("✅ 并行插入 500 条记录，耗时: %v\n", elapsed)
	}

	// ==================== 最终统计 ====================
	fmt.Println("\n📈 最终统计...")

	totalCount, _ := users.CountAll()
	activeCount, _ := users.Count(mgo.M{"status": "active"})
	pendingCount, _ := users.Count(mgo.M{"status": "pending"})

	fmt.Printf("✅ 总用户数: %d\n", totalCount)
	fmt.Printf("✅ 活跃用户: %d\n", activeCount)
	fmt.Printf("✅ 待激活用户: %d\n", pendingCount)

	fmt.Println("\n✨ 高级功能示例完成！")
}
