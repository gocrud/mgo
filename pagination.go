package mgo

// ==================== 分页功能 ====================

// PageResult 分页结果
//
// 示例：
//
//	page, err := users.Find().Page(1, 20)
//	fmt.Printf("Total: %d, Pages: %d\n", page.Total, page.Pages)
//	for _, user := range page.Items {
//	    fmt.Println(user.Name)
//	}
type PageResult[T any] struct {
	Items   []*T  // 当前页数据
	Total   int64 // 总记录数
	Page    int   // 当前页码（从 1 开始）
	PerPage int   // 每页数量
	Pages   int   // 总页数
}

// Page 分页查询
//
// 参数：
//   - page: 页码（从 1 开始）
//   - perPage: 每页数量
//
// 示例：
//
//	page, err := users.Find().
//	    Where("status", "active").
//	    Page(1, 20)
func (q *Query[T]) Page(page, perPage int, opts ...PageOption) (*PageResult[T], error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000 // 最大限制
	}

	// 解析选项
	options := &PageOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 计算 skip
	skip := int64((page - 1) * perPage)
	limit := int64(perPage)

	// 获取总数（如果未禁用）
	var total int64
	var err error
	if !options.DisableCount {
		total, err = q.Clone().Count()
		if err != nil {
			return nil, WrapError(err, "failed to count for pagination")
		}
	}

	// 获取当前页数据
	items, err := q.Clone().Skip(skip).Limit(limit).All()
	if err != nil {
		return nil, WrapError(err, "failed to get page items")
	}

	// 计算总页数
	pages := 0
	if total > 0 {
		pages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return &PageResult[T]{
		Items:   items,
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Pages:   pages,
	}, nil
}

// Paginate Paginate 是 Page 的别名
//
// 示例：
//
//	page, err := users.Find().Paginate(1, 20)
func (q *Query[T]) Paginate(page, perPage int, opts ...PageOption) (*PageResult[T], error) {
	return q.Page(page, perPage, opts...)
}

// ==================== 简化分页方法 ====================

// SimplePage 简化的分页（只包含数据和当前页信息）
type SimplePage[T any] struct {
	Items   []*T // 当前页数据
	Page    int  // 当前页码
	PerPage int  // 每页数量
	HasMore bool // 是否有下一页
}

// SimplePaginate 简化的分页（不统计总数，性能更好）
//
// 示例：
//
//	page, err := users.Find().SimplePaginate(1, 20)
func (q *Query[T]) SimplePaginate(page, perPage int) (*SimplePage[T], error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}

	// 多查询一条来判断是否还有下一页
	skip := int64((page - 1) * perPage)
	limit := int64(perPage + 1)

	items, err := q.Clone().Skip(skip).Limit(limit).All()
	if err != nil {
		return nil, WrapError(err, "failed to get page items")
	}

	hasMore := len(items) > perPage
	if hasMore {
		items = items[:perPage]
	}

	return &SimplePage[T]{
		Items:   items,
		Page:    page,
		PerPage: perPage,
		HasMore: hasMore,
	}, nil
}

// ==================== 游标分页 ====================

// CursorPage 游标分页结果
type CursorPage[T any] struct {
	Items      []*T   // 当前页数据
	NextCursor string // 下一页游标
	HasMore    bool   // 是否有下一页
}

// CursorPaginate 使用游标的分页（适用于大数据量）
//
// 示例：
//
//	// 第一页
//	page, err := users.Find().CursorPaginate("", 20)
//
//	// 下一页
//	page, err = users.Find().CursorPaginate(page.NextCursor, 20)
func (q *Query[T]) CursorPaginate(cursor string, perPage int) (*CursorPage[T], error) {
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}

	// 如果有游标，添加游标条件
	if cursor != "" {
		// TODO: 解析游标并添加查询条件
		// 这需要根据实际的游标格式来实现
	}

	// 多查询一条来判断是否还有下一页
	limit := int64(perPage + 1)
	items, err := q.Clone().Limit(limit).All()
	if err != nil {
		return nil, WrapError(err, "failed to get cursor page items")
	}

	hasMore := len(items) > perPage
	if hasMore {
		items = items[:perPage]
	}

	// 生成下一页游标
	var nextCursor string
	if hasMore && len(items) > 0 {
		// TODO: 根据最后一条记录生成游标
		// 这需要根据实际需求来实现
		nextCursor = "next_cursor_placeholder"
	}

	return &CursorPage[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// ==================== 分页辅助方法 ====================

// HasNext 是否有下一页
func (p *PageResult[T]) HasNext() bool {
	return p.Page < p.Pages
}

// HasPrev 是否有上一页
func (p *PageResult[T]) HasPrev() bool {
	return p.Page > 1
}

// NextPage 下一页页码
func (p *PageResult[T]) NextPage() int {
	if p.HasNext() {
		return p.Page + 1
	}
	return p.Page
}

// PrevPage 上一页页码
func (p *PageResult[T]) PrevPage() int {
	if p.HasPrev() {
		return p.Page - 1
	}
	return p.Page
}

// IsEmpty 是否为空
func (p *PageResult[T]) IsEmpty() bool {
	return len(p.Items) == 0
}

// Count 当前页记录数
func (p *PageResult[T]) Count() int {
	return len(p.Items)
}
