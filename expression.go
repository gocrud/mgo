package mgo

import "go.mongodb.org/mongo-driver/v2/bson"

// ExprBuilder 表达式构建器，用于构建 MongoDB 聚合表达式
//
// 示例：
//
//	// 算术运算
//	expr := Exp.Add(F("price"), F("tax"))
//
//	// 条件表达式
//	expr := Exp.Cond(
//	    Exp.Gt(F("age"), 18),
//	    "adult",
//	    "minor",
//	)
type ExprBuilder struct{}

// Exp 全局表达式构建器实例
var Exp = &ExprBuilder{}

// ===== 算术表达式操作符 =====

// Add 加法 ($add)
//
// 示例：
//
//	// 计算总价: price + tax
//	expr := Exp.Add(F("price"), F("tax"))
//
//	// 多个值相加
//	expr := Exp.Add(F("item1"), F("item2"), F("item3"))
func (eb *ExprBuilder) Add(values ...any) Expr {
	return &expr{makeD("$add", unwrapExprs(values))}
}

// Sub 减法 ($subtract)
//
// 示例：
//
//	// 计算折扣后价格: price - discount
//	expr := Exp.Sub(F("price"), F("discount"))
func (eb *ExprBuilder) Sub(a, b any) Expr {
	return &expr{makeD("$subtract", []any{unwrap(a), unwrap(b)})}
}

// Mul 乘法 ($multiply)
//
// 示例：
//
//	// 计算总价: price * quantity
//	expr := Exp.Mul(F("price"), F("quantity"))
//
//	// 计算百分比: (value / total) * 100
//	expr := Exp.Mul(Exp.Div(F("value"), F("total")), 100)
func (eb *ExprBuilder) Mul(values ...any) Expr {
	return &expr{makeD("$multiply", unwrapExprs(values))}
}

// Div 除法 ($divide)
//
// 示例：
//
//	// 计算平均值: total / count
//	expr := Exp.Div(F("total"), F("count"))
func (eb *ExprBuilder) Div(dividend, divisor any) Expr {
	return &expr{makeD("$divide", []any{unwrap(dividend), unwrap(divisor)})}
}

// Mod 取模 ($mod)
//
// 示例：
//
//	// 判断奇偶: value % 2
//	expr := Exp.Mod(F("value"), 2)
func (eb *ExprBuilder) Mod(dividend, divisor any) Expr {
	return &expr{makeD("$mod", []any{unwrap(dividend), unwrap(divisor)})}
}

// Abs 绝对值 ($abs)
//
// 示例：
//
//	// 计算差值的绝对值
//	expr := Exp.Abs(Exp.Sub(F("value1"), F("value2")))
func (eb *ExprBuilder) Abs(value any) Expr {
	return &expr{makeD("$abs", unwrap(value))}
}

// Ceil 向上取整 ($ceil)
//
// 示例：
//
//	expr := Exp.Ceil(F("price")) // 9.1 => 10
func (eb *ExprBuilder) Ceil(value any) Expr {
	return &expr{makeD("$ceil", unwrap(value))}
}

// Floor 向下取整 ($floor)
//
// 示例：
//
//	expr := Exp.Floor(F("price")) // 9.9 => 9
func (eb *ExprBuilder) Floor(value any) Expr {
	return &expr{makeD("$floor", unwrap(value))}
}

// Round 四舍五入 ($round)
//
// 示例：
//
//	// 保留2位小数
//	expr := Exp.Round(F("price"), 2)
func (eb *ExprBuilder) Round(number any, place int) Expr {
	return &expr{makeD("$round", []any{unwrap(number), place})}
}

// Sqrt 平方根 ($sqrt)
//
// 示例：
//
//	expr := Exp.Sqrt(F("area"))
func (eb *ExprBuilder) Sqrt(value any) Expr {
	return &expr{makeD("$sqrt", unwrap(value))}
}

// Pow 幂运算 ($pow)
//
// 示例：
//
//	// 计算平方: value ^ 2
//	expr := Exp.Pow(F("value"), 2)
func (eb *ExprBuilder) Pow(number, exponent any) Expr {
	return &expr{makeD("$pow", []any{unwrap(number), unwrap(exponent)})}
}

// ExpFunc 自然指数 ($exp)
//
// 示例：
//
//	expr := Exp.ExpFunc(F("rate"))
func (eb *ExprBuilder) ExpFunc(value any) Expr {
	return &expr{makeD("$exp", unwrap(value))}
}

