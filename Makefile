.PHONY: run-api stop-api run-web stop-web run-mobile stop-mobile run-all stop-all help

DEVICE ?=

# ============================================================
# sa-api (Backend API)
# ============================================================

run-api:
	@if pgrep -f "go-build.*server" > /dev/null 2>&1; then \
		echo "sa-api is already running, skipping..."; \
	else \
		echo "Starting sa-api (Go)..."; \
		cd sa-api && docker-compose up -d postgres redis; \
		echo "Waiting for PostgreSQL & Redis..."; \
		sleep 3; \
		cd sa-api && go run ./cmd/server & \
		echo "==> sa-api running at http://localhost:8080"; \
	fi

stop-api:
	@echo "Stopping sa-api..."
	@-pkill -f "go run ./cmd/server" 2>/dev/null
	@-pkill -f "go-build.*server" 2>/dev/null
	@cd sa-api && docker-compose down
	@echo "sa-api stopped."

# ============================================================
# sa-web (Admin Portal) — depends on sa-api
# ============================================================

run-web: run-api
	@echo "Starting sa-web (Next.js)..."
	@cd sa-web && npm run dev &
	@echo "==> sa-web running at http://localhost:3000"

stop-web:
	@echo "Stopping sa-web..."
	@-pkill -f "next dev" 2>/dev/null
	@-pkill -f "next-server" 2>/dev/null
	@echo "sa-web stopped."

# ============================================================
# sa-mb (Mobile App) — depends on sa-api
# ============================================================

run-mobile: run-api
	@echo "Starting sa-mb (Flutter)..."
	@cd sa-mb && flutter run $(if $(DEVICE),-d $(DEVICE),) &
	@echo "==> sa-mb running on connected device/emulator"

stop-mobile:
	@echo "Stopping sa-mb..."
	@-pkill -f "flutter run" 2>/dev/null
	@-pkill -f "flutter_tools" 2>/dev/null
	@echo "sa-mb stopped."

# ============================================================
# Run / Stop all
# ============================================================

run-all: run-api run-web run-mobile

stop-all: stop-web stop-mobile stop-api

# ============================================================
# Help
# ============================================================

help:
	@echo "Usage:"
	@echo "  make run-api        - Start sa-api only"
	@echo "  make stop-api       - Stop sa-api + Docker services"
	@echo "  make run-web        - Start sa-api + sa-web"
	@echo "  make stop-web       - Stop sa-web only"
	@echo "  make run-mobile     - Start sa-api + sa-mb (auto-detect device)"
	@echo "  make stop-mobile    - Stop sa-mb only"
	@echo "  make run-all        - Start everything"
	@echo "  make stop-all       - Stop everything"
	@echo ""
	@echo "Options:"
	@echo "  DEVICE=<id>         - Specify Flutter device (e.g. make run-mobile DEVICE=chrome)"
