.PHONY: dev web build test test-unit test-integration test-e2e test-e2e-go

dev:
	go run .

web:
	cd web && npm run dev

build:
	cd web && npm run build
	gf build -ew

# Unit tests (no database required).
test-unit:
	go test -short -count=1 ./internal/... ./utility/...

# Integration tests (requires docker-compose postgres on port 55432).
test-integration:
	@docker compose up -d postgres && sleep 2
	go test -tags=integration -count=1 -timeout 5m -v ./internal/testutil/integration/...

# All tests.
test: test-unit test-integration

# Go E2E tests (requires docker-compose postgres on port 55432).
test-e2e-go:
	@docker compose up -d postgres && sleep 2
	go test -tags=e2e -p=1 -count=1 -timeout 15m -v ./test/e2e/...

# Default E2E entrypoint now uses the Go test harness.
test-e2e: test-e2e-go
