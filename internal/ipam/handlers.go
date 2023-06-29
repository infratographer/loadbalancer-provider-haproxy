package ipam

import (
	"context"

	"go.uber.org/zap"

	"go.infratographer.com/ipam-api/pkg/ipamclient"
)

// TODO: need to add ipam client

// RequestAddress will request an address for the specified loadbalancer
func RequestAddress(ctx context.Context, c *ipamclient.Client, logger *zap.SugaredLogger, blockID string, lbID string, lbOwner string) (string, error) {
	logger.Debugw("requesting address", "loadbalancer_id", lbID)

	ip, err := c.CreateIPAddressFromBlock(ctx, blockID, lbID, lbOwner, false)
	if err != nil {
		logger.Debugw("unable to request ip address from IP block", "error", err, "block id", blockID, "load balancer", lbID)
		return "", err
	}

	return ip.IPAddress.IPAddress.IP, nil
}

// ReleaseAddress will release an address from the specified loadbalancer
func ReleaseAddress(ctx context.Context, c *ipamclient.Client, logger *zap.SugaredLogger, lbID string) error {
	// TODO: actually release address
	logger.Debugw("releasing address", "loadbalancer_id", lbID)
	return nil
}
