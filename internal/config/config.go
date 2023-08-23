// Package config provides a struct to store the applications config
package config

import (
	"go.infratographer.com/x/oauth2x"

	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/loggingx"
	"go.infratographer.com/x/otelx"
)

// AppConfig contains the application configuration structure.
var AppConfig struct {
	Logging loggingx.Config
	Events  events.Config
	Server  echox.Config
	Tracing otelx.Config
	OIDC    OIDCClientConfig
}

// EventsConfig stores the configuration for a load-balancer-api events config
// type EventsConfig struct {
// 	events.Config `mapstructure:",squash"`
// }

// OIDCClientConfig stores the configuration for an OIDC client
type OIDCClientConfig struct {
	Client oauth2x.Config
}
