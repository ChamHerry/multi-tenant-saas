#!/usr/bin/env bash
# e2e_account_lockout.sh — Docker-backed end-to-end test for account lockout and admin unlock.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-account-lockout}"
SCENARIO_LOG="${LOG_DIR}/account-lockout.log"
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
grant_platform_admin() { compose exec -T app ./repomind platform-admin-grant --user "$1" --role super_admin --actor "$1" >/dev/null; }

print_failure_context() {
  {
    echo "--- docker compose ps ---"
    compose ps || true
    echo
    echo "--- docker compose logs (tail=180) ---"
    compose logs --tail=180 app postgres redis || true
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

assert_auth_failure() {
  local id="$1" actual="$2" bodyfile="$3"
  if [[ "$actual" != "401" && "$actual" != "403" ]]; then
    log "[FAIL] ${id}: expected 401/403, got ${actual}; body=$(cat "$bodyfile")"
    exit 1
  fi
  log "[PASS] ${id}: HTTP ${actual}"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
password="AccountLockoutPassword123!"
wrong_password="WrongAccountLockout123!"
admin_email="account-lockout-admin-${stamp}@example.test"
target_email="account-lockout-target-${stamp}@example.test"
admin_cookie="${LOG_DIR}/admin.cookies"
target_cookie="${LOG_DIR}/target.cookies"

ensure_stack_ready

log "[E2E] register first user as platform admin and target user"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${admin_email}\",\"password\":\"${password}\",\"display_name\":\"Account Lockout Admin\"}" "$b" -c "$admin_cookie")"; assert_status ADMIN_REGISTER 200 "$s" "$b"
admin_id="$(json_value "$b" 'j["data"]["user"]["id"]')"
[[ -n "$admin_id" ]] || { log "[FAIL] missing admin user id"; exit 1; }
grant_platform_admin "$admin_id"
log "[PASS] ADMIN_GRANTED_SUPER_ADMIN"
admin_csrf="$(csrf_from_cookie_jar "$admin_cookie")"
[[ -n "$admin_csrf" ]] || { log "[FAIL] missing admin CSRF cookie"; exit 1; }
log "[PASS] ADMIN_CSRF"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" -b "$admin_cookie")"; assert_status ADMIN_ACCESS 200 "$s" "$b"
assert_json_equals ADMIN_IS_SUPER_ADMIN "$b" 'j["data"]["platform_admin"]["role"]' super_admin

b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${target_email}\",\"password\":\"${password}\",\"display_name\":\"Account Lockout Target\"}" "$b" -c "$target_cookie")"; assert_status TARGET_REGISTER 200 "$s" "$b"
target_id="$(json_value "$b" 'j["data"]["user"]["id"]')"
[[ -n "$target_id" ]] || { log "[FAIL] missing target user id"; exit 1; }

log "[E2E] wrong password attempts trigger server-side lockout"
for i in 1 2 3 4 5; do
  b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${target_email}\",\"password\":\"${wrong_password}\"}" "$b")"
  assert_auth_failure "WRONG_PASSWORD_${i}" "$s" "$b"
done
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${target_email}\",\"password\":\"${password}\"}" "$b")"
assert_auth_failure LOCKED_CORRECT_PASSWORD_REJECTED "$s" "$b"

log "[E2E] admin unlock endpoint clears the lock and is idempotent"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/admin/users/${target_id}/unlock" "" "$b" -b "$admin_cookie" -c "$admin_cookie" -H "X-CSRF-Token: ${admin_csrf}")"; assert_status ADMIN_UNLOCK 200 "$s" "$b"
assert_json_equals ADMIN_UNLOCK_OK "$b" 'j["data"]["ok"]' true
b="$(bodyfile)"; s="$(http_request POST "/api/v1/admin/users/${target_id}/unlock" "" "$b" -b "$admin_cookie" -c "$admin_cookie" -H "X-CSRF-Token: ${admin_csrf}")"; assert_status ADMIN_UNLOCK_IDEMPOTENT 200 "$s" "$b"

log "[E2E] target can log in after unlock"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${target_email}\",\"password\":\"${password}\"}" "$b" -c "$target_cookie")"; assert_status TARGET_LOGIN_AFTER_UNLOCK 200 "$s" "$b"
assert_json_equals TARGET_LOGIN_EMAIL "$b" 'j["data"]["user"]["email"]' "$target_email"

log "[E2E] account-lockout scenarios passed target=${target_email}"
