package mgo

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"time"
)

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

// CursorData 游标数据（公开供聚合使用）
type CursorData struct {
	Values    map[string]interface{} `json:"values"`    // 排序字段值
	ID        string                 `json:"id"`        // 文档ID（十六进制字符串）
	Direction string                 `json:"direction"` // 方向: "next" 或 "prev"
	Version   string                 `json:"v"`         // 游标版本
}

// cursorData 内部类型别名
type cursorData = CursorData

// CursorPage 游标分页结果
//
// 支持双向翻页，提供前后游标
//
// 示例：
//
//	page, err := users.Find().
//	    Desc("created_at").
//	    CursorPage("", 20)
//
//	// 下一页
//	nextPage, _ := users.Find().Desc("created_at").CursorPage(page.NextCursor, 20)
//
//	// 上一页
//	prevPage, _ := users.Find().Desc("created_at").CursorPage(page.PrevCursor, 20)
type CursorPage[T any] struct {
	Items      []*T   // 当前页数据
	NextCursor string // 下一页游标
	PrevCursor string // 上一页游标
	HasMore    bool   // 是否有下一页
}

// CursorPage 使用游标的分页（适用于大数据量，支持双向翻页）
//
// 参数：
//   - cursor: 游标字符串（空字符串表示第一页）
//   - perPage: 每页数量
//
// 特性：
//   - 自动从查询中提取排序字段
//   - 无排序时默认按 _id 降序
//   - 支持多字段排序
//   - 提供前后游标实现双向翻页
//   - 游标解析失败时返回第一页
//
// 示例：
//
//	// 第一页（默认按 _id 降序）
//	page, err := users.Find().
//	    Where("status", "active").
//	    CursorPage("", 20)
//
//	// 自定义排序
//	page, err := users.Find().
//	    Desc("created_at").
//	    CursorPage("", 50)
//
//	// 多字段排序
//	page, err := users.Find().
//	    Asc("age").
//	    Desc("created_at").
//	    CursorPage(cursor, 30)
//
//	// 下一页
//	nextPage, _ := users.Find().Desc("created_at").CursorPage(page.NextCursor, 20)
//
//	// 上一页
//	prevPage, _ := users.Find().Desc("created_at").CursorPage(page.PrevCursor, 20)
func (q *Query[T]) CursorPage(cursor string, perPage int) (*CursorPage[T], error) {
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 1000 {
		perPage = 1000
	}

	// 克隆查询避免修改原查询
	query := q.Clone()

	// 如果没有排序，默认按 _id 降序
	if len(query.sort) == 0 {
		query.sort = D{{Key: "_id", Value: -1}}
	}

	// 解析游标并添加过滤条件
	var cursorDir string
	if cursor != "" {
		data, err := decodeCursor(cursor)
		if err != nil {
			// 游标解析失败，忽略并返回第一页（用户友好）
			// 可以选择记录日志
		} else {
			cursorDir = data.Direction
			// 根据游标方向反转排序（prev时需要反向查询）
			if data.Direction == "prev" {
				query = query.Clone()
				// 反转排序方向
				newSort := make(D, len(query.sort))
				for i, elem := range query.sort {
					if o, ok := elem.Value.(int); ok {
						newSort[i] = E{Key: elem.Key, Value: -o}
					} else {
						newSort[i] = elem
					}
				}
				query.sort = newSort
			}

			// 构建游标过滤条件
			filter := query.buildCursorFilter(data)
			query = query.Filter(filter)
		}
	}

	// 多查询一条来判断是否还有下一页
	limit := int64(perPage + 1)
	items, err := query.Limit(limit).All()
	if err != nil {
		return nil, WrapError(err, "failed to get cursor page items")
	}

	// 如果是反向查询，需要反转结果顺序
	if cursorDir == "prev" {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	hasMore := len(items) > perPage
	if hasMore {
		items = items[:perPage]
	}

	// 生成游标
	var nextCursor, prevCursor string

	// 生成下一页游标（从最后一条记录）
	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = encodeCursor(lastItem, q.sort, "next")
	}

	// 生成上一页游标（从第一条记录）
	// 只有在不是第一页时才生成
	if cursor != "" && len(items) > 0 {
		firstItem := items[0]
		prevCursor = encodeCursor(firstItem, q.sort, "prev")
	}

	return &CursorPage[T]{
		Items:      items,
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
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

// ==================== 游标分页辅助函数 ====================

// encodeCursor 将记录编码为游标字符串
func encodeCursor[T any](item *T, sortDoc D, direction string) string {
	return EncodeCursorFromItem(item, sortDoc, direction)
}

// EncodeCursorFromItem 将记录编码为游标字符串（公开供聚合使用）
func EncodeCursorFromItem[T any](item *T, sortDoc D, direction string) string {
	if item == nil {
		return ""
	}

	data := &cursorData{
		Values:    make(map[string]interface{}),
		Direction: direction,
		Version:   "v1",
	}

	// 提取排序字段值
	itemValue := reflect.ValueOf(item).Elem()
	itemType := itemValue.Type()

	// 构建字段名映射（struct字段名 -> bson字段名）
	fieldMap := make(map[string]string)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		bsonName := GetBSONFieldName(field)
		fieldMap[bsonName] = field.Name
	}

	// 提取排序字段的值
	for _, elem := range sortDoc {
		sortField := elem.Key
		structFieldName, ok := fieldMap[sortField]
		if !ok {
			continue
		}

		fieldValue := itemValue.FieldByName(structFieldName)
		if !fieldValue.IsValid() {
			continue
		}

		// 获取字段值并序列化
		val := fieldValue.Interface()

		// 特殊处理不同类型
		switch v := val.(type) {
		case ObjectID:
			data.Values[sortField] = v.Hex()
		case time.Time:
			data.Values[sortField] = v.Format(time.RFC3339Nano)
		case *time.Time:
			if v != nil {
				data.Values[sortField] = v.Format(time.RFC3339Nano)
			}
		default:
			data.Values[sortField] = val
		}
	}

	// 提取 _id 字段（必需）
	if idField := itemValue.FieldByName(fieldMap["_id"]); idField.IsValid() {
		if id, ok := idField.Interface().(ObjectID); ok {
			data.ID = id.Hex()
		}
	}

	// JSON 序列化 + Base64 编码
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(jsonData)
}

// decodeCursor 解析游标字符串
func decodeCursor(cursor string) (*cursorData, error) {
	return DecodeCursorData(cursor)
}

// DecodeCursorData 解析游标字符串（公开供聚合使用）
func DecodeCursorData(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, WrapError(nil, "empty cursor")
	}

	// Base64 解码
	jsonData, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, WrapError(err, "invalid cursor format")
	}

	// JSON 反序列化
	var data cursorData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, WrapError(err, "invalid cursor data")
	}

	// 验证版本
	if data.Version != "v1" {
		return nil, WrapError(nil, "unsupported cursor version")
	}

	// 转换特殊类型
	for field, val := range data.Values {
		// 尝试解析时间字符串
		if strVal, ok := val.(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, strVal); err == nil {
				data.Values[field] = t
			}
		}
	}

	return &data, nil
}

