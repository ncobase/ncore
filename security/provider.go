package security

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/security/jwt"
)

// ProviderSet is the wire provider set for the security package.
// It provides JWT TokenManager and other security-related components.
//
// Usage:
//
//	wire.Build(
//	    config.ProviderSet,
//	    security.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(
	jwt.ProviderSet,
)
