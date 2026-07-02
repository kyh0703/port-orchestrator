# port-orchestrator

Internal orchestration service for room-based voice conversations.

`port-orchestrator` receives service-authenticated dispatches from `port-api`,
coordinates orchestrator-managed participants with `port-media`, and reports
lifecycle events back to `port-api`.

## Run

```bash
ORCHESTRATOR_SERVICE_TOKEN=dev-token \
PORT_API_SERVICE_TOKEN=dev-api-token \
go run ./cmd/orchestrator
```

## Endpoints

- `GET /healthz`
- `POST /internal/v1/dispatches`

Dispatch requests require:

```text
Authorization: Bearer <ORCHESTRATOR_SERVICE_TOKEN>
```

## Verify

```bash
go test ./...
make build
```
