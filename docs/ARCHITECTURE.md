# Architecture

## Repository Boundary

This repository owns the server-internal `port-gateway` orchestration layer.

- Gateway source: repository runtime packages, when implemented.
- Gateway docs: `docs/**`
- API repository: `../port-api`
- Media server repository: `../port-media`
- Web client repository: `../port-web`

## Purpose

`port-gateway` coordinates server-side participants and call legs for
room-based voice conversations after `../port-api` has authenticated the user
and reserved a media room.

Gateway is not a public web API. It receives internal dispatch from
`../port-api`, attaches gateway-managed services to `../port-media`, and reports
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
- Web client: `../port-web` talks to API and media only. It never calls gateway.

### Data Boundaries

- Gateway receives internal dispatch from API after room reservation.
- Gateway receives only service-scoped credentials and participant tokens needed
  to attach server-side participants.
- Gateway joins or coordinates media participants through `../port-media`.
- Gateway reports durable lifecycle and result events to `../port-api`.
- Gateway does not persist durable conversation history as the source of truth.
- Gateway does not stream events directly to `../port-web`.
- Gateway does not forward SFU media itself; media forwarding remains in
  `../port-media`.

## Responsibilities

### Gateway Owns

- Receiving API dispatch for a reserved conversation/session.
- Validating internal service authentication on dispatch and callbacks.
- Attaching AgentWorker to a media room as an agent participant.
- Starting and stopping best-effort recording egress.
- Coordinating recording completion/failure metadata reports.
- Handling future SIP/PSTN inbound call legs.
- Handling future SIP/PSTN outbound call legs.
- Joining SIP/gateway media participants to `../port-media` when required.
- Reporting lifecycle events back to API.
- Retrying or failing orchestration work according to explicit policy.

### Gateway Does Not Own

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
5. API mints internal service participant tokens for gateway-managed
   participants.
6. API sends internal dispatch to gateway with `conversationId`, `sessionId`,
   `roomId`, media signaling context, and service participant tokens.
7. Gateway attaches the AgentWorker to the media room.
8. Gateway starts best-effort recording egress when configured.
9. Gateway reports `agent.started` and `recording.started` to API.
10. Agent, recorder, and future SIP participants exchange media through
    `../port-media`.
11. Gateway reports lifecycle, completion, and failure events to API.
12. API persists durable status/history/recording metadata and fans out
    owner-scoped browser events.

## SIP Inbound Flow

1. Gateway receives an inbound SIP `INVITE`.
2. Gateway authenticates and validates the SIP trunk/source according to gateway
   policy.
3. Gateway calls API with service auth to create or attach a conversation and
   session for the inbound call.
4. API creates durable conversation/session records and reserves a media room.
5. API returns internal media context and service participant tokens to gateway.
6. Gateway joins the media room as a SIP/gateway participant through
   `../port-media`.
7. Gateway connects the SIP call leg to the room media path.
8. Gateway reports `sip.connected` to API.
9. When the call ends, gateway reports `sip.ended` to API.
10. Recording results and agent lifecycle events are reported to API through the
    same event intake boundary.

## SIP Outbound Flow

1. API or another trusted internal service requests outbound call orchestration.
2. Gateway validates service auth and policy.
3. Gateway obtains or receives API-created conversation/session and media room
   context.
4. Gateway starts the outbound SIP call leg.
5. Gateway joins the media room as a SIP/gateway participant through
   `../port-media`.
6. Gateway reports SIP lifecycle events to API.

## Recording Flow

1. Gateway receives recording configuration through API dispatch.
2. Gateway starts recording egress as a best-effort service participant or egress
   worker.
3. Gateway reports `recording.started` to API after the egress starts.
4. Gateway reports `recording.completed` with result metadata when the egress
   finishes.
5. Gateway reports `recording.failed` with failure reason when the egress fails.
6. API owns user-facing recording metadata persistence.
7. A dedicated recording/storage service may own encryption, retention, and
   object storage.

## Event Ownership

### `port-gateway` to `port-api`

- `agent.started`
- `agent.failed`
- `sip.connected`
- `sip.ended`
- `recording.started`
- `recording.completed`
- `recording.failed`

### `port-gateway` to `port-media`

- Media room join/signaling as gateway-managed service participants.
- No durable persistence contract.

### `port-gateway` to `port-web`

- No direct browser transport.

## API Integration

- Dispatch auth: service-to-service authentication.
- Dispatch payload:
  - `conversationId`
  - `sessionId`
  - `roomId`
  - `mediaSignalingUrl`
  - service participant tokens for agent, recorder, SIP, or gateway roles
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
- Direct browser gateway calls.
- Recording storage, encryption, retention, or playback ownership.
- Replacing `../port-media` as the media plane.
