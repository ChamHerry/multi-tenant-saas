#!/usr/bin/env bash
# e2e_email_verification.sh — Docker-backed end-to-end test for email verification.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-email-verification}"
SCENARIO_LOG="${LOG_DIR}/email-verification.log"
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

extract_verify_token_from_logs() {
  local email="$1"
  compose logs --no-color app 2>/dev/null \
    | grep -F "verification email to ${email}" \
    | grep -F "verify URL:" \
    | tail -1 \
    | sed -E 's/.*[?&]token=([^[:space:]]+).*/\1/'
}

wait_for_verify_token() {
  local email="$1" token=""
  for _ in $(seq 1 100); do
    token="$(extract_verify_token_from_logs "$email" || true)"
    if [[ -n "$token" && "$token" != *"verification email"* ]]; then
      printf '%s' "$token"
      return 0
    fi
    sleep 0.25
  done
  log "[FAIL] no verification token found in app logs for ${email}"
  exit 1
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
password="EmailVerifyPassword123!"
email="email-verify-e2e-${stamp}@example.test"
cooldown_email="email-verify-cooldown-${stamp}@example.test"
main_cookie="${LOG_DIR}/main.cookies"
cooldown_cookie="${LOG_DIR}/cooldown.cookies"

ensure_stack_ready

log "[E2E] register user and capture no-SMTP verification URL"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${email}\",\"password\":\"${password}\",\"display_name\":\"Email Verify E2E\"}" "$b" -c "$main_cookie")"; assert_status REGISTER 200 "$s" "$b"
assert_json_equals REGISTER_EMAIL "$b" 'j["data"]["user"]["email"]' "$email"
assert_json_equals REGISTER_EMAIL_UNVERIFIED "$b" 'j["data"]["user"].get("email_verified", False)' false
verify_token="$(wait_for_verify_token "$email")"
[[ -n "$verify_token" ]] || { log "[FAIL] empty verify token"; exit 1; }
log "[PASS] VERIFY_TOKEN_CAPTURED"

log "[E2E] verify email token updates current session user"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-email "{\"token\":\"${verify_token}\"}" "$b")"; assert_status VERIFY_EMAIL 200 "$s" "$b"
assert_json_equals VERIFY_EMAIL_OK "$b" 'j["data"]["ok"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "$b" -b "$main_cookie")"; assert_status ME_AFTER_VERIFY 200 "$s" "$b"
assert_json_equals ME_EMAIL_VERIFIED "$b" 'j["data"]["user"].get("email_verified")' true

log "[E2E] replay and invalid verify tokens are rejected without 500"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-email "{\"token\":\"${verify_token}\"}" "$b")"; assert_status VERIFY_REPLAY_REJECTED 404 "$s" "$b"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-email '{"token":"not-a-real-token"}' "$b")"; assert_status VERIFY_INVALID_REJECTED 404 "$s" "$b"

log "[E2E] resend is enumeration-safe for missing email"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/resend-verification '{"email":"missing-email-e2e@example.test"}' "$b")"; assert_status RESEND_MISSING_ENUM_SAFE 200 "$s" "$b"
assert_json_equals RESEND_MISSING_OK "$b" 'j["data"]["ok"]' true

log "[E2E] resend cooldown returns HTTP 429 for a fresh unverified token"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${cooldown_email}\",\"password\":\"${password}\",\"display_name\":\"Email Cooldown E2E\"}" "$b" -c "$cooldown_cookie")"; assert_status COOLDOWN_REGISTER 200 "$s" "$b"
wait_for_verify_token "$cooldown_email" >/dev/null
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/resend-verification "{\"email\":\"${cooldown_email}\"}" "$b")"; assert_status RESEND_COOLDOWN 429 "$s" "$b"

log "[E2E] email-verification scenarios passed user=${email}"
