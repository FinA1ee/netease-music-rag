# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.25-bookworm AS builder

WORKDIR /src/backend

ENV CGO_ENABLED=0
ENV GOOS=linux

# Copy module files first for better layer caching.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Build a small, reproducible Linux binary.
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ── Stage 2: run ──────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app

# ca-certificates is required for outbound TLS requests.
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/server ./server

EXPOSE 8080

CMD ["./server"]
