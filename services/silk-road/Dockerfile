FROM golang:1.26.3-alpine AS builder

WORKDIR /src

COPY packages/contracts ./packages/contracts
COPY services/silk-road ./services/silk-road

WORKDIR /src/services/silk-road
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.23.4

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8083

ENTRYPOINT ["/app/server"]