// Ln 自然对数 ($ln)
//
// 示例：
//
//	expr := Exp.Ln(F("value"))
func (eb *ExprBuilder) Ln(value any) Expr {
	return &expr{makeD("$ln", unwrap(value))}
}

// Log 对数 ($log)
//
// 示例：
//
//	// log10(value)
//	expr := Exp.Log(F("value"), 10)
func (eb *ExprBuilder) Log(number, base any) Expr {
	return &expr{makeD("$log", []any{unwrap(number), unwrap(base)})}
}

// Log10 以10为底的对数 ($log10)
//
// 示例：
//
//	expr := Exp.Log10(F("value"))
func (eb *ExprBuilder) Log10(value any) Expr {
	return &expr{makeD("$log10", unwrap(value))}
}

// Trunc 截断小数 ($trunc)
//
// 示例：
//
//	// 保留2位小数（截断）
//	expr := Exp.Trunc(F("price"), 2)
func (eb *ExprBuilder) Trunc(number any, place int) Expr {
	return &expr{makeD("$trunc", []any{unwrap(number), place})}
}

// ===== 比较表达式操作符 =====

// Eq 等于 ($eq)
//
// 示例：
//
//	// status == "active"
//	expr := Exp.Eq(F("status"), "active")
func (eb *ExprBuilder) Eq(a, b any) Expr {
	return &expr{makeD("$eq", []any{unwrap(a), unwrap(b)})}
}

// Ne 不等于 ($ne)
//
// 示例：
//
//	// status != "deleted"
//	expr := Exp.Ne(F("status"), "deleted")
func (eb *ExprBuilder) Ne(a, b any) Expr {
	return &expr{makeD("$ne", []any{unwrap(a), unwrap(b)})}
}

// Gt 大于 ($gt)
//
// 示例：
//
//	// age > 18
//	expr := Exp.Gt(F("age"), 18)
func (eb *ExprBuilder) Gt(a, b any) Expr {
	return &expr{makeD("$gt", []any{unwrap(a), unwrap(b)})}
}

// Gte 大于等于 ($gte)
//
// 示例：
//
//	// age >= 18
//	expr := Exp.Gte(F("age"), 18)
func (eb *ExprBuilder) Gte(a, b any) Expr {
	return &expr{makeD("$gte", []any{unwrap(a), unwrap(b)})}
}

// Lt 小于 ($lt)
//
// 示例：
//
//	// age < 65
//	expr := Exp.Lt(F("age"), 65)
func (eb *ExprBuilder) Lt(a, b any) Expr {
	return &expr{makeD("$lt", []any{unwrap(a), unwrap(b)})}
}

// Lte 小于等于 ($lte)
//
// 示例：
//
//	// age <= 65
//	expr := Exp.Lte(F("age"), 65)
func (eb *ExprBuilder) Lte(a, b any) Expr {
	return &expr{makeD("$lte", []any{unwrap(a), unwrap(b)})}
}

// Cmp 比较 ($cmp)
//
// 示例：
//
//	// 返回 -1, 0, 或 1
//	expr := Exp.Cmp(F("value1"), F("value2"))
func (eb *ExprBuilder) Cmp(a, b any) Expr {
	return &expr{makeD("$cmp", []any{unwrap(a), unwrap(b)})}
}

// ===== 逻辑表达式操作符 =====

// And 逻辑与 ($and)
//
// 示例：
//
//	// age >= 18 AND status == "active"
//	expr := Exp.And(
//	    Exp.Gte(F("age"), 18),
//	    Exp.Eq(F("status"), "active"),
//	)
func (eb *ExprBuilder) And(conditions ...any) Expr {
	return &expr{makeD("$and", unwrapExprs(conditions))}
}

// Or 逻辑或 ($or)
//
// 示例：
//
//	// vip == true OR level >= 5
//	expr := Exp.Or(
//	    Exp.Eq(F("vip"), true),
//	    Exp.Gte(F("level"), 5),
//	)
func (eb *ExprBuilder) Or(conditions ...any) Expr {
	return &expr{makeD("$or", unwrapExprs(conditions))}
}

