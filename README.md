# port-gateway

Internal orchestration service for room-based voice conversations.

`port-gateway` receives service-authenticated dispatches from `port-api`,
coordinates gateway-managed participants with `port-media`, and reports
lifecycle events back to `port-api`.

## Run

```bash
GATEWAY_SERVICE_TOKEN=dev-token \
PORT_API_SERVICE_TOKEN=dev-api-token \
PORT_RECORD_BASE_URL=http://localhost:8081 \
PORT_RECORD_INTERNAL_API_TOKEN=dev-record-token \
go run ./cmd/gateway
```

Recording dispatches call `POST {PORT_RECORD_BASE_URL}/api/recordings/start`
with `Authorization: Bearer <PORT_RECORD_INTERNAL_API_TOKEN>`.

## Endpoints

- `GET /healthz`
- `POST /internal/v1/dispatches`

Dispatch requests require:

```text
Authorization: Bearer <GATEWAY_SERVICE_TOKEN>
```

## Verify

```bash
go test ./...
make build
```
