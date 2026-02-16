# ── Build stage ──
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /bin/server ./cmd/server

# ── Runtime stage ──
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=builder /bin/server /app/server

# Copy templates, static assets, and seed data
COPY web/ /app/web/
COPY data/projects.json /app/data/projects.json
COPY data/content/ /app/data/content/

# Data directory is a volume mount point — DBs live here and survive redeploys
VOLUME /app/data

EXPOSE 8080

ENTRYPOINT ["/app/server"]