// Not 逻辑非 ($not)
//
// 示例：
//
//	// !(status == "deleted")
//	expr := Exp.Not(Exp.Eq(F("status"), "deleted"))
func (eb *ExprBuilder) Not(condition any) Expr {
	return &expr{makeD("$not", unwrap(condition))}
}

// ===== 条件表达式 =====

// Cond 三元条件 ($cond)
//
// 示例：
//
//	// age >= 18 ? "adult" : "minor"
//	expr := Exp.Cond(
//	    Exp.Gte(F("age"), 18),
//	    "adult",
//	    "minor",
//	)
func (eb *ExprBuilder) Cond(condition, trueValue, falseValue any) Expr {
	return &expr{makeD("$cond", bson.D{
		{Key: "if", Value: unwrap(condition)},
		{Key: "then", Value: unwrap(trueValue)},
		{Key: "else", Value: unwrap(falseValue)},
	})}
}

// IfNull 空值处理 ($ifNull)
//
// 示例：
//
//	// 如果 email 为 null，返回 "no-email"
//	expr := Exp.IfNull(F("email"), "no-email")
func (eb *ExprBuilder) IfNull(checkExpr, replacementValue any) Expr {
	return &expr{makeD("$ifNull", []any{unwrap(checkExpr), unwrap(replacementValue)})}
}

// SwitchCase Switch 分支
type SwitchCase struct {
	Case Expr
	Then any
}

// Switch 多条件分支 ($switch)
//
// 示例：
//
//	expr := Exp.Switch(
//	    []mgo.SwitchCase{
//	        {Exp.Lt(F("score"), 60), "F"},
//	        {Exp.Lt(F("score"), 70), "D"},
//	        {Exp.Lt(F("score"), 80), "C"},
//	        {Exp.Lt(F("score"), 90), "B"},
//	    },
//	    "A",  // default
//	)
func (eb *ExprBuilder) Switch(cases []SwitchCase, defaultValue any) Expr {
	branches := make([]bson.D, len(cases))
	for i, c := range cases {
		branches[i] = bson.D{
			{Key: "case", Value: c.Case.Build()},
			{Key: "then", Value: unwrap(c.Then)},
		}
	}
	return &expr{makeD("$switch", bson.D{
		{Key: "branches", Value: branches},
		{Key: "default", Value: unwrap(defaultValue)},
	})}
}

// ===== 字符串表达式操作符 =====

// Concat 字符串连接 ($concat)
//
// 示例：
//
//	// first_name + " " + last_name
//	expr := Exp.Concat(F("first_name"), " ", F("last_name"))
func (eb *ExprBuilder) Concat(values ...any) Expr {
	return &expr{makeD("$concat", unwrapExprs(values))}
}

// Substr 子字符串 ($substr)
//
// 示例：
//
//	// 从位置0开始，截取5个字符
//	expr := Exp.Substr(F("text"), 0, 5)
func (eb *ExprBuilder) Substr(str any, start, length int) Expr {
	return &expr{makeD("$substr", []any{unwrap(str), start, length})}
}

// SubstrBytes 字节子字符串 ($substrBytes)
//
// 示例：
//
//	expr := Exp.SubstrBytes(F("text"), 0, 10)
func (eb *ExprBuilder) SubstrBytes(str any, start, length int) Expr {
	return &expr{makeD("$substrBytes", []any{unwrap(str), start, length})}
}

// SubstrCP 码点子字符串 ($substrCP)
//
// 示例：
//
//	expr := Exp.SubstrCP(F("text"), 0, 5)
func (eb *ExprBuilder) SubstrCP(str any, start, length int) Expr {
	return &expr{makeD("$substrCP", []any{unwrap(str), start, length})}
}

// ToLower 转小写 ($toLower)
//
// 示例：
//
//	expr := Exp.ToLower(F("email"))
func (eb *ExprBuilder) ToLower(str any) Expr {
	return &expr{makeD("$toLower", unwrap(str))}
}

// ToUpper 转大写 ($toUpper)
//
// 示例：
//
//	expr := Exp.ToUpper(F("name"))
func (eb *ExprBuilder) ToUpper(str any) Expr {
	return &expr{makeD("$toUpper", unwrap(str))}
}

// StrLenBytes 字节长度 ($strLenBytes)
//
// 示例：
//
//	expr := Exp.StrLenBytes(F("text"))
func (eb *ExprBuilder) StrLenBytes(str any) Expr {
	return &expr{makeD("$strLenBytes", unwrap(str))}
}

