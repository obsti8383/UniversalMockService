.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: fmt
	go vet ./...
.PHONY:vet

build: vet
	GOOS=linux GOARCH=amd64 go build .
	GOOS=windows GOARCH=amd64 go build .
.PHONY:build

