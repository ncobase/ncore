package connection

import (
	"context"
	"errors"
	"io"
	"ncobase/common/data/config"
	"ncobase/common/data/elastic"
	"ncobase/common/logger"
)

// newElasticsearchClient creates a new Elasticsearch client
func newElasticsearchClient(conf *config.Elasticsearch) (*elastic.Client, error) {
	if conf == nil || len(conf.Addresses) == 0 {
		logger.Infof(context.Background(), "Elasticsearch configuration is nil or empty")
		return nil, errors.New("elasticsearch configuration is nil or empty")
	}

	es, err := elastic.NewClient(conf.Addresses, conf.Username, conf.Password)
	if err != nil {
		logger.Errorf(context.Background(), "Elasticsearch client creation error: %v", err)
		return nil, err
	}

	res, err := es.GetClient().Info()
	if err != nil {
		logger.Errorf(context.Background(), "Elasticsearch connect error: %v", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logger.Errorf(context.Background(), "Elasticsearch response body close error: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		logger.Errorf(context.Background(), "Elasticsearch info error: %s", res.Status())
		return nil, errors.New(res.Status())
	}

	logger.Infof(context.Background(), "Elasticsearch connected")

	return es, nil
}
