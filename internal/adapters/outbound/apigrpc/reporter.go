package apigrpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiv1 "github.com/kyh0703/port-contracts/gen/go/port/api/v1"
	"github.com/kyh0703/port-orchestrator/internal/domain/session"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReporterConfig struct {
	MaxAttempts int
	RetryDelay  time.Duration
	Timeout     time.Duration
}

type Reporter struct {
	cfg    ReporterConfig
	client apiv1.ApiEventServiceClient
}

func NewReporter(cfg ReporterConfig, client apiv1.ApiEventServiceClient) *Reporter {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	return &Reporter{cfg: cfg, client: client}
}

func (r *Reporter) Report(ctx context.Context, event session.LifecycleEvent) error {
	if r.client == nil {
		return errors.New("api event client is nil")
	}

	var lastErr error
	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		err := r.send(ctx, event)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) || attempt == r.cfg.MaxAttempts {
			break
		}
		timer := time.NewTimer(r.cfg.RetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("report event %s: %w", event.EventID, lastErr)
}

func (r *Reporter) send(ctx context.Context, event session.LifecycleEvent) error {
	if r.cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.cfg.Timeout)
		defer cancel()
	}

	_, err := r.client.RecordGatewayEvent(ctx, toRequest(event))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	return nil
}

func toRequest(event session.LifecycleEvent) *apiv1.RecordGatewayEventRequest {
	payload := make(map[string]string, len(event.Payload)+1)
	for key, value := range event.Payload {
		payload[key] = value
	}
	if event.ParticipantID != "" {
		payload["participantId"] = event.ParticipantID
	}

	return &apiv1.RecordGatewayEventRequest{
		EventId:        event.EventID,
		EventType:      toProtoEventType(event.Type),
		ConversationId: event.ConversationID,
		SessionId:      event.SessionID,
		RoomId:         event.RoomID,
		OccurredAt:     timestamppb.New(event.OccurredAt),
		Payload:        payload,
	}
}

func toProtoEventType(eventType session.EventType) apiv1.GatewayLifecycleEventType {
	switch eventType {
	case session.EventAgentStarted:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_AGENT_STARTED
	case session.EventAgentFailed:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_AGENT_FAILED
	case session.EventRecordingStarted:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_RECORDING_STARTED
	case session.EventRecordingCompleted:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_RECORDING_COMPLETED
	case session.EventRecordingFailed:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_RECORDING_FAILED
	case session.EventSIPConnected:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_SIP_CONNECTED
	case session.EventSIPEnded:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_SIP_ENDED
	default:
		return apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_UNSPECIFIED
	}
}

func isRetryable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.ResourceExhausted, codes.Unavailable:
		return true
	default:
		return false
	}
}
