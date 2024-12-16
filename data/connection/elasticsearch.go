package connection

import (
	"errors"
	"fmt"
	"io"
	"ncobase/common/data/config"
	"ncobase/common/data/elastic"
)

// newElasticsearchClient creates a new Elasticsearch client
func newElasticsearchClient(conf *config.Elasticsearch) (*elastic.Client, error) {
	if conf == nil || len(conf.Addresses) == 0 {
		return nil, errors.New("elasticsearch configuration is nil or empty")
	}

	es, err := elastic.NewClient(conf.Addresses, conf.Username, conf.Password)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client creation error: %w", err)
	}

	res, err := es.GetClient().Info()
	if err != nil {
		return nil, fmt.Errorf("elasticsearch connect error: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch info error: %s", res.Status())
	}

	return es, nil
}
