.PHONY: up down down-reset proto lint lint-fmt generate coverage test

generate:
	go generate ./...

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
	docker compose down

down-reset:
	docker compose down -v

coverage:
	go test ./... -coverprofile=cover.out && \
    grep -v '_easyjson.go:' cover.out | \
    grep -v 'service_mock_test.go:' | \
    grep -v 'repo_mocks_test.go:' | \
    grep -v 'docs/docs.go:' | \
    grep -v 'tools/easyjson-gen/' | \
    grep -v '\.pb\.go:' | \
    grep -v '_grpc\.pb\.go:' | \
    grep -v 'swagger\.json:' | \
    grep -v 'swagger\.yaml:' | \
    grep -v '_generated' \
    > cover.filtered.out && \
    go tool cover -func=cover.filtered.out | tail -n 1

test:
	env GOCACHE=/private/tmp/eshkere-go-build-cache go test ./...

lint:
	golangci-lint run

swagger:
	swag init -g main.go -d ./cmd/eshkere,./internal/app,./internal/handler,./internal/handler/v1,./internal/handler/v1/dto,./internal/models,./internal/handler/middleware,./pkg/httpx

# пример создания новой миграции, замени add_new_type на имя миграции
migrations-create-example:
	migrate create -digits 6 -ext sql -dir db/migrations -seq add_new_type
