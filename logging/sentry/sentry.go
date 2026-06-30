package sentry

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

type SentryOptions struct {
	Dsn         string
	Name        string
	Release     string
	Environment string
	SampleRate  float64
}

// NewSentry initializes the Sentry SDK.
func NewSentry(opt *SentryOptions) error {
	if opt == nil {
		return nil
	}
	if opt.SampleRate < 0 || opt.SampleRate > 1 {
		return fmt.Errorf("sentry sample rate must be between 0 and 1")
	}

	clientOptions := sentry.ClientOptions{
		Dsn:              opt.Dsn,
		AttachStacktrace: true,
		ServerName:       opt.Name,
		Release:          opt.Release,
		Environment:      opt.Environment,
	}
	if opt.SampleRate > 0 {
		clientOptions.SampleRate = opt.SampleRate
	}

	return sentry.Init(clientOptions)
}
