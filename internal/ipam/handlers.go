package ipam

import (
	"context"

	"go.uber.org/zap"

	"go.infratographer.com/ipam-api/pkg/ipamclient"
)

// TODO: need to add ipam client

// RequestAddress will request an address for the specified loadbalancer
func RequestAddress(ctx context.Context, c *ipamclient.Client, logger *zap.SugaredLogger, blockID string, lbID string, lbOwner string) (string, error) {
	logger.Infow("requesting address", "loadbalancer_id", lbID)

	ip, err := c.CreateIPAddressFromBlock(ctx, blockID, lbID, lbOwner, false)
	if err != nil {
		logger.Debugw("unable to request ip address from IP block", "error", err, "block id", blockID, "load balancer", lbID)
		return "", err
	}

	logger.Infow("address requested", "loadbalancer_id", lbID, "ip_address", ip.IPAddress.IPAddress.IP)

	return ip.IPAddress.IPAddress.IP, nil
}

// ReleaseAddress will release an address from the specified loadbalancer
func ReleaseAddress(ctx context.Context, c *ipamclient.Client, logger *zap.SugaredLogger, lbID string) error {
	logger.Infow("releasing address", "loadbalancer_id", lbID)

	// TODO: this is a hack as we don't have a way to get the IP address from the loadbalancer as
	// it has already been deleted. We're using an entities query directly against IPAM to get what
	// we need. Ideally we could query the supergraph for this, but need to look into softdeletes so
	// that the LB data is still available after a delete.
	addr, err := c.GetIPAddresses(ctx, lbID)
	if err != nil {
		logger.Debugw("unable to get ip address for loadbalancer", "error", err, "load balancer", lbID)
	}

	for _, ip := range addr.Entities[0].IPAddressableFragment.IPAddresses {
		if _, err := c.DeleteIPAddress(ctx, ip.ID); err != nil {
			logger.Debugw("unable to release ip address from loadbalancer", "error", err, "load balancer", lbID, "IPAddressID", ip.ID, "IPAddress", ip.IP)
			return err
		}

		logger.Infow("released ip address from loadbalancer", "error", err, "load balancer", lbID, "IPAddressID", ip.ID, "IPAddress", ip.IP)
	}

	return nil
}
