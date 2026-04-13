.PHONY: db-up db-down dev

db-up:
	docker compose up -d

# Поднимает Postgres, Redis и миграции, затем запускает API (нужен .env с POSTGRES_* и REDIS_HOST=localhost).
dev: db-up
	go run ./cmd/eshkere -config ./config/config.yaml

db-down:
	docker compose down

coverage:
	go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | tail -n 1

swagger:
	swag init -g main.go -d ./cmd/eshkere,./internal/app,./internal/handler,./internal/handler/dto,./internal/models,./internal/middleware,./pkg/httpx

# пример создания новой миграции, замените add_new_type на имя миграции
migrations-create-example:
	migrate create -digits 6 -ext sql -dir db/migrations -seq add_new_type