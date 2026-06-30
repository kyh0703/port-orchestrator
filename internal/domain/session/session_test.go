package session

import (
	"errors"
	"testing"
)

func TestDispatchValidateRequiresAgentTokenWhenAgentAttachRequested(t *testing.T) {
	dispatch := Dispatch{
		ConversationID:    "conv_1",
		SessionID:         "sess_1",
		RoomID:            "room_1",
		MediaSignalingURL: "ws://media",
		Capabilities: Capabilities{
			AttachAgent: true,
		},
	}

	err := dispatch.Validate()
	if !errors.Is(err, ErrInvalidDispatch) {
		t.Fatalf("expected ErrInvalidDispatch, got %v", err)
	}
}

func TestDispatchValidateAcceptsRequestedAgentAndRecordingParticipants(t *testing.T) {
	dispatch := Dispatch{
		ConversationID:    "conv_1",
		SessionID:         "sess_1",
		RoomID:            "room_1",
		MediaSignalingURL: "ws://media",
		Agent: ParticipantToken{
			ParticipantID: "agent_1",
			Token:         "agent_token",
		},
		Recorder: ParticipantToken{
			ParticipantID: "recorder_1",
			Token:         "recorder_token",
		},
		Capabilities: Capabilities{
			AttachAgent:    true,
			StartRecording: true,
		},
		Recording: RecordingRequest{
			Enabled: true,
		},
	}

	if err := dispatch.Validate(); err != nil {
		t.Fatalf("expected valid dispatch, got %v", err)
	}
}
