package connection

import (
	"context"
	"errors"
	"ncobase/common/config"
	"ncobase/common/log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// newNeo4jClient creates a new Neo4j client
func newNeo4jClient(conf *config.Neo4j) (neo4j.DriverWithContext, error) {
	if conf == nil || conf.URI == "" {
		log.Infof(context.Background(), "Neo4j configuration is nil or empty")
		return nil, errors.New("neo4j configuration is nil or empty")
	}

	driver, err := neo4j.NewDriverWithContext(conf.URI, neo4j.BasicAuth(conf.Username, conf.Password, ""))
	if err != nil {
		log.Errorf(context.Background(), "Neo4j connect error: %v", err)
		return nil, err
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		log.Errorf(context.Background(), "Neo4j verify connectivity error: %v", err)
		return nil, err
	}

	log.Infof(context.Background(), "Neo4j connected")

	return driver, nil
}
