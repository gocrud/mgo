package mgo

func Eq(key string, value any) E {
	return E{Key: key, Value: value}
}

func Ne(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$ne", Value: value}}}
}

func Gt(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$gt", Value: value}}}
}

func Gte(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$gte", Value: value}}}
}

func Lt(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$lt", Value: value}}}
}

func Lte(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$lte", Value: value}}}
}

func In(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$in", Value: value}}}
}

func Nin(key string, value any) E {
	return E{Key: key, Value: D{{Key: "$nin", Value: value}}}
}

// Field helper
type Field string

func F(name string) Field {
	return Field(name)
}

func (f Field) Eq(val any) E {
	return Eq(string(f), val)
}

func (f Field) Ne(val any) E {
	return Ne(string(f), val)
}

func (f Field) Gt(val any) E {
	return Gt(string(f), val)
}

func (f Field) Gte(val any) E {
	return Gte(string(f), val)
}

func (f Field) Lt(val any) E {
	return Lt(string(f), val)
}

func (f Field) Lte(val any) E {
	return Lte(string(f), val)
}

func (f Field) In(val any) E {
	return In(string(f), val)
}

func (f Field) Nin(val any) E {
	return Nin(string(f), val)
}
