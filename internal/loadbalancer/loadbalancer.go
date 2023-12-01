// Package loadbalancer provides functions and types for inspecting loadbalancers
package loadbalancer

import (
	"context"
	"errors"

	lbapi "go.infratographer.com/load-balancer-api/pkg/client"
	"go.infratographer.com/load-balancer-api/pkg/metadata"
	"go.infratographer.com/x/gidx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/config"
)

// NewLoadBalancer will create a new loadbalancer object
func NewLoadBalancer(ctx context.Context, logger *zap.SugaredLogger, client *lbapi.Client, subj gidx.PrefixedID, adds []gidx.PrefixedID) (*LoadBalancer, error) {
	l := new(LoadBalancer)
	l.isLoadBalancer(subj, adds)

	ctx, span := otel.Tracer(instrumentationName).Start(ctx, "NewLoadBalancer")
	defer span.End()

	if l.LbType != TypeNoLB {
		log := logger.With("loadbalancer", l.LoadBalancerID)

		data, err := client.GetLoadBalancer(ctx, l.LoadBalancerID.String())
		if err != nil {
			log.Errorw("failed to get loadbalancer from API", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			if errors.Is(err, lbapi.ErrLBNotfound) || errors.Is(err, lbapi.ErrPermissionDenied) {
				log.Infow("ignoring event", "error", err)
				return nil, ErrIgnoreEvent
			}

			// some other error, check metadata-api for load balancer state
			lbMetadata, mErr := client.NodeMetadata(ctx, l.LoadBalancerID.String())
			if mErr != nil {
				log.Errorw("failed to get loadbalancer metadata from node-resolver", "error", mErr)
				return nil, err
			}

			status, mErr := metadata.GetLoadbalancerStatus(lbMetadata.Statuses, config.AppConfig.Metadata.StatusNamespaceID, metadata.LoadBalancerAPISource)
			if mErr != nil {
				log.Errorw("failed to find loadbalancer metadata from node-resolver", "error", mErr, "statusNamespaceID", config.AppConfig.Metadata.StatusNamespaceID)
				return nil, err
			}

			span.SetAttributes(attribute.String("loadbalancerState", string(status.State)))

			if status.State == metadata.LoadBalancerStateTerminating || status.State == metadata.LoadBalancerStateDeleted {
				log.Infow("ignoring event", "loadbalancerState", status.State)
				return nil, ErrIgnoreEvent
			}

			return nil, err
		}

		l.LbData = data
	}

	return l, nil
}

func (l *LoadBalancer) isLoadBalancer(subj gidx.PrefixedID, adds []gidx.PrefixedID) {
	check, lbID := getLBFromAddSubjs(adds)

	switch {
	case subj.Prefix() == LBPrefix:
		l.LoadBalancerID = subj
		l.LbType = TypeLB

		return
	case check:
		l.LoadBalancerID = lbID
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
