package orchestration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
	"github.com/kyh0703/port-orchestrator/internal/ports"
)

type Clock func() time.Time

type Dependencies struct {
	Media    ports.MediaConnector
	Agent    ports.AgentAttacher
	Recorder ports.Recorder
	Reporter ports.EventReporter
	Clock    Clock
}

type Service struct {
	media    ports.MediaConnector
	agent    ports.AgentAttacher
	recorder ports.Recorder
	reporter ports.EventReporter
	clock    Clock
}

func NewService(deps Dependencies) *Service {
	clock := deps.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Service{
		media:    deps.Media,
		agent:    deps.Agent,
		recorder: deps.Recorder,
		reporter: deps.Reporter,
		clock:    clock,
	}
}

func (s *Service) HandleDispatch(ctx context.Context, dispatch session.Dispatch) error {
	if err := dispatch.Validate(); err != nil {
		return err
	}
	if s.reporter == nil {
		return errors.New("orchestration reporter is nil")
	}

	var tasks []func(context.Context) error
	if dispatch.Capabilities.AttachAgent {
		tasks = append(tasks, func(ctx context.Context) error {
			return s.attachAgent(ctx, dispatch)
		})
	}
	if dispatch.Recording.Enabled || dispatch.Capabilities.StartRecording {
		tasks = append(tasks, func(ctx context.Context) error {
			return s.startRecording(ctx, dispatch)
		})
	}
	if len(tasks) == 0 {
		return nil
	}

	errCh := make(chan error, len(tasks))
	var wg sync.WaitGroup
	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task(ctx); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (s *Service) attachAgent(ctx context.Context, dispatch session.Dispatch) error {
	if s.media == nil {
		return errors.New("media connector is nil")
	}
	if s.agent == nil {
		return errors.New("agent attacher is nil")
	}

	if err := s.media.JoinParticipant(ctx, dispatch.AgentJoin()); err != nil {
		return s.reportAgentFailure(ctx, dispatch, err)
	}

	attachment := session.AgentAttachment{
		ConversationID:    dispatch.ConversationID,
		SessionID:         dispatch.SessionID,
		RoomID:            dispatch.RoomID,
		MediaSignalingURL: dispatch.MediaSignalingURL,
		ParticipantID:     dispatch.Agent.ParticipantID,
		ParticipantToken:  dispatch.Agent.Token,
	}
	if err := s.agent.Attach(ctx, attachment); err != nil {
		return s.reportAgentFailure(ctx, dispatch, err)
	}

	event := session.NewLifecycleEvent(
		s.eventID(dispatch, session.EventAgentStarted, dispatch.Agent.ParticipantID),
		session.EventAgentStarted,
		dispatch,
		dispatch.Agent.ParticipantID,
		s.clock(),
		nil,
	)
	if err := s.reporter.Report(ctx, event); err != nil {
		return fmt.Errorf("report agent started: %w", err)
	}
	return nil
}

func (s *Service) reportAgentFailure(ctx context.Context, dispatch session.Dispatch, cause error) error {
	event := session.NewLifecycleEvent(
		s.eventID(dispatch, session.EventAgentFailed, dispatch.Agent.ParticipantID),
		session.EventAgentFailed,
		dispatch,
		dispatch.Agent.ParticipantID,
		s.clock(),
		map[string]string{"reason": cause.Error()},
	)
	if err := s.reporter.Report(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("agent attach: %w", cause), fmt.Errorf("report agent failed: %w", err))
	}
	return fmt.Errorf("agent attach: %w", cause)
}

func (s *Service) startRecording(ctx context.Context, dispatch session.Dispatch) error {
	if s.media == nil {
		return errors.New("media connector is nil")
	}
	if s.recorder == nil {
		return errors.New("recorder is nil")
	}

	if err := s.media.JoinParticipant(ctx, dispatch.RecorderJoin()); err != nil {
		return s.reportRecordingFailure(ctx, dispatch, err)
	}

	start := session.RecordingStart{
		ConversationID:    dispatch.ConversationID,
		SessionID:         dispatch.SessionID,
		RoomID:            dispatch.RoomID,
		MediaSignalingURL: dispatch.MediaSignalingURL,
		ParticipantID:     dispatch.Recorder.ParticipantID,
		ParticipantToken:  dispatch.Recorder.Token,
	}
	if err := s.recorder.Start(ctx, start); err != nil {
		return s.reportRecordingFailure(ctx, dispatch, err)
	}

	event := session.NewLifecycleEvent(
		s.eventID(dispatch, session.EventRecordingStarted, dispatch.Recorder.ParticipantID),
		session.EventRecordingStarted,
		dispatch,
		dispatch.Recorder.ParticipantID,
		s.clock(),
		nil,
	)
	if err := s.reporter.Report(ctx, event); err != nil {
		return fmt.Errorf("report recording started: %w", err)
	}
	return nil
}

func (s *Service) reportRecordingFailure(ctx context.Context, dispatch session.Dispatch, cause error) error {
	event := session.NewLifecycleEvent(
		s.eventID(dispatch, session.EventRecordingFailed, dispatch.Recorder.ParticipantID),
		session.EventRecordingFailed,
		dispatch,
		dispatch.Recorder.ParticipantID,
		s.clock(),
		map[string]string{"reason": cause.Error()},
	)
	if err := s.reporter.Report(ctx, event); err != nil {
		return errors.Join(fmt.Errorf("recording start: %w", cause), fmt.Errorf("report recording failed: %w", err))
	}
	return fmt.Errorf("recording start: %w", cause)
}

func (s *Service) eventID(dispatch session.Dispatch, eventType session.EventType, participantID string) string {
	return fmt.Sprintf("%s:%s:%s", dispatch.SessionID, eventType, participantID)
}
