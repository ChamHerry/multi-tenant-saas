#!/usr/bin/env bash
# e2e_system_config.sh — Docker-backed API e2e for platform system configuration.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
APP_SERVICE="${APP_SERVICE:-app}"
DB_SERVICE="${DB_SERVICE:-postgres}"
REDIS_SERVICE="${REDIS_SERVICE:-redis}"
APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-180}"
READY_INTERVAL="${READY_INTERVAL:-2}"
KEEP_STACK="${E2E_KEEP_STACK:-1}"

LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-system-config}"
SCENARIO_LOG="${LOG_DIR}/system-config.log"
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
appctl() { compose exec -T "$APP_SERVICE" ./repomind "$@"; }
psql_query() { compose exec -T -e PGPASSWORD=secret "$DB_SERVICE" psql -U saas_template -d saas_template -Atc "$1"; }
source "${ROOT_DIR}/hack/lib/e2e_auth_session.sh"

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

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

cleanup_keys=()
cleanup() {
  local status=$?
  for key in "${cleanup_keys[@]:-}"; do
    psql_query "DELETE FROM public.system_config WHERE key='${key}';" >/dev/null 2>&1 || true
    compose exec -T "$REDIS_SERVICE" redis-cli DEL "config:${key}" >/dev/null 2>&1 || true
  done
  if (( status != 0 )); then
    log "[E2E] failure detected, dumping compose context"
    print_failure_context
  fi
  if [[ "$KEEP_STACK" == "0" ]]; then
    manage_docker down >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

ensure_stack_ready() {
  case "$E2E_DOCKER_ACTION" in
    skip) log "[E2E] reuse existing docker stack" ;;
    start|up|restart) log "[E2E] manage-docker action: $E2E_DOCKER_ACTION"; manage_docker "$E2E_DOCKER_ACTION" ;;
    *) echo "unsupported E2E_DOCKER_ACTION=$E2E_DOCKER_ACTION (expected start|up|restart|skip)" >&2; exit 2 ;;
  esac
  log "[E2E] wait for /readyz"
  manage_docker ready >/dev/null
  ensure_setup_completed
  e2e_configure_http_cookies
}

create_user() {
  local tag="$1"
  local email="${tag}@example.test"
  e2e_create_password_user "$tag"
}

