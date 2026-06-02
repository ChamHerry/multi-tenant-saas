#!/usr/bin/env bash
# e2e_totp_2fa.sh — Docker-backed end-to-end test for TOTP two-factor auth.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-totp}"
SCENARIO_LOG="${LOG_DIR}/totp-2fa.log"
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
value = eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len})
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

totp_code() {
  local secret="$1"
  python3 - "$secret" <<'PY'
import base64, hashlib, hmac, struct, sys, time
secret = sys.argv[1].strip().replace(' ', '').upper()
secret += '=' * ((8 - len(secret) % 8) % 8)
key = base64.b32decode(secret)
counter = int(time.time()) // 30
msg = struct.pack('>Q', counter)
digest = hmac.new(key, msg, hashlib.sha1).digest()
offset = digest[-1] & 0x0F
code = (struct.unpack('>I', digest[offset:offset+4])[0] & 0x7fffffff) % 1000000
print(f'{code:06d}')
PY
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
email="totp-e2e-${stamp}@example.test"
password="TotpPassword123!"
auth_cookie="${LOG_DIR}/auth.cookies"
login_cookie="${LOG_DIR}/login.cookies"
backup_cookie="${LOG_DIR}/backup.cookies"

ensure_stack_ready

log "[E2E] register password user"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${email}\",\"password\":\"${password}\",\"display_name\":\"TOTP E2E\"}" "$b" -c "$auth_cookie")"; assert_status REGISTER 200 "$s" "$b"
assert_json_equals REGISTER_EMAIL "$b" 'j["data"]["user"]["email"]' "$email"
csrf_token="$(csrf_from_cookie_jar "$auth_cookie")"
[[ -n "$csrf_token" ]] || { log "[FAIL] missing CSRF cookie after register"; exit 1; }
log "[PASS] REGISTER_CSRF"

log "[E2E] setup and enable TOTP"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/setup "{\"password\":\"${password}\"}" "$b" -b "$auth_cookie" -c "$auth_cookie" -H "X-CSRF-Token: ${csrf_token}")"; assert_status TOTP_SETUP 200 "$s" "$b"
secret="$(json_value "$b" 'j["data"]["secret"]')"
[[ -n "$secret" ]] || { log "[FAIL] empty TOTP secret"; exit 1; }
log "[PASS] TOTP_SECRET_PRESENT"
code="$(totp_code "$secret")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/enable "{\"code\":\"${code}\"}" "$b" -b "$auth_cookie" -c "$auth_cookie" -H "X-CSRF-Token: ${csrf_token}")"; assert_status TOTP_ENABLE 200 "$s" "$b"
assert_json_equals TOTP_ENABLE_OK "$b" 'j["data"]["ok"]' true
assert_json_equals TOTP_BACKUP_COUNT "$b" 'len(j["data"]["backup_codes"])' 10
backup_code="$(json_value "$b" 'j["data"]["backup_codes"][0]')"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/totp/status "" "$b" -b "$auth_cookie")"; assert_status TOTP_STATUS 200 "$s" "$b"
assert_json_equals TOTP_STATUS_ENABLED "$b" 'j["data"]["enabled"]' true


