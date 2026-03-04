.PHONY: build run dev docker-build docker-up docker-down clean frontend

# Build the Go binary
build:
	go build -o bin/apisto ./cmd/apisto/

# Run locally (requires TimescaleDB and Mosquitto running)
run: build
	./bin/apisto

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Build the frontend
frontend:
	cd web && npm install && npm run build

# Build Docker image
docker-build:
	docker build -t apisto .

# Start full stack with Docker Compose
docker-up:
	docker compose up -d

# Stop Docker Compose stack
docker-down:
	docker compose down

# Stop and remove volumes
docker-clean:
	docker compose down -v

# View logs
logs:
	docker compose logs -f

# Run Go tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...
	gofmt -s -w .

# Tidy modules
tidy:
	go mod tidy

clean:
	rm -rf bin/ web/dist/
