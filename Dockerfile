# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

# download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \ -ldflags="-s -w"  -o /app/bin/app ./cmd

# ── Final stage ────────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS final

WORKDIR /app

# copy binary from builder
COPY --from=builder /app/bin/app .

# run as non-root user (distroless provides nonroot user)
USER nonroot:nonroot

EXPOSE 3000

ENTRYPOINT ["/app/app", "server"]