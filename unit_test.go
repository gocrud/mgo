package mgo

import (
	"testing"
	"time"
)

// ==================== 类型别名测试 ====================

func TestObjectID(t *testing.T) {
	t.Run("NewObjectID", func(t *testing.T) {
		id := NewObjectID()
		if id == NilObjectID {
			t.Error("NewObjectID 返回零值")
		}
	})

	t.Run("ObjectIDFromHex", func(t *testing.T) {
		validHex := "507f1f77bcf86cd799439011"
		id, err := ObjectIDFromHex(validHex)
		if err != nil {
			t.Errorf("ObjectIDFromHex 失败: %v", err)
		}

		if id.Hex() != validHex {
			t.Errorf("期望 %s，得到 %s", validHex, id.Hex())
		}
	})

	t.Run("IsValidObjectID", func(t *testing.T) {
		tests := []struct {
			hex   string
			valid bool
		}{
			{"507f1f77bcf86cd799439011", true},
			{"invalid", false},
			{"507f1f77bcf86cd799439011extra", false},
			{"", false},
		}

		for _, tt := range tests {
			result := IsValidObjectID(tt.hex)
			if result != tt.valid {
				t.Errorf("IsValidObjectID(%s) = %v，期望 %v", tt.hex, result, tt.valid)
			}
		}
	})

	t.Run("MustObjectIDFromHex", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustObjectIDFromHex 应该 panic")
			}
		}()

		MustObjectIDFromHex("invalid")
	})
}

func TestHelpers(t *testing.T) {
	t.Run("ToSnakeCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"UserProfile", "user_profile"},
			{"TestUser", "test_user"},
			{"ID", "i_d"},
			{"HTTPServer", "h_t_t_p_server"},
		}

		for _, tt := range tests {
			result := ToSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToSnakeCase(%s) = %s，期望 %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("ToCamelCase", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"user_profile", "UserProfile"},
			{"test_user", "TestUser"},
			{"id", "Id"},
		}

		for _, tt := range tests {
			result := ToCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToCamelCase(%s) = %s，期望 %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("Pluralize", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"User", "users"},
			{"City", "cities"},
			{"Person", "people"},
			{"Child", "children"},
			{"Box", "boxes"},
			{"Status", "statuses"},
		}

		for _, tt := range tests {
			result := Pluralize(tt.input)
			if result != tt.expected {
				t.Errorf("Pluralize(%s) = %s，期望 %s", tt.input, result, tt.expected)
			}
		}
	})

	t.Run("InferCollectionName", func(t *testing.T) {
		tests := []struct {
			typeName string
			expected string
		}{
			{"User", "users"},
			{"UserProfile", "user_profiles"},
			{"OrderItem", "order_items"},
		}

		for _, tt := range tests {
			result := InferCollectionName(tt.typeName)
			if result != tt.expected {
				t.Errorf("InferCollectionName(%s) = %s，期望 %s", tt.typeName, result, tt.expected)
			}
		}
	})
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"2024-01-01", true},
		{"2024-01-01 15:04:05", true},
		{"2024-01-01T15:04:05Z", true},
		{"2024-01-01T15:04:05", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		_, err := ParseTime(tt.input)
		if (err == nil) != tt.valid {
			t.Errorf("ParseTime(%s) 错误状态不符，期望 valid=%v", tt.input, tt.valid)
		}
	}
}

func TestNormalizeValue(t *testing.T) {
	t.Run("StringTime", func(t *testing.T) {
		result := NormalizeValue("2024-01-01 08:00:00")
		if _, ok := result.(time.Time); !ok {
			t.Error("字符串时间应该转换为 time.Time")
		}
	})

	t.Run("TimeValue", func(t *testing.T) {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		localTime := time.Date(2024, 1, 1, 8, 0, 0, 0, loc)

		result := NormalizeValue(localTime)
		utcTime, ok := result.(time.Time)
		if !ok {
			t.Fatal("应该返回 time.Time")
		}

		if utcTime.Location() != time.UTC {
			t.Error("时间应该转换为 UTC")
		}
	})

	t.Run("NonTimeValue", func(t *testing.T) {
		result := NormalizeValue("regular string")
		if result != "regular string" {
			t.Error("普通字符串不应该被转换")
		}

		result = NormalizeValue(123)
		if result != 123 {
			t.Error("数字不应该被转换")
		}
	})
}

