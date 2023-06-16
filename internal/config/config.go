// Package config provides a struct to store the applications config
package config

import (
	"go.infratographer.com/loadbalancer-manager-haproxy/x/oauth2x"

	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/loggingx"
	"go.infratographer.com/x/otelx"
)

// AppConfig contains the application configuration structure.
var AppConfig struct {
	Logging loggingx.Config
	Events  EventsConfig
	Server  echox.Config
	Tracing otelx.Config
	OIDC    oauth2x.Config
}

// EventsConfig stores the configuration for a load-balancer-api events config
type EventsConfig struct {
	Subscriber events.SubscriberConfig
}
