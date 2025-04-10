package connection

import (
	"context"
	"errors"
	"fmt"
	"github.com/ncobase/ncore/pkg/data/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// newNeo4jClient creates a new Neo4j client
func newNeo4jClient(conf *config.Neo4j) (neo4j.DriverWithContext, error) {
	if conf == nil || conf.URI == "" {
		return nil, errors.New("neo4j configuration is nil or empty")
	}

	driver, err := neo4j.NewDriverWithContext(conf.URI, neo4j.BasicAuth(conf.Username, conf.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j connect error: %w", err)
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		return nil, fmt.Errorf("neo4j verify connectivity error: %w", err)
	}

	return driver, nil
}
