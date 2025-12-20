package mgo

// ==================== 泛型模型工厂 ====================

// Model 创建泛型集合
//
// 自动推断集合名称（从类型名）
//
// 示例：
//
//	type User struct {
//	    ID   mgo.ObjectID `bson:"_id,omitempty"`
//	    Name string       `bson:"name"`
//	}
//
//	// 自动推断集合名为 "users"
//	users := mgo.Model[User](db)
//
//	// 显式指定集合名
//	users := mgo.Model[User](db, "app_users")
func Model[T any](source interface{}, collectionName ...string) *TypedCollection[T] {
	var name string

	// 确定集合名
	if len(collectionName) > 0 && collectionName[0] != "" {
		name = collectionName[0]
	} else {
		// 自动推断集合名
		var zero T
		name = InferCollectionNameFromType(zero)
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
