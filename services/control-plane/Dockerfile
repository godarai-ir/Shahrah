FROM public.ecr.aws/docker/library/golang:1.26.3-alpine AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src
RUN apk --no-cache add git

COPY packages/contracts ./packages/contracts
COPY services/control-plane/go.mod services/control-plane/go.sum ./services/control-plane/

WORKDIR /src/services/control-plane
ENV GOWORK=off
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /src
COPY services/control-plane ./services/control-plane

WORKDIR /src/services/control-plane
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /out/metrics-dlq-replayer ./cmd/metrics-dlq-replayer

FROM public.ecr.aws/docker/library/alpine:3.23.4

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server
COPY --from=builder /out/metrics-dlq-replayer /app/metrics-dlq-replayer
COPY --from=builder /src/services/control-plane/configs /app/configs

EXPOSE 4000

ENTRYPOINT ["/app/server"]
