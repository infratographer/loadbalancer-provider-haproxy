package server

import (
	"context"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.infratographer.com/loadbalancer-manager-haproxy/pkg/lbapi"
	"go.infratographer.com/x/echox"
	"go.infratographer.com/x/events"
	"go.uber.org/zap"
)

// Server holds options for server connectivity and settings
type Server struct {
	APIClient        *lbapi.Client
	Context          context.Context
	Debug            bool
	Echo             *echox.Server
	Locations        []string
	Logger           *zap.SugaredLogger
	SubscriberConfig events.SubscriberConfig
	Topics           []string

	ChangeChannels []<-chan *message.Message
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
		go s.ProcessChange(ch)
	}

	return nil
}

func (s *Server) ConfigureSubscribers() error {
	var ch []<-chan *message.Message

	for _, topic := range s.Topics {
		s.Logger.Debugw("subscribing to topic", "topic", topic)

		csub, err := events.NewSubscriber(s.SubscriberConfig)
		if err != nil {
			s.Logger.Errorw("unable to create change subscriber", "error", err, "topic", topic)
			return errSubscriberCreate
		}

		c, err := csub.SubscribeChanges(s.Context, topic)
		if err != nil {
			s.Logger.Errorw("unable to subscribe to change topic", "error", err, "topic", topic, "type", "change")
			return errSubscriptionCreate
		}

		ch = append(ch, c)
	}

	s.ChangeChannels = ch

	return nil
}
