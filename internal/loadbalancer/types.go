package loadbalancer

import (
	"github.com/prometheus/client_golang/prometheus"
	lbapi "go.infratographer.com/load-balancer-api/pkg/client"
	"go.infratographer.com/x/gidx"
)

// instrumentationName is a unique package name used for tracing
const instrumentationName = "go.infratographer.com/loadbalancer-provider-haproxy/loadbalancer"

const (
	LBPrefix = "loadbal"

	// TypeLB indicates that the subject of a message is a loadbalancer
	TypeLB = 1
	// TypeAssocLB indicates that the loadbalancer was found in associated subjects
	TypeAssocLB = 2
	// TypeNoLB indicates that a loadbalancer was not found in the message
	TypeNoLB = 0
)

type LoadBalancer struct {
	LoadBalancerID gidx.PrefixedID
	LbData         *lbapi.LoadBalancer
	LbType         int
	PromGauge      prometheus.Gauge
}
