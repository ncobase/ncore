package connection

import (
	"context"
	"errors"
	"ncobase/common/data/config"
	"ncobase/common/logger"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// newNeo4jClient creates a new Neo4j client
func newNeo4jClient(conf *config.Neo4j) (neo4j.DriverWithContext, error) {
	if conf == nil || conf.URI == "" {
		logger.Infof(context.Background(), "Neo4j configuration is nil or empty")
		return nil, errors.New("neo4j configuration is nil or empty")
	}

	driver, err := neo4j.NewDriverWithContext(conf.URI, neo4j.BasicAuth(conf.Username, conf.Password, ""))
	if err != nil {
		logger.Errorf(context.Background(), "Neo4j connect error: %v", err)
		return nil, err
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		logger.Errorf(context.Background(), "Neo4j verify connectivity error: %v", err)
		return nil, err
	}

	logger.Infof(context.Background(), "Neo4j connected")

	return driver, nil
}
