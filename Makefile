db-up:
	docker compose up -d

db-down:
	docker compose down

coverage:
	go test ./... -coverpkg=./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -n 1

swagger:
	swag init -d ./cmd/eshkere,./internal/app,./internal/handler,./internal/middleware,./pkg

MIGRATE_IMAGE ?= migrate/migrate:v4.18.1

# пример создания новой миграции, замените add_new_type на имя миграции
migrations-create-example:
	docker run --rm -v "$(CURDIR)/db/migrations:/migrations" $(MIGRATE_IMAGE) create -ext sql -dir /migrations -seq -digits 6 add_new_type
