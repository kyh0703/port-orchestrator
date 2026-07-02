# Roadmap

## v1

Build the first internal `port-orchestrator` service slice:

- Receive service-authenticated API dispatch.
- Attach an agent participant to `../port-media` through an outbound port.
- Start best-effort recording orchestration when requested.
- Report lifecycle events back to `../port-api`.
- Keep durable persistence, browser APIs, WebRTC forwarding, and recording
  storage outside orchestrator.

## Deferred

- SIP/PSTN inbound runtime.
- SIP/PSTN outbound runtime.
- Real media WebSocket attachment implementation.
- Recording storage, retention, encryption, and playback.

