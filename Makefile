all: build

build:
	go build -o bin/tower tower/*.go

test:
	sudo -E env "PATH=$PATH" go test -v -race ./...