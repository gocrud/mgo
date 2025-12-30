package mgo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FindBuilder struct {
	coll       *Collection
	filter     D
	sort       D
	projection D
	limit      int64
	skip       int64
	ctx        context.Context
	unscoped   bool
	err        error
}

func NewFindBuilder(coll *Collection) *FindBuilder {
	return &FindBuilder{
		coll:   coll,
		filter: D{},
		ctx:    context.Background(),
	}
}

func (b *FindBuilder) Ctx(ctx context.Context) *FindBuilder {
	b.ctx = ctx
	return b
}

func (b *FindBuilder) Where(filters ...E) *FindBuilder {
	b.filter = append(b.filter, filters...)
	return b
}

func (b *FindBuilder) SortAsc(fields ...string) *FindBuilder {
	for _, f := range fields {
		b.sort = append(b.sort, E{Key: f, Value: 1})
	}
	return b
}

func (b *FindBuilder) SortDesc(fields ...string) *FindBuilder {
	for _, f := range fields {
		b.sort = append(b.sort, E{Key: f, Value: -1})
	}
	return b
}

func (b *FindBuilder) Limit(limit int64) *FindBuilder {
	b.limit = limit
	return b
}

func (b *FindBuilder) Skip(skip int64) *FindBuilder {
	b.skip = skip
	return b
}

func (b *FindBuilder) Select(fields ...string) *FindBuilder {
	for _, f := range fields {
		b.projection = append(b.projection, E{Key: f, Value: 1})
	}
	return b
}

func (b *FindBuilder) Unscoped() *FindBuilder {
	b.unscoped = true
	return b
}

func (b *FindBuilder) buildFilter() D {
	filter := b.filter
	if b.coll.softDelete && !b.unscoped {
		filter = append(filter, E{Key: "deleted_at", Value: nil})
	}
	return filter
}

func (b *FindBuilder) Count() (int64, error) {
	if b.err != nil {
		return 0, b.err
	}
	return b.coll.CountDocuments(b.ctx, b.buildFilter())
}

func (b *FindBuilder) All(results any) error {
	if b.err != nil {
		return b.err
	}
	opts := options.Find()
	if len(b.sort) > 0 {
		opts.SetSort(b.sort)
	}
	if len(b.projection) > 0 {
		opts.SetProjection(b.projection)
	}
	if b.limit > 0 {
		opts.SetLimit(b.limit)
	}
	if b.skip > 0 {
		opts.SetSkip(b.skip)
	}

	cursor, err := b.coll.Collection.Find(b.ctx, b.buildFilter(), opts)
	if err != nil {
		return err
	}
	defer cursor.Close(b.ctx)

	return cursor.All(b.ctx, results)
}

func (b *FindBuilder) One(result any) error {
	if b.err != nil {
		return b.err
	}
	opts := options.FindOne()
	if len(b.sort) > 0 {
		opts.SetSort(b.sort)
	}
	if len(b.projection) > 0 {
		opts.SetProjection(b.projection)
	}

	err := b.coll.FindOne(b.ctx, b.buildFilter(), opts).Decode(result)
	if err != nil {
		return err
	}
	return nil
}

type PaginatedResult struct {
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	Page       int64 `json:"page"`
	Size       int64 `json:"size"`
	List       any   `json:"list"`
}

// Paginate returns a Paginator which has an All(results any) method.
func (b *FindBuilder) Paginate(page, size int64) *FindPaginator {
	return &FindPaginator{
		builder: b,
		page:    page,
		size:    size,
	}
}

type FindPaginator struct {
	builder *FindBuilder
	page    int64
	size    int64
}

func (p *FindPaginator) All(results any) (*PaginatedResult, error) {
	b := p.builder
	if b.err != nil {
		return nil, b.err
	}
	page := p.page
	size := p.size

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	filter := b.buildFilter()
	total, err := b.coll.CountDocuments(b.ctx, filter)
	if err != nil {
		return nil, err
	}

	b.limit = size
	b.skip = (page - 1) * size

	if err := b.All(results); err != nil {
		return nil, err
	}

	totalPages := (total + size - 1) / size

	return &PaginatedResult{
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		Size:       size,
		List:       nil, // List is populated in 'results' pointer
	}, nil
}

func (b *FindBuilder) Seek(lastItem any) *FindBuilder {
	if len(b.sort) == 0 {
		return b
	}

	data, err := bson.Marshal(lastItem)
	if err != nil {
		b.err = err
		return b
	}
	var doc D
	bson.Unmarshal(data, &doc)

	if len(b.sort) == 1 {
		s := b.sort[0]
		key := s.Key
		// Handle int, int32, int64 for direction
		var dir int
		switch v := s.Value.(type) {
		case int:
			dir = v
		case int32:
			dir = int(v)
		case int64:
			dir = int(v)
		}

		var val any
		for _, e := range doc {
			if e.Key == key {
				val = e.Value
				break
			}
		}

		var op string
		if dir == 1 {
			op = "$gt"
		} else {
			op = "$lt"
		}

		b.filter = append(b.filter, E{Key: key, Value: D{{Key: op, Value: val}}})
	}
	return b
}

// Helper to check if error is NoDocuments
func IsNoDocuments(err error) bool {
	return err == mongo.ErrNoDocuments
}
