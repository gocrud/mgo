package mgo

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"time"
)

// ==================== 非泛型分页功能 ====================

// UntypedPageResult 非泛型分页结果
//
// 示例：
//
//	page, err := coll.Find().PageList(1, 20, &users)
//	fmt.Printf("Total: %d, Pages: %d\n", page.Total, page.Pages)
type UntypedPageResult struct {
	Total   int64 // 总记录数
	Page    int   // 当前页码（从 1 开始）
	PerPage int   // 每页数量
	Pages   int   // 总页数
}

// UntypedSimplePageList 非泛型简化分页结果
type UntypedSimplePageList struct {
	Page    int  // 当前页码
	PerPage int  // 每页数量
	HasMore bool // 是否有下一页
}

// UntypedCursorPage 非泛型游标分页结果
type UntypedCursorPage struct {
	NextCursor string // 下一页游标
	PrevCursor string // 上一页游标
	HasMore    bool   // 是否有下一页
}

// PageList 分页查询（非泛型版本）
//
// 参数：
//   - page: 页码（从 1 开始）
//   - perPage: 每页数量
//   - results: 接收结果的切片指针
//
// 示例：
//
//	var users []User
//	page, err := coll.Find().
//	    Where("status", "active").
//	    PageList(1, 20, &users)
func (q *UntypedQuery) PageList(page, perPage int, results interface{}, opts ...PageOption) (*UntypedPageResult, error) {
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
	if err := q.Clone().Skip(skip).Limit(limit).All(results); err != nil {
		return nil, WrapError(err, "failed to get page items")
	}

	// 计算总页数
	pages := 0
	if total > 0 {
		pages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return &UntypedPageResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Pages:   pages,
	}, nil
}

// SimplePageList 简化的分页（不统计总数，性能更好）
//
// 示例：
//
//	var users []User
//	page, err := coll.Find().SimplePageList(1, 20, &users)
func (q *UntypedQuery) SimplePageList(page, perPage int, results interface{}) (*UntypedSimplePageList, error) {
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

	// 创建临时切片来接收结果
	tempResults := reflect.New(reflect.TypeOf(results).Elem()).Interface()
	if err := q.Clone().Skip(skip).Limit(limit).All(tempResults); err != nil {
		return nil, WrapError(err, "failed to get page items")
	}

	// 检查是否有更多数据
	tempValue := reflect.ValueOf(tempResults).Elem()
	hasMore := tempValue.Len() > perPage
	if hasMore {
		// 截取到 perPage 数量
		tempValue.Set(tempValue.Slice(0, perPage))
	}

	// 将结果复制到用户提供的切片
	reflect.ValueOf(results).Elem().Set(tempValue)

	return &UntypedSimplePageList{
		Page:    page,
		PerPage: perPage,
		HasMore: hasMore,
	}, nil
}

