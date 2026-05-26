# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/eshkere ./cmd/eshkere
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/auth ./cmd/auth
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/profile ./cmd/profile
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/analytics-consumer ./cmd/analytics-consumer

FROM alpine:3.22 AS auth
RUN adduser -D appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/auth /app/auth

EXPOSE 50051 9101

CMD ["/app/auth"]

FROM alpine:3.22 AS profile
RUN adduser -D appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/profile /app/profile

EXPOSE 50052 9102

CMD ["/app/profile"]

FROM alpine:3.22 AS analytics-consumer
RUN adduser -D appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/analytics-consumer /app/analytics-consumer
COPY config/config.yaml /app/config/config.yaml

CMD ["/app/analytics-consumer", "-config", "/app/config/config.yaml"]

FROM alpine:3.22 AS eshkere
RUN adduser -D appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/eshkere /app/eshkere
COPY config/config.yaml /app/config/config.yaml

EXPOSE 8000 9100

CMD ["/app/eshkere", "-config", "/app/config/config.yaml"]
