.PHONY: all gofmt golint govet test

all: gofmt golint govet test cover

gofmt:
	gofmt -s=true -d=true -l=true .

golint:
	golint .

govet:
	go tool vet -all -v=true *.go

test:
	go test -v -race -cpu=1,2,4 -coverprofile=coverage.out -covermode=atomic

cover:
	go tool cover -html=coverage.out -o coverage.html
