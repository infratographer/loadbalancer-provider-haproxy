package server

import (
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"

	"golang.org/x/exp/slices"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/loadbalancer"
)

func (s *Server) ProcessChange(messages <-chan *message.Message) {
	for msg := range messages {
		m, err := events.UnmarshalChangeMessage(msg.Payload)
		if err != nil {
			s.Logger.Errorw("unable to unmarshal change message", "error", err, "messageID", msg.UUID, "message", msg.Payload)
			msg.Nack()
		}

		if slices.ContainsFunc(m.AdditionalSubjectIDs, s.LocationCheck) || len(s.Locations) == 0 {
			lb, err := loadbalancer.NewLoadBalancer(s.Context, s.Logger, s.APIClient, m.SubjectID, m.AdditionalSubjectIDs)
			if err != nil {
				s.Logger.Errorw("unable to initialize loadbalancer", "error", err, "messageID", msg.UUID, "message", msg.Payload)
				msg.Nack()
			}

			if lb.LbType != loadbalancer.TypeNoLB {
				switch {
				case m.EventType == string(events.CreateChangeType) && lb.LbType == loadbalancer.TypeLB:
					s.Logger.Debugw("requesting address for loadbalancer", "loadbalancer", lb.LoadBalancerID.String())

					if err := s.processLoadBalancerChangeCreate(lb); err != nil {
						s.Logger.Errorw("handler unable to request address for loadbalancer", "error", err, "loadbalancer", lb.LoadBalancerID.String())
						msg.Nack()
					}
				case m.EventType == string(events.DeleteChangeType) && lb.LbType == loadbalancer.TypeLB:
					s.Logger.Debugw("releasing address from loadbalancer", "loadbalancer", lb.LoadBalancerID.String())

					if err := s.processLoadBalancerChangeDelete(lb); err != nil {
						s.Logger.Errorw("handler unable to release address from loadbalancer", "error", err, "loadbalancer", lb.LoadBalancerID.String())
						msg.Nack()
					}
				default:
					s.Logger.Debugw("Ignoring event", "loadbalancer", lb.LoadBalancerID.String(), "message", msg.Payload)
				}
			}
		}
		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}

func (s *Server) LocationCheck(i gidx.PrefixedID) bool {
	for _, s := range s.Locations {
		if strings.HasSuffix(i.String(), s) {
			return true
		}
	}

	return false
}
