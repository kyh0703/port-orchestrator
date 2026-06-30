# Gateway Core

## Goal

- Build the first runnable Go `port-gateway` service for internal API dispatch,
  agent attach orchestration, optional recording start, and API lifecycle
  reporting.

## References

- docs/STATE.md
- docs/ROADMAP.md
- docs/ARCHITECTURE.md
- docs/v1/designs/2026-06-30-v1-gateway-core.md

## Workspace

- Branch: feat/v1-gateway-core
- Base: main
- Isolation: required
- Created by: planning via docs lifecycle

## Task Graph

### Task T1

- [ ] Complete
- Goal: Initialize the Go module and implement domain models, ports, config,
  HTTP inbound adapter, orchestration use case, outbound API reporter, stub
  agent/recording/media adapters, and focused tests for the first dispatch
  flow.
- Depends on:
  - none
- Write Scope:
  - go.mod
  - go.sum
  - Makefile
  - README.md
  - cmd/gateway/**
  - internal/**
- Read Context:
  - docs/ARCHITECTURE.md
  - docs/v1/designs/2026-06-30-v1-gateway-core.md
- Checks:
  - go test ./...
  - go build ./cmd/gateway
- Parallel-safe: no

## Notes

- Keep adapters replaceable. Do not let HTTP DTOs leak into domain.
- Do not add database persistence.
- Keep SIP as a deferred boundary unless needed by the first dispatch flow.

