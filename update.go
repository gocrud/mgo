package mgo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// UpdateBuilder 更新构建器
//
// 用于构建 MongoDB 更新操作，完全消除 $set、$inc 等操作符字符串。
//
// 使用示例：
//
//	// 基础更新
//	update := mgo.Update().
//	    Set("name", "张三").
//	    Set("age", 25).
//	    Inc("visits", 1)
//
//	// 数组操作
//	update := mgo.Update().
//	    Push("tags", "new-tag").
//	    Pull("old_tags", "obsolete").
//	    AddToSet("skills", "golang")
//
//	// 组合更新
//	update := mgo.Update().
//	    Set("status", "active").
//	    Inc("level", 1).
//	    CurrentDate("updated_at").
//	    SetOnInsert("created_at", time.Now())
type UpdateBuilder struct {
	updates map[string]bson.D
}

// Update 创建新的更新构建器
//
// 示例：
//
//	update := mgo.Update()
func Update() *UpdateBuilder {
	return &UpdateBuilder{
		updates: make(map[string]bson.D),
	}
}

// Set 设置字段值 ($set)
//
// 示例：
//
//	update.Set("name", "张三")
//	update.Set("age", 25)
//	update.Set("user.email", "zhangsan@example.com")
func (u *UpdateBuilder) Set(field string, value any) *UpdateBuilder {
	u.addOperator("$set", field, value)
	return u
}

// SetM 批量设置多个字段 ($set)
//
// 示例：
//
//	update.SetM(M{
//	    "name": "张三",
//	    "age": 25,
//	    "status": "active",
//	})
func (u *UpdateBuilder) SetM(fields M) *UpdateBuilder {
	for field, value := range fields {
		u.addOperator("$set", field, value)
	}
	return u
}

// Unset 删除字段 ($unset)
//
// 示例：
//
//	update.Unset("temp_field")
//	update.Unset("user.old_field")
func (u *UpdateBuilder) Unset(fields ...string) *UpdateBuilder {
	for _, field := range fields {
		u.addOperator("$unset", field, "")
	}
	return u
}

// Inc 递增字段值 ($inc)
//
// 示例：
//
//	update.Inc("visits", 1)
//	update.Inc("level", 2)
//	update.Inc("score", -10)  // 负数表示递减
func (u *UpdateBuilder) Inc(field string, value any) *UpdateBuilder {
	u.addOperator("$inc", field, value)
	return u
}

// Mul 乘以字段值 ($mul)
//
// 示例：
//
//	update.Mul("price", 1.1)    // 价格提高10%
//	update.Mul("discount", 0.8) // 折扣降低20%
func (u *UpdateBuilder) Mul(field string, value any) *UpdateBuilder {
	u.addOperator("$mul", field, value)
	return u
}

// Min 设置字段最小值 ($min)
//
// 只有当新值小于当前值时才更新
//
// 示例：
//
//	update.Min("price", 100)      // 价格不低于100
//	update.Min("min_score", 60)   // 最低分不低于60
func (u *UpdateBuilder) Min(field string, value any) *UpdateBuilder {
	u.addOperator("$min", field, value)
	return u
}

// Max 设置字段最大值 ($max)
//
// 只有当新值大于当前值时才更新
//
// 示例：
//
//	update.Max("stock", 1000)     // 库存不超过1000
//	update.Max("max_score", 100)  // 最高分不超过100
func (u *UpdateBuilder) Max(field string, value any) *UpdateBuilder {
	u.addOperator("$max", field, value)
	return u
}

// Rename 重命名字段 ($rename)
//
// 示例：
//
//	update.Rename("old_name", "new_name")
//	update.Rename("user.old_field", "user.new_field")
func (u *UpdateBuilder) Rename(oldField, newField string) *UpdateBuilder {
	u.addOperator("$rename", oldField, newField)
	return u
}

