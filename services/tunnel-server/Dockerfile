FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.work ./
COPY packages/contracts ./packages/contracts
COPY services/tunnel-server ./services/tunnel-server

WORKDIR /src/services/tunnel-server
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8082

ENTRYPOINT ["/app/server"]
