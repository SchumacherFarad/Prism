.PHONY: all build run clean test dev docker-build docker-up docker-down

# Default target
all: build

# Build backend
build:
	cd backend && go build -o bin/prism ./cmd/prism

# Run backend locally
run: build
	cd backend && ./bin/prism

# Run backend in development mode
dev:
	cd backend && go run ./cmd/prism

# Run frontend development server
dev-frontend:
	cd frontend && npm run dev

# Run both backend and frontend in development
dev-all:
	$(MAKE) -j2 dev dev-frontend

# Clean build artifacts
clean:
	rm -rf backend/bin
	rm -rf frontend/.next
	rm -rf frontend/node_modules

# Run tests
test:
	cd backend && go test ./...
	cd frontend && npm test

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Install dependencies
deps:
	cd backend && go mod download
	cd frontend && npm install

# Format code
fmt:
	cd backend && go fmt ./...

# Lint code
lint:
	cd backend && go vet ./...
