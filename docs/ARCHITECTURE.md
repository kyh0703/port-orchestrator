# Architecture

## Repository Boundary

This repository owns the server-internal `port-orchestrator` orchestration layer.

- Orchestrator source: repository runtime packages, when implemented.
- Orchestrator docs: `docs/**`
- API repository: `../port-api`
- Media server repository: `../port-media`
- Web client repository: `../port-web`

## Purpose

`port-orchestrator` coordinates server-side participants and call legs for
room-based voice conversations after `../port-api` has authenticated the user
and reserved a media room.

Orchestrator is not a public web API. It receives internal dispatch from
`../port-api`, attaches orchestrator-managed services to `../port-media`, and reports
lifecycle events back to API-owned durable persistence and user-facing fanout.

LiveKit is a structural reference only. This architecture does not introduce a
LiveKit service or SDK dependency.

## Shared Boundaries

### Core Domains

- API server: `../port-api` owns user authentication, authorization, durable
  conversation creation, participant token issuance, owner-scoped SSE,
  user-facing status, history APIs, and durable persistence.
- Media server: `../port-media` owns SFU rooms, participants, tracks,
  WebSocket signaling/control, WebRTC forwarding, and live media runtime state.
- Gateway: this repository owns API dispatch handling, AgentWorker attach,
  recording egress orchestration, future SIP/PSTN inbound and outbound call
  legs, and lifecycle reporting to API.
- Web client: `../port-web` talks to API and media only. It never calls orchestrator.

### Data Boundaries

- Orchestrator receives internal dispatch from API after room reservation.
- Orchestrator receives only service-scoped credentials and participant tokens needed
  to attach server-side participants.
- Orchestrator joins or coordinates media participants through `../port-media`.
- Orchestrator reports durable lifecycle and result events to `../port-api`.
- Orchestrator does not persist durable conversation history as the source of truth.
- Orchestrator does not stream events directly to `../port-web`.
- Orchestrator does not forward SFU media itself; media forwarding remains in
  `../port-media`.

## Responsibilities

### Orchestrator Owns

- Receiving API dispatch for a reserved conversation/session.
- Validating internal service authentication on dispatch and callbacks.
- Attaching AgentWorker to a media room as an agent participant.
- Starting and stopping best-effort recording egress.
- Coordinating recording completion/failure metadata reports.
- Handling future SIP/PSTN inbound call legs.
- Handling future SIP/PSTN outbound call legs.
- Joining SIP/orchestrator media participants to `../port-media` when required.
- Reporting lifecycle events back to API.
- Retrying or failing orchestration work according to explicit policy.

### Orchestrator Does Not Own

- Public browser APIs.
- User authentication source of truth.
- Durable conversation persistence.
- Durable transcript/history APIs.
- Owner-scoped browser SSE.
- SFU room forwarding, track routing, or WebRTC peer connections.
- Browser WebRTC signaling.
- Recording storage, encryption, retention, or playback APIs.
- API-issued user participant token responses.

## Primary Conversation Flow

1. Browser calls `POST /api/v1/conversations` on `../port-api`.
2. API authenticates the user and creates a durable conversation.
3. API reserves a room in `../port-media`.
4. API mints the user participant token for the browser.
5. API mints internal service participant tokens for orchestrator-managed
   participants.
6. API sends internal dispatch to orchestrator with `conversationId`, `sessionId`,
   `roomId`, media signaling context, and service participant tokens.
7. Orchestrator attaches the AgentWorker to the media room.
8. Orchestrator starts best-effort recording egress when configured.
9. Orchestrator reports `agent.started` and `recording.started` to API.
10. Agent, recorder, and future SIP participants exchange media through
    `../port-media`.
11. Orchestrator reports lifecycle, completion, and failure events to API.
12. API persists durable status/history/recording metadata and fans out
    owner-scoped browser events.

## SIP Inbound Flow

1. Orchestrator receives an inbound SIP `INVITE`.
2. Orchestrator authenticates and validates the SIP trunk/source according to orchestrator
   policy.
3. Gateway calls API with service auth to create or attach a conversation and
   session for the inbound call.
4. API creates durable conversation/session records and reserves a media room.
5. API returns internal media context and service participant tokens to orchestrator.
6. Orchestrator joins the media room as a SIP/orchestrator participant through
   `../port-media`.
7. Gateway connects the SIP call leg to the room media path.
8. Orchestrator reports `sip.connected` to API.
9. When the call ends, orchestrator reports `sip.ended` to API.
10. Recording results and agent lifecycle events are reported to API through the
    same event intake boundary.

## SIP Outbound Flow

1. API or another trusted internal service requests outbound call orchestration.
2. Orchestrator validates service auth and policy.
3. Orchestrator obtains or receives API-created conversation/session and media room
   context.
4. Orchestrator starts the outbound SIP call leg.
5. Orchestrator joins the media room as a SIP/orchestrator participant through
   `../port-media`.
6. Orchestrator reports SIP lifecycle events to API.

## Recording Flow

1. Orchestrator receives recording configuration through API dispatch.
2. Orchestrator starts recording egress as a best-effort service participant or egress
   worker.
3. Orchestrator reports `recording.started` to API after the egress starts.
4. Orchestrator reports `recording.completed` with result metadata when the egress
   finishes.
5. Orchestrator reports `recording.failed` with failure reason when the egress fails.
6. API owns user-facing recording metadata persistence.
7. A dedicated recording/storage service may own encryption, retention, and
   object storage.

## Event Ownership

### `port-orchestrator` to `port-api`

- `agent.started`
- `agent.failed`
- `sip.connected`
- `sip.ended`
- `recording.started`
- `recording.completed`
- `recording.failed`

### `port-orchestrator` to `port-media`

- Media room join/signaling as orchestrator-managed service participants.
- No durable persistence contract.

### `port-orchestrator` to `port-web`

- No direct browser transport.

## API Integration

- Dispatch auth: service-to-service authentication.
- Dispatch payload:
  - `conversationId`
  - `sessionId`
  - `roomId`
  - `mediaSignalingUrl`
  - service participant tokens for agent, recorder, SIP, or orchestrator roles
  - requested capabilities, such as agent attach or recording egress
- Event callback auth: service-to-service authentication.
- Event payload:
  - stable event id
  - event type
  - `conversationId`
  - `sessionId`
  - `roomId`
  - participant id when applicable
  - occurred timestamp
  - status/result/failure payload

## Non-Goals

- Public web API.
- Browser authentication or user session management.
- Durable conversation source of truth.
- Durable transcript/history storage.
- Owner-scoped browser SSE.
- SFU forwarding or WebRTC peer connection ownership.
- API response shaping for browser room creation.
- Direct browser orchestrator calls.
- Recording storage, encryption, retention, or playback ownership.
- Replacing `../port-media` as the media plane.
