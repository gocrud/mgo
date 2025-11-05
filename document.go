package mgo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Doc 文档构建器
//
// 用于优雅地构建 MongoDB 文档，避免直接使用 bson.D 或 bson.M
//
// 使用示例：
//
//	// 创建文档
//	doc := mgo.NewDoc().
//	    Set("name", "张三").
//	    Set("age", 25).
//	    Set("email", "zhangsan@example.com")
//
//	// 嵌套文档
//	doc := mgo.NewDoc().
//	    Set("user", mgo.NewDoc().
//	        Set("name", "张三").
//	        Set("email", "zhangsan@example.com")).
//	    Set("tags", []string{"tag1", "tag2"})
type Doc struct {
	fields bson.D
}

// NewDoc 创建新的文档构建器
//
// 示例：
//
//	doc := mgo.NewDoc()
func NewDoc() *Doc {
	return &Doc{
		fields: make(bson.D, 0),
	}
}

// Set 设置字段值
//
// 示例：
//
//	doc.Set("name", "张三")
//	doc.Set("age", 25)
//	doc.Set("user.email", "zhangsan@example.com")
func (d *Doc) Set(field string, value any) *Doc {
	d.fields = append(d.fields, bson.E{Key: field, Value: value})
	return d
}

// SetIf 条件设置字段值
//
// 只有当 condition 为 true 时才设置字段
//
// 示例：
//
//	doc.SetIf(name != "", "name", name).
//	    SetIf(age > 0, "age", age)
func (d *Doc) SetIf(condition bool, field string, value any) *Doc {
	if condition {
		d.fields = append(d.fields, bson.E{Key: field, Value: value})
	}
	return d
}

