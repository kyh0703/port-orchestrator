# port-gateway

Internal orchestration service for room-based voice conversations.

`port-gateway` receives service-authenticated dispatches from `port-api`,
coordinates gateway-managed participants with `port-media`, and reports
lifecycle events back to `port-api`.

## Run

```bash
GATEWAY_SERVICE_TOKEN=dev-token \
PORT_API_SERVICE_TOKEN=dev-api-token \
go run ./cmd/gateway
```

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
