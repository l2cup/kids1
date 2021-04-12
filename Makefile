lint:
	go fmt ./...
	go vet ./...
	golint ./...

build:
	go build -o ./cmd/wc ./cmd/

run:
	PROP_CONFIG_PATH=./cmd/config.properties ./cmd/wc

staticcheck:
	staticcheck ./...

all: lint staticcheck build run
