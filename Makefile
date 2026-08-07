.PHONY: build run install remove uninstall

build:
	@echo "Building btenl..."
	go build -o btenl ./cmd/btenl
	@echo "Building btenld..."
	go build -o btenld ./cmd/btenld
	@echo "Build complete."

run: build
	@echo "Starting daemon..."
	./btenl start

install:
	@echo "Installing btenl..."
	go install ./cmd/btenl
	@echo "Installing btenld..."
	go install ./cmd/btenld
	@echo "Install complete."

remove:
	@echo "Removing local binaries..."
	rm -f btenl btenld
	@echo "Removed."

uninstall:
	@echo "Uninstalling btenl..."
	rm -f $(shell go env GOPATH)/bin/btenl
	@echo "Uninstalling btenld..."
	rm -f $(shell go env GOPATH)/bin/btenld
	@echo "Uninstalled."
