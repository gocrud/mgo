package mgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InsertBuilder struct {
	coll *Collection
	docs []any
	ctx  context.Context
}

func NewInsertBuilder(coll *Collection) *InsertBuilder {
	return &InsertBuilder{
		coll: coll,
		ctx:  context.Background(),
	}
}

func (b *InsertBuilder) Ctx(ctx context.Context) *InsertBuilder {
	b.ctx = ctx
	return b
}

func (b *InsertBuilder) Doc(doc any) *InsertBuilder {
	b.docs = append(b.docs, doc)
	return b
}

func (b *InsertBuilder) Docs(docs ...any) *InsertBuilder {
	b.docs = append(b.docs, docs...)
	return b
}

func (b *InsertBuilder) One() (*mongo.InsertOneResult, error) {
	if len(b.docs) == 0 {
		return nil, nil
	}
	doc := b.docs[0]

	if b.coll.autoTime {
		now := time.Now()
		applyTimeHook(doc, now)
	}

	return b.coll.InsertOne(b.ctx, doc)
}

func (b *InsertBuilder) Many() (*mongo.InsertManyResult, error) {
	if len(b.docs) == 0 {
		return nil, nil
	}

	if b.coll.autoTime {
		now := time.Now()
		for i := range b.docs {
			applyTimeHook(b.docs[i], now)
		}
	}

	return b.coll.InsertMany(b.ctx, b.docs)
}

func applyTimeHook(doc any, t time.Time) {
	if h, ok := doc.(TimeHook); ok {
		h.SetCreatedAt(t)
		h.SetUpdatedAt(t)
	}
}
