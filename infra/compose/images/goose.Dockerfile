FROM golang:1.24-alpine AS builder

ARG GOOSE_VERSION=v3.24.1

RUN apk add --no-cache git ca-certificates && \
    go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/goose /usr/local/bin/goose

ENTRYPOINT ["goose"]
