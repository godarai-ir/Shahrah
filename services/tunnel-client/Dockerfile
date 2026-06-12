FROM public.ecr.aws/docker/library/golang:1.26.3-alpine AS builder

WORKDIR /src

COPY packages/contracts ./packages/contracts
COPY services/tunnel-client/go.mod ./services/tunnel-client/go.mod

WORKDIR /src/services/tunnel-client
ENV GOWORK=off
RUN go mod download

WORKDIR /src
COPY services/tunnel-client ./services/tunnel-client

WORKDIR /src/services/tunnel-client
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM public.ecr.aws/docker/library/alpine:3.23.4

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8081

ENTRYPOINT ["/app/server"]
