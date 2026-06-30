package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kyh0703/port-gateway/internal/domain/session"
	"github.com/kyh0703/port-gateway/internal/ports"
)

type Config struct {
	ServiceToken string
}

type Server struct {
	cfg          Config
	orchestrator ports.Orchestrator
	logger       *slog.Logger
}

func NewServer(cfg Config, orchestrator ports.Orchestrator, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		cfg:          cfg,
		orchestrator: orchestrator,
		logger:       logger,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /internal/v1/dispatches", s.dispatch)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	if s.orchestrator == nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "orchestrator unavailable"})
		return
	}

	var request dispatchRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	dispatch := request.toDomain()
	if err := s.orchestrator.HandleDispatch(r.Context(), dispatch); err != nil {
		if errors.Is(err, session.ErrInvalidDispatch) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		s.logger.Error("dispatch orchestration failed", "error", err)
		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "dispatch failed"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"conversationId": dispatch.ConversationID,
		"sessionId":      dispatch.SessionID,
		"roomId":         dispatch.RoomID,
		"status":         "accepted",
	})
}

func (s *Server) authorized(r *http.Request) bool {
	if s.cfg.ServiceToken == "" {
		return false
	}
	value := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return false
	}
	token := strings.TrimPrefix(value, prefix)
	return subtle.ConstantTimeCompare([]byte(token), []byte(s.cfg.ServiceToken)) == 1
}

type dispatchRequest struct {
	ConversationID    string                 `json:"conversationId"`
	SessionID         string                 `json:"sessionId"`
	RoomID            string                 `json:"roomId"`
	MediaSignalingURL string                 `json:"mediaSignalingUrl"`
	ServiceTokens     serviceTokensRequest   `json:"serviceTokens"`
	Capabilities      capabilitiesRequest    `json:"capabilities"`
	Recording         recordingConfigRequest `json:"recording"`
}

type serviceTokensRequest struct {
	Agent    participantTokenRequest `json:"agent"`
	Recorder participantTokenRequest `json:"recorder"`
}

type participantTokenRequest struct {
	ParticipantID string `json:"participantId"`
	Token         string `json:"participantToken"`
}

type capabilitiesRequest struct {
	AttachAgent    bool `json:"attachAgent"`
	StartRecording bool `json:"startRecording"`
}

type recordingConfigRequest struct {
	Enabled bool `json:"enabled"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (r dispatchRequest) toDomain() session.Dispatch {
	return session.Dispatch{
		ConversationID:    r.ConversationID,
		SessionID:         r.SessionID,
		RoomID:            r.RoomID,
		MediaSignalingURL: r.MediaSignalingURL,
		Agent: session.ParticipantToken{
			ParticipantID: r.ServiceTokens.Agent.ParticipantID,
			Token:         r.ServiceTokens.Agent.Token,
		},
		Recorder: session.ParticipantToken{
			ParticipantID: r.ServiceTokens.Recorder.ParticipantID,
			Token:         r.ServiceTokens.Recorder.Token,
		},
		Capabilities: session.Capabilities{
			AttachAgent:    r.Capabilities.AttachAgent,
			StartRecording: r.Capabilities.StartRecording,
		},
		Recording: session.RecordingRequest{
			Enabled: r.Recording.Enabled,
		},
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
