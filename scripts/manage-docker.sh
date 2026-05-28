#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
APP_PORT="${APP_PORT:-8000}"
APP_URL="${APP_URL:-http://127.0.0.1:${APP_PORT}}"
READY_TIMEOUT="${READY_TIMEOUT:-60}"
READY_INTERVAL="${READY_INTERVAL:-2}"

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

usage() {
  cat <<'EOF'
Usage: scripts/manage-docker.sh [command] [args]

Commands:
  start|up       Build and start app + PostgreSQL in background (default)
  stop|down      Stop and remove containers
  restart        Stop, rebuild and start services
  status|ps      Show service status
  logs           Follow service logs
  build          Build the app image
  ready          Check /readyz
  psql           Open psql in the PostgreSQL container
  clean          Stop services and remove Compose volumes
  help           Show this help

Extra args are passed to the underlying docker compose command.
EOF
}

check_ready() {
  local deadline=$((SECONDS + READY_TIMEOUT))
  local status=1

  until curl -fsS "${APP_URL}/readyz"; do
    status=$?
    if (( SECONDS >= deadline )); then
      printf '\nready check timed out after %ss: %s/readyz\n' "$READY_TIMEOUT" "$APP_URL" >&2
      return "$status"
    fi
    sleep "$READY_INTERVAL"
  done
  printf '\n'
}

cmd="${1:-start}"
if [[ $# -gt 0 ]]; then
  shift
fi

case "$cmd" in
  start|up)
    compose up -d --build --remove-orphans "$@"
    ;;
  stop|down)
    compose down --remove-orphans "$@"
    ;;
  restart)
    compose down --remove-orphans
    compose up -d --build --remove-orphans "$@"
    ;;
  status|ps)
    compose ps "$@"
    ;;
  logs)
    compose logs -f --tail=100 "$@"
    ;;
  build)
    compose build "$@"
    ;;
  ready|health)
    check_ready
    ;;
  psql|db)
    compose exec postgres psql -U saas_template -d saas_template "$@"
    ;;
  clean)
    compose down -v --remove-orphans "$@"
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
