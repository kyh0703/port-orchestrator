# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25-bookworm

FROM golang:${GO_VERSION} AS build
WORKDIR /src

# 의존성 캐시 레이어 (go.mod/go.sum 만 먼저 복사)
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/orchestrator ./cmd/orchestrator

# distroless: shell 없는 최소 런타임, nonroot 로 실행
FROM gcr.io/distroless/static-debian12:nonroot AS runner
WORKDIR /
COPY --from=build /out/orchestrator /orchestrator

ENV ORCHESTRATOR_HTTP_ADDR=:8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/orchestrator"]
