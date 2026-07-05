# port-orchestrator

Internal orchestration service for room-based voice conversations.

`port-orchestrator` receives dispatches from `port-api`, coordinates
orchestrator-managed participants with `port-media`, and reports lifecycle
events back to `port-api` through the shared gRPC contracts.

## Run

```bash
PORT_API_GRPC_ADDR=localhost:50051 \
go run ./cmd/orchestrator
```

## Endpoints

- `GET /healthz`
- `POST /internal/v1/dispatches`

## Verify

```bash
go test ./...
make build
```
