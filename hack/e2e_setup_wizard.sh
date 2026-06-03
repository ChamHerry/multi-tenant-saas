#!/usr/bin/env bash
# e2e_setup_wizard.sh — Docker-backed end-to-end test for the first-run setup wizard.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-180}"
READY_INTERVAL="${READY_INTERVAL:-2}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
APP_SERVICE="${APP_SERVICE:-app}"
DB_SERVICE="${DB_SERVICE:-postgres}"
REDIS_SERVICE="${REDIS_SERVICE:-redis}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-setup-wizard}"
SCENARIO_LOG="${LOG_DIR}/setup-wizard.log"
FAILURE_LOG="${LOG_DIR}/compose-failure.log"
mkdir -p "$LOG_DIR"
: >"$SCENARIO_LOG"
: >"$FAILURE_LOG"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }; }
need docker
need curl
need python3
need bash

log() { echo "$*" | tee -a "$SCENARIO_LOG"; }
compose() { docker compose -f "$COMPOSE_FILE" "$@"; }
manage_docker() { COMPOSE_FILE="$COMPOSE_FILE" APP_PORT="$APP_PORT" READY_TIMEOUT="$READY_TIMEOUT" READY_INTERVAL="$READY_INTERVAL" ./scripts/manage-docker.sh "$@"; }
psql_query() { compose exec -T -e PGPASSWORD=secret "$DB_SERVICE" psql -U saas_template -d saas_template -Atc "$1"; }
redis_del() { compose exec -T "$REDIS_SERVICE" redis-cli DEL "$@" >/dev/null 2>&1 || true; }

print_failure_context() {
  {
    echo "--- docker compose ps ---"
    compose ps || true
    echo
    echo "--- docker compose logs (tail=180) ---"
    compose logs --tail=180 "$APP_SERVICE" "$DB_SERVICE" "$REDIS_SERVICE" || true
  } >"$FAILURE_LOG" 2>&1
  cat "$FAILURE_LOG" >&2 || true
}

cleanup() {
  local status=$?
  if (( status != 0 )); then
    log "[E2E] failure detected, dumping compose context"
    print_failure_context
  fi
}
trap cleanup EXIT

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

ensure_stack_ready() {
  case "$E2E_DOCKER_ACTION" in
    skip) log "[E2E] reuse existing docker stack" ;;
    start|up|restart) log "[E2E] manage-docker action: $E2E_DOCKER_ACTION"; manage_docker "$E2E_DOCKER_ACTION" ;;
    *) echo "unsupported E2E_DOCKER_ACTION=$E2E_DOCKER_ACTION (expected start|up|restart|skip)" >&2; exit 2 ;;
  esac
  log "[E2E] wait for /readyz"
  manage_docker ready >/dev/null
}

http_request() {
  local method="$1" path="$2" body="$3" outfile="$4"
  shift 4
  local args=(-sS -o "$outfile" -w "%{http_code}" -X "$method")
  if [[ -n "$body" ]]; then args+=(-H "Content-Type: application/json" -d "$body"); fi
  args+=("$@" "${BASE_URL}${path}")
  curl "${args[@]}"
}

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j = json.load(f)
value = eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len, "any": any, "all": all, "sum": sum})
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
else:
    print(value)
PY
}

assert_status() {
  local id="$1" expected="$2" actual="$3" bodyfile="$4"
  if [[ "$actual" != "$expected" ]]; then
    log "[FAIL] ${id}: expected HTTP ${expected}, got ${actual}; body=$(cat "$bodyfile")"
    exit 1
  fi
  log "[PASS] ${id}: HTTP ${actual}"
}

assert_json_equals() {
  local id="$1" file="$2" expr="$3" expected="$4"
  local actual
  actual="$(json_value "$file" "$expr")"
  if [[ "$actual" != "$expected" ]]; then
    log "[FAIL] ${id}: expected ${expr}=${expected}, got ${actual}; body=$(cat "$file")"
    exit 1
  fi
  log "[PASS] ${id}: ${expr}=${actual}"
}

assert_body_not_contains() {
  local id="$1" file="$2" needle="$3"
  if [[ -n "$needle" ]] && grep -Fq "$needle" "$file"; then
    log "[FAIL] ${id}: response leaked forbidden value"
    exit 1
  fi
  log "[PASS] ${id}: forbidden value absent"
}

csrf_from_cookie_jar() {
  local jar="$1"
  awk '$0 !~ /^#/ && $6 == "saas_template_csrf" {print $7}' "$jar" | tail -1
}

