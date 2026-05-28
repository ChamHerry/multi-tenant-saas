#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8000}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-55432}"
PGUSER="${PGUSER:-saas_template}"
PGPASSWORD="${PGPASSWORD:-secret}"
PGDATABASE="${PGDATABASE:-saas_template}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e-menu-permissions}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/repomind-server.log"
SCENARIO_LOG="${LOG_DIR}/menu-permissions.log"
: >"${SCENARIO_LOG}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }; }
need go
need curl
need python3

psql_query() {
  local sql="$1"
  if command -v psql >/dev/null 2>&1; then
    psql -Atc "$sql"
    return
  fi
  if command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' | grep -qx repomind-pg; then
    docker exec -e PGPASSWORD="${PGPASSWORD}" repomind-pg psql -U "${PGUSER}" -d "${PGDATABASE}" -Atc "$sql"
    return
  fi
  echo "missing dependency: psql or docker container repomind-pg" >&2
  exit 1
}

server_pid=""
cleanup() {
  if [[ -n "${server_pid}" ]] && kill -0 "${server_pid}" >/dev/null 2>&1; then
    kill "${server_pid}" >/dev/null 2>&1 || true
    wait "${server_pid}" >/dev/null 2>&1 || true
  fi
  lsof -tiTCP:8000 -sTCP:LISTEN 2>/dev/null | xargs -r kill >/dev/null 2>&1 || true
}
trap cleanup EXIT

log() { echo "$*" | tee -a "${SCENARIO_LOG}"; }

start_server() {
  if lsof -tiTCP:8000 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "port 8000 is already in use; stop the existing local server before running e2e" >&2
    exit 1
  fi
  (cd "${ROOT_DIR}" && go run . >"${SERVER_LOG}" 2>&1 & echo $! >"${LOG_DIR}/server.pid")
  server_pid="$(cat "${LOG_DIR}/server.pid")"
  for _ in $(seq 1 60); do
    if curl -sS -f "${BASE_URL}/healthz" >/dev/null 2>&1; then return; fi
    if ! kill -0 "${server_pid}" >/dev/null 2>&1; then
      echo "server exited during startup" >&2
      tail -160 "${SERVER_LOG}" >&2 || true
      exit 1
    fi
    sleep 0.25
  done
  echo "server did not become ready" >&2
  tail -160 "${SERVER_LOG}" >&2 || true
  exit 1
}

create_user() {
  local tag="$1"
  local email="${tag}@example.test"
  (cd "${ROOT_DIR}" && go run . user-upsert --provider password --auth-id "${email}" --email "${email}" --name "${tag}" --email-verified >/dev/null)
  psql_query "SELECT id FROM public.users WHERE email='${email}' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1"
}

http_request() {
  local method="$1" path="$2" body="$3" outfile="$4"
  shift 4
  local args=(-sS -o "${outfile}" -w "%{http_code}" -X "${method}")
  if [[ -n "${body}" ]]; then args+=(-H "Content-Type: application/json" -d "${body}"); fi
  args+=("$@" "${BASE_URL}${path}")
  curl "${args[@]}"
}

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j=json.load(f)
value=eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len})
if isinstance(value, bool): print("true" if value else "false")
elif value is None: print("")
else: print(value)
PY
}

assert_status() {
  local id="$1" expected="$2" actual="$3" bodyfile="$4"
  if [[ "${actual}" != "${expected}" ]]; then
    log "[FAIL] ${id}: expected HTTP ${expected}, got ${actual}; body=$(cat "${bodyfile}")"
    exit 1
  fi
  log "[PASS] ${id}: HTTP ${actual}"
}

assert_json_equals() {
  local id="$1" file="$2" expr="$3" expected="$4"
  local actual
  actual="$(json_value "${file}" "${expr}")"
  if [[ "${actual}" != "${expected}" ]]; then
    log "[FAIL] ${id}: expected ${expr}=${expected}, got ${actual}; body=$(cat "${file}")"
    exit 1
  fi
  log "[PASS] ${id}: ${expr}=${actual}"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
owner_tag="menu-owner-${stamp}"
admin_tag="menu-admin-${stamp}"
viewer_tag="menu-viewer-${stamp}"
support_tag="menu-support-${stamp}"
tenant_slug="menu-${stamp}"

log "[E2E] setup users"
owner_id="$(create_user "${owner_tag}")"
admin_id="$(create_user "${admin_tag}")"
viewer_id="$(create_user "${viewer_tag}")"
support_id="$(create_user "${support_tag}")"
psql_query "INSERT INTO public.platform_admins(user_id, role, status, created_at, updated_at) VALUES ('${support_id}', 'support', 'active', now(), now()) ON CONFLICT (user_id) DO UPDATE SET role='support', status='active', updated_at=now()" >/dev/null

log "[E2E] start server"
start_server

b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Menu ${stamp}\",\"slug\":\"${tenant_slug}\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_CREATE 200 "${s}" "${b}"
tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"

b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/members" "{\"user_id\":\"${admin_id}\",\"role\":\"admin\",\"status\":\"active\"}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status MEMBER_ADD_ADMIN 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/members" "{\"user_id\":\"${viewer_id}\",\"role\":\"viewer\",\"status\":\"active\"}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status MEMBER_ADD_VIEWER 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status OWNER_ACCESS 200 "${s}" "${b}"
assert_json_equals OWNER_HAS_TENANT_MANAGE "${b}" '"tenant:manage" in j["data"]["tenants"][0]["permissions"]' true
assert_json_equals OWNER_NO_APIKEY_MANAGE "${b}" '"api_key:manage" in j["data"]["tenants"][0]["permissions"]' false
assert_json_equals OWNER_PLATFORM_NULL "${b}" 'j["data"]["platform_admin"] is None' true

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -H "X-User-ID: ${admin_id}")"; assert_status ADMIN_ACCESS 200 "${s}" "${b}"
assert_json_equals ADMIN_HAS_MEMBER_MANAGE "${b}" '"member:manage" in j["data"]["tenants"][0]["permissions"]' true
assert_json_equals ADMIN_NO_TENANT_MANAGE "${b}" '"tenant:manage" in j["data"]["tenants"][0]["permissions"]' false

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -H "X-User-ID: ${viewer_id}")"; assert_status VIEWER_ACCESS 200 "${s}" "${b}"
assert_json_equals VIEWER_HAS_TENANT_READ "${b}" '"tenant:read" in j["data"]["tenants"][0]["permissions"]' true
assert_json_equals VIEWER_NO_APIKEY_MANAGE "${b}" '"api_key:manage" in j["data"]["tenants"][0]["permissions"]' false

b="$(bodyfile)"; s="$(http_request GET /api/v1/api-keys "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status PERSONAL_APIKEY_ROUTE_ALLOWED 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -H "X-User-ID: ${support_id}")"; assert_status SUPPORT_ACCESS 200 "${s}" "${b}"
assert_json_equals SUPPORT_HAS_USER_READ "${b}" '"platform:user:read" in j["data"]["platform_admin"]["permissions"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/plans "" "${b}" -H "X-User-ID: ${support_id}")"; assert_status ADMIN_PLANS_REMOVED 404 "${s}" "${b}"

log "[E2E] all menu permission scenarios passed tenant=${tenant_id}"
