.DEFAULT_GOAL := all

SRCS := $(shell find . -type f -name '*.go' | grep -v './vendor')
LDFLAGS := -ldflags="-s -w -extldflags \"-static\""
BUILDFLAGS := -a -tags netgo -installsuffix netgo

.PHONY: test clean all

test:
	go test -cover -v ./...

clean:
	rm -rf ./bin

all: bin/linux/amd64/eth-detector bin/linux/arm64/eth-detector

bin/linux/amd64/eth-detector: $(SRCS)
	mkdir -p bin/linux/amd64
	GOOS=linux GOARCH=amd64 go build $(BUILDFLAGS) $(LDFLAGS) -o bin/linux/amd64/eth-detector main.go

bin/linux/arm64/eth-detector: $(SRCS)
	mkdir -p bin/linux/arm64
	GOOS=linux GOARCH=arm64 go build $(BUILDFLAGS) $(LDFLAGS) -o bin/linux/arm64/eth-detector main.go

