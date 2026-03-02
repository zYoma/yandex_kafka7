FROM golang:1.25.4-bookworm AS builder

WORKDIR /app

ENV GOPATH=/app/.go
ENV PATH="${PATH}:${GOPATH}/bin}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/yandex-kafka -ldflags "-s -w" ./cmd/shop && \
    go build -o /app/yandex-kafka-client -ldflags "-s -w" ./cmd/client && \
    go build -o /app/yandex-kafka-analytic -ldflags "-s -w" ./cmd/analytic_client

FROM debian:12-slim AS runtime

WORKDIR /app

COPY --from=builder /app/yandex-kafka /app/yandex-kafka
COPY --from=builder /app/yandex-kafka-client /app/yandex-kafka-client
COPY --from=builder /app/yandex-kafka-analytic /app/yandex-kafka-analytic
RUN chmod +x /app/yandex-kafka /app/yandex-kafka-client /app/yandex-kafka-analytic

USER 65535
