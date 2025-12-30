package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AggregateBuilder struct {
	coll     *Collection
	pipeline A
	ctx      context.Context
}

func NewAggregateBuilder(coll *Collection) *AggregateBuilder {
	return &AggregateBuilder{
		coll:     coll,
		pipeline: A{},
		ctx:      context.Background(),
	}
}

func (b *AggregateBuilder) Ctx(ctx context.Context) *AggregateBuilder {
	b.ctx = ctx
	return b
}

func (b *AggregateBuilder) Match(filters ...E) *AggregateBuilder {
	match := D{}
	match = append(match, filters...)
	b.pipeline = append(b.pipeline, D{{Key: "$match", Value: match}})
	return b
}

func (b *AggregateBuilder) Group(id any, accumulators ...E) *AggregateBuilder {
	group := D{{Key: "_id", Value: id}}
	group = append(group, accumulators...)
	b.pipeline = append(b.pipeline, D{{Key: "$group", Value: group}})
	return b
}

func (b *AggregateBuilder) SortAsc(fields ...string) *AggregateBuilder {
	sort := D{}
	for _, f := range fields {
		sort = append(sort, E{Key: f, Value: 1})
	}
	b.pipeline = append(b.pipeline, D{{Key: "$sort", Value: sort}})
	return b
}

func (b *AggregateBuilder) SortDesc(fields ...string) *AggregateBuilder {
	sort := D{}
	for _, f := range fields {
		sort = append(sort, E{Key: f, Value: -1})
	}
	b.pipeline = append(b.pipeline, D{{Key: "$sort", Value: sort}})
	return b
}

func (b *AggregateBuilder) Limit(limit int64) *AggregateBuilder {
	b.pipeline = append(b.pipeline, D{{Key: "$limit", Value: limit}})
	return b
}

func (b *AggregateBuilder) Skip(skip int64) *AggregateBuilder {
	b.pipeline = append(b.pipeline, D{{Key: "$skip", Value: skip}})
	return b
}

func (b *AggregateBuilder) Lookup(from, localField, foreignField, as string) *AggregateBuilder {
	b.pipeline = append(b.pipeline, D{{Key: "$lookup", Value: D{
		{Key: "from", Value: from},
		{Key: "localField", Value: localField},
		{Key: "foreignField", Value: foreignField},
		{Key: "as", Value: as},
	}}})
	return b
}

// Join is a shortcut for Lookup + Unwind (Left Join 1:1)
// It assumes foreignField is "_id" if not specified (simplified version in docs uses 3 args)
func (b *AggregateBuilder) Join(from, localField, as string) *AggregateBuilder {
	b.Lookup(from, localField, "_id", as)
	b.pipeline = append(b.pipeline, D{{Key: "$unwind", Value: D{
		{Key: "path", Value: "$" + as},
		{Key: "preserveNullAndEmptyArrays", Value: true},
	}}})
	return b
}

func (b *AggregateBuilder) Pipeline(stages A) *AggregateBuilder {
	b.pipeline = append(b.pipeline, stages...)
	return b
}

func (b *AggregateBuilder) All(results any) error {
	cursor, err := b.coll.Collection.Aggregate(b.ctx, b.pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(b.ctx)
	return cursor.All(b.ctx, results)
}

func (b *AggregateBuilder) One(result any) error {
	b.Limit(1)
	cursor, err := b.coll.Collection.Aggregate(b.ctx, b.pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(b.ctx)

	if !cursor.Next(b.ctx) {
		if err := cursor.Err(); err != nil {
			return err
		}
		return mongo.ErrNoDocuments
	}
	return cursor.Decode(result)
}

// Re-implementing Paginate to match README usage pattern:
// .Paginate(1, 20).All(&list)

type AggregatePaginator struct {
	builder *AggregateBuilder
	page    int64
	size    int64
}

func (b *AggregateBuilder) Paginate(page, size int64) *AggregatePaginator {
	return &AggregatePaginator{
		builder: b,
		page:    page,
		size:    size,
	}
}

func (p *AggregatePaginator) All(results any) (*PaginatedResult, error) {
	b := p.builder
	page := p.page
	size := p.size

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	// Count pipeline
	countPipeline := append(A{}, b.pipeline...)
	countPipeline = append(countPipeline, D{{Key: "$count", Value: "total"}})

	var countRes []struct {
		Total int64 `bson:"total"`
	}
	cursor, err := b.coll.Collection.Aggregate(b.ctx, countPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(b.ctx)

	if err := cursor.All(b.ctx, &countRes); err != nil {
		return nil, err
	}

	var total int64
	if len(countRes) > 0 {
		total = countRes[0].Total
	}

	// Data pipeline
	b.Skip((page - 1) * size)
	b.Limit(size)

	if err := b.All(results); err != nil {
		return nil, err
	}

	totalPages := (total + size - 1) / size

	return &PaginatedResult{
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		Size:       size,
		List:       nil, // List is populated in 'results'
	}, nil
}
