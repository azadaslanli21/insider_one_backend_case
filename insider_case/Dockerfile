# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (layer-cached unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o football-simulator ./cmd/api

# ── Stage 2: Minimal runtime image ────────────────────────────────────────────
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/football-simulator .

EXPOSE 8080

CMD ["./football-simulator"]