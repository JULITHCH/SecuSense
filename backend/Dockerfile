FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod file and source code
COPY . .

# Download dependencies and build
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata python3 py3-pip \
    && pip3 install --break-system-packages edge-tts

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/migrations ./migrations

# Create directory for generated audio files
RUN mkdir -p ./generated/audio

EXPOSE 8080

CMD ["./main"]