// Build 构建为 bson.D
//
// 示例：
//
//	doc := NewDoc().Set("name", "张三").Set("age", 25)
//	bsonD := doc.Build()
func (d *Doc) Build() bson.D {
	return d.fields
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	doc := NewDoc().Set("name", "张三").Set("age", 25)
//	bsonM := doc.BuildM()
func (d *Doc) BuildM() bson.M {
	result := make(bson.M, len(d.fields))
	for _, field := range d.fields {
		result[field.Key] = field.Value
	}
	return result
}

// Projection 投影构建器
//
// 用于构建查询投影，指定返回哪些字段
//
// 使用示例：
//
//	// 包含指定字段
//	proj := mgo.NewProjection().
//	    Include("name", "age", "email")
//
//	// 排除指定字段
//	proj := mgo.NewProjection().
//	    Exclude("password", "secret")
//
//	// 混合使用（注意：除了 _id，不能同时包含和排除）
//	proj := mgo.NewProjection().
//	    Include("name", "age").
//	    ExcludeID()
//
//	// 数组切片
//	proj := mgo.NewProjection().
//	    Include("name").
//	    Slice("items", 10)  // 只返回前10个元素
//
//	// 使用表达式
//	proj := mgo.NewProjection().
//	    Include("name").
//	    SetExpr("fullName", mgo.Exp.Concat(mgo.F("firstName"), " ", mgo.F("lastName")))
type Projection struct {
	fields bson.D
}

// NewProjection 创建新的投影构建器
//
// 示例：
//
//	proj := mgo.NewProjection()
func NewProjection() *Projection {
	return &Projection{
		fields: make(bson.D, 0),
	}
}

// Include 包含指定字段
//
// 示例：
//
//	proj.Include("name", "age", "email")
//	proj.Include("user.name", "user.email")
func (p *Projection) Include(fields ...string) *Projection {
	for _, field := range fields {
		p.fields = append(p.fields, bson.E{Key: field, Value: 1})
	}
	return p
}

// Exclude 排除指定字段
//
// 注意：不能同时使用 Include 和 Exclude（_id 除外）
//
// 示例：
//
//	proj.Exclude("password", "secret", "internal")
func (p *Projection) Exclude(fields ...string) *Projection {
	for _, field := range fields {
		p.fields = append(p.fields, bson.E{Key: field, Value: 0})
	}
	return p
}

// IncludeID 显式包含 _id 字段
//
// 默认情况下 _id 会自动包含，此方法用于明确表达意图
//
// 示例：
//
//	proj.Exclude("password").IncludeID()
func (p *Projection) IncludeID() *Projection {
	p.fields = append(p.fields, bson.E{Key: "_id", Value: 1})
	return p
}

// ExcludeID 排除 _id 字段
//
// 示例：
//
//	proj.Include("name", "age").ExcludeID()
func (p *Projection) ExcludeID() *Projection {
	p.fields = append(p.fields, bson.E{Key: "_id", Value: 0})
	return p
}

// Slice 数组切片投影
//
// limit: 返回的元素数量（正数从前取，负数从后取）
// skip: 跳过的元素数量（可选）
//
// 示例：
//
//	// 返回前10个元素
//	proj.Slice("items", 10)
//
//	// 跳过前5个，返回接下来的10个
//	proj.SliceWithSkip("items", 5, 10)
func (p *Projection) Slice(field string, limit int) *Projection {
	p.fields = append(p.fields, bson.E{Key: field, Value: bson.D{{Key: "$slice", Value: limit}}})
	return p
}

// SliceWithSkip 带跳过的数组切片投影
//
// 示例：
//
//	proj.SliceWithSkip("comments", 10, 5)  // 跳过前10条，返回接下来5条
func (p *Projection) SliceWithSkip(field string, skip, limit int) *Projection {
	p.fields = append(p.fields, bson.E{Key: field, Value: bson.D{{Key: "$slice", Value: bson.A{skip, limit}}}})
	return p
}

// ElemMatch 数组元素匹配投影
//
// 只返回匹配条件的第一个数组元素
//
// 示例：
//
//	// 只返回评分>4的第一条评论
//	proj.ElemMatch("comments", mgo.Filter().Gt("rating", 4))
func (p *Projection) ElemMatch(field string, filter *FilterBuilder) *Projection {
	p.fields = append(p.fields, bson.E{Key: field, Value: bson.D{{Key: "$elemMatch", Value: filter.BuildM()}}})
	return p
}

// SetExpr 使用表达式设置字段
//
// 在投影中计算新字段
//
// 示例：
//
//	// 计算全名
//	proj.SetExpr("fullName", mgo.Exp.Concat(mgo.F("firstName"), " ", mgo.F("lastName")))
//
//	// 计算折扣价
//	proj.SetExpr("finalPrice", mgo.Exp.Sub(mgo.F("price"), mgo.F("discount")))
func (p *Projection) SetExpr(field string, expr Expr) *Projection {
	p.fields = append(p.fields, bson.E{Key: field, Value: expr.Build()})
	return p
}

// Meta 文本搜索元数据投影
//
// 返回文本搜索的相关性分数
//
// 示例：
//
//	proj.Meta("score", "textScore")
func (p *Projection) Meta(field string, metaType string) *Projection {
	p.fields = append(p.fields, bson.E{Key: field, Value: bson.D{{Key: "$meta", Value: metaType}}})
	return p
}

// Build 构建为 bson.D
//
// 示例：
//
//	proj := NewProjection().Include("name", "age")
//	bsonD := proj.Build()
func (p *Projection) Build() bson.D {
	return p.fields
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	proj := NewProjection().Include("name", "age")
//	bsonM := proj.BuildM()
func (p *Projection) BuildM() bson.M {
	result := make(bson.M, len(p.fields))
	for _, field := range p.fields {
		result[field.Key] = field.Value
	}
	return result
}

// Sort 排序构建器
//
// 用于构建查询排序
//
// 使用示例：
//
//	// 升序排序
//	sort := mgo.NewSort().Asc("name", "age")
//
//	// 降序排序
//	sort := mgo.NewSort().Desc("created_at")
//
//	// 混合排序
//	sort := mgo.NewSort().
//	    Desc("priority").      // 先按优先级降序
//	    Asc("created_at")      // 再按创建时间升序
//
//	// 文本搜索分数排序
//	sort := mgo.NewSort().TextScore("score")
type Sort struct {
	fields bson.D
}

// NewSort 创建新的排序构建器
//
// 示例：
//
//	sort := mgo.NewSort()
func NewSort() *Sort {
	return &Sort{
		fields: make(bson.D, 0),
	}
}

// Asc 升序排序
//
// 示例：
//
//	sort.Asc("name", "age", "created_at")
func (s *Sort) Asc(fields ...string) *Sort {
	for _, field := range fields {
		s.fields = append(s.fields, bson.E{Key: field, Value: 1})
	}
	return s
}

// Desc 降序排序
//
// 示例：
//
//	sort.Desc("created_at", "updated_at", "priority")
func (s *Sort) Desc(fields ...string) *Sort {
	for _, field := range fields {
		s.fields = append(s.fields, bson.E{Key: field, Value: -1})
	}
	return s
}

// TextScore 按文本搜索分数排序
//
// 示例：
//
//	sort.TextScore("score")
func (s *Sort) TextScore(field string) *Sort {
	s.fields = append(s.fields, bson.E{Key: field, Value: bson.D{{Key: "$meta", Value: "textScore"}}})
	return s
}

// Build 构建为 bson.D
//
// 示例：
//
//	sort := NewSort().Desc("priority").Asc("name")
//	bsonD := sort.Build()
func (s *Sort) Build() bson.D {
	return s.fields
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	sort := NewSort().Desc("priority").Asc("name")
//	bsonM := sort.BuildM()
func (s *Sort) BuildM() bson.M {
	result := make(bson.M, len(s.fields))
	for _, field := range s.fields {
		result[field.Key] = field.Value
	}
	return result
}

// Index 索引构建器
//
// 用于构建索引定义
//
// 使用示例：
//
//	// 单字段索引
//	idx := mgo.NewIndex().Asc("name")
//
//	// 复合索引
//	idx := mgo.NewIndex().Asc("category").Desc("created_at")
//
//	// 文本索引
//	idx := mgo.NewIndex().Text("title", "content")
//
//	// 地理空间索引
//	idx := mgo.NewIndex().Geo2DSphere("location")
type Index struct {
	fields bson.D
}

// NewIndex 创建新的索引构建器
//
// 示例：
//
//	idx := mgo.NewIndex()
func NewIndex() *Index {
	return &Index{
		fields: make(bson.D, 0),
	}
}

// Asc 升序索引
//
// 示例：
//
//	idx.Asc("name", "age")
func (i *Index) Asc(fields ...string) *Index {
	for _, field := range fields {
		i.fields = append(i.fields, bson.E{Key: field, Value: 1})
	}
	return i
}

// Desc 降序索引
//
// 示例：
//
//	idx.Desc("created_at", "priority")
func (i *Index) Desc(fields ...string) *Index {
	for _, field := range fields {
		i.fields = append(i.fields, bson.E{Key: field, Value: -1})
	}
	return i
}

// Text 文本索引
//
// 示例：
//
//	idx.Text("title", "content", "description")
func (i *Index) Text(fields ...string) *Index {
	for _, field := range fields {
		i.fields = append(i.fields, bson.E{Key: field, Value: "text"})
	}
	return i
}

// Geo2D 2D 地理空间索引
//
// 示例：
//
//	idx.Geo2D("location")
func (i *Index) Geo2D(field string) *Index {
	i.fields = append(i.fields, bson.E{Key: field, Value: "2d"})
	return i
}

// Geo2DSphere 2DSphere 地理空间索引
//
// 示例：
//
//	idx.Geo2DSphere("location")
func (i *Index) Geo2DSphere(field string) *Index {
	i.fields = append(i.fields, bson.E{Key: field, Value: "2dsphere"})
	return i
}

// Hashed 哈希索引
//
// 示例：
//
//	idx.Hashed("user_id")
func (i *Index) Hashed(field string) *Index {
	i.fields = append(i.fields, bson.E{Key: field, Value: "hashed"})
	return i
}

// Build 构建为 bson.D
//
// 示例：
//
//	idx := NewIndex().Asc("category").Desc("created_at")
//	bsonD := idx.Build()
func (i *Index) Build() bson.D {
	return i.fields
}

// BuildM 构建为 bson.M
//
// 示例：
//
//	idx := NewIndex().Asc("category").Desc("created_at")
//	bsonM := idx.BuildM()
func (i *Index) BuildM() bson.M {
	result := make(bson.M, len(i.fields))
	for _, field := range i.fields {
		result[field.Key] = field.Value
	}
	return result
}