// CurrentDate 设置当前日期 ($currentDate)
//
// timestamp 参数为 true 时使用 timestamp 类型，否则使用 date 类型
//
// 示例：
//
//	update.CurrentDate("updated_at", false)  // date 类型
//	update.CurrentDate("last_seen", true)    // timestamp 类型
func (u *UpdateBuilder) CurrentDate(field string, timestamp bool) *UpdateBuilder {
	var value any = true
	if timestamp {
		value = bson.D{{Key: "$type", Value: "timestamp"}}
	}
	u.addOperator("$currentDate", field, value)
	return u
}

// SetOnInsert 插入时设置字段值 ($setOnInsert)
//
// 只在 upsert 操作且文档不存在时设置字段值
//
// 示例：
//
//	update.SetOnInsert("created_at", time.Now())
//	update.SetOnInsert("version", 1)
func (u *UpdateBuilder) SetOnInsert(field string, value any) *UpdateBuilder {
	u.addOperator("$setOnInsert", field, value)
	return u
}

// Push 向数组末尾添加元素 ($push)
//
// 示例：
//
//	update.Push("tags", "new-tag")
//	update.Push("items", M{"id": 1, "name": "item1"})
func (u *UpdateBuilder) Push(field string, value any) *UpdateBuilder {
	u.addOperator("$push", field, value)
	return u
}

// PushEach 向数组末尾添加多个元素 ($push + $each)
//
// 示例：
//
//	update.PushEach("tags", []string{"tag1", "tag2", "tag3"})
//	update.PushEach("items", []M{
//	    {"id": 1, "name": "item1"},
//	    {"id": 2, "name": "item2"},
//	})
func (u *UpdateBuilder) PushEach(field string, values ...any) *UpdateBuilder {
	u.addOperator("$push", field, bson.D{{Key: "$each", Value: values}})
	return u
}

// PushSlice 向数组末尾添加多个元素（限制、排序） ($push + $each + $slice + $sort)
//
// slice: 限制数组大小（正数从前保留，负数从后保留，0 清空数组）
// sort: 排序方式（1 升序，-1 降序，也可以是字段名）
//
// 示例：
//
//	// 添加并只保留最近10条记录
//	update.PushSlice("history", -10, -1, item1, item2, item3)
//
//	// 添加并按分数排序，只保留前5名
//	update.PushSlice("top_scores", 5, -1, score1, score2, score3)
func (u *UpdateBuilder) PushSlice(field string, slice int, sort any, values ...any) *UpdateBuilder {
	pushDoc := bson.D{{Key: "$each", Value: values}}
	if slice != 0 {
		pushDoc = append(pushDoc, bson.E{Key: "$slice", Value: slice})
	}
	if sort != nil {
		pushDoc = append(pushDoc, bson.E{Key: "$sort", Value: sort})
	}
	u.addOperator("$push", field, pushDoc)
	return u
}

// PushPosition 在指定位置插入元素 ($push + $each + $position)
//
// position: 插入位置（0 表示开头，-1 表示末尾前）
//
// 示例：
//
//	// 在数组开头插入
//	update.PushPosition("items", 0, "item1", "item2")
//
//	// 在倒数第二个位置插入
//	update.PushPosition("items", -1, "new-item")
func (u *UpdateBuilder) PushPosition(field string, position int, values ...any) *UpdateBuilder {
	u.addOperator("$push", field, bson.D{
		{Key: "$each", Value: values},
		{Key: "$position", Value: position},
	})
	return u
}

// Pull 从数组中删除匹配的元素 ($pull)
//
// 示例：
//
//	update.Pull("tags", "old-tag")
//	update.Pull("items", M{"status": "deleted"})
func (u *UpdateBuilder) Pull(field string, value any) *UpdateBuilder {
	u.addOperator("$pull", field, value)
	return u
}

// PullAll 从数组中删除多个值 ($pullAll)
//
// 示例：
//
//	update.PullAll("tags", []string{"tag1", "tag2", "tag3"})
//	update.PullAll("ids", []int{1, 2, 3, 4, 5})
func (u *UpdateBuilder) PullAll(field string, values ...any) *UpdateBuilder {
	u.addOperator("$pullAll", field, values)
	return u
}

