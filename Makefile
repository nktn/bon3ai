.PHONY: build install test clean

BINARY_NAME := bon3

build:
	go build -o $(BINARY_NAME)

install:
	go build -o $(BINARY_NAME)
	mv $(BINARY_NAME) $(GOPATH)/bin/

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
