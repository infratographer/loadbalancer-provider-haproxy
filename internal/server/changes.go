package server

import (
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/ipam"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/loadbalancer"
)

func (s *Server) processLoadBalancerChangeCreate(lb *loadbalancer.LoadBalancer) error {
	if err := ipam.RequestAddress(s.Context, s.Logger, lb.LoadBalancerID.String()); err != nil {
		return err
	}

	return nil
}

func (s *Server) processLoadBalancerChangeDelete(lb *loadbalancer.LoadBalancer) error {
	if err := ipam.ReleaseAddress(s.Context, s.Logger, lb.LoadBalancerID.String()); err != nil {
		return err
	}

	return nil
}
