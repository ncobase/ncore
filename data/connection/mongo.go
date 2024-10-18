package connection

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// newMongoClient creates a new MongoDB client
func newMongoClient(conf *config.MongoDB) (*mongo.Client, error) {
	if conf == nil || conf.URI == "" {
		log.Infof(context.Background(), "MongoDB configuration is nil or empty")
		return nil, nil
	}

	clientOptions := options.Client().ApplyURI(conf.URI)
	if conf.Username != "" && conf.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: conf.Username,
			Password: conf.Password,
		})
	}

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Errorf(context.Background(), "MongoDB connect error: %v", err)
		return nil, err
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Errorf(context.Background(), "MongoDB ping error: %v", err)
		return nil, err
	}

	log.Infof(context.Background(), "MongoDB connected")

	return client, nil
}
