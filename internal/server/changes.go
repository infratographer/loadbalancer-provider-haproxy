package server

import (
	"time"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/ipam"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/loadbalancer"

	"go.infratographer.com/x/events"
)

func (s *Server) processLoadBalancerChangeCreate(lb *loadbalancer.LoadBalancer) error {
	// for now, limit to one IP address per loadbalancer
	if len(lb.LbData.LoadBalancer.IPAddresses) == 0 {
		if ip, err := ipam.RequestAddress(s.Context, s.IPAMClient, s.Logger, s.IPBlock, lb.LoadBalancerID.String(), lb.LbData.LoadBalancer.Owner.ID); err != nil {
			return err
		} else {
			msg := events.EventMessage{
				EventType: "ip-address.assigned",
				SubjectID: lb.LoadBalancerID,
				Timestamp: time.Now().UTC(),
			}

			if err := s.Publisher.PublishEvent(s.Context, "load-balancer", msg); err != nil {
				s.Logger.Debugw("failed to publish event", "error", err, "ip", ip, "loadbalancer", lb.LoadBalancerID, "block", s.IPBlock)
				return err
			}
		}
	}

	return nil
}

func (s *Server) processLoadBalancerChangeDelete(lb *loadbalancer.LoadBalancer) error {
	if err := ipam.ReleaseAddress(s.Context, s.IPAMClient, s.Logger, lb.LoadBalancerID.String()); err != nil {
		return err
	}

	msg := events.EventMessage{
		EventType: "ip-address.unassigned",
		SubjectID: lb.LoadBalancerID,
		Timestamp: time.Now().UTC(),
	}

	if err := s.Publisher.PublishEvent(s.Context, "load-balancer", msg); err != nil {
		s.Logger.Debugw("failed to publish event", "error", err, "loadbalancer", lb.LoadBalancerID, "block", s.IPBlock)
		return err
	}

	return nil
}
