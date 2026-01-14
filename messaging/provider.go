package messaging

import (
	"github.com/google/wire"
	"github.com/ncobase/ncore/messaging/email"
)

// ProviderSet is the wire provider set for the messaging package.
// It provides email Sender and other messaging-related components.
//
// Usage:
//
//	wire.Build(
//	    messaging.ProviderSet,
//	    // ... other providers
//	)
var ProviderSet = wire.NewSet(
	email.ProviderSet,
)
