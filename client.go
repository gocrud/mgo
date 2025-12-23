package mgo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ==================== 客户端创建函数 ====================

// Open 连接到 MongoDB 并返回数据库实例
//
// URI 格式：mongodb://[username:password@]host[:port]/database
//
// 示例：
//
//	db, err := mgo.Open("mongodb://localhost/myapp")
//	if err != nil {
//	    return err
//	}
//	defer db.Close()
func Open(uri string, opts ...ClientOption) (*Database, error) {
	// 解析 URI 提取数据库名
	dbName := extractDatabaseName(uri)
	if dbName == "" {
		return nil, fmt.Errorf("mgo: invalid URI, database name not found")
	}

	// 构建客户端选项
	clientOpts := &clientOptions{
		uri:         uri,
		timeout:     10 * time.Second,
		retryWrites: true,
		retryReads:  true,
		ctx:         context.Background(),
	}

	for _, opt := range opts {
		opt(clientOpts)
	}

	// 创建客户端
	mongoOpts := buildClientOptions(clientOpts)
	client, err := mongo.Connect(mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("mgo: failed to connect: %w", err)
	}

	// Ping 验证连接
	ctx := clientOpts.ctx
	if clientOpts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, clientOpts.timeout)
		defer cancel()
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("mgo: failed to ping: %w", err)
	}

	// 返回数据库实例
	return &Database{
		client: client,
		db:     client.Database(dbName),
		name:   dbName,
	}, nil
}

// MustOpen 连接到 MongoDB，失败时 panic
//
// 示例：
//
//	db := mgo.MustOpen("mongodb://localhost/myapp")
//	defer db.Close()
func MustOpen(uri string, opts ...ClientOption) *Database {
	db, err := Open(uri, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

// Connect 使用多个选项连接到 MongoDB
//
// 示例：
//
//	db, err := mgo.Connect(
//	    mgo.URI("mongodb://localhost/myapp"),
//	    mgo.Timeout(10*time.Second),
//	    mgo.MaxPoolSize(100),
//	)
func Connect(opts ...ClientOption) (*Database, error) {
	clientOpts := &clientOptions{
		timeout:     10 * time.Second,
		retryWrites: true,
		retryReads:  true,
		ctx:         context.Background(),
	}

	for _, opt := range opts {
		opt(clientOpts)
	}

	if clientOpts.uri == "" {
		return nil, fmt.Errorf("mgo: URI is required")
	}

	return Open(clientOpts.uri, opts...)
}

// From 从现有的 mongo.Client 创建数据库实例
//
// 示例：
//
//	nativeClient, _ := mongo.Connect(opts)
//	db := mgo.From(nativeClient, "myapp")
func From(client *mongo.Client, dbName string) *Database {
	return &Database{
		client: client,
		db:     client.Database(dbName),
		name:   dbName,
	}
}

// ==================== Client 类型（支持跨库操作）====================

// Client MongoDB 客户端（支持跨库操作）
//
// 用于需要访问多个数据库的场景，支持跨库事务
//
// 示例：
//
//	client, err := mgo.OpenClient("mongodb://localhost")
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	// 访问多个数据库
//	db1 := client.Database("db1")
//	db2 := client.Database("db2")
type Client struct {
	client *mongo.Client
	ctx    context.Context
}

// OpenClient 连接到 MongoDB 并返回 Client 实例（支持跨库）
//
// URI 不需要指定数据库名（如果有会被忽略）
//
// 示例：
//
//	client, err := mgo.OpenClient("mongodb://localhost")
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	// 访问不同数据库
//	accountsDB := client.Database("accounts")
//	logsDB := client.Database("logs")
func OpenClient(uri string, opts ...ClientOption) (*Client, error) {
	// 移除 URI 中的数据库名（如果存在）
	// OpenClient 用于跨库操作，不需要默认数据库
	cleanURI := removeDBNameFromURI(uri)

	// 构建客户端选项
	clientOpts := &clientOptions{
		uri:         cleanURI,
		timeout:     10 * time.Second,
		retryWrites: true,
		retryReads:  true,
		ctx:         context.Background(),
	}

	for _, opt := range opts {
		opt(clientOpts)
	}

	// 创建客户端
	mongoOpts := buildClientOptions(clientOpts)
	client, err := mongo.Connect(mongoOpts)
	if err != nil {
		return nil, fmt.Errorf("mgo: failed to connect: %w", err)
	}

	// Ping 验证连接
	ctx := clientOpts.ctx
	if clientOpts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, clientOpts.timeout)
		defer cancel()
	}

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("mgo: failed to ping: %w", err)
	}

	return &Client{
		client: client,
		ctx:    clientOpts.ctx,
	}, nil
}

// removeDBNameFromURI 移除 URI 中的数据库名（保留查询参数）
func removeDBNameFromURI(uri string) string {
	// 查找 / 和 ? 的位置
	lastSlash := strings.LastIndex(uri, "/")
	queryStart := strings.Index(uri, "?")

	if lastSlash == -1 {
		return uri
	}

	// 如果有查询参数
	if queryStart != -1 && queryStart > lastSlash {
		// 移除数据库名，保留查询参数
		return uri[:lastSlash+1] + uri[queryStart:]
	}

	// 没有查询参数，检查最后一个 / 后是否有数据库名
	afterSlash := uri[lastSlash+1:]
	if afterSlash != "" && !strings.Contains(afterSlash, "@") {
		// 有数据库名，移除它
		return uri[:lastSlash+1]
	}

	return uri
}

