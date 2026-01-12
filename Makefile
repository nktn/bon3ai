.PHONY: build install test clean

BINARY_NAME := bon3

build:
	go build -o $(BINARY_NAME)

install:
	go build -o $(shell go env GOPATH)/bin/$(BINARY_NAME)

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
