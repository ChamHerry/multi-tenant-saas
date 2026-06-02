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

LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e-menu-permissions}"
SCENARIO_LOG="${LOG_DIR}/menu-permissions.log"
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

enable_dev_header_auth() {
  psql_query "UPDATE public.system_config SET value='true', updated_at=now() WHERE key='auth.devHeader.enabled';" >/dev/null
  psql_query "UPDATE public.system_config SET value='local', updated_at=now() WHERE key='server.env';" >/dev/null
  compose exec -T redis redis-cli DEL config:auth.devHeader.enabled config:server.env >/dev/null 2>&1 || true
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
  enable_dev_header_auth
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
owner_tag="menu-owner-${stamp}"
admin_tag="menu-admin-${stamp}"
viewer_tag="menu-viewer-${stamp}"
support_tag="menu-support-${stamp}"
tenant_a_slug="menu-a-${stamp}"
tenant_b_slug="menu-b-${stamp}"

ensure_stack_ready

log "[E2E] setup users via app container CLI"
owner_id="$(create_user "$owner_tag")"
admin_id="$(create_user "$admin_tag")"
viewer_id="$(create_user "$viewer_tag")"
support_id="$(create_user "$support_tag")"
appctl platform-admin-grant --user "$support_id" --role support --actor "$support_id" >/dev/null

log "[E2E] ready check"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "$b")"; assert_status READYZ 200 "$s" "$b"

log "[E2E] create two tenants for owner"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Menu A ${stamp}\",\"slug\":\"${tenant_a_slug}\"}" "$b" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_A_CREATE 200 "$s" "$b"
tenant_a_id="$(json_value "$b" 'j["data"]["tenant"]["id"]')"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Menu B ${stamp}\",\"slug\":\"${tenant_b_slug}\"}" "$b" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_B_CREATE 200 "$s" "$b"
tenant_b_id="$(json_value "$b" 'j["data"]["tenant"]["id"]')"

log "[E2E] add tenant A admin/viewer members"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_a_id}/members" "{\"user_id\":\"${admin_id}\",\"role\":\"admin\",\"status\":\"active\"}" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_a_id}")"; assert_status MEMBER_ADD_ADMIN 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_a_id}/members" "{\"user_id\":\"${viewer_id}\",\"role\":\"viewer\",\"status\":\"active\"}" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_a_id}")"; assert_status MEMBER_ADD_VIEWER 200 "$s" "$b"

log "[E2E] verify /me/* endpoints stay tenant-free"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" -H "X-User-ID: ${owner_id}")"; assert_status OWNER_ACCESS 200 "$s" "$b"
assert_json_equals OWNER_HAS_TENANT_MANAGE "$b" "any(t['tenant_id'] == '${tenant_a_id}' and 'tenant:manage' in t['permissions'] for t in j['data']['tenants'])" true
assert_json_equals OWNER_PLATFORM_NULL "$b" 'j["data"]["platform_admin"] is None' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/invitations "" "$b" -H "X-User-ID: ${owner_id}")"; assert_status OWNER_MY_INVITATIONS 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/security-events "" "$b" -H "X-User-ID: ${owner_id}")"; assert_status OWNER_SECURITY_EVENTS 200 "$s" "$b"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" -H "X-User-ID: ${admin_id}")"; assert_status ADMIN_ACCESS 200 "$s" "$b"
assert_json_equals ADMIN_HAS_MEMBER_MANAGE "$b" "any(t['tenant_id'] == '${tenant_a_id}' and 'member:manage' in t['permissions'] for t in j['data']['tenants'])" true
assert_json_equals ADMIN_NO_TENANT_MANAGE "$b" "any(t['tenant_id'] == '${tenant_a_id}' and 'tenant:manage' in t['permissions'] for t in j['data']['tenants'])" false

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" -H "X-User-ID: ${viewer_id}")"; assert_status VIEWER_ACCESS 200 "$s" "$b"
assert_json_equals VIEWER_HAS_TENANT_READ "$b" "any(t['tenant_id'] == '${tenant_a_id}' and 'tenant:read' in t['permissions'] for t in j['data']['tenants'])" true
assert_json_equals VIEWER_NO_MEMBER_READ_EXTRA "$b" "any(t['tenant_id'] == '${tenant_a_id}' and 'member:manage' in t['permissions'] for t in j['data']['tenants'])" false

log "[E2E] verify legacy API key route is removed"
b="$(bodyfile)"; s="$(http_request GET /api/v1/api-keys "" "$b" -H "X-User-ID: ${owner_id}")"; assert_status LEGACY_APIKEY_ROUTE_REMOVED 404 "$s" "$b"

log "[E2E] verify tenant-scoped API key flow"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_a_id}/api-keys" "{\"name\":\"menu-key\",\"scopes\":[\"user:read\",\"user:tenant:read\",\"tenant:read\"]}" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_a_id}")"; assert_status TENANT_APIKEY_CREATE 200 "$s" "$b"
raw_key="$(json_value "$b" 'j["data"]["raw_key"]')"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_a_id}/api-keys" "" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_a_id}")"; assert_status TENANT_A_APIKEY_LIST 200 "$s" "$b"
assert_json_equals TENANT_A_APIKEY_VISIBLE "$b" 'len(j["data"]["api_keys"]) == 1' true
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_b_id}/api-keys" "" "$b" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_b_id}")"; assert_status TENANT_B_APIKEY_LIST 200 "$s" "$b"
assert_json_equals TENANT_B_APIKEY_HIDDEN "$b" 'len(j["data"]["api_keys"]) == 0' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_a_id}")"; assert_status APIKEY_TENANT_CONTEXT_OK 200 "$s" "$b"
assert_json_equals APIKEY_TENANT_CONTEXT_MATCH "$b" 'j["data"]["tenant_context"]["tenant_id"]' "$tenant_a_id"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_TENANT_REQUIRED 400 "$s" "$b"
assert_json_equals APIKEY_TENANT_REQUIRED_CODE "$b" 'j["code"]' TENANT_REQUIRED
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "$b" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_b_id}")"; assert_status APIKEY_CROSS_TENANT_FORBIDDEN 403 "$s" "$b"
assert_json_equals APIKEY_CROSS_TENANT_CODE "$b" 'j["code"]' TENANT_GRANT_FORBIDDEN

log "[E2E] verify platform routes do not require tenant and reject api-key identities"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "$b" -H "X-User-ID: ${support_id}")"; assert_status SUPPORT_ACCESS 200 "$s" "$b"
assert_json_equals SUPPORT_HAS_PLATFORM_USER_READ "$b" '"platform:user:read" in j["data"]["platform_admin"]["permissions"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "$b" -H "X-User-ID: ${support_id}")"; assert_status PLATFORM_TENANTS_NO_TENANT 200 "$s" "$b"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "$b" -H "Authorization: Bearer ${raw_key}")"; assert_status PLATFORM_APIKEY_REJECTED 403 "$s" "$b"
assert_json_equals PLATFORM_APIKEY_REJECTED_CODE "$b" 'j["code"]' PLATFORM_ADMIN_SESSION_REQUIRED
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/plans "" "$b" -H "X-User-ID: ${support_id}")"; assert_status ADMIN_PLANS_REMOVED 404 "$s" "$b"

log "[E2E] all menu permission scenarios passed tenant_a=${tenant_a_id} tenant_b=${tenant_b_id}"
