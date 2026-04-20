.PHONY: build run test fmt vet clean

BINARY := bit-tracker
PKG    := ./...

build:
	go build -o $(BINARY) ./cmd/bit-tracker

run: build
	./$(BINARY)

test:
	go test $(PKG)

fmt:
	gofmt -w .

vet:
	go vet $(PKG)

clean:
	rm -f $(BINARY)
	rm -rf dist