http_request() {
  local method="$1" path="$2" body="$3" outfile="$4"
  shift 4
  local args=(-sS -o "$outfile" -w "%{http_code}" -X "$method")
  if [[ -n "$body" ]]; then
    args+=(-H "Content-Type: application/json" -d "$body")
  fi
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

assert_authz_failure() {
  local id="$1" actual="$2" bodyfile="$3"
  if [[ "$actual" != "401" && "$actual" != "403" ]]; then
    log "[FAIL] ${id}: expected HTTP 401/403, got ${actual}; body=$(cat "$bodyfile")"
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
  if grep -Fq "$needle" "$file"; then
    log "[FAIL] ${id}: response leaked forbidden value"
    exit 1
  fi
  log "[PASS] ${id}: forbidden value absent"
}

assert_sql_equals() {
  local id="$1" sql="$2" expected="$3"
  local actual
  actual="$(psql_query "$sql")"
  if [[ "$actual" != "$expected" ]]; then
    log "[FAIL] ${id}: expected sql=${expected}, got ${actual}"
    exit 1
  fi
  log "[PASS] ${id}: ${actual}"
}

ensure_setup_completed() {
  local b s requires setup_email setup_password payload
  b="$(bodyfile)"; s="$(http_request GET /api/v1/setup/state "" "$b")"; assert_status SETUP_STATE_CHECK 200 "$s" "$b"
  requires="$(json_value "$b" 'j["data"]["requires_setup"]')"
  if [[ "$requires" != "true" ]]; then
    log "[E2E] setup already initialized"
    return
  fi

  log "[E2E] complete setup bootstrap for system-config scenario"
  setup_email="system-config-bootstrap-${stamp}@example.test"
  setup_password="SetupPassword12345!"
  payload="$(printf '{"admin":{"email":"%s","password":"%s","display_name":"System Config Bootstrap"},"runtime":{"web_base_url":"%s","generate_session_secret":true,"generate_api_key_secret":true}}' "$setup_email" "$setup_password" "$BASE_URL")"
  b="$(bodyfile)"; s="$(http_request POST /api/v1/setup/complete "$payload" "$b")"; assert_status SETUP_BOOTSTRAP_COMPLETE 200 "$s" "$b"
  assert_json_equals SETUP_BOOTSTRAP_INITIALIZED "$b" 'j["data"]["initialized"]' true
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
admin_tag="system-config-admin-${stamp}"
support_tag="system-config-support-${stamp}"
user_tag="system-config-user-${stamp}"
bool_key="e2e.systemConfig.${stamp}.bool"
secret_key="e2e.systemConfig.${stamp}.secret"
secret_plaintext="system-config-secret-${stamp}-never-leak"
cleanup_keys=("$bool_key" "$secret_key")

ensure_stack_ready

log "[E2E] setup platform users"
admin_id="$(create_user "$admin_tag")"
support_id="$(create_user "$support_tag")"
user_id="$(create_user "$user_tag")"
appctl platform-admin-grant --user "$admin_id" --role super_admin --actor "$admin_id" >/dev/null
appctl platform-admin-grant --user "$support_id" --role support --actor "$admin_id" >/dev/null
admin_cookie="${LOG_DIR}/admin.cookies"
support_cookie="${LOG_DIR}/support.cookies"
user_cookie="${LOG_DIR}/user.cookies"
e2e_login_user_tag "$admin_tag" "$admin_cookie"
e2e_login_user_tag "$support_tag" "$support_cookie"
e2e_login_user_tag "$user_tag" "$user_cookie"
admin_csrf="$(e2e_csrf_from_cookie_jar "$admin_cookie")"
support_csrf="$(e2e_csrf_from_cookie_jar "$support_cookie")"
user_csrf="$(e2e_csrf_from_cookie_jar "$user_cookie")"
admin_auth=(-b "$admin_cookie" -c "$admin_cookie" -H "X-CSRF-Token: ${admin_csrf}")
support_auth=(-b "$support_cookie" -c "$support_cookie" -H "X-CSRF-Token: ${support_csrf}")
user_auth=(-b "$user_cookie" -c "$user_cookie" -H "X-CSRF-Token: ${user_csrf}")

log "[E2E] ready and admin access"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "$b")"; assert_status READYZ 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" "${admin_auth[@]}")"; assert_status ADMIN_ACCESS 200 "$s" "$b"
assert_json_equals ADMIN_HAS_CONFIG_MANAGE "$b" '"platform:config:manage" in j["data"]["platform_admin"]["permissions"]' true

log "[E2E] list, verify runtime secret, create and update bool config"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/system-config?limit=5 "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_LIST_INITIAL 200 "$s" "$b"
raw_session_secret="$(psql_query "SELECT value FROM public.system_config WHERE key='auth.session.secret'")"
[[ -n "$raw_session_secret" ]] || { log "[FAIL] missing runtime auth.session.secret"; exit 1; }
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/system-config/auth.session.secret "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_RUNTIME_SECRET_GET 200 "$s" "$b"
assert_body_not_contains CONFIG_RUNTIME_SECRET_MASKED "$b" "$raw_session_secret"
assert_json_equals CONFIG_RUNTIME_SECRET_TYPE "$b" 'j["data"]["config"]["value_type"]' secret
assert_json_equals CONFIG_RUNTIME_SECRET_VALUE_EMPTY "$b" 'j["data"]["config"]["value"]' ""
b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${bool_key}" '{"value":"true","value_provided":true,"value_type":"bool","description":"Docker e2e bool"}' "$b" "${admin_auth[@]}")"; assert_status CONFIG_BOOL_CREATE 200 "$s" "$b"
assert_json_equals CONFIG_BOOL_VALUE "$b" 'j["data"]["config"]["value"]' true
assert_json_equals CONFIG_BOOL_TYPE "$b" 'j["data"]["config"]["value_type"]' bool
assert_json_equals CONFIG_BOOL_CATEGORY "$b" 'j["data"]["config"]["category"]' e2e
assert_sql_equals CONFIG_BOOL_DB "SELECT value FROM public.system_config WHERE key='${bool_key}'" true

b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${bool_key}" '{"value":"false","value_provided":true,"value_type":"bool","description":"Docker e2e bool updated"}' "$b" "${admin_auth[@]}")"; assert_status CONFIG_BOOL_UPDATE 200 "$s" "$b"
assert_json_equals CONFIG_BOOL_UPDATED_VALUE "$b" 'j["data"]["config"]["value"]' false
assert_sql_equals CONFIG_BOOL_DB_UPDATED "SELECT value FROM public.system_config WHERE key='${bool_key}'" false

