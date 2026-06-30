package ports

import (
	"context"

	"github.com/kyh0703/port-gateway/internal/domain/session"
)

type Orchestrator interface {
	HandleDispatch(ctx context.Context, dispatch session.Dispatch) error
}

type MediaConnector interface {
	JoinParticipant(ctx context.Context, join session.ParticipantJoin) error
}

type AgentAttacher interface {
	Attach(ctx context.Context, attachment session.AgentAttachment) error
}

type Recorder interface {
	Start(ctx context.Context, start session.RecordingStart) error
}

type EventReporter interface {
	Report(ctx context.Context, event session.LifecycleEvent) error
}
