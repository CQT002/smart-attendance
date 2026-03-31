.PHONY: run-web stop-web run-mobile stop-mobile help

# ============================================================
# sa-api + sa-web (Backend API + Admin Portal)
# ============================================================

run-web:
	@echo "Starting sa-api (Go) + sa-web (Next.js)..."
	@cd sa-api && docker-compose up -d postgres redis
	@echo "Waiting for PostgreSQL & Redis..."
	@sleep 3
	@cd sa-api && go run ./cmd/server &
	@cd sa-web && npm run dev &
	@echo ""
	@echo "==> sa-api running at http://localhost:8080"
	@echo "==> sa-web running at http://localhost:3000"

stop-web:
	@echo "Stopping sa-api + sa-web..."
	@-pkill -f "go run ./cmd/server" 2>/dev/null
	@-pkill -f "go-build.*server" 2>/dev/null
	@-pkill -f "next dev" 2>/dev/null
	@-pkill -f "next-server" 2>/dev/null
	@cd sa-api && docker-compose down
	@echo "All stopped."

# ============================================================
# sa-api + sa-mb (Backend API + Mobile App)
# ============================================================

run-mobile:
	@echo "Starting sa-api (Go) + sa-mb (Flutter)..."
	@cd sa-api && docker-compose up -d postgres redis
	@echo "Waiting for PostgreSQL & Redis..."
	@sleep 3
	@cd sa-api && go run ./cmd/server &
	@cd sa-mb && flutter run -d macos &
	@echo ""
	@echo "==> sa-api running at http://localhost:8080"
	@echo "==> sa-mb running on connected device/emulator"

stop-mobile:
	@echo "Stopping sa-api + sa-mb..."
	@-pkill -f "go run ./cmd/server" 2>/dev/null
	@-pkill -f "go-build.*server" 2>/dev/null
	@-pkill -f "flutter run" 2>/dev/null
	@-pkill -f "flutter_tools" 2>/dev/null
	@cd sa-api && docker-compose down
	@echo "All stopped."

# ============================================================
# Help
# ============================================================

help:
	@echo "Usage:"
	@echo "  make run-web      - Start sa-api + sa-web"
	@echo "  make stop-web     - Stop sa-api + sa-web"
	@echo "  make run-mobile   - Start sa-api + sa-mb"
	@echo "  make stop-mobile  - Stop sa-api + sa-mb"
