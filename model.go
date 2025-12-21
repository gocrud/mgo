package mgo

// ==================== 泛型模型工厂 ====================

// Namer 接口用于获取集合名称
//
// 实现此接口的类型可以用于 Model 泛型函数
//
// 示例：
//
//	type User struct {
//	    ID   mgo.ObjectID `bson:"_id,omitempty"`
//	    Name string       `bson:"name"`
//	}
//
//	func (User) Name() string {
//	    return "users"
//	}
type Namer interface {
	// CollName 返回集合名称
	CollName() string
}

// Model 创建泛型集合
//
// T 必须实现 Namer 接口
// collectionName 参数可以强制覆盖接口返回的集合名称
//
// 示例：
//
//	type User struct {
//	    ID   mgo.ObjectID `bson:"_id,omitempty"`
//	    Name string       `bson:"name"`
//	}
//
//	func (User) Name() string {
//	    return "users"
//	}
//
//	// 使用接口方法返回的集合名 "users"
//	users := mgo.Model[User](db)
//
//	// 强制覆盖为 "app_users"
//	users := mgo.Model[User](db, "app_users")
func Model[T Namer](source interface{}, collectionName ...string) *TypedCollection[T] {
	var name string

	// 确定集合名
	if len(collectionName) > 0 && collectionName[0] != "" {
		// collectionName 参数强制覆盖接口方法
		name = collectionName[0]
	} else {
		// 从 Namer 接口获取集合名
		var zero T
		name = zero.CollName()
	}

	// 根据 source 类型创建 TypedCollection
	switch src := source.(type) {
	case *Database:
		return newTypedCollection[T](src, src.db.Collection(name))
	case *Session:
		db := src.Database()
		return newTypedCollection[T](db, db.db.Collection(name))
	case *Collection:
		return newTypedCollection[T](src.db, src.coll)
	default:
		panic("mgo: invalid source type for Model")
	}
}
