package recordhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyh0703/port-gateway/internal/domain/session"
)

type RecorderConfig struct {
	BaseURL      string
	StartPath    string
	ServiceToken string
}

type Recorder struct {
	cfg    RecorderConfig
	client *http.Client
}

func NewRecorder(cfg RecorderConfig, client *http.Client) *Recorder {
	if cfg.StartPath == "" {
		cfg.StartPath = "/api/recordings/start"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Recorder{cfg: cfg, client: client}
}

func (r *Recorder) Start(ctx context.Context, start session.RecordingStart) error {
	if strings.TrimSpace(r.cfg.BaseURL) == "" {
		return errors.New("port-record base url is empty")
	}

	body, err := json.Marshal(finalizeRequest{
		AccountID: start.ConversationID,
		CallID:    start.SessionID,
	})
	if err != nil {
		return fmt.Errorf("marshal recording start request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create recording start request: %w", err)
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
		return fmt.Errorf("post recording start: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("port-record returned status %d", resp.StatusCode)
}

func (r *Recorder) endpoint() string {
	return strings.TrimRight(r.cfg.BaseURL, "/") + "/" + strings.TrimLeft(r.cfg.StartPath, "/")
}

type finalizeRequest struct {
	AccountID string `json:"account_id"`
	CallID    string `json:"call_id"`
}
