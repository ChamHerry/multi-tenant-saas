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
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/repomind-frontend-server.log"
SCENARIO_LOG="${LOG_DIR}/frontend-spa-scenarios.log"
: >"${SCENARIO_LOG}"

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }
}
need go
need curl
need python3
need npm

psql_query() {
  local sql="$1"
  if command -v psql >/dev/null 2>&1; then
    psql -Atc "$sql"
    return
  fi
  if command -v docker >/dev/null 2>&1; then
    for container in multi-tenant-saas-postgres; do
      if docker ps --format '{{.Names}}' | grep -qx "${container}"; then
        docker exec -e PGPASSWORD="${PGPASSWORD}" "${container}" psql -U "${PGUSER}" -d "${PGDATABASE}" -Atc "$sql"
        return
      fi
    done
  fi
  echo "missing dependency: psql or docker container multi-tenant-saas-postgres" >&2
  exit 1
}
source "${ROOT_DIR}/hack/lib/e2e_auth_session.sh"

server_pid=""
cleanup() {
  if [[ -n "${server_pid}" ]] && kill -0 "${server_pid}" >/dev/null 2>&1; then
    kill "${server_pid}" >/dev/null 2>&1 || true
    wait "${server_pid}" >/dev/null 2>&1 || true
  fi
  # go run may leave the compiled child process listening briefly; the script
  # refused pre-existing listeners at startup, so any remaining :8000 listener
  # here belongs to this e2e run.
  lsof -tiTCP:8000 -sTCP:LISTEN 2>/dev/null | xargs -r kill >/dev/null 2>&1 || true
}
trap cleanup EXIT

log() {
  echo "$*" | tee -a "${SCENARIO_LOG}"
}

wait_for_free_port() {
  for _ in $(seq 1 120); do
    if ! lsof -tiTCP:8000 -sTCP:LISTEN >/dev/null 2>&1; then
      return
    fi
    sleep 0.25
  done
  echo "port 8000 is already in use; stop the existing local server before running e2e" >&2
  exit 1
}

start_server() {
  wait_for_free_port
  (cd "${ROOT_DIR}" && go run . >"${SERVER_LOG}" 2>&1 & echo $! >"${LOG_DIR}/frontend-server.pid")
  server_pid="$(cat "${LOG_DIR}/frontend-server.pid")"
  for _ in $(seq 1 60); do
    if curl -sS -f "${BASE_URL}/healthz" >/dev/null 2>&1; then
      return
    fi
    if ! kill -0 "${server_pid}" >/dev/null 2>&1; then
      echo "server exited during startup" >&2
      tail -120 "${SERVER_LOG}" >&2 || true
      exit 1
    fi
    sleep 0.25
  done
  echo "server did not become ready" >&2
  tail -120 "${SERVER_LOG}" >&2 || true
  exit 1
}

create_user() {
  local tag="$1"
  local email="${tag}@example.test"
  e2e_create_password_user "$tag"
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

assert_contains() {
  local id="$1" file="$2" needle="$3"
  if ! grep -Fq "${needle}" "${file}"; then
    log "[FAIL] ${id}: expected body to contain ${needle}; body=$(head -c 300 "${file}")"
    exit 1
  fi
  log "[PASS] ${id}: body contains ${needle}"
}

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j=json.load(f)
value=eval(sys.argv[2], {"__builtins__": {}}, {"j": j})
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
else:
    print(value)
PY
}

bodyfile() {
  mktemp "${LOG_DIR}/frontend-body.XXXXXX"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
owner_email="fe-owner-${stamp}"
viewer_email="fe-viewer-${stamp}"
tenant_slug="fe-${stamp}"
tenant_name="Frontend ${stamp}"

log "[E2E] build frontend"
(cd "${ROOT_DIR}/web" && npm run build >/dev/null)

log "[E2E] configure local runtime for HTTP e2e"
e2e_configure_local_runtime

log "[E2E] setup users"
owner_id="$(create_user "${owner_email}")"
viewer_id="$(create_user "${viewer_email}")"

log "[E2E] start server"
start_server

b="$(bodyfile)"; s="$(http_request GET / "" "${b}")"; assert_status SPA_ROOT 200 "${s}" "${b}"; assert_contains SPA_ROOT_HTML "${b}" "SaaS Template Console"
b="$(bodyfile)"; s="$(http_request GET /tenants "" "${b}")"; assert_status SPA_TENANTS_FALLBACK 200 "${s}" "${b}"; assert_contains SPA_TENANTS_HTML "${b}" "SaaS Template Console"
b="$(bodyfile)"; s="$(http_request GET /members "" "${b}")"; assert_status SPA_MEMBERS_FALLBACK 200 "${s}" "${b}"; assert_contains SPA_MEMBERS_HTML "${b}" "SaaS Template Console"
b="$(bodyfile)"; s="$(http_request GET /api-keys "" "${b}")"; assert_status SPA_PERSONAL_API_KEYS_FALLBACK 200 "${s}" "${b}"; assert_contains SPA_PERSONAL_API_KEYS_HTML "${b}" "SaaS Template Console"
b="$(bodyfile)"; s="$(http_request GET /assets/not-found.js "" "${b}")"; assert_status ASSET_MISSING_NOT_FALLBACK 404 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}")"; assert_status API_NOT_FALLBACK 401 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /healthz "" "${b}")"; assert_status HEALTHZ_NOT_FALLBACK 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ_NOT_FALLBACK 200 "${s}" "${b}"
owner_cookie="${LOG_DIR}/frontend-owner.cookies"
e2e_login_user_tag "$owner_email" "$owner_cookie"
owner_csrf="$(e2e_csrf_from_cookie_jar "$owner_cookie")"
owner_auth=(-b "$owner_cookie" -c "$owner_cookie" -H "X-CSRF-Token: ${owner_csrf}")

create_payload="{\"name\":\"${tenant_name}\",\"slug\":\"${tenant_slug}\"}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "${create_payload}" "${b}" "${owner_auth[@]}")"; assert_status TENANT_CREATE_FOR_UI 200 "${s}" "${b}"
tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}" "${owner_auth[@]}")"; assert_status DEV_LOGIN_ME 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/tenants "" "${b}" "${owner_auth[@]}")"; assert_status DEV_LOGIN_TENANTS 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_CONTEXT 200 "${s}" "${b}"

add_member_payload="{\"user_id\":\"${viewer_id}\",\"role\":\"viewer\",\"status\":\"active\"}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/members" "${add_member_payload}" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status MEMBER_ADD 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/members" "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status MEMBER_LIST 200 "${s}" "${b}"

personal_key_payload="{\"name\":\"frontend-smoke\",\"scopes\":[\"tenant:read\",\"member:read\"]}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" "${personal_key_payload}" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status PERSONAL_APIKEY_CREATE 200 "${s}" "${b}"
raw_key="$(json_value "${b}" 'j["data"]["raw_key"]')"
if [[ "${raw_key}" != saas_* ]]; then
  log "[FAIL] PERSONAL_APIKEY_RAW_PREFIX: unexpected raw key ${raw_key}"
  exit 1
fi
log "[PASS] PERSONAL_APIKEY_RAW_PREFIX"

log "[E2E] frontend SPA scenarios passed tenant=${tenant_id} owner=${owner_id} viewer=${viewer_id}"
