FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/eshkere ./cmd/eshkere

FROM alpine:3.22
RUN adduser -D appuser
USER appuser
WORKDIR /app

COPY --from=build /bin/eshkere /app/eshkere
COPY config/config.yaml /app/config/config.yaml

EXPOSE 8000

CMD ["/app/eshkere", "-config", "/app/config/config.yaml"]
