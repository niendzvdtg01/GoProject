# syntax=docker/dockerfile:1.7

# ── builder ─────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder
WORKDIR /src

# Cache modules layer separately so source edits don't bust the dep download.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend ./cmd

# ── runtime ─────────────────────────────────────────────────────────────────
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=builder /out/backend ./backend
USER app
EXPOSE 8080

# Lightweight healthcheck: any 2xx/3xx/4xx from the API root means the server
# is up; only network failures (curl exit != 0) trip the unhealthy state.
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- --tries=1 --timeout=2 http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["./backend"]