log "[E2E] API key with user:security:read can read but cannot mutate TOTP settings"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"TOTP E2E API Key Org\",\"slug\":\"totp-e2e-api-key-${stamp}\"}" "$b" -b "$auth_cookie" -c "$auth_cookie" -H "X-CSRF-Token: ${csrf_token}")"; assert_status API_KEY_TENANT_CREATE 200 "$s" "$b"
tenant_id="$(json_value "$b" 'j["data"]["tenant"]["id"]')"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" '{"name":"TOTP Read Key","scopes":["user:security:read","tenant:read"]}' "$b" -b "$auth_cookie" -c "$auth_cookie" -H "X-Tenant-ID: ${tenant_id}" -H "X-CSRF-Token: ${csrf_token}")"; assert_status API_KEY_CREATE 200 "$s" "$b"
raw_key="$(json_value "$b" 'j["data"]["raw_key"]')"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/totp/status "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status API_KEY_TOTP_STATUS_READ 200 "$s" "$b"
for write_case in \
  "setup:/api/v1/me/totp/setup:{\"password\":\"${password}\"}" \
  "enable:/api/v1/me/totp/enable:{\"code\":\"123456\"}" \
  "disable:/api/v1/me/totp/disable:{\"password\":\"${password}\"}" \
  "regenerate:/api/v1/me/totp/backup-codes/regenerate:{\"password\":\"${password}\"}"; do
  IFS=':' read -r case_name case_path case_body <<<"$write_case"
  b="$(bodyfile)"; s="$(http_request POST "$case_path" "$case_body" "$b" -H "Authorization: Bearer ${raw_key}")"
  assert_status "API_KEY_TOTP_${case_name}_DENIED" 403 "$s" "$b"
done

log "[E2E] login requires TOTP and does not create a session before verification"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$login_cookie")"; assert_status LOGIN_REQUIRES_TOTP 200 "$s" "$b"
assert_json_equals LOGIN_REQUIRES_2FA "$b" 'j["data"].get("requires_2fa")' true
totp_token="$(json_value "$b" 'j["data"].get("totp_token", "")')"
[[ -n "$totp_token" ]] || { log "[FAIL] missing temporary TOTP token"; exit 1; }
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$login_cookie")"; assert_status PRE_VERIFY_SESSION_REJECTED 401 "$s" "$b"

log "[E2E] verify TOTP code creates session"
code="$(totp_code "$secret")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${totp_token}\",\"code\":\"${code}\"}" "$b" -b "$login_cookie" -c "$login_cookie")"; assert_status TOTP_VERIFY 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/auth/session "" "$b" -b "$login_cookie")"; assert_status POST_VERIFY_SESSION 200 "$s" "$b"
assert_json_equals POST_VERIFY_EMAIL "$b" 'j["data"]["user"]["email"]' "$email"

log "[E2E] backup code login works once and reports remaining count"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$backup_cookie")"; assert_status BACKUP_LOGIN_REQUIRES_TOTP 200 "$s" "$b"
backup_token="$(json_value "$b" 'j["data"].get("totp_token", "")')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${backup_token}\",\"code\":\"${backup_code}\"}" "$b" -b "$backup_cookie" -c "$backup_cookie")"; assert_status BACKUP_VERIFY 200 "$s" "$b"
assert_json_equals BACKUP_REMAINING "$b" 'j["data"].get("backup_codes_remaining")' 9

log "[E2E] used backup code cannot be reused"
reuse_cookie="${LOG_DIR}/backup-reuse.cookies"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${email}\",\"password\":\"${password}\"}" "$b" -c "$reuse_cookie")"; assert_status BACKUP_REUSE_LOGIN_REQUIRES_TOTP 200 "$s" "$b"
reuse_token="$(json_value "$b" 'j["data"].get("totp_token", "")')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${reuse_token}\",\"code\":\"${backup_code}\"}" "$b" -b "$reuse_cookie" -c "$reuse_cookie")"; assert_status BACKUP_REUSE_REJECTED 403 "$s" "$b"



