FROM golang:1.26-alpine AS builder

ARG GOPROXY=https://goproxy.io,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /src
COPY packages/contracts ./packages/contracts
COPY services/rag/go.mod services/rag/go.sum ./services/rag/
WORKDIR /src/services/rag
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

WORKDIR /src
COPY services/rag ./services/rag
WORKDIR /src/services/rag
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/rag ./cmd/rag

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/rag /rag
USER 65532:65532
ENTRYPOINT ["/rag"]