func TestFilterBuilders(t *testing.T) {
	t.Run("Eq", func(t *testing.T) {
		filter := Eq("status", "active")
		if len(filter) != 1 {
			t.Error("Eq 应该生成一个条件")
		}
		if filter["status"] != "active" {
			t.Error("Eq 条件值错误")
		}
	})

	t.Run("Gt", func(t *testing.T) {
		filter := Gt("age", 18)
		if len(filter) != 1 {
			t.Error("Gt 应该生成一个条件")
		}

		ageFilter, ok := filter["age"].(M)
		if !ok {
			t.Fatal("Gt 应该生成嵌套 map")
		}

		if ageFilter["$gt"] != 18 {
			t.Error("Gt 操作符错误")
		}
	})

	t.Run("In", func(t *testing.T) {
		filter := In("status", "active", "pending")
		statusFilter, ok := filter["status"].(M)
		if !ok {
			t.Fatal("In 应该生成嵌套 map")
		}

		values, ok := statusFilter["$in"].([]interface{})
		if !ok {
			t.Fatal("In 应该包含值数组")
		}

		if len(values) != 2 {
			t.Errorf("期望 2 个值，得到 %d 个", len(values))
		}
	})

	t.Run("Between", func(t *testing.T) {
		filter := Between("age", 18, 60)
		ageFilter, ok := filter["age"].(M)
		if !ok {
			t.Fatal("Between 应该生成嵌套 map")
		}

		if ageFilter["$gte"] != 18 || ageFilter["$lte"] != 60 {
			t.Error("Between 范围值错误")
		}
	})
}

func TestLogicFilters(t *testing.T) {
	t.Run("And", func(t *testing.T) {
		filter := And(
			Eq("status", "active"),
			Gt("age", 18),
		)

		if _, ok := filter["$and"]; !ok {
			t.Error("And 应该生成 $and 操作符")
		}
	})

	t.Run("Or", func(t *testing.T) {
		filter := Or(
			Eq("status", "active"),
			Eq("status", "pending"),
		)

		if _, ok := filter["$or"]; !ok {
			t.Error("Or 应该生成 $or 操作符")
		}
	})

	t.Run("SingleCondition", func(t *testing.T) {
		filter := And(Eq("status", "active"))
		// 单个条件应该直接返回，不包装 $and
		if filter["status"] != "active" {
			t.Error("单个条件不应该包装")
		}
	})
}

func TestErrorFunctions(t *testing.T) {
	t.Run("IsNoDocuments", func(t *testing.T) {
		if IsNoDocuments(nil) {
			t.Error("nil 不应该是 NoDocuments")
		}

		if !IsNoDocuments(ErrNoDocuments) {
			t.Error("ErrNoDocuments 应该被识别")
		}
	})

	t.Run("WrapError", func(t *testing.T) {
		err := WrapError(ErrNotFound, "failed to find user")
		if err == nil {
			t.Error("WrapError 不应该返回 nil")
		}

		errStr := err.Error()
		if len(errStr) == 0 {
			t.Error("错误信息不应该为空")
		}
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("email", "email is required")
		if err == nil {
			t.Error("NewValidationError 不应该返回 nil")
		}

		verr, ok := err.(*ValidationError)
		if !ok {
			t.Fatal("应该返回 ValidationError 类型")
		}

		if verr.Field != "email" {
			t.Errorf("期望字段 'email'，得到 '%s'", verr.Field)
		}
	})
}

func TestChunkSlice(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7}

	chunks := ChunkSlice(items, 3)
	if len(chunks) != 3 {
		t.Errorf("期望 3 个块，得到 %d 个", len(chunks))
	}

	if len(chunks[0]) != 3 {
		t.Errorf("第一块应该是 3 个元素，得到 %d 个", len(chunks[0]))
	}

	if len(chunks[2]) != 1 {
		t.Errorf("最后一块应该是 1 个元素，得到 %d 个", len(chunks[2]))
	}
}

func TestMergeMaps(t *testing.T) {
	map1 := M{"a": 1, "b": 2}
	map2 := M{"b": 3, "c": 4}

	result := MergeMaps(map1, map2)

	if result["a"] != 1 {
		t.Error("a 的值应该是 1")
	}

	if result["b"] != 3 {
		t.Error("b 的值应该被覆盖为 3")
	}

	if result["c"] != 4 {
		t.Error("c 的值应该是 4")
	}
}

func TestParseOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{">", "$gt"},
		{">=", "$gte"},
		{"<", "$lt"},
		{"<=", "$lte"},
		{"!=", "$ne"},
		{"=", "$eq"},
		{"in", "$in"},
		{"like", "$regex"},
		{"$custom", "$custom"},
		{"unknown", "$eq"},
	}

	for _, tt := range tests {
		result := ParseOperator(tt.input)
		if result != tt.expected {
			t.Errorf("ParseOperator(%s) = %s，期望 %s", tt.input, result, tt.expected)
		}
	}
}

