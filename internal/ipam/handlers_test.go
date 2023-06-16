package ipam_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/ipam"
	"go.infratographer.com/x/gidx"
	"go.uber.org/zap"
)

func TestRequestAddress(t *testing.T) {
	err := ipam.RequestAddress(context.TODO(), zap.NewNop().Sugar(), gidx.MustNewID("loadbal").String())
	assert.NoError(t, err)
}

func TestReleaseAddress(t *testing.T) {
	err := ipam.ReleaseAddress(context.TODO(), zap.NewNop().Sugar(), gidx.MustNewID("loadbal").String())
	assert.NoError(t, err)
}
