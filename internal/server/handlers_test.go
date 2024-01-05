package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.infratographer.com/ipam-api/pkg/ipamclient"
	lbapi "go.infratographer.com/load-balancer-api/pkg/client"

	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"
	"go.uber.org/zap"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/server"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/testutils/mock"
)

func TestLocationCheck(t *testing.T) {
	lb, _ := gidx.Parse("testloc-abcd1234")

	srv := server.Server{
		Locations: []string{"abcd1234", "defg5678"},
	}

	check := srv.LocationCheck(lb)
	assert.Equal(t, true, check)

	lb, _ = gidx.Parse("testloc-efgh5678")
	check = srv.LocationCheck(lb)
	assert.Equal(t, false, check)
}

func TestProcessChange(t *testing.T) { //nolint:govet
	id := gidx.MustNewID("loadbal")

	api := mock.DummyAPI(id.String())
	api.Start()

	ipamapi := mock.DummyIPAMAPI(id.String())
	ipamapi.Start()

	eSrv, err := echox.NewServer(zap.NewNop(), echox.Config{}, nil)
	if err != nil {
		errPanic("unable to initialize echo server", err)
	}

	loc, err := gidx.Parse("testloc-abcd1234")
	if err != nil {
		errPanic("unable to parse location", err)
	}

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
		APIClient:        lbapi.NewClient(api.URL),
		IPAMClient:       ipamclient.New(ipamapi.URL),
		Context:          context.TODO(),
		Echo:             eSrv,
		EventsConnection: conn,
		Locations:        []string{"abcd1234"},
		Logger:           zap.NewNop().Sugar(),
		ChangeTopics:     []string{"*.load-balancer"},
	}

	// TODO: check that namespace does not exist
	// TODO: check that release does not exist

	// publish a message to the change channel
	_, err = srv.EventsConnection.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.CreateChangeType),
		SubjectID:            id,
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
	})
	if err != nil {
		errPanic("unable to publish change", err)
	}

	err = srv.ConfigureSubscribers()
	if err != nil {
		errPanic("unable to configure subscribers", err)
	}

	go srv.ListenChanges(srv.ChangeChannels[0])

	_, err = srv.EventsConnection.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.UpdateChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
		SubjectID:            id,
	})
	if err != nil {
		errPanic("unable to publish change", err)
	}

	// TODO: check that namespace exists
	// TODO: check that release exists
	// TODO: verify some update, maybe with values file

	_, err = srv.EventsConnection.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.UpdateChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{id, loc},
		SubjectID:            gidx.MustNewID("loadprt"),
	})
	if err != nil {
		errPanic("unable to publish change", err)
	}

	// TODO: verify some update exists

	_, err = srv.EventsConnection.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.DeleteChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
		SubjectID:            id,
	})
	if err != nil {
		errPanic("unable to publish change", err)
	}
}
