#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# ── Configuration (env-overridable) ──────────────────────────────────────────
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
COMPOSE_DEV_FILE="${COMPOSE_DEV_FILE:-docker-compose.dev.yml}"
PROJECT_NAME="${PROJECT_NAME:-mtsaas}"
STATE_DIR="${STATE_DIR:-$ROOT_DIR/.local}"
COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:-$STATE_DIR/compose.env}"

APP_PORT="${APP_PORT:-8000}"
APP_URL="${APP_URL:-http://127.0.0.1:${APP_PORT}}"
POSTGRES_HOST_PORT="${POSTGRES_HOST_PORT:-55432}"
POSTGRES_USER="${POSTGRES_USER:-saas_template}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-secret}"
POSTGRES_DB="${POSTGRES_DB:-saas_template}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"

# ── Helpers ──────────────────────────────────────────────────────────────────

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

compose_dev() {
  docker compose -f "$COMPOSE_FILE" -f "$COMPOSE_DEV_FILE" "$@"
}

prepare_env() {
  mkdir -p "$STATE_DIR"
  cat > "$COMPOSE_ENV_FILE" <<EOF
APP_PORT=$APP_PORT
EOF
}

check_ready() {
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl not found; skip health check."
    return 0
  fi

  local deadline=$((SECONDS + READY_TIMEOUT))

  echo "Waiting for app (timeout ${READY_TIMEOUT}s)..."
  until curl -fsS "${APP_URL}/readyz" 2>/dev/null; do
    if (( SECONDS >= deadline )); then
      printf '\nready check timed out after %ss: %s/readyz\n' "$READY_TIMEOUT" "$APP_URL" >&2
      compose ps >&2 || true
      echo "Recent app logs:" >&2
      compose logs --tail=60 app >&2 || true
      return 1
    fi
    sleep "$READY_INTERVAL"
  done
  printf '\n'
}

print_summary() {
  echo ""
  echo "============================================"
  echo "  multi-tenant-saas local stack"
  echo "============================================"
  echo "  App:         ${APP_URL}  (API + SPA)"
  echo "  Health:      ${APP_URL}/readyz"
  echo "  Swagger:     ${APP_URL}/swagger"
  echo "  PostgreSQL:  postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:${POSTGRES_HOST_PORT}/${POSTGRES_DB}"
  echo "  Compose:     ${COMPOSE_FILE}"
  echo "  Env file:    ${COMPOSE_ENV_FILE}"
  echo "  State dir:   ${STATE_DIR}"
  echo "============================================"
}

print_dev_summary() {
  echo ""
  echo "============================================"
  echo "  multi-tenant-saas DEV stack"
  echo "============================================"
  echo "  App:         ${APP_URL}  (API + SPA — single port)"
  echo "  Health:      ${APP_URL}/readyz"
  echo "  Swagger:     ${APP_URL}/swagger"
  echo "  PostgreSQL:  postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:${POSTGRES_HOST_PORT}/${POSTGRES_DB}"
  echo ""
  echo "  Source mounted at /app inside the dev container."
  echo "  Go backend serves both API and built SPA (web/dist)."
  echo ""
  echo "  After Go changes:    bash scripts/manage-docker.sh dev-restart"
  echo "  After SPA changes:   bash scripts/manage-docker.sh dev-rebuild-spa"
  echo "============================================"
}

# ── Usage ────────────────────────────────────────────────────────────────────

usage() {
  cat <<'EOF'
Usage: scripts/manage-docker.sh [command] [args]

Commands:
  start|up         Build and start app + PostgreSQL in background (default)
  dev              Fast dev mode: all-in-Docker with live source mount
  dev-down         Stop dev stack
  dev-logs         Follow dev container logs
  dev-restart      Restart Go backend (picks up Go source changes)
  dev-rebuild-spa  Rebuild SPA inside dev container (after frontend changes)
  stop|down        Stop and remove containers
  restart          Stop, rebuild and start (production mode)
  status|ps        Show service status
  logs             Follow service logs (pass service names to filter)
  build            Build Docker images
  ready            Check /readyz
  psql             Open psql in the PostgreSQL container
  clean            Stop containers and remove volumes
  help             Show this help

Fast dev mode (dev):
  Starts PostgreSQL + a dev container with the full project mounted at /app.
  The Go backend compiles on startup and serves both the API and the built
  SPA on a single port. No separate Vite process — the SPA is built once
  and served as static files by Go (see internal/cmd/static.go).

  The dev container uses manifest/docker/config.yaml so the DB connection
  points to postgres:5432 on the Docker network.

  The regular 'app' service is stopped when dev mode starts (port conflict).

Environment overrides:
  APP_PORT=8000              App HTTP port
  POSTGRES_HOST_PORT=55432   PostgreSQL host port
  COMPOSE_FILE               Path to docker-compose file
  COMPOSE_DEV_FILE           Path to dev compose overlay
  STATE_DIR                  State/config directory (default: .local)
  READY_TIMEOUT=120          Health check timeout (seconds; higher for first build)
EOF
}

