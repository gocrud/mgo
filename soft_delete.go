package mgo

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SoftDeleteConfig 软删除配置
type SoftDeleteConfig struct {
	// 软删除字段名，默认 "deleted_at"
	Field string
	// 是否启用软删除
	Enabled bool
}

// defaultSoftDeleteConfig 默认软删除配置
func defaultSoftDeleteConfig() *SoftDeleteConfig {
	return &SoftDeleteConfig{
		Field:   "deleted_at",
		Enabled: false,
	}
}

// CollectionOption Collection 配置选项
type CollectionOption func(*Collection)

// WithSoftDelete 启用软删除
//
// 示例：
//
//	// 使用默认字段名 "deleted_at"
//	coll := mgo.NewCollection(mongoCollection, mgo.WithSoftDelete())
//
//	// 自定义字段名
//	coll := mgo.NewCollection(mongoCollection, mgo.WithSoftDelete("removed_at"))
func WithSoftDelete(field ...string) CollectionOption {
	return func(c *Collection) {
		c.softDelete.Enabled = true
		if len(field) > 0 && field[0] != "" {
			c.softDelete.Field = field[0]
		}
	}
}

// buildSoftDeleteFilter 构建软删除过滤条件
// includeDeleted: false - 排除已删除, true - 包含已删除, nil - 仅已删除
func (c *Collection) buildSoftDeleteFilter(includeDeleted *bool) bson.D {
	if !c.softDelete.Enabled {
		return bson.D{}
	}

	// 包含已删除：不添加过滤条件
	if includeDeleted != nil && *includeDeleted {
		return bson.D{}
	}

	// 仅已删除：字段存在
	if includeDeleted == nil {
		return bson.D{{Key: c.softDelete.Field, Value: bson.M{"$exists": true}}}
	}

	// 默认：排除已删除（字段不存在或为 nil）
	return bson.D{{Key: c.softDelete.Field, Value: bson.M{"$exists": false}}}
}

// buildSoftDeleteUpdate 构建软删除更新操作
func (c *Collection) buildSoftDeleteUpdate() bson.D {
	return bson.D{{Key: "$set", Value: bson.D{{Key: c.softDelete.Field, Value: time.Now()}}}}
}

// buildRestoreUpdate 构建恢复更新操作
func (c *Collection) buildRestoreUpdate() bson.D {
	return bson.D{{Key: "$unset", Value: bson.D{{Key: c.softDelete.Field, Value: ""}}}}
}