// StrLenCP 码点长度 ($strLenCP)
//
// 示例：
//
//	expr := Exp.StrLenCP(F("text"))
func (eb *ExprBuilder) StrLenCP(str any) Expr {
	return &expr{makeD("$strLenCP", unwrap(str))}
}

// Strcasecmp 不区分大小写比较 ($strcasecmp)
//
// 示例：
//
//	expr := Exp.Strcasecmp(F("name1"), F("name2"))
func (eb *ExprBuilder) Strcasecmp(str1, str2 any) Expr {
	return &expr{makeD("$strcasecmp", []any{unwrap(str1), unwrap(str2)})}
}

// Split 分割字符串 ($split)
//
// 示例：
//
//	// 按逗号分割
//	expr := Exp.Split(F("tags"), ",")
func (eb *ExprBuilder) Split(str, delimiter any) Expr {
	return &expr{makeD("$split", []any{unwrap(str), unwrap(delimiter)})}
}

// Trim 去除首尾字符 ($trim)
//
// 示例：
//
//	expr := Exp.Trim(F("text"), " ")
func (eb *ExprBuilder) Trim(input, chars any) Expr {
	return &expr{makeD("$trim", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "chars", Value: unwrap(chars)},
	})}
}

// Ltrim 去除左边字符 ($ltrim)
//
// 示例：
//
//	expr := Exp.Ltrim(F("text"), " ")
func (eb *ExprBuilder) Ltrim(input, chars any) Expr {
	return &expr{makeD("$ltrim", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "chars", Value: unwrap(chars)},
	})}
}

// Rtrim 去除右边字符 ($rtrim)
//
// 示例：
//
//	expr := Exp.Rtrim(F("text"), " ")
func (eb *ExprBuilder) Rtrim(input, chars any) Expr {
	return &expr{makeD("$rtrim", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "chars", Value: unwrap(chars)},
	})}
}

// ReplaceOne 替换第一个匹配 ($replaceOne)
//
// 示例：
//
//	expr := Exp.ReplaceOne(F("text"), "old", "new")
func (eb *ExprBuilder) ReplaceOne(input, find, replacement any) Expr {
	return &expr{makeD("$replaceOne", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "find", Value: unwrap(find)},
		{Key: "replacement", Value: unwrap(replacement)},
	})}
}

// ReplaceAll 替换所有匹配 ($replaceAll)
//
// 示例：
//
//	expr := Exp.ReplaceAll(F("text"), "old", "new")
func (eb *ExprBuilder) ReplaceAll(input, find, replacement any) Expr {
	return &expr{makeD("$replaceAll", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "find", Value: unwrap(find)},
		{Key: "replacement", Value: unwrap(replacement)},
	})}
}

// RegexFind 正则查找 ($regexFind)
//
// 示例：
//
//	expr := Exp.RegexFind(F("text"), "[0-9]+", "i")
func (eb *ExprBuilder) RegexFind(input, regex, options any) Expr {
	return &expr{makeD("$regexFind", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "regex", Value: unwrap(regex)},
		{Key: "options", Value: unwrap(options)},
	})}
}

// RegexFindAll 正则查找所有 ($regexFindAll)
//
// 示例：
//
//	expr := Exp.RegexFindAll(F("text"), "[0-9]+", "i")
func (eb *ExprBuilder) RegexFindAll(input, regex, options any) Expr {
	return &expr{makeD("$regexFindAll", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "regex", Value: unwrap(regex)},
		{Key: "options", Value: unwrap(options)},
	})}
}

// RegexMatch 正则匹配 ($regexMatch)
//
// 示例：
//
//	expr := Exp.RegexMatch(F("email"), "^[a-z]+@[a-z]+\\.[a-z]+$", "i")
func (eb *ExprBuilder) RegexMatch(input, regex, options any) Expr {
	return &expr{makeD("$regexMatch", bson.D{
		{Key: "input", Value: unwrap(input)},
		{Key: "regex", Value: unwrap(regex)},
		{Key: "options", Value: unwrap(options)},
	})}
}

// ToString 转字符串 ($toString)
//
// 示例：
//
//	expr := Exp.ToString(F("user_id"))
func (eb *ExprBuilder) ToString(value any) Expr {
	return &expr{makeD("$toString", unwrap(value))}
}
