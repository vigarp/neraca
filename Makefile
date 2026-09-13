.PHONY: all dev-backend dev-frontend build build-frontend build-backend run \
        test-backend test-frontend test \
        lint-backend lint-frontend lint \
        format-backend format-frontend format format-check \
        docker-build docker-up docker-down clean

# Environment setup
GOPATH ?= /usr/local/go/bin
PATH := $(GOPATH):$(PATH)

all: lint test build

# Menjalankan backend Go (dengan auto reload jika ada air, atau go run biasa)
dev-backend:
	go run ./cmd/server

# Menjalankan Vite dev server untuk frontend (HMR di port 5173)
dev-frontend:
	cd web && npm run dev

# Package list excluding node_modules
PKGS = $$(go list ./... | grep -v '/node_modules/')

# Testing
test-backend:
	go test -v -race $(PKGS)

test-frontend:
	cd web && npm run test

test: test-backend test-frontend

# Linting & Static Analysis
lint-backend:
	go vet $(PKGS)

lint-frontend:
	cd web && npm run lint

lint: lint-backend lint-frontend

# Formatting
format-backend:
	gofmt -w -s .

format-frontend:
	cd web && npm run format

format: format-backend format-frontend

format-check:
	@echo "Checking Go formatting..."
	@test -z "$$(gofmt -s -l . | grep -v 'web/dist' | tee /dev/stderr)"
	@echo "Checking Frontend formatting..."
	cd web && npm run format:check

# Build frontend Vue ke folder web/dist
build-frontend:
	cd web && npm run build

# Build binary Go (meng-embed hasil build web/dist)
build-backend:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/neraca ./cmd/server

# Build keseluruhan (frontend + backend embed)
build: build-frontend build-backend

# Jalankan hasil binary lokal
run: build
	./bin/neraca

# Docker commands
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/
	rm -rf web/dist/
