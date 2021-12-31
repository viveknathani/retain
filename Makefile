build:
	go build -o ./bin/ ./cmd/server/
	go build -o ./bin/ ./cmd/client/

test:
	go test -v ./...
	