// MustOpenClient 连接到 MongoDB，失败时 panic
//
// 示例：
//
//	client := mgo.MustOpenClient("mongodb://localhost")
//	defer client.Close()
func MustOpenClient(uri string, opts ...ClientOption) *Client {
	client, err := OpenClient(uri, opts...)
	if err != nil {
		panic(err)
	}
	return client
}

// ConnectClient 使用多个选项连接
//
// 示例：
//
//	client, err := mgo.ConnectClient(
//	    mgo.URI("mongodb://localhost"),
//	    mgo.Timeout(10*time.Second),
//	    mgo.MaxPoolSize(100),
//	)
func ConnectClient(opts ...ClientOption) (*Client, error) {
	clientOpts := &clientOptions{
		timeout:     10 * time.Second,
		retryWrites: true,
		retryReads:  true,
		ctx:         context.Background(),
	}

	for _, opt := range opts {
		opt(clientOpts)
	}

	if clientOpts.uri == "" {
		return nil, fmt.Errorf("mgo: URI is required")
	}

	return OpenClient(clientOpts.uri, opts...)
}

// Database 获取指定数据库
//
// 示例：
//
//	db := client.Database("myapp")
func (c *Client) Database(name string) *Database {
	return &Database{
		client: c.client,
		db:     c.client.Database(name),
		name:   name,
	}
}

// Close 关闭客户端连接
//
// 示例：
//
//	defer client.Close()
func (c *Client) Close() error {
	return c.client.Disconnect(c.ctx)
}

// Disconnect Close 的别名
//
// 示例：
//
//	defer client.Disconnect()
func (c *Client) Disconnect() error {
	return c.Close()
}

// Ping 测试连接
//
// 示例：
//
//	if err := client.Ping(); err != nil {
//	    log.Fatal("连接失败:", err)
//	}
func (c *Client) Ping() error {
	return c.client.Ping(c.ctx, nil)
}

// Context 获取默认上下文
//
// 示例：
//
//	ctx := client.Context()
func (c *Client) Context() context.Context {
	return getContext(c.ctx)
}

// WithContext 设置默认上下文并返回新的 Client
//
// 示例：
//
//	ctx := context.WithTimeout(context.Background(), 5*time.Second)
//	client = client.WithContext(ctx)
func (c *Client) WithContext(ctx context.Context) *Client {
	return &Client{
		client: c.client,
		ctx:    ctx,
	}
}

// ListDatabaseNames 列出所有数据库名称
//
// 示例：
//
//	names, err := client.ListDatabaseNames()
//	for _, name := range names {
//	    fmt.Println(name)
//	}
func (c *Client) ListDatabaseNames() ([]string, error) {
	return c.client.ListDatabaseNames(c.ctx, M{})
}

// Native 返回原生 mongo.Client
//
// 示例：
//
//	nativeClient := client.Native()
func (c *Client) Native() *mongo.Client {
	return c.client
}

// ==================== Client 事务支持 ====================

// Transaction 执行跨库事务
//
// 示例：
//
//	err := client.Transaction(ctx, func(sess *mgo.ClientSession) error {
//	    // 访问多个数据库
//	    accountsDB := sess.Database("accounts")
//	    logsDB := sess.Database("logs")
//
//	    users := mgo.Model[User](accountsDB)
//	    logs := mgo.Model[Log](logsDB)
//
//	    // 跨库操作
//	    if err := users.Find().ID(userID).Inc("balance", -amount).Update(sess.Context()); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    if _, err := logs.Insert(log); err != nil {
//	        return err  // 自动回滚
//	    }
//
//	    return nil  // 自动提交
//	})
func (c *Client) Transaction(fn func(*ClientSession) error) error {
	ctx := c.Context()
	// 启动会话
	session, err := c.client.StartSession()
	if err != nil {
		return WrapError(err, "failed to start session")
	}
	defer session.EndSession(ctx)

	// 执行事务回调
	callback := func(sessCtx context.Context) (interface{}, error) {
		sess := &ClientSession{
			client:  c.client,
			session: session,
			ctx:     sessCtx,
		}
		return nil, fn(sess)
	}

	_, err = session.WithTransaction(ctx, callback)
	return err
}

// ==================== 辅助函数 ====================

// extractDatabaseName 从 URI 中提取数据库名
func extractDatabaseName(uri string) string {
	// 移除前缀
	uri = strings.TrimPrefix(uri, "mongodb://")
	uri = strings.TrimPrefix(uri, "mongodb+srv://")

	// 查找 / 后的数据库名
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return ""
	}

	// 数据库名可能包含查询参数
	dbName := parts[len(parts)-1]
	if idx := strings.Index(dbName, "?"); idx != -1 {
		dbName = dbName[:idx]
	}

	return dbName
}
