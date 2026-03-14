.PHONY: build dev test clean dashboard

# Build the dashboard then the Go binary with embedded frontend
build: dashboard
	go build -o maguro cmd/api/main.go

# Build just the frontend
dashboard:
	cd maguro-dashboard && npm install --silent && npm run build

# Development: run API server (requires pre-built dashboard or use dev-frontend separately)
dev:
	go run cmd/api/main.go

# Development: run frontend with hot reload (proxies API to :8080)
dev-frontend:
	cd maguro-dashboard && npm run dev

# Run all backend tests
test:
	go test ./internal/tests/... -v -count=1

# Run crypto unit tests
test-crypto:
	go test ./internal/crypto/... -v

# Clean build artifacts
clean:
	rm -f maguro
	rm -rf maguro-dashboard/dist
