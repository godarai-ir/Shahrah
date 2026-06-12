FROM public.ecr.aws/docker/library/golang:1.26.3-alpine AS builder

WORKDIR /src

COPY packages/contracts ./packages/contracts
COPY services/tunnel-server/go.mod ./services/tunnel-server/go.mod

WORKDIR /src/services/tunnel-server
ENV GOWORK=off
RUN go mod download

WORKDIR /src
COPY services/tunnel-server ./services/tunnel-server

WORKDIR /src/services/tunnel-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM public.ecr.aws/docker/library/alpine:3.23.4

WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /out/server /app/server

EXPOSE 8082

ENTRYPOINT ["/app/server"]
