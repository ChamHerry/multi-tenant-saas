#!/usr/bin/env bash
set -euo pipefail

BASE_PORT="${BASE_PORT:-18081}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${BASE_PORT}}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-55432}"
PGUSER="${PGUSER:-repomind}"
PGPASSWORD="${PGPASSWORD:-secret}"
PGDATABASE="${PGDATABASE:-repomind}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/default-tenant-server.log"
SCENARIO_LOG="${LOG_DIR}/default-tenant-scenarios.log"
CONFIG_FILE="${LOG_DIR}/default-tenant-config.yaml"
: >"${SCENARIO_LOG}"

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }
}
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
  lsof -tiTCP:"${BASE_PORT}" -sTCP:LISTEN 2>/dev/null | xargs -r kill >/dev/null 2>&1 || true
}
trap cleanup EXIT

log() {
  echo "$*" | tee -a "${SCENARIO_LOG}"
}

write_config() {
  python3 - "${ROOT_DIR}/manifest/config/config.yaml" "${CONFIG_FILE}" "${BASE_PORT}" <<'PY'
from pathlib import Path
import sys
src, dst, port = sys.argv[1:4]
text = Path(src).read_text()
text = text.replace('address:     ":8000"', f'address:     ":{port}"')
text = text.replace('registrationEnabled: false', 'registrationEnabled: true')
Path(dst).write_text(text)
PY
}

start_server() {
  if lsof -tiTCP:"${BASE_PORT}" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "port ${BASE_PORT} is already in use; choose another BASE_PORT" >&2
    exit 1
  fi
  write_config
  (cd "${ROOT_DIR}" && GF_GCFG_FILE="${CONFIG_FILE}" go run . >"${SERVER_LOG}" 2>&1 & echo $! >"${LOG_DIR}/default-tenant-server.pid")
  server_pid="$(cat "${LOG_DIR}/default-tenant-server.pid")"
  for _ in $(seq 1 120); do
    if curl -sS -f "${BASE_URL}/healthz" >/dev/null 2>&1; then
      return
    fi
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

http_request() {
  local method="$1" path="$2" body="$3" outfile="$4"
  shift 4
  local args=(-sS -o "${outfile}" -w "%{http_code}" -X "${method}")
  if [[ -n "${body}" ]]; then
    args+=(-H "Content-Type: application/json" -d "${body}")
  fi
  args+=("$@" "${BASE_URL}${path}")
  curl "${args[@]}"
}

assert_status() {
  local id="$1" expected="$2" actual="$3" bodyfile="$4"
  if [[ "${actual}" != "${expected}" ]]; then
    log "[FAIL] ${id}: expected HTTP ${expected}, got ${actual}; body=$(cat "${bodyfile}")"
    exit 1
  fi
  log "[PASS] ${id}: HTTP ${actual}"
}

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j=json.load(f)
value=eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len})
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
else:
    print(value)
PY
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

assert_sql_equals() {
  local id="$1" sql="$2" expected="$3"
  local actual
  actual="$(psql_query "${sql}")"
  if [[ "${actual}" != "${expected}" ]]; then
    log "[FAIL] ${id}: expected sql=${expected}, got ${actual}"
    exit 1
  fi
  log "[PASS] ${id}: ${actual}"
}

bodyfile() {
  mktemp "${LOG_DIR}/default-tenant-body.XXXXXX"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
email="default-tenant-${stamp}@example.test"
password="0123456789abcde-${stamp}"
name="Default Organization ${stamp}"

log "[E2E] start server on ${BASE_URL}"
start_server

b="$(bodyfile)"; s="$(http_request GET /healthz "" "${b}")"; assert_status HEALTHZ 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ 200 "${s}" "${b}"; assert_json_equals READYZ_OK "${b}" 'j["ok"]' true

register_payload="{\"email\":\"${email}\",\"password\":\"${password}\",\"display_name\":\"${name}\"}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "${register_payload}" "${b}")"; assert_status REGISTER 200 "${s}" "${b}"
user_id="$(json_value "${b}" 'j["data"]["user"]["id"]')"
assert_json_equals REGISTER_EMAIL "${b}" 'j["data"]["user"]["email"]' "${email}"

b="$(bodyfile)"; s="$(http_request POST /api/v1/auth/register "${register_payload}" "${b}")"; assert_status REGISTER_DUPLICATE 409 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -H "X-User-ID: ${user_id}")"; assert_status ACCESS_DEFAULT_TENANT 200 "${s}" "${b}"; assert_json_equals ACCESS_TENANT_COUNT "${b}" 'len(j["data"]["tenants"])' 1
tenant_id="$(json_value "${b}" 'j["data"]["tenants"][0]["tenant_id"]')"
tenant_slug="$(json_value "${b}" 'j["data"]["tenants"][0]["tenant_slug"]')"
assert_json_equals ACCESS_DEFAULT_ROLE "${b}" 'j["data"]["tenants"][0]["role"]' owner
assert_json_equals ACCESS_NO_SCHEMA_FIELD "${b}" '"schema_name" in j["data"]["tenants"][0]' false
assert_sql_equals DEFAULT_TENANT_ROW "SELECT count(*) FROM public.tenants t JOIN public.tenant_memberships tm ON tm.tenant_id=t.id WHERE t.id='${tenant_id}' AND tm.user_id='${user_id}' AND tm.role='owner' AND t.metadata->>'default_personal_tenant'='true'" 1
assert_sql_equals LEGACY_ORGANIZATIONS_DROPPED "SELECT COALESCE(to_regclass('public.organizations')::text, '')" ""
assert_sql_equals TENANT_SCHEMA_COLUMNS_REMOVED "SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='tenants' AND column_name IN ('schema_name','graph_name','tenant_schema_version','tenant_schema_dirty','tenant_schema_checked_at')" 0

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/tenants "" "${b}" -H "X-User-ID: ${user_id}")"; assert_status ME_TENANTS 200 "${s}" "${b}"; assert_json_equals ME_TENANT_COUNT "${b}" 'len(j["data"]["tenants"])' 1
assert_json_equals ME_TENANT_SLUG "${b}" 'j["data"]["tenants"][0]["tenant_slug"]' "${tenant_slug}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "X-User-ID: ${user_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_CONTEXT_PUBLIC_MODEL 200 "${s}" "${b}"
assert_json_equals TENANT_CONTEXT_ID "${b}" 'j["data"]["tenant_context"]["tenant_id"]' "${tenant_id}"
assert_json_equals TENANT_CONTEXT_NO_SCHEMA_FIELD "${b}" '"schema_name" in j["data"]["tenant_context"]' false
assert_json_equals TENANT_CONTEXT_NO_GRAPH_FIELD "${b}" '"graph_name" in j["data"]["tenant_context"]' false
assert_sql_equals NO_TENANT_SCHEMA_CREATED "SELECT count(*) FROM information_schema.schemata WHERE schema_name LIKE 'tenant_%'" 0

log "[E2E] default tenant public-schema model passed user=${user_id} tenant=${tenant_id}"
