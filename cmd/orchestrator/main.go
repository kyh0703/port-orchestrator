package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiv1 "github.com/kyh0703/port-contracts/gen/go/port/api/v1"
	"github.com/kyh0703/port-orchestrator/internal/adapters/inbound/httpapi"
	"github.com/kyh0703/port-orchestrator/internal/adapters/outbound/apigrpc"
	"github.com/kyh0703/port-orchestrator/internal/adapters/outbound/stub"
	"github.com/kyh0703/port-orchestrator/internal/application/orchestration"
	"github.com/kyh0703/port-orchestrator/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid orchestrator config", "error", err)
		os.Exit(1)
	}

	apiConn, err := grpc.NewClient(cfg.APIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("create api grpc client", "error", err)
		os.Exit(1)
	}
	defer apiConn.Close()

	reporter := apigrpc.NewReporter(apigrpc.ReporterConfig{
		MaxAttempts: cfg.APIReportMaxAttempts,
		RetryDelay:  cfg.APIReportRetryDelay,
		Timeout:     cfg.APIReportTimeout,
	}, apiv1.NewApiEventServiceClient(apiConn))

	orchestrator := orchestration.NewService(orchestration.Dependencies{
		Media:    stub.NewMediaConnector(logger),
		Agent:    stub.NewAgentAttacher(logger),
		Recorder: stub.NewRecorder(logger),
		Reporter: reporter,
		Clock:    time.Now,
	})

	handler := httpapi.NewServer(orchestrator, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("orchestrator listening", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("orchestrator shutting down", "signal", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("orchestrator failed", "error", err)
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("orchestrator shutdown failed", "error", err)
		os.Exit(1)
	}
}
