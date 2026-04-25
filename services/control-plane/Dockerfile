FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY packages/contracts ./packages/contracts
COPY services/control-plane ./services/control-plane

WORKDIR /src/services/control-plane
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /out/server ./cmd/server

FROM alpine:3.19

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server
COPY --from=builder /src/services/control-plane/configs /app/configs

EXPOSE 4000

ENTRYPOINT ["/app/server"]
