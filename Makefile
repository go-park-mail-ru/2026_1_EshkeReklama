coverage:
	go test ./... -coverpkg=./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -n 1

swagger:
	swag init -d ./cmd/eshkere,./internal/app,./internal/handler,./internal/middleware,./pkg
