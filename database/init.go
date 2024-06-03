package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func InitMongoDB() error {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	Client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}
	return nil
}
