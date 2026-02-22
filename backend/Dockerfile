# Dev stage
FROM golang:1.25.3-alpine AS dev
RUN go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["air"]

# Build stage
FROM golang:1.25.3-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/spotiscan

# Production stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates && \
    adduser -D -s /sbin/nologin appuser

WORKDIR /home/appuser
COPY --from=builder --chown=appuser:appuser /build/server .
COPY --from=builder --chown=appuser:appuser /build/bot_data.db .
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["./server"]
