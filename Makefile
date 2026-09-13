.PHONY: all dev-backend dev-frontend build build-frontend build-backend run docker-build docker-up docker-down clean

# Environment setup
GOPATH ?= /usr/local/go/bin
PATH := $(GOPATH):$(PATH)

all: build

# Menjalankan backend Go (dengan auto reload jika ada air, atau go run biasa)
dev-backend:
	go run ./cmd/server

# Menjalankan Vite dev server untuk frontend (HMR di port 5173)
dev-frontend:
	cd web && npm run dev

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
