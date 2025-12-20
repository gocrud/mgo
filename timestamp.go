package mgo

import (
	"reflect"
	"time"
)

// ==================== 自动时间戳功能 ====================

// applyTimestamps 应用时间戳（实现）
//
// isInsert: true 表示插入操作，false 表示更新操作
func applyTimestamps(doc interface{}, config *TimestampConfig, isInsert bool) {
	if config == nil || !config.Enabled {
		return
	}

	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	now := time.Now().UTC()

	// 插入时设置 created_at 和 updated_at
	if isInsert {
		setTimeField(val, config.CreatedField, now)
		setTimeField(val, config.UpdatedField, now)
	} else {
		// 更新时只设置 updated_at
		setTimeField(val, config.UpdatedField, now)
	}
}

// setTimeField 设置时间字段
func setTimeField(val reflect.Value, fieldName string, t time.Time) {
	// 查找字段（通过 bson 标签）
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		bsonName := GetBSONFieldName(field)

		if bsonName == fieldName {
			fieldVal := val.Field(i)
			if !fieldVal.CanSet() {
				continue
			}

			// 根据字段类型设置值
			switch fieldVal.Kind() {
			case reflect.Struct:
				if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
					fieldVal.Set(reflect.ValueOf(t))
				}
			case reflect.Ptr:
				if fieldVal.Type() == reflect.TypeOf((*time.Time)(nil)) {
					fieldVal.Set(reflect.ValueOf(&t))
				}
			case reflect.Int64:
				// Unix timestamp (seconds)
				fieldVal.SetInt(t.Unix())
			case reflect.Int:
				// Unix timestamp (seconds)
				fieldVal.SetInt(t.Unix())
			}
			break
		}
	}
}

// ==================== 时间戳辅助函数 ====================

// TouchTimestamps 手动更新时间戳
//
// 示例：
//
//	mgo.TouchTimestamps(&user, config)
func TouchTimestamps(doc interface{}, config *TimestampConfig) {
	applyTimestamps(doc, config, false)
}

// SetCreatedTimestamp 设置创建时间戳
//
// 示例：
//
//	mgo.SetCreatedTimestamp(&user, config)
func SetCreatedTimestamp(doc interface{}, config *TimestampConfig) {
	if config == nil || !config.Enabled {
		return
	}

	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	setTimeField(val, config.CreatedField, time.Now().UTC())
}

// SetUpdatedTimestamp 设置更新时间戳
//
// 示例：
//
//	mgo.SetUpdatedTimestamp(&user, config)
func SetUpdatedTimestamp(doc interface{}, config *TimestampConfig) {
	if config == nil || !config.Enabled {
		return
	}

	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return
	}

	setTimeField(val, config.UpdatedField, time.Now().UTC())
}

// ==================== 时间戳检查 ====================

// HasTimestamps 检查类型是否包含时间戳字段
//
// 示例：
//
//	var user User
//	has := mgo.HasTimestamps(user, "created_at", "updated_at")
func HasTimestamps(doc interface{}, createdField, updatedField string) bool {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return false
	}

	hasCreated := false
	hasUpdated := false

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		bsonName := GetBSONFieldName(field)

		if bsonName == createdField {
			hasCreated = true
		}
		if bsonName == updatedField {
			hasUpdated = true
		}

		if hasCreated && hasUpdated {
			return true
		}
	}

	return false
}

// GetTimestampFields 获取时间戳字段名
//
// 返回 (createdField, updatedField)
//
// 示例：
//
//	created, updated := mgo.GetTimestampFields(User{})
func GetTimestampFields(doc interface{}) (string, string) {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", ""
	}

	var createdField, updatedField string

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		bsonName := GetBSONFieldName(field)
		fieldType := field.Type

		// 检查是否为时间类型
		isTimeField := false
		if fieldType == reflect.TypeOf(time.Time{}) ||
			fieldType == reflect.TypeOf((*time.Time)(nil)) {
			isTimeField = true
		}

		if !isTimeField {
			continue
		}

		// 根据字段名推断用途
		name := field.Name
		if name == "CreatedAt" || bsonName == "created_at" {
			createdField = bsonName
		} else if name == "UpdatedAt" || bsonName == "updated_at" {
			updatedField = bsonName
		}
	}

	return createdField, updatedField
}
