package stub

import (
	"context"
	"log/slog"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
)

type AgentAttacher struct {
	logger *slog.Logger
}

func NewAgentAttacher(logger *slog.Logger) *AgentAttacher {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentAttacher{logger: logger}
}

func (a *AgentAttacher) Attach(_ context.Context, attachment session.AgentAttachment) error {
	a.logger.Info(
		"agent attach requested",
		"conversation_id", attachment.ConversationID,
		"session_id", attachment.SessionID,
		"room_id", attachment.RoomID,
		"participant_id", attachment.ParticipantID,
		"media_signaling_url", attachment.MediaSignalingURL,
	)
	return nil
}
