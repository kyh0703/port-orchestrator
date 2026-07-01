package recordhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyh0703/port-gateway/internal/domain/session"
)

func TestRecorderStartCallsPortRecordStart(t *testing.T) {
	var gotAuth string
	var gotBody finalizeRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/recordings/start" {
			t.Fatalf("path = %q, want /api/recordings/start", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	recorder := NewRecorder(RecorderConfig{
		BaseURL:      server.URL,
		ServiceToken: "record-token",
	}, server.Client())

	err := recorder.Start(context.Background(), session.RecordingStart{
		ConversationID: "conv_1",
		SessionID:      "sess_1",
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if gotAuth != "Bearer record-token" {
		t.Fatalf("Authorization = %q, want Bearer record-token", gotAuth)
	}
	if gotBody.AccountID != "conv_1" || gotBody.CallID != "sess_1" {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestRecorderStartReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	recorder := NewRecorder(RecorderConfig{
		BaseURL:      server.URL,
		ServiceToken: "wrong-token",
	}, server.Client())

	err := recorder.Start(context.Background(), session.RecordingStart{
		ConversationID: "conv_1",
		SessionID:      "sess_1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