// PullFilter 根据条件从数组中删除元素 ($pull + 条件)
//
// 示例：
//
//	// 删除价格大于1000的商品
//	update.PullFilter("items", Filter().Gt("price", 1000))
//
//	// 删除已过期的项
//	update.PullFilter("list", Filter().Lt("expire_at", time.Now()))
func (u *UpdateBuilder) PullFilter(field string, filter *FilterBuilder) *UpdateBuilder {
	u.addOperator("$pull", field, filter.BuildM())
	return u
}

// Pop 删除数组第一个或最后一个元素 ($pop)
//
// position: 1 删除最后一个元素，-1 删除第一个元素
//
// 示例：
//
//	update.Pop("items", 1)   // 删除最后一个
//	update.Pop("items", -1)  // 删除第一个
func (u *UpdateBuilder) Pop(field string, position int) *UpdateBuilder {
	u.addOperator("$pop", field, position)
	return u
}

// AddToSet 向数组添加元素（去重） ($addToSet)
//
// 如果元素已存在则不添加
//
// 示例：
//
//	update.AddToSet("tags", "unique-tag")
//	update.AddToSet("user_ids", 12345)
func (u *UpdateBuilder) AddToSet(field string, value any) *UpdateBuilder {
	u.addOperator("$addToSet", field, value)
	return u
}

// AddToSetEach 向数组添加多个元素（去重） ($addToSet + $each)
//
// 示例：
//
//	update.AddToSetEach("tags", "tag1", "tag2", "tag3")
//	update.AddToSetEach("skills", "golang", "python", "rust")
func (u *UpdateBuilder) AddToSetEach(field string, values ...any) *UpdateBuilder {
	u.addOperator("$addToSet", field, bson.D{{Key: "$each", Value: values}})
	return u
}

// Bit 位运算更新 ($bit)
//
// operation: "and", "or", "xor"
//
// 示例：
//
//	update.Bit("flags", "or", 4)   // flags |= 4
//	update.Bit("flags", "and", 3)  // flags &= 3
//	update.Bit("flags", "xor", 5)  // flags ^= 5
func (u *UpdateBuilder) Bit(field string, operation string, value int) *UpdateBuilder {
	u.addOperator("$bit", field, bson.D{{Key: operation, Value: value}})
	return u
}

// Build 构建为 bson.D
//
// 示例：
//
//	update := Update().Set("name", "张三").Inc("age", 1)
//	bsonD := update.Build()
func (u *UpdateBuilder) Build() bson.D {
	result := make(bson.D, 0, len(u.updates))
	for op, fields := range u.updates {
		result = append(result, bson.E{Key: op, Value: fields})
	}
	return result
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	update := Update().Set("name", "张三").Inc("age", 1)
//	bsonM := update.BuildM()
func (u *UpdateBuilder) BuildM() bson.M {
	result := make(bson.M, len(u.updates))
	for op, fields := range u.updates {
		opMap := make(bson.M, len(fields))
		for _, field := range fields {
			opMap[field.Key] = field.Value
		}
		result[op] = opMap
	}
	return result
}

// Clone 克隆更新构建器
//
// 示例：
//
//	update1 := Update().Set("name", "张三")
//	update2 := update1.Clone().Inc("age", 1)
func (u *UpdateBuilder) Clone() *UpdateBuilder {
	newUpdate := Update()
	for op, fields := range u.updates {
		newUpdate.updates[op] = make(bson.D, len(fields))
		copy(newUpdate.updates[op], fields)
	}
	return newUpdate
}

// addOperator 添加操作符
func (u *UpdateBuilder) addOperator(operator, field string, value any) {
	if _, exists := u.updates[operator]; !exists {
		u.updates[operator] = make(bson.D, 0)
	}
	u.updates[operator] = append(u.updates[operator], bson.E{Key: field, Value: value})
}
