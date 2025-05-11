# Variables
BINARY_NAME=wallet-app
MAIN_FILE=cmd/server/main.go

# Phony targets
.PHONY: all build run test clean

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

# Run the application
run:
	go run $(MAIN_FILE)

# Run tests
test:
	go test ./...

# Clean up generated files
clean:
	rm -f $(BINARY_NAME)
