all: build

build:
	go build -o tower cmd/*.go

test:
	sudo -E env "PATH=$PATH" go test -v -race ./...