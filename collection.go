package mgo

import "go.mongodb.org/mongo-driver/v2/mongo"

type Collection struct {
	*mongo.Collection
	db         *Database
	autoTime   bool
	softDelete bool
}

func (c *Collection) AutoTime() *Collection {
	c.autoTime = true
	return c
}

func (c *Collection) SoftDelete() *Collection {
	c.softDelete = true
	return c
}

func (c *Collection) Find() *FindBuilder {
	return NewFindBuilder(c)
}

func (c *Collection) Insert() *InsertBuilder {
	return NewInsertBuilder(c)
}

func (c *Collection) Update() *UpdateBuilder {
	return NewUpdateBuilder(c)
}

func (c *Collection) Delete() *DeleteBuilder {
	return NewDeleteBuilder(c)
}

func (c *Collection) Aggregate() *AggregateBuilder {
	return NewAggregateBuilder(c)
}
