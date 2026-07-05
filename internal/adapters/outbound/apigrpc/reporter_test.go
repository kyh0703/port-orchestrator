package apigrpc

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/kyh0703/port-contracts/gen/go/port/api/v1"
	"github.com/kyh0703/port-orchestrator/internal/domain/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestReporterCallsGatewayEventContractWithoutInternalAuth(t *testing.T) {
	client := &capturingClient{}
	reporter := NewReporter(ReporterConfig{MaxAttempts: 1}, client)

	err := reporter.Report(context.Background(), session.LifecycleEvent{
		EventID:        "event_1",
		Type:           session.EventAgentStarted,
		ConversationID: "conv_1",
		SessionID:      "sess_1",
		RoomID:         "room_1",
		ParticipantID:  "agent_1",
		OccurredAt:     time.Unix(10, 0),
		Payload:        map[string]string{"key": "value"},
	})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	if client.request.GetEventType() != apiv1.GatewayLifecycleEventType_GATEWAY_LIFECYCLE_EVENT_TYPE_AGENT_STARTED {
		t.Fatalf("event type = %v", client.request.GetEventType())
	}
	if client.request.GetConversationId() != "conv_1" {
		t.Fatalf("conversation id = %q", client.request.GetConversationId())
	}
	if client.request.GetPayload()["participantId"] != "agent_1" {
		t.Fatalf("participantId payload = %q", client.request.GetPayload()["participantId"])
	}
	if len(client.metadata) != 0 {
		t.Fatalf("metadata = %v, want no internal auth metadata", client.metadata)
	}
}

func TestReporterRetriesTransientGRPCFailure(t *testing.T) {
	client := &capturingClient{
		errs: []error{status.Error(codes.Unavailable, "try again"), nil},
	}
	reporter := NewReporter(ReporterConfig{
		MaxAttempts: 2,
		RetryDelay:  time.Nanosecond,
	}, client)

	if err := reporter.Report(context.Background(), testEvent()); err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if client.attempts != 2 {
		t.Fatalf("attempts = %d, want 2", client.attempts)
	}
}

func TestReporterDoesNotRetryInvalidArgument(t *testing.T) {
	client := &capturingClient{
		errs: []error{status.Error(codes.InvalidArgument, "bad event")},
	}
	reporter := NewReporter(ReporterConfig{
		MaxAttempts: 3,
		RetryDelay:  time.Nanosecond,
	}, client)

	err := reporter.Report(context.Background(), testEvent())
	if err == nil {
		t.Fatal("expected error")
	}
	if client.attempts != 1 {
		t.Fatalf("attempts = %d, want 1", client.attempts)
	}
}

func TestReporterAppliesPerAttemptTimeout(t *testing.T) {
	client := &capturingClient{blockUntilDone: true}
	reporter := NewReporter(ReporterConfig{
		MaxAttempts: 1,
		Timeout:     5 * time.Millisecond,
	}, client)

	startedAt := time.Now()
	err := reporter.Report(context.Background(), testEvent())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(startedAt); elapsed > 80*time.Millisecond {
		t.Fatalf("Report() took %s, expected reporter timeout", elapsed)
	}
}

type capturingClient struct {
	request        *apiv1.RecordGatewayEventRequest
	metadata       metadata.MD
	errs           []error
	attempts       int
	blockUntilDone bool
}

func (c *capturingClient) RecordGatewayEvent(ctx context.Context, request *apiv1.RecordGatewayEventRequest, _ ...grpc.CallOption) (*apiv1.RecordGatewayEventResponse, error) {
	c.attempts++
	c.request = request
	c.metadata, _ = metadata.FromOutgoingContext(ctx)
	if c.blockUntilDone {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if len(c.errs) >= c.attempts {
		return nil, c.errs[c.attempts-1]
	}
	return &apiv1.RecordGatewayEventResponse{}, nil
}

func testEvent() session.LifecycleEvent {
	return session.LifecycleEvent{
		EventID:        "event_1",
		Type:           session.EventAgentStarted,
		ConversationID: "conv_1",
		SessionID:      "sess_1",
		RoomID:         "room_1",
		OccurredAt:     time.Unix(10, 0),
	}
}
