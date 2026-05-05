.PHONY: up down proto

proto:
	protoc \
		--go_out=. \
		--go_opt=module=eshkere \
		--go-grpc_out=. \
		--go-grpc_opt=module=eshkere \
		proto/auth/v1/auth.proto \
		proto/profile/v1/profile.proto

up:
	docker compose up -d --build

down:
	docker compose down -v

coverage:
	go test ./... -coverprofile=cover.out && go tool cover -func=cover.out | tail -n 1

swagger:
	swag init -g main.go -d ./cmd/eshkere,./internal/app,./internal/handler,./internal/handler/v1,./internal/handler/v1/dto,./internal/models,./internal/handler/middleware,./pkg/httpx

# пример создания новой миграции, замените add_new_type на имя миграции
migrations-create-example:
	migrate create -digits 6 -ext sql -dir db/migrations -seq add_new_type
