TOWER_BIN = bin/tower

all: build

build:
	go build -o $(TOWER_BIN) tower/*.go

test:
	sudo -E env "PATH=$PATH" go test -v -race ./...

.PHONY : clean
clean:
	-rm $(TOWER_BIN)