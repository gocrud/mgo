package mgo

import (
	"reflect"
	"strings"
	"time"
	"unicode"
)

// ==================== 字符串辅助函数 ====================

// ToSnakeCase 将驼峰命名转换为蛇形命名
//
// 示例：
//
//	result := mgo.ToSnakeCase("UserProfile") // "user_profile"
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ToCamelCase 将蛇形命名转换为驼峰命名
//
// 示例：
//
//	result := mgo.ToCamelCase("user_profile") // "UserProfile"
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteRune(unicode.ToUpper(rune(part[0])))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}

// Pluralize 将单词转换为复数形式（简单规则）
//
// 示例：
//
//	result := mgo.Pluralize("User") // "users"
//	result := mgo.Pluralize("City") // "cities"
func Pluralize(s string) string {
	s = strings.ToLower(s)

	// 特殊情况
	special := map[string]string{
		"person": "people",
		"child":  "children",
		"man":    "men",
		"woman":  "women",
	}
	if plural, ok := special[s]; ok {
		return plural
	}

	// 一般规则
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		// city -> cities
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "z") || strings.HasSuffix(s, "ch") ||
		strings.HasSuffix(s, "sh") {
		// box -> boxes, buzz -> buzzes
		return s + "es"
	}

	return s + "s"
}

// ==================== 集合名推断 ====================

// InferCollectionName 从类型名推断集合名
//
// 规则：
//  1. 类型名转换为蛇形命名
//  2. 转换为复数形式
//
// 示例：
//
//	User -> users
//	OrderItem -> order_items
//	UserProfile -> user_profiles
func InferCollectionName(typeName string) string {
	// 移除指针标记
	typeName = strings.TrimPrefix(typeName, "*")

	// 转换为蛇形命名
	snake := ToSnakeCase(typeName)

	// 转换为复数
	return Pluralize(snake)
}

// InferCollectionNameFromType 从类型推断集合名
//
// 示例：
//
//	var user User
//	name := mgo.InferCollectionNameFromType(user) // "users"
func InferCollectionNameFromType(v interface{}) string {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return InferCollectionName(t.Name())
}

// ==================== 时间辅助函数 ====================

// ParseTime 解析时间字符串
//
// 支持的格式：
//   - "2006-01-02"
//   - "2006-01-02 15:04:05"
//   - "2006-01-02T15:04:05Z"
//   - RFC3339
//
// 示例：
//
//	t, err := mgo.ParseTime("2024-01-01 08:00:00")
func ParseTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, ErrInvalidOperation
}

// MustParseTime 解析时间字符串，失败时 panic
//
// 示例：
//
//	t := mgo.MustParseTime("2024-01-01 08:00:00")
func MustParseTime(s string) time.Time {
	t, err := ParseTime(s)
	if err != nil {
		panic(err)
	}
	return t
}

// NormalizeValue 标准化值（自动转换时间为 UTC）
//
// 示例：
//
//	value := mgo.NormalizeValue("2024-01-01 08:00:00")
//	// 返回 UTC 时间
func NormalizeValue(v interface{}) interface{} {
	switch val := v.(type) {
	case string:
		// 尝试解析为时间
		if t, err := ParseTime(val); err == nil {
			return t.In(time.Local).UTC()
		}
		return val
	case time.Time:
		return val.UTC()
	default:
		return v
	}
}

// ==================== 反射辅助函数 ====================

// GetBSONFieldName 获取字段的 BSON 名称
//
// 示例：
//
//	type User struct {
//	    Name string `bson:"name"`
//	}
//	field, _ := reflect.TypeOf(User{}).FieldByName("Name")
//	name := mgo.GetBSONFieldName(field) // "name"
func GetBSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("bson")
	if tag == "" {
		return ToSnakeCase(field.Name)
	}

	// 解析 bson 标签
	parts := strings.Split(tag, ",")
	if parts[0] == "-" {
		return ""
	}
	if parts[0] != "" {
		return parts[0]
	}

	return ToSnakeCase(field.Name)
}

