package agg

import "go.mongodb.org/mongo-driver/v2/bson"

type E = bson.E
type D = bson.D

func Count(as string) E {
	return E{Key: as, Value: D{{Key: "$sum", Value: 1}}}
}

func Sum(as, field string) E {
	return E{Key: as, Value: D{{Key: "$sum", Value: field}}}
}

func Avg(as, field string) E {
	return E{Key: as, Value: D{{Key: "$avg", Value: field}}}
}

func Min(as, field string) E {
	return E{Key: as, Value: D{{Key: "$min", Value: field}}}
}

func Max(as, field string) E {
	return E{Key: as, Value: D{{Key: "$max", Value: field}}}
}

func First(as, field string) E {
	return E{Key: as, Value: D{{Key: "$first", Value: field}}}
}

func Last(as, field string) E {
	return E{Key: as, Value: D{{Key: "$last", Value: field}}}
}

func Push(as, field string) E {
	return E{Key: as, Value: D{{Key: "$push", Value: field}}}
}

func AddToSet(as, field string) E {
	return E{Key: as, Value: D{{Key: "$addToSet", Value: field}}}
}
