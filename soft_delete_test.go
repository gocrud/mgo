package mgo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// TestSoftDelete 测试软删除功能
func TestSoftDelete(t *testing.T) {
	// 测试软删除配置
	t.Run("SoftDeleteConfig", func(t *testing.T) {
		config := defaultSoftDeleteConfig()
		if config.Field != "deleted_at" {
			t.Errorf("默认字段应为 deleted_at, got %s", config.Field)
		}
		if config.Enabled {
			t.Errorf("默认应该禁用软删除")
		}
	})

	// 测试 WithSoftDelete 选项
	t.Run("WithSoftDelete", func(t *testing.T) {
		// 测试默认字段
		opt := WithSoftDelete()
		coll := &Collection{softDelete: defaultSoftDeleteConfig()}
		opt(coll)

		if !coll.softDelete.Enabled {
			t.Error("软删除应该被启用")
		}
		if coll.softDelete.Field != "deleted_at" {
			t.Errorf("默认字段应为 deleted_at, got %s", coll.softDelete.Field)
		}

		// 测试自定义字段
		opt = WithSoftDelete("removed_at")
		coll = &Collection{softDelete: defaultSoftDeleteConfig()}
		opt(coll)

		if coll.softDelete.Field != "removed_at" {
			t.Errorf("自定义字段应为 removed_at, got %s", coll.softDelete.Field)
		}
	})
}

// TestBuildSoftDeleteFilter 测试构建软删除过滤器
func TestBuildSoftDeleteFilter(t *testing.T) {
	coll := &Collection{softDelete: &SoftDeleteConfig{
		Field:   "deleted_at",
		Enabled: true,
	}}

	tests := []struct {
		name           string
		includeDeleted *bool
		expected       int // 期望的过滤条件数量
	}{
		{
			name:           "默认排除已删除",
			includeDeleted: nil,
			expected:       1, // {deleted_at: {$exists: false}}
		},
		{
			name: "包含已删除",
			includeDeleted: func() *bool {
				v := true
				return &v
			}(),
			expected: 0, // 无过滤条件
		},
		{
			name: "仅已删除",
			includeDeleted: func() *bool {
				v := false
				return &v
			}(),
			expected: 1, // {deleted_at: {$exists: true}}
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := coll.buildSoftDeleteFilter(tt.includeDeleted)
			if len(filter) != tt.expected {
				t.Errorf("期望 %d 个过滤条件, got %d", tt.expected, len(filter))
			}
		})
	}

	// 测试未启用软删除
	t.Run("未启用软删除", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{Enabled: false}}
		filter := coll.buildSoftDeleteFilter(nil)
		if len(filter) != 0 {
			t.Error("未启用软删除时不应有过滤条件")
		}
	})
}

// TestBuildSoftDeleteUpdate 测试构建软删除更新
func TestBuildSoftDeleteUpdate(t *testing.T) {
	coll := &Collection{softDelete: &SoftDeleteConfig{
		Field:   "deleted_at",
		Enabled: true,
	}}

	update := coll.buildSoftDeleteUpdate()
	if len(update) != 1 {
		t.Errorf("期望 1 个更新操作, got %d", len(update))
	}

	// 检查是否包含 $set
	var found bool
	for _, elem := range update {
		if elem.Key == "$set" {
			found = true
			setDoc, ok := elem.Value.(bson.D)
			if !ok {
				t.Error("$set 值应为 bson.D 类型")
			}
			if len(setDoc) != 1 || setDoc[0].Key != "deleted_at" {
				t.Error("$set 应包含 deleted_at 字段")
			}
			_, ok = setDoc[0].Value.(time.Time)
			if !ok {
				t.Error("deleted_at 应为 time.Time 类型")
			}
		}
	}

	if !found {
		t.Error("更新操作应包含 $set")
	}
}

// TestBuildRestoreUpdate 测试构建恢复更新
func TestBuildRestoreUpdate(t *testing.T) {
	coll := &Collection{softDelete: &SoftDeleteConfig{
		Field:   "deleted_at",
		Enabled: true,
	}}

	update := coll.buildRestoreUpdate()
	if len(update) != 1 {
		t.Errorf("期望 1 个更新操作, got %d", len(update))
	}

	// 检查是否包含 $unset
	var found bool
	for _, elem := range update {
		if elem.Key == "$unset" {
			found = true
			unsetDoc, ok := elem.Value.(bson.D)
			if !ok {
				t.Error("$unset 值应为 bson.D 类型")
			}
			if len(unsetDoc) != 1 || unsetDoc[0].Key != "deleted_at" {
				t.Error("$unset 应包含 deleted_at 字段")
			}
		}
	}

	if !found {
		t.Error("恢复操作应包含 $unset")
	}
}

