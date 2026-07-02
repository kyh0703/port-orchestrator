package stub

import (
	"context"
	"log/slog"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
)

type Recorder struct {
	logger *slog.Logger
}

func NewRecorder(logger *slog.Logger) *Recorder {
	if logger == nil {
		logger = slog.Default()
	}
	return &Recorder{logger: logger}
}

func (r *Recorder) Start(_ context.Context, start session.RecordingStart) error {
	r.logger.Info(
		"recording start requested",
		"conversation_id", start.ConversationID,
		"session_id", start.SessionID,
		"room_id", start.RoomID,
		"participant_id", start.ParticipantID,
		"media_signaling_url", start.MediaSignalingURL,
	)
	return nil
}
