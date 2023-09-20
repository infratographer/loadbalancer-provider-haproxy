package server

import (
	"context"

	lbapi "go.infratographer.com/load-balancer-api/pkg/client"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.uber.org/zap"

	"go.infratographer.com/ipam-api/pkg/ipamclient"
)

// instrumentationName is a unique package name used for tracing
const instrumentationName = "go.infratographer.com/loadbalancer-provider-haproxy/server"

// Server holds options for server connectivity and settings
type Server struct {
	APIClient        *lbapi.Client
	IPAMClient       *ipamclient.Client
	Context          context.Context
	Debug            bool
	Echo             *echox.Server
	IPBlock          string
	Locations        []string
	Logger           *zap.SugaredLogger
	Publisher        *events.Publisher
	EventsConnection events.Connection
	ChangeTopics     []string

	ChangeChannels []<-chan events.Message[events.ChangeMessage]
}

// Run will start the server queue connections and healthcheck endpoints
func (s *Server) Run(ctx context.Context) error {
	go func() {
		if err := s.Echo.Run(); err != nil {
			s.Logger.Error("unable to start healthcheck server", zap.Error(err))
		}
	}()

	s.Logger.Infow("starting subscribers")

	if err := s.ConfigureSubscribers(); err != nil {
		s.Logger.Errorw("unable to configure subscribers", "error", err)
		return err
	}

	for _, ch := range s.ChangeChannels {
		go s.ListenChanges(ch)
	}

	return nil
}

func (s *Server) ConfigureSubscribers() error {
	var ch []<-chan events.Message[events.ChangeMessage]

	for _, topic := range s.ChangeTopics {
		s.Logger.Debugw("subscribing to topic", "topic", topic)

		c, err := s.EventsConnection.SubscribeChanges(s.Context, topic)
		if err != nil {
			s.Logger.Errorw("unable to subscribe to change topic", "error", err, "topic", topic, "type", "change")
			return errSubscriptionCreate
		}

		ch = append(ch, c)
	}

	s.ChangeChannels = ch

	return nil
}
