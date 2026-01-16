// Package all provides backward compatibility by importing all drivers at once.
//
// This package is for existing projects that need a quick migration path.
// Import this package to register all available drivers automatically:
//
//	import _ "github.com/ncobase/ncore/data/all"
//
// WARNING: This will include ALL drivers in your binary, resulting in larger
// binary sizes. For new projects, import only the drivers you need:
//
//	import (
//	    _ "github.com/ncobase/ncore/data/postgres"
//	    _ "github.com/ncobase/ncore/data/redis"
//	)
package all

// Package all imports all database, cache, search, and messaging drivers for convenience.
// This is useful for applications that want to support all drivers without explicitly importing each one.
//
// Usage:
//
//	import _ "github.com/ncobase/ncore/data/all"
import (
	// Database drivers
	_ "github.com/ncobase/ncore/data/mongodb"
	_ "github.com/ncobase/ncore/data/mysql"
	_ "github.com/ncobase/ncore/data/neo4j"
	_ "github.com/ncobase/ncore/data/postgres"
	_ "github.com/ncobase/ncore/data/redis"
	_ "github.com/ncobase/ncore/data/sqlite"

	// Search drivers
	_ "github.com/ncobase/ncore/data/elasticsearch"
	_ "github.com/ncobase/ncore/data/meilisearch"
	_ "github.com/ncobase/ncore/data/opensearch"

	// Messaging drivers
	_ "github.com/ncobase/ncore/data/kafka"
	_ "github.com/ncobase/ncore/data/rabbitmq"
)
