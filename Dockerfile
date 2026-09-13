# Stage 1: Build Frontend (Vue 3 + Vite)
FROM node:24-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 2: Build Backend (Go + Embed Frontend)
FROM golang:alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Salin hasil build frontend dari stage 1 ke folder web/dist
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/neraca ./cmd/server

# Stage 3: Minimal Production Image (< 30 MB)
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

RUN mkdir -p /app/data

COPY --from=backend-builder /app/neraca /app/neraca

ENV PORT=8088
ENV DB_PATH=/app/data/neraca.db
ENV ENV=production

EXPOSE 8088
VOLUME ["/app/data"]

CMD ["/app/neraca"]