b="$(bodyfile)"; s="$(http_request GET "/api/v1/admin/system-config?query=${stamp}&category=e2e" "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_LIST_FILTERED 200 "$s" "$b"
assert_json_equals CONFIG_LIST_HAS_BOOL "$b" "any(item['key'] == '${bool_key}' for item in j['data']['items'])" true

log "[E2E] invalid input is rejected"
b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${bool_key}" '{"value":"enabled","value_provided":true,"value_type":"bool","description":"bad"}' "$b" "${admin_auth[@]}")"; assert_status CONFIG_INVALID_BOOL 400 "$s" "$b"
b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/e2e.systemConfig.${stamp}.json" '{"value":"{bad}","value_provided":true,"value_type":"json","description":"bad"}' "$b" "${admin_auth[@]}")"; assert_status CONFIG_INVALID_JSON 400 "$s" "$b"

log "[E2E] secret config is encrypted and masked"
b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${secret_key}" "{\"value\":\"${secret_plaintext}\",\"value_provided\":true,\"value_type\":\"secret\",\"description\":\"Docker e2e secret\"}" "$b" "${admin_auth[@]}")"; assert_status CONFIG_SECRET_CREATE 200 "$s" "$b"
assert_body_not_contains CONFIG_SECRET_RESPONSE_MASKED "$b" "$secret_plaintext"
assert_json_equals CONFIG_SECRET_VALUE_EMPTY "$b" 'j["data"]["config"]["value"]' ""
assert_json_equals CONFIG_SECRET_MASK "$b" 'j["data"]["config"]["masked_value"]' "********"
assert_json_equals CONFIG_SECRET_HAS_VALUE "$b" 'j["data"]["config"]["has_value"]' true
assert_sql_equals CONFIG_SECRET_DB_ENCRYPTED "SELECT CASE WHEN value <> '${secret_plaintext}' AND value <> '' THEN 'true' ELSE 'false' END FROM public.system_config WHERE key='${secret_key}'" true

b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${secret_key}" '{"value":"","value_provided":false,"value_type":"secret","description":"Docker e2e secret description only"}' "$b" "${admin_auth[@]}")"; assert_status CONFIG_SECRET_PRESERVE 200 "$s" "$b"
assert_body_not_contains CONFIG_SECRET_PRESERVE_MASKED "$b" "$secret_plaintext"
assert_sql_equals CONFIG_SECRET_DB_STILL_ENCRYPTED "SELECT CASE WHEN value <> '${secret_plaintext}' AND value <> '' THEN 'true' ELSE 'false' END FROM public.system_config WHERE key='${secret_key}'" true

log "[E2E] audit log and permission boundaries"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/admin/audit-logs?action=system_config.updated&resource_type=system_config&resource_id=${secret_key}&limit=20" "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_AUDIT_LIST 200 "$s" "$b"
assert_body_not_contains CONFIG_AUDIT_NO_SECRET "$b" "$secret_plaintext"
assert_json_equals CONFIG_AUDIT_FOUND "$b" 'len(j["data"]["logs"]) >= 1' true

b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/system-config?limit=1 "" "$b" "${support_auth[@]}")"; assert_status SUPPORT_CONFIG_READ 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request PUT "/api/v1/admin/system-config/${bool_key}" '{"value":"true","value_provided":true,"value_type":"bool","description":"support attack"}' "$b" "${support_auth[@]}")"; assert_status SUPPORT_CONFIG_WRITE_FORBIDDEN 403 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/system-config?limit=1 "" "$b" "${user_auth[@]}")"; assert_authz_failure NON_ADMIN_CONFIG_READ "$s" "$b"

log "[E2E] delete non-protected config and protect core keys"
b="$(bodyfile)"; s="$(http_request DELETE /api/v1/admin/system-config/auth.session.secret "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_DELETE_PROTECTED 400 "$s" "$b"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/admin/system-config/${bool_key}" "" "$b" "${admin_auth[@]}")"; assert_status CONFIG_DELETE_BOOL 200 "$s" "$b"
assert_sql_equals CONFIG_BOOL_DELETED "SELECT count(*) FROM public.system_config WHERE key='${bool_key}'" 0

log "[E2E] auth/session regression smoke"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "$b" "${admin_auth[@]}")"; assert_status AUTH_SESSION_ME 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "$b")"; assert_status AUTH_MISSING_REJECTED 401 "$s" "$b"

log "[E2E] system-config scenarios passed admin=${admin_id} support=${support_id}"
