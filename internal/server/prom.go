package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const subsystem = "loadbalancer_provider_haproxy"

var numberIPsRequestedAndReleasedGauge = promauto.NewGauge(
	prometheus.GaugeOpts{
		Subsystem: subsystem,
		Name:      "ips_requested_and_released",
		Help:      "Count of IPs requested and released",
	},
)
