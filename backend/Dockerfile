# Build stage
FROM golang:1.25-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 go build -o server ./cmd/server

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates && \
    adduser -D appuser

WORKDIR /home/appuser
COPY --from=backend --chown=appuser:appuser /build/server .

USER appuser
EXPOSE 8080

CMD ["./server"]