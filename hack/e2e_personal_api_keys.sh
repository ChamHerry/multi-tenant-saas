#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
APP_SERVICE="${APP_SERVICE:-app}"
DB_SERVICE="${DB_SERVICE:-postgres}"
APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-60}"
READY_INTERVAL="${READY_INTERVAL:-2}"
KEEP_STACK="${E2E_KEEP_STACK:-1}"

LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/saas-template-e2e-personal-api-keys}"
SCENARIO_LOG="${LOG_DIR}/scenario.log"
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

compose() {
  docker compose -f "$COMPOSE_FILE" "$@"
}

manage_docker() {
  COMPOSE_FILE="$COMPOSE_FILE" APP_PORT="$APP_PORT" READY_TIMEOUT="$READY_TIMEOUT" READY_INTERVAL="$READY_INTERVAL" ./scripts/manage-docker.sh "$@"
}

appctl() {
  compose exec -T "$APP_SERVICE" ./repomind "$@"
}

psql_query() {
  local sql="$1"
  compose exec -T -e PGPASSWORD=secret "$DB_SERVICE" psql -U saas_template -d saas_template -Atc "$sql"
}

print_failure_context() {
  {
    echo "--- docker compose ps ---"
    compose ps || true
    echo
    echo "--- docker compose logs (tail=120) ---"
    compose logs --tail=120 "$APP_SERVICE" "$DB_SERVICE" || true
  } >"$FAILURE_LOG" 2>&1
  cat "$FAILURE_LOG" >&2 || true
}

cleanup() {
  local status=$?
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
    skip)
      log "[E2E] reuse existing docker stack"
      ;;
    start|up|restart)
      log "[E2E] manage-docker action: $E2E_DOCKER_ACTION"
      manage_docker "$E2E_DOCKER_ACTION"
      ;;
    *)
      echo "unsupported E2E_DOCKER_ACTION=$E2E_DOCKER_ACTION (expected start|up|restart|skip)" >&2
      exit 2
      ;;
  esac
  log "[E2E] wait for /readyz"
  manage_docker ready >/dev/null
}

create_user() {
  local tag="$1"
  local email="${tag}@example.test"
  appctl user-upsert --provider password --auth-id "$email" --email "$email" --name "$tag" --email-verified >/dev/null
  psql_query "SELECT id FROM public.users WHERE email='${email}' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1"
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

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j = json.load(f)
value = eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len, "any": any, "all": all})
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

stamp="$(date +%Y%m%d%H%M%S)-$$"
owner_tag="personal-owner-${stamp}"
outsider_tag="personal-outsider-${stamp}"
tenant_slug="personal-${stamp}"
other_slug="personal-other-${stamp}"

ensure_stack_ready

log "[E2E] setup users via app container CLI"
owner_id="$(create_user "$owner_tag")"
outsider_id="$(create_user "$outsider_tag")"

log "[E2E] ready check"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "$b")"; assert_status READYZ 200 "$s" "$b"

log "[E2E] create owner and outsider tenants"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Personal ${stamp}\",\"slug\":\"${tenant_slug}\"}" "$b" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_CREATE 200 "$s" "$b"
tenant_id="$(json_value "$b" 'j["data"]["tenant"]["id"]')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Other ${stamp}\",\"slug\":\"${other_slug}\"}" "$b" -H "X-User-ID: ${outsider_id}")"; assert_status OTHER_TENANT_CREATE 200 "$s" "$b"
other_tenant_id="$(json_value "$b" 'j["data"]["tenant"]["id"]')"

log "[E2E] legacy personal API key route is removed"
b="$(bodyfile)"; s="$(http_request GET /api/v1/api-keys "" "$b" -H "X-User-ID: ${owner_id}")"; assert_status LEGACY_PERSONAL_ROUTE_REMOVED 404 "$s" "$b"

log "[E2E] create tenant-scoped API key for owner tenant"
create_key_payload="{\"name\":\"personal-reader\",\"scopes\":[\"user:read\",\"tenant:read\"]}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" "$create_key_payload" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_SCOPED_KEY_CREATE 200 "$s" "$b"
raw_key="$(json_value "$b" 'j["data"]["raw_key"]')"
api_key_id="$(json_value "$b" 'j["data"]["api_key"]["id"]')"
if [[ "$raw_key" != saas_* ]]; then
  log "[FAIL] KEY_PREFIX: unexpected raw key ${raw_key}"
  exit 1
fi
log "[PASS] KEY_PREFIX"

log "[E2E] owner tenant can list the key; outsider tenant cannot"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/api-keys" "" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status OWNER_TENANT_LIST_KEYS 200 "$s" "$b"
assert_json_equals OWNER_TENANT_LIST_COUNT "$b" 'len(j["data"]["api_keys"]) == 1' true
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${other_tenant_id}/api-keys" "" "$b" -H "X-User-ID: ${outsider_id}" -H "X-Tenant-ID: ${other_tenant_id}")"; assert_status OUTSIDER_TENANT_LIST_KEYS 200 "$s" "$b"
assert_json_equals OUTSIDER_TENANT_LIST_EMPTY "$b" 'len(j["data"]["api_keys"]) == 0' true

log "[E2E] API key can read /me and owner tenant context, but not outsider tenant"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_ME 200 "$s" "$b"
assert_json_equals APIKEY_ME_ID "$b" 'j["data"]["user"]["id"]' "$owner_id"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_TENANT_REQUIRED 400 "$s" "$b"
assert_json_equals APIKEY_TENANT_REQUIRED_CODE "$b" 'j["code"]' TENANT_REQUIRED
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_TENANT_GRANTED 200 "$s" "$b"
assert_json_equals APIKEY_TENANT_ROLE "$b" 'j["data"]["tenant_context"]["role"]' owner
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${other_tenant_id}")"; assert_status APIKEY_TENANT_NOT_MEMBER 403 "$s" "$b"
assert_json_equals APIKEY_TENANT_NOT_MEMBER_CODE "$b" 'j["code"]' TENANT_FORBIDDEN

log "[E2E] scope intersection blocks member list when key lacks member:read"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/members" "" "$b" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_SCOPE_INTERSECTION 403 "$s" "$b"

log "[E2E] platform routes reject API key identity even if owner is platform admin"
appctl platform-admin-grant --user "$owner_id" --role super_admin --actor "$owner_id" >/dev/null
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_PLATFORM_ADMIN_FORBIDDEN 403 "$s" "$b"
assert_json_equals APIKEY_PLATFORM_ADMIN_FORBIDDEN_CODE "$b" 'j["code"]' PLATFORM_ADMIN_SESSION_REQUIRED

log "[E2E] revoke tenant-scoped key and confirm it can no longer authenticate"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/tenants/${tenant_id}/api-keys/${api_key_id}" "" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_SCOPED_KEY_REVOKE 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status KEY_REVOKED_REJECTED 401 "$s" "$b"

log "[E2E] tenant-scoped personal API key scenarios passed tenant=${tenant_id} owner=${owner_id}"
