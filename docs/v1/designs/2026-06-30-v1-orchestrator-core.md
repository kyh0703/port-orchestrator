---
feature: orchestrator-core
status: plan_ready
created_at: 2026-06-30T00:00:00+09:00
---

# Orchestrator Core

## Context / Inputs

- `docs/ARCHITECTURE.md` defines `port-orchestrator` as the internal orchestration
  layer between `../port-api` and `../port-media`.
- `../port-api` owns user auth, durable conversation state, token issuance, and
  event persistence.
- `../port-media` owns SFU rooms, signaling, participants, tracks, and live
  media forwarding.
- Orchestrator must be written in Go with a practical DDD + Hexagonal structure.

## Problem Statement

Create the first runnable orchestrator service that can receive internal dispatch,
validate service authentication, coordinate agent and recording attach work
through ports, and report lifecycle events back to API without taking ownership
of durable state or media forwarding.

## Decision Drivers

- Keep orchestrator as an internal service, not a public browser API.
- Use Go concurrency for session-scoped orchestration, cancellation, timeouts,
  and retryable outbound callbacks.
- Keep domain rules independent from HTTP, WebSocket, process, and API clients.
- Avoid overbuilding SIP and recorder runtime before contracts are concrete.

## Options Considered

### Option A: Go stdlib HTTP + small ports

- Minimal dependencies.
- Easy to test orchestration with fake outbound ports.
- Fits internal service scope.
- Requires manually wiring dependencies.

### Option B: Fiber + fx template

- Provides full app framework and DI conventions.
- Adds many unused pieces for this slice: DB, auth scaffolding, DTO layering,
  validators, and code generation.

## Recommended Option

Use Option A. Build a small Go service with stdlib HTTP, domain/application
packages, inbound/outbound ports, and stub outbound adapters. Add only the
runtime dependencies needed by the first slice.

## Scope Decision

In:

- Go module and runnable `cmd/orchestrator`.
- Service auth middleware for dispatch.
- `POST /internal/v1/dispatches`.
- `GET /healthz`.
- Domain model for dispatch, participant tokens, lifecycle event types, and
  recording request.
- Application orchestrator that runs agent attach and optional recording start.
- API lifecycle reporter with retry policy.
- Fake/stub media, agent, and recorder adapters suitable for local smoke tests.
- Unit tests for validation, auth, orchestration, and retry behavior.

Out:

- Public browser API.
- Durable database.
- WebRTC/SFU forwarding.
- Real SIP signaling.
- Real recording storage.
- OpenAI provider runtime.

Deferred:

- Real media WebSocket participant attach adapter.
- Real AgentWorker process or network protocol.
- Real recorder egress adapter.
- SIP inbound/outbound call legs.

## Open Questions

- Exact `port-api` event intake endpoint is not finalized. Use a configurable
  path with a documented default.
- Exact `port-media` service participant signaling contract is not finalized.
  Keep it behind an outbound port.

## Plan Handoff

### Scope for Planning

Create the first `orchestrator-core` implementation as one vertical slice: config,
HTTP inbound adapter, domain/application ports, orchestration service, outbound
API reporter, stub agent/recording/media adapters, and tests.

### Success Criteria

- `go test ./...` passes.
- `go build ./cmd/orchestrator` passes.
- Dispatch without valid service auth returns `401`.
- Dispatch with `attachAgent: true` reports `agent.started` after agent attach.
- Dispatch with `recording.enabled: true` reports `recording.started`.
- Agent attach failure reports `agent.failed`.
- API event reporter retries transient failures with bounded attempts.

### Suggested Validation

- `go test ./...`
- `go build ./cmd/orchestrator`

### Parallelization Hints

- Sequential implementation is recommended for the first slice because module
  layout, ports, application service, and HTTP adapter share contracts.

