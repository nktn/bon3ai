.PHONY: build clean test run install uninstall

BINARY_NAME=bon3ai
BUILD_DIR=bin
INSTALL_DIR=/usr/local/bin

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v ./...

install: build
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

uninstall:
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)

.DEFAULT_GOAL := build
