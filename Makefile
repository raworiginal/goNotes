.PHONY: build run clean help

build:
	go build -o bin/goNotes ./cmd/api

run:
	go run ./cmd/api

clean:
	rm -rf bin/

help:
	@echo "Available targets:"
	@echo "  make build    - Build the project to bin/goNotes"
	@echo "  make run      - Run the project with go run"
	@echo "  make clean    - Remove the bin/ directory"
	@echo "  make help     - Show this help message"
