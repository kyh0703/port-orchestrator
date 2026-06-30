package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kyh0703/port-gateway/internal/adapters/inbound/httpapi"
	"github.com/kyh0703/port-gateway/internal/adapters/outbound/apihttp"
	"github.com/kyh0703/port-gateway/internal/adapters/outbound/stub"
	"github.com/kyh0703/port-gateway/internal/application/orchestration"
	"github.com/kyh0703/port-gateway/internal/config"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid gateway config", "error", err)
		os.Exit(1)
	}

	apiClient := &http.Client{
		Timeout: cfg.APIReportTimeout,
	}

	reporter := apihttp.NewReporter(apihttp.ReporterConfig{
		BaseURL:      cfg.APIBaseURL,
		EventPath:    cfg.APIEventPath,
		ServiceToken: cfg.APIServiceToken,
		MaxAttempts:  cfg.APIReportMaxAttempts,
		RetryDelay:   cfg.APIReportRetryDelay,
	}, apiClient)

	orchestrator := orchestration.NewService(orchestration.Dependencies{
		Media:    stub.NewMediaConnector(logger),
		Agent:    stub.NewAgentAttacher(logger),
		Recorder: stub.NewRecorder(logger),
		Reporter: reporter,
		Clock:    time.Now,
	})

	handler := httpapi.NewServer(httpapi.Config{
		ServiceToken: cfg.ServiceToken,
	}, orchestrator, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("gateway listening", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("gateway shutting down", "signal", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("gateway failed", "error", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("gateway shutdown failed", "error", err)
		os.Exit(1)
	}
}
