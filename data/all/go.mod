module github.com/ncobase/ncore/data/all

go 1.25.3

require (
	github.com/ncobase/ncore/data/elasticsearch v0.2.0
	github.com/ncobase/ncore/data/kafka v0.2.0
	github.com/ncobase/ncore/data/meilisearch v0.2.0
	github.com/ncobase/ncore/data/mongodb v0.2.0
	github.com/ncobase/ncore/data/mysql v0.2.0
	github.com/ncobase/ncore/data/neo4j v0.2.0
	github.com/ncobase/ncore/data/opensearch v0.2.0
	github.com/ncobase/ncore/data/postgres v0.2.0
	github.com/ncobase/ncore/data/rabbitmq v0.2.0
	github.com/ncobase/ncore/data/redis v0.2.0
	github.com/ncobase/ncore/data/sqlite v0.2.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/elastic/elastic-transport-go/v8 v8.8.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.19.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/mattn/go-sqlite3 v1.14.33 // indirect
	github.com/meilisearch/meilisearch-go v0.35.1 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/ncobase/ncore/data v0.1.23 // indirect
	github.com/ncobase/ncore/utils v0.1.24 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4 // indirect
	github.com/opensearch-project/opensearch-go/v4 v4.6.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.23 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/redis/go-redis/v9 v9.17.2 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/ncobase/ncore/data/elasticsearch => ../elasticsearch
	github.com/ncobase/ncore/data/kafka => ../kafka
	github.com/ncobase/ncore/data/meilisearch => ../meilisearch
	github.com/ncobase/ncore/data/mongodb => ../mongodb
	github.com/ncobase/ncore/data/mysql => ../mysql
	github.com/ncobase/ncore/data/neo4j => ../neo4j
	github.com/ncobase/ncore/data/opensearch => ../opensearch
	github.com/ncobase/ncore/data/postgres => ../postgres
	github.com/ncobase/ncore/data/rabbitmq => ../rabbitmq
	github.com/ncobase/ncore/data/redis => ../redis
	github.com/ncobase/ncore/data/sqlite => ../sqlite
)
