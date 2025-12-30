package mgo

import "go.mongodb.org/mongo-driver/v2/bson"

// Type Aliases for BSON primitives to avoid importing the official driver package.
type D = bson.D
type E = bson.E
type M = bson.M
type A = bson.A
