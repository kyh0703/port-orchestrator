package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyh0703/port-gateway/internal/domain/session"
)

func TestDispatchRejectsMissingServiceAuth(t *testing.T) {
	server := NewServer(Config{ServiceToken: "secret"}, &capturingOrchestrator{}, nil)
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/dispatches", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestDispatchAcceptsAuthorizedRequest(t *testing.T) {
	orchestrator := &capturingOrchestrator{}
	server := NewServer(Config{ServiceToken: "secret"}, orchestrator, nil)
	body := `{
		"conversationId":"conv_1",
		"sessionId":"sess_1",
		"roomId":"room_1",
		"mediaSignalingUrl":"ws://media",
		"serviceTokens":{
			"agent":{"participantId":"agent_1","participantToken":"agent_token"},
			"recorder":{"participantId":"recorder_1","participantToken":"recorder_token"}
		},
		"capabilities":{"attachAgent":true,"startRecording":true},
		"recording":{"enabled":true}
	}`
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/dispatches", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()

	server.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d body=%s", response.Code, http.StatusAccepted, response.Body.String())
	}
	if orchestrator.dispatch.SessionID != "sess_1" {
		t.Fatalf("session id = %q, want sess_1", orchestrator.dispatch.SessionID)
	}
	if !orchestrator.dispatch.Capabilities.AttachAgent {
		t.Fatal("expected attach agent capability")
	}
}

type capturingOrchestrator struct {
	dispatch session.Dispatch
	err      error
}

func (o *capturingOrchestrator) HandleDispatch(_ context.Context, dispatch session.Dispatch) error {
	o.dispatch = dispatch
	return o.err
}
