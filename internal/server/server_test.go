package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
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

	config := events.Config{
		NATS: events.NATSConfig{
			URL:             nats.Server.ClientURL(),
			SubscribePrefix: "com.infratographer.testing",
			PublishPrefix:   "com.infratographer.testing",
			Source:          "loadbalancerproviderhaproxy",
		},
	}

	conn, err := events.NewConnection(config)
	if err != nil {
		errPanic("unable to create connection", err)
	}

	srv := server.Server{
		APIClient:    lbapi.NewClient(api.URL),
		Context:      context.TODO(),
		Echo:         eSrv,
		Connection:   conn,
		Locations:    []string{"abcd1234"},
		Logger:       zap.NewNop().Sugar(),
		ChangeTopics: []string{"*.load-balancer"},
	}

	err = srv.Run(srv.Context)

	assert.Nil(t, err)
}
