# ── Build stage ──────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cài đặt dependencies trước để tận dụng Docker layer cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build với tối ưu hóa cho production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/bin/server ./cmd/server

# ── Runtime stage ─────────────────────────────────────
# Dùng distroless để giảm attack surface và kích thước image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary từ build stage
COPY --from=builder /app/bin/server .

EXPOSE 8080

ENTRYPOINT ["/app/server"]
