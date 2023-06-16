package loadbalancer

import (
	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"
	"go.infratographer.com/x/gidx"
)

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
	LbData         *lbapi.GetLoadBalancer
	LbType         int
}
