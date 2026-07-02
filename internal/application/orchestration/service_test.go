package orchestration

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
)

func TestHandleDispatchReportsAgentAndRecordingStarted(t *testing.T) {
	reporter := &recordingReporter{}
	service := NewService(Dependencies{
		Media:    &recordingMedia{},
		Agent:    &recordingAgent{},
		Recorder: &recordingRecorder{},
		Reporter: reporter,
		Clock: func() time.Time {
			return time.Unix(10, 0)
		},
	})

	err := service.HandleDispatch(context.Background(), validDispatch())
	if err != nil {
		t.Fatalf("HandleDispatch() error = %v", err)
	}

	eventTypes := reporter.eventTypes()
	expected := []session.EventType{session.EventAgentStarted, session.EventRecordingStarted}
	if !sameEventSet(eventTypes, expected) {
		t.Fatalf("event types = %v, want %v", eventTypes, expected)
	}
}

func TestHandleDispatchPassesMediaContextToAgentAndRecorder(t *testing.T) {
	agent := &recordingAgent{}
	recorder := &recordingRecorder{}
	service := NewService(Dependencies{
		Media:    &recordingMedia{},
		Agent:    agent,
		Recorder: recorder,
		Reporter: &recordingReporter{},
		Clock: func() time.Time {
			return time.Unix(10, 0)
		},
	})

	err := service.HandleDispatch(context.Background(), validDispatch())
	if err != nil {
		t.Fatalf("HandleDispatch() error = %v", err)
	}

	if agent.attachment.MediaSignalingURL != "ws://media" {
		t.Fatalf("agent media signaling url = %q, want ws://media", agent.attachment.MediaSignalingURL)
	}
	if agent.attachment.ParticipantToken != "agent_token" {
		t.Fatalf("agent participant token = %q, want agent_token", agent.attachment.ParticipantToken)
	}
	if recorder.start.MediaSignalingURL != "ws://media" {
		t.Fatalf("recorder media signaling url = %q, want ws://media", recorder.start.MediaSignalingURL)
	}
	if recorder.start.ParticipantToken != "recorder_token" {
		t.Fatalf("recorder participant token = %q, want recorder_token", recorder.start.ParticipantToken)
	}
}

func TestHandleDispatchReportsAgentFailure(t *testing.T) {
	reporter := &recordingReporter{}
	service := NewService(Dependencies{
		Media: &recordingMedia{
			err: errors.New("media down"),
		},
		Agent:    &recordingAgent{},
		Recorder: &recordingRecorder{},
		Reporter: reporter,
		Clock: func() time.Time {
			return time.Unix(10, 0)
		},
	})

	err := service.HandleDispatch(context.Background(), validDispatch())
	if err == nil {
		t.Fatal("expected error")
	}

	if !containsEventType(reporter.eventTypes(), session.EventAgentFailed) {
		t.Fatalf("events = %v, want agent.failed", reporter.eventTypes())
	}
}

func TestEventIDIsStableForSameLifecycleEvent(t *testing.T) {
	service := NewService(Dependencies{
		Clock: func() time.Time {
			return time.Unix(10, 0)
		},
	})
	dispatch := validDispatch()

	first := service.eventID(dispatch, session.EventAgentStarted, "agent_1")
	second := service.eventID(dispatch, session.EventAgentStarted, "agent_1")

	if first != second {
		t.Fatalf("event ids differ: %q != %q", first, second)
	}
}

type recordingMedia struct {
	err error
}

func (m *recordingMedia) JoinParticipant(context.Context, session.ParticipantJoin) error {
	return m.err
}

type recordingAgent struct {
	err        error
	attachment session.AgentAttachment
}

func (a *recordingAgent) Attach(_ context.Context, attachment session.AgentAttachment) error {
	a.attachment = attachment
	return a.err
}

type recordingRecorder struct {
	err   error
	start session.RecordingStart
}

func (r *recordingRecorder) Start(_ context.Context, start session.RecordingStart) error {
	r.start = start
	return r.err
}

type recordingReporter struct {
	mu     sync.Mutex
	events []session.LifecycleEvent
	err    error
}

func (r *recordingReporter) Report(_ context.Context, event session.LifecycleEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return r.err
}

func (r *recordingReporter) eventTypes() []session.EventType {
	r.mu.Lock()
	defer r.mu.Unlock()
	types := make([]session.EventType, 0, len(r.events))
	for _, event := range r.events {
		types = append(types, event.Type)
	}
	return types
}

func validDispatch() session.Dispatch {
	return session.Dispatch{
		ConversationID:    "conv_1",
		SessionID:         "sess_1",
		RoomID:            "room_1",
		MediaSignalingURL: "ws://media",
		Agent: session.ParticipantToken{
			ParticipantID: "agent_1",
			Token:         "agent_token",
		},
		Recorder: session.ParticipantToken{
			ParticipantID: "recorder_1",
			Token:         "recorder_token",
		},
		Capabilities: session.Capabilities{
			AttachAgent:    true,
			StartRecording: true,
		},
		Recording: session.RecordingRequest{
			Enabled: true,
		},
	}
}

func sameEventSet(got, want []session.EventType) bool {
	if len(got) != len(want) {
		return false
	}
	gotCounts := map[session.EventType]int{}
	wantCounts := map[session.EventType]int{}
	for _, eventType := range got {
		gotCounts[eventType]++
	}
	for _, eventType := range want {
		wantCounts[eventType]++
	}
	return reflect.DeepEqual(gotCounts, wantCounts)
}

func containsEventType(types []session.EventType, want session.EventType) bool {
	for _, eventType := range types {
		if eventType == want {
			return true
		}
	}
	return false
}
