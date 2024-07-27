package config

import "github.com/spf13/viper"

// Sentry config struct
type Sentry struct {
	Endpoint string
}

// getSentryConfig get sentry config
func getSentryConfig(v *viper.Viper) *Sentry {
	return &Sentry{
		Endpoint: v.GetString("observes.sentry.endpoint"),
	}
}

// Tracer config struct
type Tracer struct {
	Endpoint string
}

// getTracerConfig get tracer config
func getTracerConfig(v *viper.Viper) *Tracer {
	return &Tracer{
		Endpoint: v.GetString("observes.tracer.endpoint"),
	}
}

// Observes config struct
type Observes struct {
	Sentry *Sentry
	Tracer *Tracer
}

// get Observes config
func getObservesConfig(v *viper.Viper) *Observes {
	return &Observes{
		Sentry: getSentryConfig(v),
		Tracer: getTracerConfig(v),
	}
}
