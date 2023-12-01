// Package loadbalancer provides functions and types for inspecting loadbalancers
package loadbalancer

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	lbapi "go.infratographer.com/load-balancer-api/pkg/client"
	"go.infratographer.com/x/gidx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// NewLoadBalancer will create a new loadbalancer object
func NewLoadBalancer(ctx context.Context, logger *zap.SugaredLogger, client *lbapi.Client, subj gidx.PrefixedID, adds []gidx.PrefixedID) (*LoadBalancer, error) {
	l := new(LoadBalancer)
	l.isLoadBalancer(subj, adds)
	l.withPromGauges()

	ctx, span := otel.Tracer(instrumentationName).Start(ctx, "NewLoadBalancer")
	defer span.End()

	if l.LbType != TypeNoLB {
		data, err := client.GetLoadBalancer(ctx, l.LoadBalancerID.String())
		if err != nil {
			logger.Errorw("unable to get loadbalancer from API", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return nil, err
		}

		l.LbData = data
	}

	return l, nil
}

func (l *LoadBalancer) isLoadBalancer(subj gidx.PrefixedID, adds []gidx.PrefixedID) {
	check, subs := getLBFromAddSubjs(adds)

	switch {
	case subj.Prefix() == LBPrefix:
		l.LoadBalancerID = subj
		l.LbType = TypeLB

		return
	case check:
		l.LoadBalancerID = subs
		l.LbType = TypeAssocLB

		return
	default:
		l.LbType = TypeNoLB
		return
	}
}

func getLBFromAddSubjs(adds []gidx.PrefixedID) (bool, gidx.PrefixedID) {
	for _, i := range adds {
		if i.Prefix() == LBPrefix {
			return true, i
		}
	}

	id := new(gidx.PrefixedID)

	return false, *id
}

func (l *LoadBalancer) withPromGauges() *LoadBalancer {
	o := prometheus.GaugeOpts{
		Name: "ips_requested_and_released",
		Help: "IP addresses requested and released by LoadBalancer " + string(l.LoadBalancerID),
	}

	l.PromGauge = promauto.NewGauge(o)

	return l
}