// GetStructFields 获取结构体的所有字段信息
//
// 示例：
//
//	fields := mgo.GetStructFields(User{})
//	for name, bsonName := range fields {
//	    fmt.Printf("%s -> %s\n", name, bsonName)
//	}
func GetStructFields(v interface{}) map[string]string {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	fields := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		bsonName := GetBSONFieldName(field)
		if bsonName != "" {
			fields[field.Name] = bsonName
		}
	}

	return fields
}

// IsZeroValue 检查值是否为零值
//
// 示例：
//
//	if mgo.IsZeroValue(user.Name) {
//	    // Name 为空字符串
//	}
func IsZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	return val.IsZero()
}

// ==================== 切片辅助函数 ====================

// ChunkSlice 将切片分块
//
// 示例：
//
//	items := []int{1, 2, 3, 4, 5}
//	chunks := mgo.ChunkSlice(items, 2)
//	// [[1, 2], [3, 4], [5]]
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}

	chunks := make([][]T, 0, (len(slice)+chunkSize-1)/chunkSize)
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ==================== Map 辅助函数 ====================

// MergeMaps 合并多个 map
//
// 示例：
//
//	result := mgo.MergeMaps(
//	    mgo.M{"a": 1, "b": 2},
//	    mgo.M{"b": 3, "c": 4},
//	)
//	// {"a": 1, "b": 3, "c": 4}
func MergeMaps(maps ...M) M {
	result := M{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// CopyMap 复制 map
//
// 示例：
//
//	original := mgo.M{"a": 1, "b": 2}
//	copy := mgo.CopyMap(original)
func CopyMap(m M) M {
	result := make(M, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ==================== ID 辅助函数 ====================

// ToObjectID 将各种类型转换为 ObjectID
//
// 支持：string, []byte, ObjectID
//
// 示例：
//
//	id, err := mgo.ToObjectID("507f1f77bcf86cd799439011")
//	id, err := mgo.ToObjectID([]byte{...})
func ToObjectID(v interface{}) (ObjectID, error) {
	switch val := v.(type) {
	case ObjectID:
		return val, nil
	case string:
		return ObjectIDFromHex(val)
	case []byte:
		if len(val) != 12 {
			return NilObjectID, ErrInvalidID
		}
		var oid ObjectID
		copy(oid[:], val)
		return oid, nil
	default:
		return NilObjectID, ErrInvalidID
	}
}

// MustToObjectID 将值转换为 ObjectID，失败时 panic
//
// 示例：
//
//	id := mgo.MustToObjectID("507f1f77bcf86cd799439011")
func MustToObjectID(v interface{}) ObjectID {
	id, err := ToObjectID(v)
	if err != nil {
		panic(err)
	}
	return id
}

// ==================== 操作符解析 ====================

// ParseOperator 解析操作符字符串
//
// 支持：>, >=, <, <=, !=, in, nin, exists, like
//
// 示例：
//
//	op := mgo.ParseOperator(">")  // "$gt"
//	op := mgo.ParseOperator("in") // "$in"
func ParseOperator(op string) string {
	operators := map[string]string{
		">":      "$gt",
		">=":     "$gte",
		"<":      "$lt",
		"<=":     "$lte",
		"!=":     "$ne",
		"=":      "$eq",
		"in":     "$in",
		"nin":    "$nin",
		"exists": "$exists",
		"regex":  "$regex",
		"like":   "$regex",
	}

	if mongoOp, ok := operators[op]; ok {
		return mongoOp
	}

	// 如果已经是 MongoDB 操作符，直接返回
	if strings.HasPrefix(op, "$") {
		return op
	}

	return "$eq"
}

// ==================== 字段值设置 ====================

// SetFieldValue 设置字段值
//
// 示例：
//
//	err := mgo.SetFieldValue(&user, "_id", objectID)
func SetFieldValue(doc interface{}, fieldName string, value interface{}) error {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ErrInvalidOperation
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		bsonName := GetBSONFieldName(field)

		if bsonName == fieldName || field.Name == fieldName {
			fieldVal := val.Field(i)
			if !fieldVal.CanSet() {
				return ErrInvalidOperation
			}

			valueVal := reflect.ValueOf(value)
			if valueVal.Type().AssignableTo(fieldVal.Type()) {
				fieldVal.Set(valueVal)
				return nil
			}
			return ErrInvalidOperation
		}
	}

	return ErrNotFound
}
