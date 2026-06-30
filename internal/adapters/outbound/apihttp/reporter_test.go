package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyh0703/port-gateway/internal/domain/session"
)

func TestReporterRetriesTransientFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.Header.Get("Authorization") != "Bearer api-secret" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	reporter := NewReporter(ReporterConfig{
		BaseURL:      server.URL,
		EventPath:    "/events",
		ServiceToken: "api-secret",
		MaxAttempts:  2,
		RetryDelay:   time.Nanosecond,
	}, server.Client())

	err := reporter.Report(context.Background(), session.LifecycleEvent{
		EventID:        "event_1",
		Type:           session.EventAgentStarted,
		ConversationID: "conv_1",
		SessionID:      "sess_1",
		RoomID:         "room_1",
		OccurredAt:     time.Unix(10, 0),
	})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestReporterDoesNotRetryClientError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	reporter := NewReporter(ReporterConfig{
		BaseURL:     server.URL,
		EventPath:   "/events",
		MaxAttempts: 3,
		RetryDelay:  time.Nanosecond,
	}, server.Client())

	err := reporter.Report(context.Background(), testEvent())
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestReporterHonorsClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := server.Client()
	client.Timeout = 5 * time.Millisecond
	reporter := NewReporter(ReporterConfig{
		BaseURL:     server.URL,
		EventPath:   "/events",
		MaxAttempts: 1,
	}, client)

	startedAt := time.Now()
	err := reporter.Report(context.Background(), testEvent())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(startedAt); elapsed > 80*time.Millisecond {
		t.Fatalf("Report() took %s, expected client timeout", elapsed)
	}
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
