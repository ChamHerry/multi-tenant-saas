.PHONY: dev web build test test-unit test-integration test-e2e

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
	go test -tags=integration -count=1 -timeout 120s -v ./internal/testutil/integration/...

# All tests.
test: test-unit test-integration

# E2E tests via bash script.
test-e2e:
	@docker compose up -d postgres && sleep 2
	bash hack/e2e_full_flow.sh
