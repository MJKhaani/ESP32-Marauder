.PHONY: all build build-frontend test fmt vet clean

all: test build

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -v ./...

build-frontend:
	cd frontend && npm install && npm run build

build: build-frontend
	go build -o marauder-studio main.go
	go build -o marauder-cli ./cmd/marauder-studio/main.go

clean:
	rm -f marauder-studio marauder-cli
	rm -rf frontend/dist
