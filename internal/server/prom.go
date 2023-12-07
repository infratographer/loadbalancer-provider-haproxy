package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const subsystem = "loadbalancer_provider_haproxy"

var (
	numberIPsRequestedGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "ips_requested_count",
			Help:      "The total number of IPs requested",
		},
	)
	numberIPsReleasedGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: subsystem,
			Name:      "ips_released_count",
			Help:      "The total number of IPs released",
		},
	)
)
