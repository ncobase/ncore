package observes

import (
	"github.com/getsentry/sentry-go"
)

type SentryOptions struct {
	Dsn         string
	Name        string
	Release     string
	Environment string
}

// NewSentry is the register sentry
func NewSentry(opt *SentryOptions) error {
	// if not exist sentry config, skip initialization
	if opt == nil {
		return nil
	}

	return sentry.Init(sentry.ClientOptions{
		Dsn:              opt.Dsn,
		AttachStacktrace: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
		ServerName:       opt.Name,
		Release:          opt.Release,
		Environment:      opt.Environment,
	})
}
