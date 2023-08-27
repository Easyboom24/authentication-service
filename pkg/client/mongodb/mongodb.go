package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context, host, port, username, password, datebase, authDB string) (db *mongo.Database, err error) {
	var mogoDBURL string
	var isAuth bool
	if username == "" && password == "" {
		mogoDBURL = fmt.Sprintf("mongodb://%s:%s", host, port) 
		isAuth = false
	} else {
		mogoDBURL = fmt.Sprintf("mongodb://%s:%s@%s:%s", username, password, host, port)
		isAuth = true
	}

	clientOptions := options.Client().ApplyURI(mogoDBURL)
	if isAuth {
		if authDB == "" {
			authDB = datebase
		}
		clientOptions.SetAuth(options.Credential{
			AuthSource: authDB,
			Username: username,
			Password: password,
		})
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Mongodb: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Failed to ping Mongodb: %v", err)
	}

	return client.Database(datebase), nil

}