log "[E2E] concurrent TOTP enable has exactly one winner"
race_email="totp-e2e-race-${stamp}@example.test"
race_cookie1="${LOG_DIR}/race1.cookies"
race_cookie2="${LOG_DIR}/race2.cookies"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${race_email}\",\"password\":\"${password}\",\"display_name\":\"TOTP Race\"}" "$b" -c "$race_cookie1")"; assert_status RACE_REGISTER 200 "$s" "$b"
race_csrf1="$(csrf_from_cookie_jar "$race_cookie1")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/setup "{\"password\":\"${password}\"}" "$b" -b "$race_cookie1" -c "$race_cookie1" -H "X-CSRF-Token: ${race_csrf1}")"; assert_status RACE_SETUP 200 "$s" "$b"
race_secret="$(json_value "$b" 'j["data"]["secret"]')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${race_email}\",\"password\":\"${password}\"}" "$b" -c "$race_cookie2")"; assert_status RACE_SECOND_SESSION_LOGIN 200 "$s" "$b"
race_csrf2="$(csrf_from_cookie_jar "$race_cookie2")"
race_code="$(totp_code "$race_secret")"
race_out1="$(bodyfile)"; race_out2="$(bodyfile)"; race_status1="${LOG_DIR}/race-enable-1.status"; race_status2="${LOG_DIR}/race-enable-2.status"
(http_request POST /api/v1/me/totp/enable "{\"code\":\"${race_code}\"}" "$race_out1" -b "$race_cookie1" -c "$race_cookie1" -H "X-CSRF-Token: ${race_csrf1}" >"$race_status1") &
pid1=$!
(http_request POST /api/v1/me/totp/enable "{\"code\":\"${race_code}\"}" "$race_out2" -b "$race_cookie2" -c "$race_cookie2" -H "X-CSRF-Token: ${race_csrf2}" >"$race_status2") &
pid2=$!
wait "$pid1" "$pid2"
race_s1="$(cat "$race_status1")"; race_s2="$(cat "$race_status2")"
race_successes=0
[[ "$race_s1" == "200" ]] && race_successes=$((race_successes + 1))
[[ "$race_s2" == "200" ]] && race_successes=$((race_successes + 1))
if [[ "$race_successes" != "1" ]]; then
  log "[FAIL] ENABLE_RACE_SINGLE_WINNER: statuses=${race_s1},${race_s2}; bodies=$(cat "$race_out1") | $(cat "$race_out2")"
  exit 1
fi
log "[PASS] ENABLE_RACE_SINGLE_WINNER: statuses=${race_s1},${race_s2}"

log "[E2E] concurrent backup code reuse has exactly one winner"
backup_race_email="totp-e2e-backup-race-${stamp}@example.test"
backup_race_cookie="${LOG_DIR}/backup-race-setup.cookies"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${backup_race_email}\",\"password\":\"${password}\",\"display_name\":\"TOTP Backup Race\"}" "$b" -c "$backup_race_cookie")"; assert_status BACKUP_RACE_REGISTER 200 "$s" "$b"
backup_race_csrf="$(csrf_from_cookie_jar "$backup_race_cookie")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/setup "{\"password\":\"${password}\"}" "$b" -b "$backup_race_cookie" -c "$backup_race_cookie" -H "X-CSRF-Token: ${backup_race_csrf}")"; assert_status BACKUP_RACE_SETUP 200 "$s" "$b"
backup_race_secret="$(json_value "$b" 'j["data"]["secret"]')"
backup_race_code="$(totp_code "$backup_race_secret")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/enable "{\"code\":\"${backup_race_code}\"}" "$b" -b "$backup_race_cookie" -c "$backup_race_cookie" -H "X-CSRF-Token: ${backup_race_csrf}")"; assert_status BACKUP_RACE_ENABLE 200 "$s" "$b"
backup_race_backup_code="$(json_value "$b" 'j["data"]["backup_codes"][0]')"
verify_cookie1="${LOG_DIR}/backup-race-verify1.cookies"; verify_cookie2="${LOG_DIR}/backup-race-verify2.cookies"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${backup_race_email}\",\"password\":\"${password}\"}" "$b" -c "$verify_cookie1")"; assert_status BACKUP_RACE_LOGIN1 200 "$s" "$b"; token1="$(json_value "$b" 'j["data"].get("totp_token", "")')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${backup_race_email}\",\"password\":\"${password}\"}" "$b" -c "$verify_cookie2")"; assert_status BACKUP_RACE_LOGIN2 200 "$s" "$b"; token2="$(json_value "$b" 'j["data"].get("totp_token", "")')"
verify_out1="$(bodyfile)"; verify_out2="$(bodyfile)"; verify_status1="${LOG_DIR}/backup-verify-1.status"; verify_status2="${LOG_DIR}/backup-verify-2.status"
(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${token1}\",\"code\":\"${backup_race_backup_code}\"}" "$verify_out1" -b "$verify_cookie1" -c "$verify_cookie1" >"$verify_status1") &
pid1=$!
(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${token2}\",\"code\":\"${backup_race_backup_code}\"}" "$verify_out2" -b "$verify_cookie2" -c "$verify_cookie2" >"$verify_status2") &
pid2=$!
wait "$pid1" "$pid2"
verify_s1="$(cat "$verify_status1")"; verify_s2="$(cat "$verify_status2")"
verify_successes=0
[[ "$verify_s1" == "200" ]] && verify_successes=$((verify_successes + 1))
[[ "$verify_s2" == "200" ]] && verify_successes=$((verify_successes + 1))
if [[ "$verify_successes" != "1" ]]; then
  log "[FAIL] BACKUP_CODE_PARALLEL_REUSE_SINGLE_WINNER: statuses=${verify_s1},${verify_s2}; bodies=$(cat "$verify_out1") | $(cat "$verify_out2")"
  exit 1
