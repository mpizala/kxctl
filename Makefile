# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
BINARY_NAME=kxctl
BINARY_PATH=cmd/$(BINARY_NAME)/

.PHONY: all build clean test fmt vet install

all: vet fmt test build

build:
	$(GOBUILD) -o $(BINARY_NAME) ./$(BINARY_PATH)

test:
	$(GOTEST) -v ./...

clean:
	rm -f $(BINARY_NAME)

fmt:
	$(GOFMT) ./...

vet:
	$(GOVET) ./...

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
