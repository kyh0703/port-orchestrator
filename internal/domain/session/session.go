package session

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidDispatch = errors.New("invalid dispatch")

type ParticipantRole string

const (
	RoleAgent    ParticipantRole = "agent"
	RoleRecorder ParticipantRole = "recorder"
	RoleSIP      ParticipantRole = "sip"
	RoleService  ParticipantRole = "service"
)

type EventType string

const (
	EventAgentStarted       EventType = "agent.started"
	EventAgentFailed        EventType = "agent.failed"
	EventRecordingStarted   EventType = "recording.started"
	EventRecordingCompleted EventType = "recording.completed"
	EventRecordingFailed    EventType = "recording.failed"
	EventSIPConnected       EventType = "sip.connected"
	EventSIPEnded           EventType = "sip.ended"
)

type Dispatch struct {
	ConversationID    string
	SessionID         string
	RoomID            string
	MediaSignalingURL string
	Agent             ParticipantToken
	Recorder          ParticipantToken
	Capabilities      Capabilities
	Recording         RecordingRequest
}

type ParticipantToken struct {
	ParticipantID string
	Token         string
}

type Capabilities struct {
	AttachAgent    bool
	StartRecording bool
}

type RecordingRequest struct {
	Enabled bool
}

type ParticipantJoin struct {
	ConversationID    string
	SessionID         string
	RoomID            string
	MediaSignalingURL string
	ParticipantID     string
	ParticipantToken  string
	Role              ParticipantRole
}

type AgentAttachment struct {
	ConversationID string
	SessionID      string
	RoomID         string
	ParticipantID  string
}

type RecordingStart struct {
	ConversationID string
	SessionID      string
	RoomID         string
	ParticipantID  string
}

type LifecycleEvent struct {
	EventID        string
	Type           EventType
	ConversationID string
	SessionID      string
	RoomID         string
	ParticipantID  string
	OccurredAt     time.Time
	Payload        map[string]string
}

func (d Dispatch) Validate() error {
	var missing []string
	if strings.TrimSpace(d.ConversationID) == "" {
		missing = append(missing, "conversationId")
	}
	if strings.TrimSpace(d.SessionID) == "" {
		missing = append(missing, "sessionId")
	}
	if strings.TrimSpace(d.RoomID) == "" {
		missing = append(missing, "roomId")
	}
	if strings.TrimSpace(d.MediaSignalingURL) == "" {
		missing = append(missing, "mediaSignalingUrl")
	}
	if d.Capabilities.AttachAgent {
		if strings.TrimSpace(d.Agent.ParticipantID) == "" {
			missing = append(missing, "agent.participantId")
		}
		if strings.TrimSpace(d.Agent.Token) == "" {
			missing = append(missing, "agent.participantToken")
		}
	}
	if d.Recording.Enabled || d.Capabilities.StartRecording {
		if strings.TrimSpace(d.Recorder.ParticipantID) == "" {
			missing = append(missing, "recorder.participantId")
		}
		if strings.TrimSpace(d.Recorder.Token) == "" {
			missing = append(missing, "recorder.participantToken")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %s", ErrInvalidDispatch, strings.Join(missing, ", "))
	}
	return nil
}

func (d Dispatch) AgentJoin() ParticipantJoin {
	return ParticipantJoin{
		ConversationID:    d.ConversationID,
		SessionID:         d.SessionID,
		RoomID:            d.RoomID,
		MediaSignalingURL: d.MediaSignalingURL,
		ParticipantID:     d.Agent.ParticipantID,
		ParticipantToken:  d.Agent.Token,
		Role:              RoleAgent,
	}
}

func (d Dispatch) RecorderJoin() ParticipantJoin {
	return ParticipantJoin{
		ConversationID:    d.ConversationID,
		SessionID:         d.SessionID,
		RoomID:            d.RoomID,
		MediaSignalingURL: d.MediaSignalingURL,
		ParticipantID:     d.Recorder.ParticipantID,
		ParticipantToken:  d.Recorder.Token,
		Role:              RoleRecorder,
	}
}

func NewLifecycleEvent(eventID string, eventType EventType, dispatch Dispatch, participantID string, occurredAt time.Time, payload map[string]string) LifecycleEvent {
	return LifecycleEvent{
		EventID:        eventID,
		Type:           eventType,
		ConversationID: dispatch.ConversationID,
		SessionID:      dispatch.SessionID,
		RoomID:         dispatch.RoomID,
		ParticipantID:  participantID,
		OccurredAt:     occurredAt.UTC(),
		Payload:        payload,
	}
}
