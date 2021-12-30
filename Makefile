build:
	go build -o ./bin/
	sudo cp ./bin/retain ~/bin/retain

test:
	go test -v ./...
	