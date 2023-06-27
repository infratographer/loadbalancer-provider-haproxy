package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/gidx"
	"go.uber.org/zap"

	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/server"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/testutils/mock"
)

func TestRun(t *testing.T) {
	id := gidx.MustNewID("loadbal")

	api := mock.DummyAPI(id.String())
	api.Start()

	eSrv, _ := echox.NewServer(zap.NewNop(), echox.Config{}, nil)

	srv := server.Server{
		APIClient:        lbapi.NewClient(api.URL),
		Context:          context.TODO(),
		Echo:             eSrv,
		Locations:        []string{"abcd1234"},
		Logger:           zap.NewNop().Sugar(),
		SubscriberConfig: SC,
		Topics:           []string{"*.load-balancer-run"},
	}

	err := srv.Run(srv.Context)

	assert.Nil(t, err)
}
