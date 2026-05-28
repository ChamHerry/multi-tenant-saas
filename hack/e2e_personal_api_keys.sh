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
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/saas-template-e2e-personal-api-keys}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/server.log"
SCENARIO_LOG="${LOG_DIR}/scenario.log"
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
  for _ in $(seq 1 80); do
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
owner_tag="personal-owner-${stamp}"
outsider_tag="personal-outsider-${stamp}"
tenant_slug="personal-${stamp}"
other_slug="personal-other-${stamp}"

log "[E2E] setup users"
owner_id="$(create_user "${owner_tag}")"
outsider_id="$(create_user "${outsider_tag}")"

log "[E2E] start server"
start_server

b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Personal ${stamp}\",\"slug\":\"${tenant_slug}\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_CREATE 200 "${s}" "${b}"
tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"

b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"Other ${stamp}\",\"slug\":\"${other_slug}\"}" "${b}" -H "X-User-ID: ${outsider_id}")"; assert_status OTHER_TENANT_CREATE 200 "${s}" "${b}"
other_tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"

create_key_payload="{\"name\":\"personal-reader\",\"scopes\":[\"user:read\",\"tenant:read\"],\"grants\":[{\"tenant_id\":\"${tenant_id}\",\"scopes\":[\"tenant:read\"]}]}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/api-keys "${create_key_payload}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status PERSONAL_KEY_CREATE 200 "${s}" "${b}"
raw_key="$(json_value "${b}" 'j["data"]["raw_key"]')"
api_key_id="$(json_value "${b}" 'j["data"]["api_key"]["id"]')"
if [[ "${raw_key}" != saas_* ]]; then
  log "[FAIL] PERSONAL_KEY_PREFIX: unexpected raw key ${raw_key}"
  exit 1
fi
log "[PASS] PERSONAL_KEY_PREFIX"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_ME 200 "${s}" "${b}"; assert_json_equals APIKEY_ME_ID "${b}" 'j["data"]["user"]["id"]' "${owner_id}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_TENANT_REQUIRED 400 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_TENANT_GRANTED 200 "${s}" "${b}"; assert_json_equals APIKEY_TENANT_ROLE "${b}" 'j["data"]["tenant_context"]["role"]' owner
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${other_tenant_id}")"; assert_status APIKEY_TENANT_NOT_MEMBER 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/members" "" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_SCOPE_INTERSECTION 403 "${s}" "${b}"

psql_query "INSERT INTO public.platform_admins(user_id, role, status, created_at, updated_at) VALUES ('${owner_id}', 'super_admin', 'active', now(), now()) ON CONFLICT (user_id) DO UPDATE SET role='super_admin', status='active', updated_at=now()" >/dev/null
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_PLATFORM_ADMIN_FORBIDDEN 403 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/api-keys/${api_key_id}" "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status PERSONAL_KEY_REVOKE 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status PERSONAL_KEY_REVOKED_REJECTED 401 "${s}" "${b}"

log "[E2E] personal API key scenarios passed tenant=${tenant_id} owner=${owner_id}"
