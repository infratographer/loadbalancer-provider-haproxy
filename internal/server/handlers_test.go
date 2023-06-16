package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/server"
	"go.infratographer.com/loadbalancer-provider-haproxy/internal/testutils/mock"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"
	"go.uber.org/zap"
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

	eSrv, err := echox.NewServer(zap.NewNop(), echox.Config{}, nil)
	if err != nil {
		errPanic("unable to initialize echo server", err)
	}

	loc, err := gidx.Parse("testloc-abcd1234")
	if err != nil {
		errPanic("unable to parse location", err)
	}

	srv := server.Server{
		APIClient:        lbapi.NewClient(api.URL),
		Context:          context.TODO(),
		Echo:             eSrv,
		Locations:        []string{"abcd1234"},
		Logger:           zap.NewNop().Sugar(),
		SubscriberConfig: SC,
		Topics:           []string{"*.load-balancer"},
	}

	// TODO: check that namespace does not exist
	// TODO: check that release does not exist

	// publish a message to the change channel
	p, _ := events.NewPublisher(PC)
	_ = p.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.CreateChangeType),
		SubjectID:            id,
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
	})

	_ = srv.ConfigureSubscribers()

	go srv.ProcessChange(srv.ChangeChannels[0])

	_ = p.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.UpdateChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
		SubjectID:            id,
	})

	// TODO: check that namespace exists
	// TODO: check that release exists
	// TODO: verify some update, maybe with values file

	_ = p.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.UpdateChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{id, loc},
		SubjectID:            gidx.MustNewID("loadprt"),
	})

	//TODO: verify some update exists

	_ = p.PublishChange(context.TODO(), "load-balancer", events.ChangeMessage{
		EventType:            string(events.DeleteChangeType),
		AdditionalSubjectIDs: []gidx.PrefixedID{loc},
		SubjectID:            id,
	})
}
