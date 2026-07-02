package apihttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyh0703/port-orchestrator/internal/domain/session"
)

type ReporterConfig struct {
	BaseURL      string
	EventPath    string
	ServiceToken string
	MaxAttempts  int
	RetryDelay   time.Duration
}

type Reporter struct {
	cfg    ReporterConfig
	client *http.Client
}

func NewReporter(cfg ReporterConfig, client *http.Client) *Reporter {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	if cfg.EventPath == "" {
		cfg.EventPath = "/internal/v1/gateway/events"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Reporter{cfg: cfg, client: client}
}

func (r *Reporter) Report(ctx context.Context, event session.LifecycleEvent) error {
	if strings.TrimSpace(r.cfg.BaseURL) == "" {
		return errors.New("api base url is empty")
	}

	var lastErr error
	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		err := r.send(ctx, event)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryable(err) {
			break
		}
		if attempt == r.cfg.MaxAttempts {
			break
		}
		timer := time.NewTimer(r.cfg.RetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("report event %s: %w", event.EventID, lastErr)
}

func (r *Reporter) send(ctx context.Context, event session.LifecycleEvent) error {
	payload := eventPayload{
		EventID:        event.EventID,
		Type:           string(event.Type),
		ConversationID: event.ConversationID,
		SessionID:      event.SessionID,
		RoomID:         event.RoomID,
		ParticipantID:  event.ParticipantID,
		OccurredAt:     event.OccurredAt,
		Payload:        event.Payload,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create event request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if r.cfg.ServiceToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.cfg.ServiceToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return retryableError{err: fmt.Errorf("post event: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return statusError{code: resp.StatusCode}
}

func (r *Reporter) endpoint() string {
	return strings.TrimRight(r.cfg.BaseURL, "/") + "/" + strings.TrimLeft(r.cfg.EventPath, "/")
}

type eventPayload struct {
	EventID        string            `json:"eventId"`
	Type           string            `json:"type"`
	ConversationID string            `json:"conversationId"`
	SessionID      string            `json:"sessionId"`
	RoomID         string            `json:"roomId"`
	ParticipantID  string            `json:"participantId,omitempty"`
	OccurredAt     time.Time         `json:"occurredAt"`
	Payload        map[string]string `json:"payload,omitempty"`
}

type retryableError struct {
	err error
}

func (e retryableError) Error() string {
	return e.err.Error()
}

func (e retryableError) Unwrap() error {
	return e.err
}

type statusError struct {
	code int
}

func (e statusError) Error() string {
	return fmt.Sprintf("api returned status %d", e.code)
}

func isRetryable(err error) bool {
	var retryable retryableError
	if errors.As(err, &retryable) {
		return true
	}

	var status statusError
	if errors.As(err, &status) {
		return status.code == http.StatusRequestTimeout ||
			status.code == http.StatusTooManyRequests ||
			status.code >= http.StatusInternalServerError
	}
	return false
}
