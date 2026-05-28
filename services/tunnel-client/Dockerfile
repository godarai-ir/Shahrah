FROM golang:1.26.3-alpine AS builder

WORKDIR /src

COPY . .
ENV GOWORK=off
ENV GOFLAGS=-mod=vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.23.4

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8081

ENTRYPOINT ["/app/server"]
