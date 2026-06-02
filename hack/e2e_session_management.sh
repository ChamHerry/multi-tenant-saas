#!/usr/bin/env bash
# e2e_session_management.sh — Docker-backed end-to-end test for user session management.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-session-management}"
SCENARIO_LOG="${LOG_DIR}/session-management.log"
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

compose() { docker compose -f docker-compose.yml "$@"; }
manage_docker() { APP_PORT="$APP_PORT" READY_TIMEOUT="$READY_TIMEOUT" READY_INTERVAL="$READY_INTERVAL" ./scripts/manage-docker.sh "$@"; }

print_failure_context() {
  {
    echo "--- docker compose ps ---"
    compose ps || true
    echo
    echo "--- docker compose logs (tail=160) ---"
    compose logs --tail=160 app postgres redis || true
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

ensure_stack_ready() {
  case "$E2E_DOCKER_ACTION" in
    skip) log "[E2E] reuse existing docker stack" ;;
    start|up|restart) log "[E2E] manage-docker action: $E2E_DOCKER_ACTION"; manage_docker "$E2E_DOCKER_ACTION" ;;
    *) echo "unsupported E2E_DOCKER_ACTION=$E2E_DOCKER_ACTION" >&2; exit 2 ;;
  esac
  log "[E2E] wait for /readyz"
  manage_docker ready >/dev/null
}

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

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
value = eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len, "any": any, "all": all, "next": next, "sum": sum})
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

csrf_from_cookie_jar() {
  local jar="$1"
  awk '$0 !~ /^#/ && $6 == "saas_template_csrf" {print $7}' "$jar" | tail -1
}

session_id_expr() {
  local current_flag="$1"
  if [[ "$current_flag" == "current" ]]; then
    printf '%s' 'next(item["session_id"] for item in j["data"]["sessions"] if item["is_current"])'
  else
    printf '%s' 'next(item["session_id"] for item in j["data"]["sessions"] if not item["is_current"])'
  fi
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
password="SessionPassword123!"
email="session-e2e-${stamp}@example.test"
main_cookie="${LOG_DIR}/main.cookies"
second_cookie="${LOG_DIR}/second.cookies"
third_cookie="${LOG_DIR}/third.cookies"
foreign_cookie="${LOG_DIR}/foreign.cookies"
csrfless_cookie="${LOG_DIR}/csrfless.cookies"

ensure_stack_ready

log "[E2E] register user and create two sessions"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${email}\",\"password\":\"${password}\",\"display_name\":\"Session E2E\"}" "$b" -c "$main_cookie")"; assert_status REGISTER 200 "$s" "$b"
main_csrf="$(csrf_from_cookie_jar "$main_cookie")"
[[ -n "$main_csrf" ]] || { log "[FAIL] missing CSRF cookie after register"; exit 1; }
log "[PASS] REGISTER_CSRF"

b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$second_cookie")"; assert_status SECOND_LOGIN 200 "$s" "$b"

log "[E2E] list active sessions with sanitized response"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/sessions "" "$b" -b "$main_cookie")"; assert_status LIST_SESSIONS 200 "$s" "$b"
assert_json_equals SESSION_COUNT "$b" 'len(j["data"]["sessions"])' 2
assert_json_equals ONE_CURRENT_SESSION "$b" 'sum(1 for item in j["data"]["sessions"] if item["is_current"])' 1
assert_json_equals NO_SECRET_LEAK "$b" 'all("secret_hash" not in item and "csrf_hash" not in item and "token" not in item for item in j["data"]["sessions"])' true
assert_json_equals MASKED_IP "$b" 'all(item["ip"] and item["ip"] != "127.0.0.1" for item in j["data"]["sessions"])' true
assert_json_equals DEVICE_FIELDS "$b" 'all(item["device"]["browser"] and item["device"]["os"] for item in j["data"]["sessions"])' true
current_session_id="$(json_value "$b" "$(session_id_expr current)")"
other_session_id="$(json_value "$b" "$(session_id_expr other)")"

log "[E2E] current session cannot be revoked through /me/sessions/{id}"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/me/sessions/${current_session_id}" "" "$b" -b "$main_cookie" -c "$main_cookie" -H "X-CSRF-Token: ${main_csrf}")"; assert_status CURRENT_REVOKE_REJECTED 400 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$main_cookie")"; assert_status CURRENT_STILL_AUTHENTICATED 200 "$s" "$b"

log "[E2E] revoke one other session and confirm only that browser is invalidated"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/me/sessions/${other_session_id}" "" "$b" -b "$main_cookie" -c "$main_cookie" -H "X-CSRF-Token: ${main_csrf}")"; assert_status REVOKE_OTHER 200 "$s" "$b"
assert_json_equals REVOKE_OTHER_OK "$b" 'j["data"]["ok"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$second_cookie")"; assert_status REVOKED_BROWSER_REJECTED 401 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$main_cookie")"; assert_status CURRENT_SURVIVES_SINGLE_REVOKE 200 "$s" "$b"

log "[E2E] batch revoke all other sessions keeps current session"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$third_cookie")"; assert_status THIRD_LOGIN_FOR_BATCH 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request DELETE /api/v1/me/sessions "" "$b" -b "$main_cookie" -c "$main_cookie" -H "X-CSRF-Token: ${main_csrf}")"; assert_status REVOKE_OTHERS 200 "$s" "$b"
assert_json_equals REVOKE_OTHERS_COUNT "$b" 'j["data"]["revoked_count"]' 1
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$third_cookie")"; assert_status BATCH_REVOKED_BROWSER_REJECTED 401 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$main_cookie")"; assert_status CURRENT_SURVIVES_BATCH_REVOKE 200 "$s" "$b"

log "[E2E] cross-user session revoke is forbidden"
foreign_email="session-foreign-${stamp}@example.test"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${foreign_email}\",\"password\":\"${password}\",\"display_name\":\"Foreign Session E2E\"}" "$b" -c "$foreign_cookie")"; assert_status FOREIGN_REGISTER 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/sessions "" "$b" -b "$foreign_cookie")"; assert_status FOREIGN_LIST 200 "$s" "$b"
foreign_session_id="$(json_value "$b" "$(session_id_expr current)")"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/me/sessions/${foreign_session_id}" "" "$b" -b "$main_cookie" -c "$main_cookie" -H "X-CSRF-Token: ${main_csrf}")"; assert_status FOREIGN_REVOKE_FORBIDDEN 403 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$foreign_cookie")"; assert_status FOREIGN_SESSION_SURVIVES 200 "$s" "$b"

log "[E2E] DELETE requires CSRF and does not revoke target session when missing"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$csrfless_cookie")"; assert_status CSRF_TARGET_LOGIN 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/sessions "" "$b" -b "$main_cookie")"; assert_status LIST_FOR_CSRF_TARGET 200 "$s" "$b"
csrf_target_id="$(json_value "$b" "$(session_id_expr other)")"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/me/sessions/${csrf_target_id}" "" "$b" -b "$main_cookie" -c "$main_cookie")"; assert_status MISSING_CSRF_REJECTED 403 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$csrfless_cookie")"; assert_status CSRF_TARGET_SURVIVES 200 "$s" "$b"

log "[E2E] session-management scenarios passed user=${email}"
