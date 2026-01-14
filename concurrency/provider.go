package concurrency

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/concurrency/worker"
)

// ProviderSet is the wire provider set for the concurrency package.
// It provides worker Pool and other concurrency-related components.
//
// Usage:
//
//	wire.Build(
//	    concurrency.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(
	worker.ProviderSet,
)
