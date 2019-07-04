package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//NewClient new mongo client
func NewClient(URL string) (c *mongo.Client, e error) {
	mongoopt := options.Client().ApplyURI(URL)
	if c, e = mongo.Connect(context.TODO(), mongoopt); e != nil {
		return
	}
	// Check the connection
	if e = c.Ping(context.TODO(), nil); e != nil {
		return nil, e
	}
	fmt.Println("Connected to MongoDB! " + URL)
	return
}
