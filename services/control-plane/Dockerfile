FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server/

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary and config
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/configs /app/configs

EXPOSE 4000

ENTRYPOINT ["/app/server"]
