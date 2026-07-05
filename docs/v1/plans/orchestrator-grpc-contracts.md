# Orchestrator gRPC Contracts

## Goal

- Replace internal HTTP/key-based API event reporting with the shared gRPC
  contract from `../contracts`, and remove internal service token handling.

## References

- docs/STATE.md
- docs/ROADMAP.md
- docs/ARCHITECTURE.md
- docs/v1/designs/2026-06-30-v1-orchestrator-core.md
- ../contracts/proto/port/api/v1/gateway_events.proto

## Workspace

- Branch: feat/orchestrator-grpc-contracts
- Base: main
- Isolation: required
- Created by: planning via docs lifecycle

## Task Graph

### Task T1

- [x] Complete
- Goal: Switch orchestrator API lifecycle reporting to the shared
  `ApiEventService.RecordGatewayEvent` gRPC contract, remove internal bearer
  token configuration/auth checks, and update focused tests/docs.
- Depends on:
  - none
- Write Scope:
  - go.mod
  - go.sum
  - README.md
  - cmd/orchestrator/**
  - internal/adapters/inbound/httpapi/**
  - internal/adapters/outbound/apihttp/**
  - internal/adapters/outbound/apigrpc/**
  - internal/config/**
- Read Context:
  - docs/ARCHITECTURE.md
  - docs/v1/designs/2026-06-30-v1-orchestrator-core.md
  - ../contracts/proto/port/api/v1/gateway_events.proto
- Checks:
  - go test ./...
  - go vet ./...
  - go build ./cmd/orchestrator
- Parallel-safe: no

## Notes

- `../contracts` currently defines the API lifecycle event gRPC service, but no
  dispatch-receive gRPC service for orchestrator. This plan keeps the existing
  dispatch HTTP endpoint and removes its internal bearer auth only.
