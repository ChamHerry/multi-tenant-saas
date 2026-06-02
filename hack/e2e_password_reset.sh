#!/usr/bin/env bash
# e2e_password_reset.sh — Docker-backed end-to-end test for forgot/reset password.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-password-reset}"
SCENARIO_LOG="${LOG_DIR}/password-reset.log"
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

extract_reset_token_from_logs() {
  local email="$1"
  compose logs --no-color app 2>/dev/null \
    | grep -F "password reset to ${email}" \
    | grep -F "reset URL:" \
    | tail -1 \
    | sed -E 's/.*[?&]token=([0-9a-f]{64}).*/\1/'
}

wait_for_reset_token() {
  local email="$1" token=""
  for _ in $(seq 1 100); do
    token="$(extract_reset_token_from_logs "$email" || true)"
    if [[ "$token" =~ ^[0-9a-f]{64}$ ]]; then
      printf '%s' "$token"
      return 0
    fi
    sleep 0.25
  done
  log "[FAIL] no password reset token found in app logs for ${email}"
  exit 1
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
old_password="PasswordResetOld123!"
new_password="PasswordResetNew123!"
email="password-reset-e2e-${stamp}@example.test"
main_cookie="${LOG_DIR}/main.cookies"
second_cookie="${LOG_DIR}/second.cookies"
new_cookie="${LOG_DIR}/new.cookies"

ensure_stack_ready

log "[E2E] register password user and create a second active session"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${email}\",\"password\":\"${old_password}\",\"display_name\":\"Password Reset E2E\"}" "$b" -c "$main_cookie")"; assert_status REGISTER 200 "$s" "$b"
assert_json_equals REGISTER_EMAIL "$b" 'j["data"]["user"]["email"]' "$email"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${old_password}\"}" "$b" -c "$second_cookie")"; assert_status SECOND_LOGIN 200 "$s" "$b"

log "[E2E] forgot password is enumeration-safe for missing email"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/forgot-password '{"email":"missing-password-e2e@example.test"}' "$b")"; assert_status FORGOT_MISSING_ENUM_SAFE 200 "$s" "$b"
assert_json_equals FORGOT_MISSING_OK "$b" 'j["data"]["ok"]' true

log "[E2E] forgot password logs a reset URL when SMTP is disabled"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/forgot-password "{\"email\":\"${email}\"}" "$b")"; assert_status FORGOT_EXISTING 200 "$s" "$b"
assert_json_equals FORGOT_EXISTING_OK "$b" 'j["data"]["ok"]' true
reset_token="$(wait_for_reset_token "$email")"
[[ -n "$reset_token" ]] || { log "[FAIL] empty reset token"; exit 1; }
log "[PASS] RESET_TOKEN_CAPTURED"

log "[E2E] reset changes password and revokes existing sessions"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/reset-password "{\"token\":\"${reset_token}\",\"new_password\":\"${new_password}\"}" "$b")"; assert_status RESET_PASSWORD 200 "$s" "$b"
assert_json_equals RESET_PASSWORD_OK "$b" 'j["data"]["ok"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$main_cookie")"; assert_status MAIN_SESSION_REVOKED 401 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$second_cookie")"; assert_status SECOND_SESSION_REVOKED 401 "$s" "$b"

log "[E2E] old password fails, new password succeeds"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${old_password}\"}" "$b")"
if [[ "$s" != "401" && "$s" != "403" ]]; then
  log "[FAIL] OLD_PASSWORD_REJECTED: expected 401/403, got ${s}; body=$(cat "$b")"
  exit 1
fi
log "[PASS] OLD_PASSWORD_REJECTED: HTTP ${s}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${new_password}\"}" "$b" -c "$new_cookie")"; assert_status NEW_PASSWORD_LOGIN 200 "$s" "$b"
assert_json_equals NEW_PASSWORD_LOGIN_EMAIL "$b" 'j["data"]["user"]["email"]' "$email"

log "[E2E] replayed reset token is rejected without another password change"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/reset-password "{\"token\":\"${reset_token}\",\"new_password\":\"ReplayPassword12345!\"}" "$b")"; assert_status RESET_REPLAY_REJECTED 403 "$s" "$b"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${new_password}\"}" "$b")"; assert_status NEW_PASSWORD_STILL_VALID 200 "$s" "$b"

log "[E2E] password-reset scenarios passed user=${email}"