func TestToObjectID(t *testing.T) {
	t.Run("FromObjectID", func(t *testing.T) {
		original := NewObjectID()
		result, err := ToObjectID(original)
		if err != nil {
			t.Errorf("ToObjectID 失败: %v", err)
		}

		if result != original {
			t.Error("ObjectID 转换应该返回原值")
		}
	})

	t.Run("FromString", func(t *testing.T) {
		hex := "507f1f77bcf86cd799439011"
		result, err := ToObjectID(hex)
		if err != nil {
			t.Errorf("从字符串转换失败: %v", err)
		}

		if result.Hex() != hex {
			t.Error("转换后的 hex 不一致")
		}
	})

	t.Run("FromInvalid", func(t *testing.T) {
		_, err := ToObjectID(12345)
		if err == nil {
			t.Error("无效类型应该返回错误")
		}
	})
}

func TestNewDecimal128(t *testing.T) {
	t.Run("FromString", func(t *testing.T) {
		dec, err := NewDecimal128("123.45")
		if err != nil {
			t.Errorf("NewDecimal128 失败: %v", err)
		}

		_ = dec
	})

	t.Run("FromFloat", func(t *testing.T) {
		dec := NewDecimal128FromFloat64(123.45)
		_ = dec
	})
}

func TestTimestampHelpers(t *testing.T) {
	type TestDoc struct {
		ID        ObjectID  `bson:"_id,omitempty"`
		Name      string    `bson:"name"`
		CreatedAt time.Time `bson:"created_at"`
		UpdatedAt time.Time `bson:"updated_at"`
	}

	t.Run("applyTimestamps Insert", func(t *testing.T) {
		doc := &TestDoc{Name: "test"}
		config := &TimestampConfig{
			CreatedField: "created_at",
			UpdatedField: "updated_at",
			Enabled:      true,
		}

		applyTimestamps(doc, config, true)

		if doc.CreatedAt.IsZero() {
			t.Error("CreatedAt 应该已设置")
		}

		if doc.UpdatedAt.IsZero() {
			t.Error("UpdatedAt 应该已设置")
		}
	})

	t.Run("applyTimestamps Update", func(t *testing.T) {
		doc := &TestDoc{Name: "test"}
		config := &TimestampConfig{
			CreatedField: "created_at",
			UpdatedField: "updated_at",
			Enabled:      true,
		}

		oldCreatedAt := doc.CreatedAt
		applyTimestamps(doc, config, false)

		if doc.CreatedAt != oldCreatedAt {
			t.Error("更新时不应该修改 CreatedAt")
		}

		if doc.UpdatedAt.IsZero() {
			t.Error("UpdatedAt 应该已设置")
		}
	})

	t.Run("GetTimestampFields", func(t *testing.T) {
		doc := TestDoc{}
		created, updated := GetTimestampFields(doc)

		if created != "created_at" {
			t.Errorf("期望 created_at，得到 %s", created)
		}

		if updated != "updated_at" {
			t.Errorf("期望 updated_at，得到 %s", updated)
		}
	})
}

func TestFieldHelpers(t *testing.T) {
	type TestDoc struct {
		Name  string `bson:"name"`
		Email string `bson:"email"`
		Age   int    `bson:"age"`
	}

	t.Run("SetFieldValue", func(t *testing.T) {
		doc := TestDoc{Name: "原名"}

		err := SetFieldValue(&doc, "name", "新名字")
		if err != nil {
			t.Errorf("SetFieldValue 失败: %v", err)
		}

		if doc.Name != "新名字" {
			t.Errorf("期望 '新名字'，得到 '%s'", doc.Name)
		}
	})
}

func TestGetStructFields(t *testing.T) {
	type TestDoc struct {
		Name       string `bson:"name"`
		Email      string `bson:"email"`
		Age        int    `bson:"age"`
		Ignored    string `bson:"-"`
		unexported string
	}

	fields := GetStructFields(TestDoc{})

	if len(fields) != 3 {
		t.Errorf("期望 3 个字段，得到 %d 个", len(fields))
	}

	if fields["Name"] != "name" {
		t.Error("Name 字段映射错误")
	}

	if _, ok := fields["Ignored"]; ok {
		t.Error("Ignored 字段不应该包含")
	}

	if _, ok := fields["unexported"]; ok {
		t.Error("未导出字段不应该包含")
	}
}

func TestMustFunctions(t *testing.T) {
	t.Run("MustParseTime", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustParseTime 应该 panic")
			}
		}()

		MustParseTime("invalid")
	})

	t.Run("MustDecimal128", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDecimal128 应该 panic")
			}
		}()

		MustDecimal128("invalid")
	})

	t.Run("MustToObjectID", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustToObjectID 应该 panic")
			}
		}()

		MustToObjectID("invalid")
	})
}

// BenchmarkHelpers 性能测试
func BenchmarkToSnakeCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToSnakeCase("UserProfileSettings")
	}
}

func BenchmarkPluralize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Pluralize("User")
	}
}

func BenchmarkInferCollectionName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InferCollectionName("UserProfile")
	}
}

func BenchmarkParseTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseTime("2024-01-01 15:04:05")
	}
}
