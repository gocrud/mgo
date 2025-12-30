package mgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DeleteBuilder struct {
	coll   *Collection
	filter D
	ctx    context.Context
	hard   bool
}

func NewDeleteBuilder(coll *Collection) *DeleteBuilder {
	return &DeleteBuilder{
		coll:   coll,
		filter: D{},
		ctx:    context.Background(),
	}
}

func (b *DeleteBuilder) Ctx(ctx context.Context) *DeleteBuilder {
	b.ctx = ctx
	return b
}

func (b *DeleteBuilder) Tx(ctx context.Context) *DeleteBuilder {
	return b.Ctx(ctx)
}

func (b *DeleteBuilder) Where(filters ...E) *DeleteBuilder {
	b.filter = append(b.filter, filters...)
	return b
}

func (b *DeleteBuilder) Hard() *DeleteBuilder {
	b.hard = true
	return b
}

func (b *DeleteBuilder) One() (*mongo.DeleteResult, error) {
	if b.coll.softDelete && !b.hard {
		// Soft delete: Update deleted_at
		res, err := b.coll.UpdateOne(b.ctx, b.filter, D{{Key: "$set", Value: D{{Key: "deleted_at", Value: time.Now()}}}})
		if err != nil {
			return nil, err
		}
		return &mongo.DeleteResult{DeletedCount: res.ModifiedCount}, nil
	}
	return b.coll.DeleteOne(b.ctx, b.filter)
}

func (b *DeleteBuilder) Many() (*mongo.DeleteResult, error) {
	if b.coll.softDelete && !b.hard {
		// Soft delete: Update deleted_at
		res, err := b.coll.UpdateMany(b.ctx, b.filter, D{{Key: "$set", Value: D{{Key: "deleted_at", Value: time.Now()}}}})
		if err != nil {
			return nil, err
		}
		return &mongo.DeleteResult{DeletedCount: res.ModifiedCount}, nil
	}
	return b.coll.DeleteMany(b.ctx, b.filter)
}
