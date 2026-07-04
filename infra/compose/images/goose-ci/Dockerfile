FROM public.ecr.aws/docker/library/golang:1.26.3-alpine AS builder

ARG GOOSE_VERSION=v3.24.1

RUN apk add --no-cache git ca-certificates && \
    for i in 1 2 3 4 5; do \
        go install -tags 'no_sqlite3 no_clickhouse no_mysql no_mssql no_vertica no_ydb no_libsql' \
            github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION} && break || sleep 5; \
    done

FROM public.ecr.aws/docker/library/alpine:3.23.4

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations

ENTRYPOINT ["goose"]
