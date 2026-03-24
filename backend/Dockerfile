# Base stage: shared dependency download
FROM golang:1.25.3-alpine AS base
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

# Source stage: base + full source
FROM base AS source
COPY . .

# Dev stage
FROM base AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
CMD ["air"]

# Scraper stage
FROM source AS scraper
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper ./cmd/scraper
CMD ["./scraper"]

# Build stage
FROM source AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/spotiscan

# Production stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates && \
    adduser -D -s /sbin/nologin appuser

WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appuser /build/server .

USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./server"]
