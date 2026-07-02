package stub

import (
	"context"
	"log/slog"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
)

type MediaConnector struct {
	logger *slog.Logger
}

func NewMediaConnector(logger *slog.Logger) *MediaConnector {
	if logger == nil {
		logger = slog.Default()
	}
	return &MediaConnector{logger: logger}
}

func (m *MediaConnector) JoinParticipant(_ context.Context, join session.ParticipantJoin) error {
	m.logger.Info(
		"media participant join requested",
		"conversation_id", join.ConversationID,
		"session_id", join.SessionID,
		"room_id", join.RoomID,
		"participant_id", join.ParticipantID,
		"role", join.Role,
		"media_signaling_url", join.MediaSignalingURL,
	)
	return nil
}
