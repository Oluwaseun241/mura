run: build
	@./bin/mura

build:
	@go build -o bin/mura main.go

test:
	@echo "Testing..."
	@go test ./cmd/api -v