reset_to_uninitialized() {
  log "[E2E] reset database to uninitialized setup state"
  psql_query "DELETE FROM public.system_setup;" >/dev/null
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('web.baseUrl','http://127.0.0.1:5173','string','Frontend base URL for invitation links',false,now(),now()) ON CONFLICT(key) DO UPDATE SET value='http://127.0.0.1:5173', value_type='string', is_encrypted=false, updated_at=now();" >/dev/null
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('auth.session.cookie.secure','false','bool','HTTP e2e cookie override',false,now(),now()) ON CONFLICT(key) DO UPDATE SET value='false', value_type='bool', is_encrypted=false, updated_at=now();" >/dev/null
  psql_query "DELETE FROM public.system_config WHERE key IN ('auth.session.secret','auth.apiKey.secret');" >/dev/null
  redis_del config:web.baseUrl config:auth.session.cookie.secure config:auth.session.secret config:auth.apiKey.secret
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
admin_email="setup-admin-${stamp}@example.test"
admin_password="SetupPassword12345!"
admin_cookie="${LOG_DIR}/admin.cookies"

ensure_stack_ready
reset_to_uninitialized

log "[E2E] uninitialized state is visible and business APIs are blocked"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "$b")"; assert_status READYZ_SETUP_REQUIRED 200 "$s" "$b"
assert_json_equals READYZ_SETUP_CODE "$b" 'j.get("code")' SYSTEM_SETUP_REQUIRED
b="$(bodyfile)"; s="$(http_request GET /api/v1/setup/state "" "$b")"; assert_status SETUP_STATE 200 "$s" "$b"
assert_json_equals SETUP_STATE_REQUIRES "$b" 'j["data"]["requires_setup"]' true
assert_json_equals SETUP_STATE_INITIALIZED "$b" 'j["data"]["initialized"]' false
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"blocked-${stamp}@example.test\",\"password\":\"${admin_password}\",\"display_name\":\"Blocked\"}" "$b")"; assert_status REGISTER_BLOCKED 503 "$s" "$b"
assert_json_equals REGISTER_BLOCKED_CODE "$b" 'j["code"]' SYSTEM_SETUP_REQUIRED

log "[E2E] complete first-run setup"
setup_payload="{\"admin\":{\"email\":\"${admin_email}\",\"password\":\"${admin_password}\",\"display_name\":\"Setup Admin\"},\"runtime\":{\"web_base_url\":\"${BASE_URL}\",\"generate_session_secret\":true,\"generate_api_key_secret\":true}}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/setup/complete "$setup_payload" "$b")"; assert_status SETUP_COMPLETE 200 "$s" "$b"
assert_json_equals SETUP_COMPLETE_INITIALIZED "$b" 'j["data"]["initialized"]' true
assert_json_equals SETUP_COMPLETE_EMAIL "$b" 'j["data"]["user"]["email"]' "$admin_email"
assert_body_not_contains SETUP_RESPONSE_NO_PASSWORD "$b" "$admin_password"
assert_body_not_contains SETUP_RESPONSE_NO_CHANGE_ME "$b" "change-me"

log "[E2E] setup completion persisted lock and encrypted secrets"
[[ "$(psql_query "SELECT COUNT(*) FROM public.system_setup WHERE id=1 AND status='initialized';")" == "1" ]] || { log "[FAIL] system_setup row missing"; exit 1; }
[[ "$(psql_query "SELECT value_type || ':' || is_encrypted::text FROM public.system_config WHERE key='auth.session.secret';")" == "secret:true" ]] || { log "[FAIL] session secret not encrypted"; exit 1; }
[[ "$(psql_query "SELECT value_type || ':' || is_encrypted::text FROM public.system_config WHERE key='auth.apiKey.secret';")" == "secret:true" ]] || { log "[FAIL] api key secret not encrypted"; exit 1; }
[[ "$(psql_query "SELECT COUNT(*) FROM public.system_config WHERE key IN ('auth.session.secret','auth.apiKey.secret') AND value ILIKE '%change-me%';")" == "0" ]] || { log "[FAIL] generated secrets still contain change-me"; exit 1; }

log "[E2E] second setup submit is rejected"
b="$(bodyfile)"; s="$(http_request POST /api/v1/setup/complete "$setup_payload" "$b")"; assert_status SETUP_SECOND_REJECTED 409 "$s" "$b"

log "[E2E] initialized admin can log in and read masked config"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${admin_email}\",\"password\":\"${admin_password}\"}" "$b" -c "$admin_cookie")"; assert_status ADMIN_LOGIN 200 "$s" "$b"
csrf_token="$(csrf_from_cookie_jar "$admin_cookie")"
[[ -n "$csrf_token" ]] || { log "[FAIL] missing CSRF cookie after login"; exit 1; }
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/system-config/auth.session.secret "" "$b" -b "$admin_cookie")"; assert_status ADMIN_GET_SESSION_SECRET 200 "$s" "$b"
assert_json_equals ADMIN_SECRET_MASKED "$b" 'j["data"]["config"]["masked_value"]' '********'
assert_json_equals ADMIN_SECRET_VALUE_EMPTY "$b" 'j["data"]["config"]["value"]' ''

b="$(bodyfile)"; s="$(http_request GET /api/v1/setup/state "" "$b")"; assert_status SETUP_STATE_AFTER_COMPLETE 200 "$s" "$b"
assert_json_equals SETUP_STATE_AFTER_COMPLETE_INITIALIZED "$b" 'j["data"]["initialized"]' true
assert_json_equals SETUP_STATE_AFTER_COMPLETE_REQUIRES "$b" 'j["data"]["requires_setup"]' false

log "[E2E] setup-wizard scenarios passed admin=${admin_email}"
