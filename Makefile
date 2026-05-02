BINARY=poly-switch

.PHONY: all build run clean vet

all: build

build:
	go build -o $(BINARY) ./cmd/poly-switch/

run:
	go run ./cmd/poly-switch/

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