// buildCursorFilter 根据游标数据构建过滤条件
func (q *Query[T]) buildCursorFilter(data *cursorData) M {
	if data == nil || len(data.Values) == 0 {
		return M{}
	}

	// 获取排序字段和方向
	sortFields := make([]string, 0, len(q.sort))
	sortOrders := make(map[string]int)
	for _, elem := range q.sort {
		sortFields = append(sortFields, elem.Key)
		if order, ok := elem.Value.(int); ok {
			sortOrders[elem.Key] = order
		} else {
			sortOrders[elem.Key] = 1 // 默认升序
		}
	}

	// 如果只有一个排序字段（常见情况），使用简化查询
	if len(sortFields) == 1 {
		field := sortFields[0]
		value := data.Values[field]
		order := sortOrders[field]

		// 根据排序方向选择操作符
		var op string
		if data.Direction == "prev" {
			// 反向查询已在外层反转排序，这里保持一致
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		} else {
			// 正向查询
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		}

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if data.ID != "" {
				if oid, err := ObjectIDFromHex(data.ID); err == nil {
					value = oid
				}
			}
		}

		return M{field: M{op: value}}
	}

	// 多字段排序：构建复杂的 $or 条件
	// 例如：ORDER BY age ASC, created_at DESC
	// 游标值：{age: 25, created_at: "2024-01-01"}
	// 条件：(age > 25) OR (age = 25 AND created_at < "2024-01-01")

	conditions := make([]M, 0)

	for i, field := range sortFields {
		value := data.Values[field]
		order := sortOrders[field]

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if data.ID != "" {
				if oid, err := ObjectIDFromHex(data.ID); err == nil {
					value = oid
				}
			}
		}

		// 根据排序方向选择操作符
		var op string
		if data.Direction == "prev" {
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		} else {
			if order > 0 {
				op = "$gt"
			} else {
				op = "$lt"
			}
		}

		// 构建条件
		condition := M{}

		// 前面的字段必须相等
		for j := 0; j < i; j++ {
			prevField := sortFields[j]
			prevValue := data.Values[prevField]

			// 特殊处理 _id
			if prevField == "_id" {
				if hexID, ok := prevValue.(string); ok {
					if oid, err := ObjectIDFromHex(hexID); err == nil {
						prevValue = oid
					}
				}
			}

			condition[prevField] = prevValue
		}

		// 当前字段使用比较操作符
		condition[field] = M{op: value}

		conditions = append(conditions, condition)
	}

	// 如果只有一个条件，直接返回
	if len(conditions) == 1 {
		return conditions[0]
	}

	// 多个条件用 $or 连接
	return M{"$or": conditions}
}

// HasPrev 是否有上一页
func (p *CursorPage[T]) HasPrev() bool {
	return p.PrevCursor != ""
}

// IsEmpty 是否为空
func (p *CursorPage[T]) IsEmpty() bool {
	return len(p.Items) == 0
}

// Count 当前页记录数
func (p *CursorPage[T]) Count() int {
	return len(p.Items)
}
