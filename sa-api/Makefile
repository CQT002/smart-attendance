.PHONY: run build test docker-up docker-down migrate tidy

# Chạy server ở môi trường development
run:
	go run ./cmd/server

# Build binary
build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

# Chạy tests
test:
	go test ./... -v -race -coverprofile=coverage.out

# Xem coverage
coverage: test
	go tool cover -html=coverage.out

# Khởi động infrastructure (PostgreSQL + Redis) bằng Docker
docker-up:
	docker-compose up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "PostgreSQL and Redis are ready!"

# Dừng tất cả Docker containers
docker-down:
	docker-compose down

# Khởi động toàn bộ stack (bao gồm API)
docker-all:
	docker-compose up -d

# Tidy go modules
tidy:
	go mod tidy

# Kiểm tra lỗi với golangci-lint
lint:
	golangci-lint run ./...

# Tạo admin user mặc định (chạy sau khi migrate)
seed:
	@echo "Seeding default admin user..."
	@echo "Email: admin@hdbank.com.vn | Password: Admin@123"

# Xem logs của API container
logs:
	docker-compose logs -f api

# Help
help:
	@echo "Available commands:"
	@echo "  make run         - Chạy server development"
	@echo "  make build       - Build binary"
	@echo "  make test        - Chạy unit tests"
	@echo "  make docker-up   - Khởi động PostgreSQL + Redis"
	@echo "  make docker-down - Dừng Docker containers"
	@echo "  make docker-all  - Khởi động toàn bộ stack"
	@echo "  make tidy        - Tidy go modules"
	@echo "  make lint        - Kiểm tra lỗi code"
