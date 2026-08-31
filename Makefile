VERSION=0.0.16
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"
all: mackerel-plugin-log-counter

.PHONY: mackerel-plugin-linux-process-status linux check lint

mackerel-plugin-log-counter: cmd/mackerel-plugin-log-counter/*.go
	go build $(LDFLAGS) -o mackerel-plugin-log-counter ./...

linux: cmd/mackerel-plugin-log-counter/*.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-log-counter ./...

check:
	go test -v ./...
	go test -race ./...

lint:
	golangci-lint run ./...