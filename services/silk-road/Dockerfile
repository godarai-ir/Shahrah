FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.work ./
COPY packages/contracts ./packages/contracts
COPY services/silk-road ./services/silk-road

WORKDIR /src/services/silk-road
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8083

ENTRYPOINT ["/app/server"]
