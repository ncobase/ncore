module github.com/ncobase/ncore/data/opensearch

go 1.25.3

require (
	github.com/ncobase/ncore/data v0.1.23
	github.com/ncobase/ncore/data/search v0.0.0
	github.com/ncobase/ncore/utils v0.1.24
	github.com/opensearch-project/opensearch-go/v4 v4.6.0
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/google/wire v0.7.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/ncobase/ncore/data => ../
	github.com/ncobase/ncore/data/search => ../search
)