// TestQueryBuilderSoftDeleteIntegration 测试 QueryBuilder 集成
func TestQueryBuilderSoftDeleteIntegration(t *testing.T) {
	ctx := context.Background()

	// 创建未启用软删除的集合
	t.Run("未启用软删除", func(t *testing.T) {
		coll := &Collection{softDelete: defaultSoftDeleteConfig()}
		qb := newQueryBuilder(coll, ctx)

		// Restore 应返回错误
		_, err := qb.Restore()
		if err != ErrSoftDeleteNotEnabled {
			t.Errorf("期望错误 %v, got %v", ErrSoftDeleteNotEnabled, err)
		}
	})

	// 测试 WithDeleted 和 OnlyDeleted
	t.Run("WithDeleted和OnlyDeleted", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{
			Field:   "deleted_at",
			Enabled: true,
		}}
		qb := newQueryBuilder(coll, ctx)

		// 默认应该排除已删除
		if qb.includeDeleted != nil {
			t.Error("默认 includeDeleted 应为 nil")
		}

		// WithDeleted
		qb.WithDeleted()
		if qb.includeDeleted == nil || !*qb.includeDeleted {
			t.Error("WithDeleted 后 includeDeleted 应为 true")
		}

		// OnlyDeleted
		qb = newQueryBuilder(coll, ctx)
		qb.OnlyDeleted()
		if qb.includeDeleted != nil {
			t.Error("OnlyDeleted 后 includeDeleted 应为 nil")
		}
	})

	// 测试 WithHardDelete
	t.Run("WithHardDelete", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{
			Field:   "deleted_at",
			Enabled: true,
		}}
		qb := newQueryBuilder(coll, ctx)

		// 默认应该是软删除
		if qb.hardDelete {
			t.Error("默认 hardDelete 应为 false")
		}

		// WithHardDelete
		qb.WithHardDelete()
		if !qb.hardDelete {
			t.Error("WithHardDelete 后 hardDelete 应为 true")
		}
	})
}

// TestBuildFilterWithSoftDelete 测试过滤器合并
func TestBuildFilterWithSoftDelete(t *testing.T) {
	ctx := context.Background()

	// 未启用软删除
	t.Run("未启用软删除", func(t *testing.T) {
		coll := &Collection{softDelete: defaultSoftDeleteConfig()}
		qb := newQueryBuilder(coll, ctx).Eq("status", "active")

		filter := qb.buildFilterWithSoftDelete()
		if len(filter) != 1 {
			t.Errorf("期望 1 个过滤条件, got %d", len(filter))
		}
		if filter["status"] != "active" {
			t.Error("应包含 status 过滤条件")
		}
	})

	// 启用软删除 - 默认排除已删除
	t.Run("启用软删除-排除已删除", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{
			Field:   "deleted_at",
			Enabled: true,
		}}
		qb := newQueryBuilder(coll, ctx).Eq("status", "active")

		filter := qb.buildFilterWithSoftDelete()
		if len(filter) != 2 {
			t.Errorf("期望 2 个过滤条件, got %d", len(filter))
		}
		if filter["status"] != "active" {
			t.Error("应包含 status 过滤条件")
		}
		if filter["deleted_at"] == nil {
			t.Error("应包含 deleted_at 过滤条件")
		}
	})

	// 启用软删除 - WithDeleted
	t.Run("启用软删除-包含已删除", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{
			Field:   "deleted_at",
			Enabled: true,
		}}
		qb := newQueryBuilder(coll, ctx).Eq("status", "active").WithDeleted()

		filter := qb.buildFilterWithSoftDelete()
		if len(filter) != 1 {
			t.Errorf("期望 1 个过滤条件, got %d", len(filter))
		}
		if filter["status"] != "active" {
			t.Error("应包含 status 过滤条件")
		}
		if filter["deleted_at"] != nil {
			t.Error("不应包含 deleted_at 过滤条件")
		}
	})

	// 启用软删除 - OnlyDeleted
	t.Run("启用软删除-仅已删除", func(t *testing.T) {
		coll := &Collection{softDelete: &SoftDeleteConfig{
			Field:   "deleted_at",
			Enabled: true,
		}}
		qb := newQueryBuilder(coll, ctx).Eq("status", "active").OnlyDeleted()

		filter := qb.buildFilterWithSoftDelete()
		if len(filter) != 2 {
			t.Errorf("期望 2 个过滤条件, got %d", len(filter))
		}
		if filter["status"] != "active" {
			t.Error("应包含 status 过滤条件")
		}
		if filter["deleted_at"] == nil {
			t.Error("应包含 deleted_at 过滤条件")
		}
	})
}
