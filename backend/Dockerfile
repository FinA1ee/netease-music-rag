# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:latest AS builder

# Allow Go to download the toolchain version required by go.mod (e.g. 1.25)
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Cache dependency downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Produce a statically-linked binary (no CGO needed for pure-Go deps)
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# ── Stage 2: run ──────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app

# ca-certificates is required for TLS calls to Gemini / Supabase
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server .

# .env is optional — all config can be injected via docker-compose environment:
# COPY .env .

EXPOSE 8080

CMD ["./server"]
