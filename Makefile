run: build
	@./bin/mura

build:
	@go build -o bin/mura main.go