# ── Subcommand dispatch ──────────────────────────────────────────────────────

cmd="${1:-start}"
if [[ $# -gt 0 ]]; then shift; fi

# Pass-through to production compose.
case "$cmd" in
  down|stop)
    compose down --remove-orphans "$@"
    exit $?
    ;;
  logs)
    compose logs -f --tail=100 "$@"
    exit $?
    ;;
  ps|status)
    compose ps "$@"
    exit $?
    ;;
  pull|config)
    compose "$cmd" "$@"
    exit $?
    ;;
esac

# Commands with custom logic.
case "$cmd" in
  # ── Production (full Docker build) ──
  start|up)
    prepare_env
    compose --env-file "$COMPOSE_ENV_FILE" up -d --build --remove-orphans "$@"
    print_summary
    check_ready
    ;;

  restart)
    compose down --remove-orphans
    prepare_env
    compose --env-file "$COMPOSE_ENV_FILE" up -d --build --remove-orphans "$@"
    print_summary
    check_ready
    ;;

  build)
    compose build "$@"
    ;;

  # ── Dev mode (all-in-Docker with live source mount) ──
  dev)
    prepare_env

    echo "==> Building dev image (Go + Node)..."
    compose_dev build dev

    echo "==> Starting PostgreSQL + dev container..."
    # Start postgres and dev; --remove-orphans stops the prod 'app' if running
    compose_dev --env-file "$COMPOSE_ENV_FILE" up -d --remove-orphans postgres dev "$@"

    print_dev_summary

    echo "Waiting for Go backend (first build includes SPA + Go compilation)..."
    _dev_deadline=$((SECONDS + READY_TIMEOUT))
    until curl -fsS "${APP_URL}/readyz" 2>/dev/null; do
      if (( SECONDS >= _dev_deadline )); then
        printf '\nGo backend not ready after %ss. Recent dev logs:\n' "$READY_TIMEOUT" >&2
        compose_dev logs --tail=60 dev >&2 || true
        echo "It may still be compiling. Check: bash scripts/manage-docker.sh dev-logs" >&2
        break
      fi
      sleep "$READY_INTERVAL"
    done
    printf '\n'
    echo "Go backend ready: ${APP_URL}/readyz"
    ;;

  dev-down)
    compose_dev down --remove-orphans "$@"
    echo "Dev stack stopped."
    ;;

  dev-logs)
    compose_dev logs -f --tail=100 dev "$@"
    ;;

  dev-restart)
    compose_dev restart dev "$@"
    echo "Dev container restarted (Go will recompile on startup)."
    ;;

  dev-rebuild-spa)
    echo "==> Rebuilding SPA inside dev container..."
    compose_dev exec dev npm run build --prefix /app/web
    echo "SPA rebuilt. Go will serve the updated files immediately."
    ;;

  ready|health)
    check_ready
    ;;

  psql|db)
    # Try dev container first, fall back to prod
    if compose_dev ps --status running dev 2>/dev/null | grep -q dev; then
      compose_dev exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" "$@"
    else
      compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" "$@"
    fi
    ;;

  clean)
    compose down -v --remove-orphans "$@" 2>/dev/null || true
    compose_dev down -v --remove-orphans "$@" 2>/dev/null || true
    rm -rf "$STATE_DIR"
    echo "Removed containers, volumes, and state dir ($STATE_DIR)."
    ;;

  help|-h|--help)
    usage
    ;;

  *)
    usage >&2
    exit 2
    ;;
esac
