coverage:
	go test ./... -coverpkg=./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -n 1