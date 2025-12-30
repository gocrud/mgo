package mgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UpdateBuilder struct {
	coll     *Collection
	filter   D
	update   D
	ctx      context.Context
	unscoped bool
}

func NewUpdateBuilder(coll *Collection) *UpdateBuilder {
	return &UpdateBuilder{
		coll:   coll,
		filter: D{},
		update: D{},
		ctx:    context.Background(),
	}
}

func (b *UpdateBuilder) Ctx(ctx context.Context) *UpdateBuilder {
	b.ctx = ctx
	return b
}

func (b *UpdateBuilder) Tx(ctx context.Context) *UpdateBuilder {
	return b.Ctx(ctx)
}

func (b *UpdateBuilder) Where(filters ...E) *UpdateBuilder {
	b.filter = append(b.filter, filters...)
	return b
}

func (b *UpdateBuilder) Set(key string, value any) *UpdateBuilder {
	b.pushOp("$set", key, value)
	return b
}

func (b *UpdateBuilder) Inc(key string, value any) *UpdateBuilder {
	b.pushOp("$inc", key, value)
	return b
}

func (b *UpdateBuilder) Push(key string, value any) *UpdateBuilder {
	b.pushOp("$push", key, value)
	return b
}

func (b *UpdateBuilder) Pull(key string, value any) *UpdateBuilder {
	b.pushOp("$pull", key, value)
	return b
}

func (b *UpdateBuilder) Unscoped() *UpdateBuilder {
	b.unscoped = true
	return b
}

func (b *UpdateBuilder) Restore() *UpdateBuilder {
	b.unscoped = true
	b.Set("deleted_at", nil)
	return b
}

func (b *UpdateBuilder) pushOp(op, key string, value any) {
	for i := range b.update {
		if b.update[i].Key == op {
			if d, ok := b.update[i].Value.(D); ok {
				b.update[i].Value = append(d, E{Key: key, Value: value})
			}
			return
		}
	}
	b.update = append(b.update, E{Key: op, Value: D{{Key: key, Value: value}}})
}

func (b *UpdateBuilder) buildFilter() D {
	filter := b.filter
	if b.coll.softDelete && !b.unscoped {
		filter = append(filter, E{Key: "deleted_at", Value: nil})
	}
	return filter
}

func (b *UpdateBuilder) One() (*mongo.UpdateResult, error) {
	if b.coll.autoTime {
		b.Set("updated_at", time.Now())
	}
	return b.coll.UpdateOne(b.ctx, b.buildFilter(), b.update)
}

func (b *UpdateBuilder) Many() (*mongo.UpdateResult, error) {
	if b.coll.autoTime {
		b.Set("updated_at", time.Now())
	}
	return b.coll.UpdateMany(b.ctx, b.buildFilter(), b.update)
}
