.PHONY: run build test clean

run:
	go run main.go

build:
	go build -o bin/taskmanager main.go

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f bin/taskmanager coverage.out coverage.html

lint:
	go vet ./...