// CursorPage 使用游标的分页（适用于大数据量，支持双向翻页）
//
// 参数：
//   - cursor: 游标字符串（空字符串表示第一页）
//   - perPage: 每页数量
//   - results: 接收结果的切片指针
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
//	var users []User
//	page, err := coll.Find().
//	    Where("status", "active").
//	    CursorPage("", 20, &users)
//
//	// 下一页
//	nextPage, _ := coll.Find().CursorPage(page.NextCursor, 20, &users)
func (q *UntypedQuery) CursorPage(cursor string, perPage int, results interface{}) (*UntypedCursorPage, error) {
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
		} else {
			cursorDir = data.Direction
			// 根据游标方向反转排序（prev时需要反向查询）
			if data.Direction == "prev" {
				query = query.Clone()
				// 反转排序方向
				newSort := make(D, len(query.sort))
				for i, elem := range query.sort {
					newSort[i] = E{
						Key:   elem.Key,
						Value: -elem.Value.(int),
					}
				}
				query.sort = newSort
			}

			// 添加游标过滤条件
			cursorFilter := buildUntypedCursorFilter(query.sort, data.Values, data.ID, data.Direction)
			query = query.Filter(cursorFilter)
		}
	}

	// 多查询一条来判断是否还有下一页
	limit := int64(perPage + 1)

	// 创建临时切片来接收结果
	tempResults := reflect.New(reflect.TypeOf(results).Elem()).Interface()
	if err := query.Limit(limit).All(tempResults); err != nil {
		return nil, WrapError(err, "failed to get cursor page items")
	}

	tempValue := reflect.ValueOf(tempResults).Elem()
	itemCount := tempValue.Len()
	hasMore := itemCount > perPage

	// 截取到 perPage 数量
	if hasMore {
		tempValue.Set(tempValue.Slice(0, perPage))
		itemCount = perPage
	}

	// 如果是 prev 方向，需要反转结果（因为我们反向查询了）
	if cursorDir == "prev" && itemCount > 0 {
		// 反转切片
		for i := 0; i < itemCount/2; i++ {
			j := itemCount - 1 - i
			temp := tempValue.Index(i).Interface()
			tempValue.Index(i).Set(tempValue.Index(j))
			tempValue.Index(j).Set(reflect.ValueOf(temp))
		}
	}

	// 将结果复制到用户提供的切片
	reflect.ValueOf(results).Elem().Set(tempValue)

	// 生成游标
	var nextCursor, prevCursor string
	if itemCount > 0 {
		// 获取第一条和最后一条记录
		firstItem := tempValue.Index(0).Interface()
		lastItem := tempValue.Index(itemCount - 1).Interface()

		// 生成下一页游标（基于最后一条记录）
		if hasMore {
			nextCursor = encodeUntypedCursor(lastItem, query.sort, "next")
		}

		// 生成上一页游标（基于第一条记录）
		if cursor != "" { // 不是第一页
			prevCursor = encodeUntypedCursor(firstItem, query.sort, "prev")
		}
	}

	return &UntypedCursorPage{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		HasMore:    hasMore,
	}, nil
}

// CursorPaginate CursorPaginate 是 CursorPage 的别名
//
// 示例：
//
//	var users []User
//	page, err := coll.Find().CursorPaginate("", 20, &users)
func (q *UntypedQuery) CursorPaginate(cursor string, perPage int, results interface{}) (*UntypedCursorPage, error) {
	return q.CursorPage(cursor, perPage, results)
}

// ==================== 辅助函数 ====================

// buildUntypedCursorFilter 根据游标数据构建过滤条件（非泛型版本）
func buildUntypedCursorFilter(sortDoc D, values map[string]interface{}, id string, direction string) M {
	if len(values) == 0 {
		return M{}
	}

	// 获取排序字段和方向
	sortFields := make([]string, 0, len(sortDoc))
	sortOrders := make(map[string]int)
	for _, elem := range sortDoc {
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
		value := values[field]
		order := sortOrders[field]

		// 根据排序方向选择操作符
		var op string
		if direction == "prev" {
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

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if id != "" {
				if oid, err := ObjectIDFromHex(id); err == nil {
					value = oid
				}
			}
		}

		return M{field: M{op: value}}
	}

	// 多字段排序：构建复杂的 $or 条件
	conditions := make([]M, 0)

	for i, field := range sortFields {
		value := values[field]
		order := sortOrders[field]

		// 特殊处理 _id 字段
		if field == "_id" {
			if hexID, ok := value.(string); ok {
				if oid, err := ObjectIDFromHex(hexID); err == nil {
					value = oid
				}
			} else if id != "" {
				if oid, err := ObjectIDFromHex(id); err == nil {
					value = oid
				}
			}
		}

		// 根据排序方向选择操作符
		var op string
		if direction == "prev" {
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
			prevValue := values[prevField]

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

// encodeUntypedCursor 将记录编码为游标字符串（非泛型版本）
func encodeUntypedCursor(item interface{}, sortDoc D, direction string) string {
	if item == nil {
		return ""
	}

	data := &CursorData{
		Values:    make(map[string]interface{}),
		Direction: direction,
		Version:   "v1",
	}

	// 使用反射提取排序字段值
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}

	if itemValue.Kind() != reflect.Struct {
		return ""
	}

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
