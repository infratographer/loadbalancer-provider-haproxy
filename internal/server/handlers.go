package server

import (
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/exp/slices"

	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"

	lbapi "go.infratographer.com/load-balancer-api/pkg/client"

	"go.infratographer.com/loadbalancer-provider-haproxy/internal/loadbalancer"
)

var (
	defaultNakDelay = time.Second * 5
)

func (s *Server) ListenChanges(messages <-chan events.Message[events.ChangeMessage]) {
	for msg := range messages {
		slogger := s.Logger.With(
			"event.message.id", msg.ID(),
			"event.message.topic", msg.Topic(),
			"event.message.source", msg.Source(),
			"event.message.timestamp", msg.Timestamp(),
			"event.message.deliveries", msg.Deliveries(),
		)

		if err := s.processChange(msg); err != nil {
			if s.MaxProcessMsgAttempts != 0 && msg.Deliveries()+1 > s.MaxProcessMsgAttempts {
				slogger.Warnw("terminating event, too many attempts")

				if termErr := msg.Term(); termErr != nil {
					slogger.Warnw("error occurred while terminating event")
				}
			} else if nakErr := msg.Nak(defaultNakDelay); nakErr != nil {
				slogger.Warnw("error occurred while naking", "error", nakErr)
			}
		} else if ackErr := msg.Ack(); ackErr != nil {
			slogger.Warnw("error occurred while acking", "error", ackErr)
		}
	}
}

func (s *Server) processChange(msg events.Message[events.ChangeMessage]) error {
	var lb *loadbalancer.LoadBalancer

	var err error

	m := msg.Message()

	ctx, span := otel.Tracer(instrumentationName).Start(m.GetTraceContext(s.Context), "processChange")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	if slices.ContainsFunc(m.AdditionalSubjectIDs, s.LocationCheck) || len(s.Locations) == 0 {
		if m.EventType != string(events.DeleteChangeType) {
			lb, err = loadbalancer.NewLoadBalancer(ctx, s.Logger, s.APIClient, m.SubjectID, m.AdditionalSubjectIDs)
			if err != nil {
				s.Logger.Errorw("unable to initialize loadbalancer", "error", err, "messageID", msg.ID(), "message", m)
				return err
			}
		} else {
			// on delete event, we need to get the location id from additional subjects
			loc := s.GetLocation(m.AdditionalSubjectIDs)

			lb = &loadbalancer.LoadBalancer{
				LoadBalancerID: m.SubjectID,
				LbType:         loadbalancer.TypeLB,
				LbData: &lbapi.LoadBalancer{
					Location: lbapi.LocationNode{
						ID: loc.String(),
					},
				},
			}
		}

		if err == nil && lb != nil && lb.LbType != loadbalancer.TypeNoLB {
			span.SetAttributes(
				attribute.String("loadbalancer.id", lb.LoadBalancerID.String()),
				attribute.String("message.event", m.EventType),
				attribute.String("message.id", msg.ID()),
				attribute.String("message.subject", m.SubjectID.String()),
			)

			switch {
			case m.EventType == string(events.CreateChangeType) && lb.LbType == loadbalancer.TypeLB:
				s.Logger.Debugw("requesting address for loadbalancer", "loadbalancer", lb.LoadBalancerID.String())

				if err := s.processLoadBalancerChangeCreate(ctx, lb); err != nil {
					s.Logger.Errorw("handler unable to request address for loadbalancer", "error", err, "loadbalancer", lb.LoadBalancerID.String())
					return err
				}
			case m.EventType == string(events.DeleteChangeType) && lb.LbType == loadbalancer.TypeLB:
				s.Logger.Debugw("releasing address from loadbalancer", "loadbalancer", lb.LoadBalancerID.String())

				if err := s.processLoadBalancerChangeDelete(ctx, lb); err != nil {
					s.Logger.Errorw("handler unable to release address from loadbalancer", "error", err, "loadbalancer", lb.LoadBalancerID.String())
					return err
				}
			default:
				s.Logger.Debugw("Ignoring event", "loadbalancer", lb.LoadBalancerID.String(), "message", m)
			}
		}
	}

	return nil
}

func (s *Server) LocationCheck(i gidx.PrefixedID) bool {
	for _, s := range s.Locations {
		if strings.HasSuffix(i.String(), s) {
			return true
		}
	}

	return false
}

func (s *Server) GetLocation(subjs []gidx.PrefixedID) gidx.PrefixedID {
	for _, subj := range subjs {
		if s.LocationCheck(subj) {
			return subj
		}
	}

	return gidx.PrefixedID("")
}
