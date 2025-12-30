package mgo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	*mongo.Client
}

func Connect(uri string, opts ...*options.ClientOptions) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri).SetBSONOptions(&options.BSONOptions{
		ObjectIDAsHexString: true,
	})
	// Prepend default options, user options can override
	finalOpts := append([]*options.ClientOptions{clientOpts}, opts...)

	client, err := mongo.Connect(finalOpts...)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

func (c *Client) Database(name string, opts ...*options.DatabaseOptions) *Database {
	// TODO: Fix options passing for V2
	return &Database{Database: c.Client.Database(name)}
}

// Tx executes a transaction.
func (c *Client) Tx(ctx context.Context, fn func(txCtx context.Context) error) error {
	session, err := c.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}

type Database struct {
	*mongo.Database
}

// Collection returns a collection.
func (d *Database) Collection(name string, opts ...*options.CollectionOptions) *Collection {
	// TODO: Fix options passing for V2
	return &Collection{
		Collection: d.Database.Collection(name),
		db:         d,
	}
}
