FROM golang:1.26.3-alpine AS builder

ARG GOOSE_VERSION=v3.24.1

RUN apk add --no-cache git ca-certificates && \
    go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.23.4

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/goose /usr/local/bin/goose

ENTRYPOINT ["goose"]
