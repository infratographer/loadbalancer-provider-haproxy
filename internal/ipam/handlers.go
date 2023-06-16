package ipam

import (
	"context"

	"go.uber.org/zap"
)

// TODO: need to add ipam client

// RequestAddress will request an address for the specified loadbalancer
func RequestAddress(ctx context.Context, logger *zap.SugaredLogger, lbID string) error {
	// TODO: actually request address
	logger.Infow("requesting address", "loadbalancer_id", lbID)
	return nil
}

// ReleaseAddress will release an address from the specified loadbalancer
func ReleaseAddress(ctx context.Context, logger *zap.SugaredLogger, lbID string) error {
	// TODO: actually release address
	logger.Infow("releasing address", "loadbalancer_id", lbID)
	return nil
}
