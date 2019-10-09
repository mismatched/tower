all: build

build:
	go build -o tower cmd/*

test:
	sudo go test -v -race ./...