fi
log "[PASS] BACKUP_CODE_PARALLEL_REUSE_SINGLE_WINNER: statuses=${verify_s1},${verify_s2}"

log "[E2E] invalid TOTP verify attempts return 429 with Retry-After"
rate_email="totp-e2e-rate-${stamp}@example.test"
rate_setup_cookie="${LOG_DIR}/rate-limit-setup.cookies"
rate_cookie="${LOG_DIR}/rate-limit.cookies"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "{\"email\":\"${rate_email}\",\"password\":\"${password}\",\"display_name\":\"TOTP Rate Limit\"}" "$b" -c "$rate_setup_cookie")"; assert_status RATE_LIMIT_REGISTER 200 "$s" "$b"
rate_csrf="$(csrf_from_cookie_jar "$rate_setup_cookie")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/setup "{\"password\":\"${password}\"}" "$b" -b "$rate_setup_cookie" -c "$rate_setup_cookie" -H "X-CSRF-Token: ${rate_csrf}")"; assert_status RATE_LIMIT_SETUP 200 "$s" "$b"
rate_secret="$(json_value "$b" 'j["data"]["secret"]')"
rate_code="$(totp_code "$rate_secret")"
b="$(bodyfile)"; s="$(http_request POST /api/v1/me/totp/enable "{\"code\":\"${rate_code}\"}" "$b" -b "$rate_setup_cookie" -c "$rate_setup_cookie" -H "X-CSRF-Token: ${rate_csrf}")"; assert_status RATE_LIMIT_ENABLE 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/login "{\"email\":\"${rate_email}\",\"password\":\"${password}\"}" "$b" -c "$rate_cookie")"; assert_status RATE_LIMIT_LOGIN_REQUIRES_TOTP 200 "$s" "$b"
rate_token="$(json_value "$b" 'j["data"].get("totp_token", "")')"
for attempt in 1 2 3 4 5; do
  b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${rate_token}\",\"code\":\"000000\"}" "$b" -b "$rate_cookie" -c "$rate_cookie")"; assert_status "RATE_LIMIT_INVALID_${attempt}" 403 "$s" "$b"
done
headers="$(bodyfile)"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/verify-totp "{\"totp_token\":\"${rate_token}\",\"code\":\"000000\"}" "$b" -D "$headers" -b "$rate_cookie" -c "$rate_cookie")"; assert_status RATE_LIMIT_429 429 "$s" "$b"
if ! grep -qi '^Retry-After:' "$headers"; then
  log "[FAIL] RATE_LIMIT_RETRY_AFTER_HEADER missing; headers=$(cat "$headers")"
  exit 1
fi
log "[PASS] RATE_LIMIT_RETRY_AFTER_HEADER"

log "[E2E] TOTP two-factor auth scenarios passed email=${email}"
