package loadbalancer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.infratographer.com/x/gidx"
	"go.uber.org/zap"

	lbapi "go.infratographer.com/load-balancer-api/pkg/client"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/loadbalancer"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/testutils/mock"
)

func TestNewLoadBalancer(t *testing.T) {
	id := gidx.MustNewID("loadbal")

	api := mock.DummyAPI(id.String())
	api.Start()

	defer api.Close()

	cli := lbapi.NewClient(api.URL)

	subj := id
	adds := []gidx.PrefixedID{gidx.MustNewID("loadprt")}

	lb, err := loadbalancer.NewLoadBalancer(context.TODO(), zap.NewNop().Sugar(), cli, subj, adds)
	assert.NotNil(t, lb)
	assert.NoError(t, err)

	assert.Equal(t, lb.LoadBalancerID, subj)
	assert.Equal(t, lb.LbType, loadbalancer.TypeLB)
	assert.NotNil(t, lb.LbData)

	subj = gidx.MustNewID("loadprt")
	adds = []gidx.PrefixedID{id}

	lb, err = loadbalancer.NewLoadBalancer(context.TODO(), zap.NewNop().Sugar(), cli, subj, adds)
	assert.NotNil(t, lb)
	assert.NoError(t, err)

	assert.Equal(t, lb.LoadBalancerID, id)
	assert.Equal(t, lb.LbType, loadbalancer.TypeAssocLB)
	assert.NotNil(t, lb.LbData)

	subj = gidx.MustNewID("loadprt")
	adds = []gidx.PrefixedID{}

	lb, err = loadbalancer.NewLoadBalancer(context.TODO(), zap.NewNop().Sugar(), cli, subj, adds)
	assert.NotNil(t, lb)
	assert.NoError(t, err)

	assert.Equal(t, lb.LoadBalancerID.String(), "")
	assert.Equal(t, lb.LbType, loadbalancer.TypeNoLB)
	assert.Nil(t, lb.LbData)

	errapi := mock.DummyErrorAPI()
	errapi.Start()

	defer errapi.Close()

	errcli := lbapi.NewClient(errapi.URL)

	lb, err = loadbalancer.NewLoadBalancer(context.TODO(), zap.NewNop().Sugar(), errcli, id, adds)
	assert.Nil(t, lb)
	assert.Error(t, err)
}
