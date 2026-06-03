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
SERVER_LOG="${LOG_DIR}/repomind-server.log"
SCENARIO_LOG="${LOG_DIR}/multitenancy-http-scenarios.log"
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

start_server() {
  if lsof -tiTCP:8000 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "port 8000 is already in use; stop the existing local server before running e2e" >&2
    exit 1
  fi
  (cd "${ROOT_DIR}" && go run . >"${SERVER_LOG}" 2>&1 & echo $! >"${LOG_DIR}/server.pid")
  server_pid="$(cat "${LOG_DIR}/server.pid")"
  for _ in $(seq 1 40); do
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

bodyfile() {
  mktemp "${LOG_DIR}/body.XXXXXX"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
owner_email="e2e-owner-${stamp}"
viewer_email="e2e-viewer-${stamp}"
outsider_email="e2e-outsider-${stamp}"
tenant_slug="e2e-${stamp}"
tenant_name="E2E ${stamp}"

log "[E2E] configure local runtime for HTTP e2e"
e2e_configure_local_runtime

log "[E2E] setup users"
owner_id="$(create_user "${owner_email}")"
viewer_id="$(create_user "${viewer_email}")"
outsider_id="$(create_user "${outsider_email}")"

log "[E2E] start server"
start_server

b="$(bodyfile)"; s="$(http_request GET /healthz "" "${b}")"; assert_status HEALTHZ 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ 200 "${s}" "${b}"; assert_json_equals READYZ_OK "${b}" 'j["ok"]' true

b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}")"; assert_status AUTH_MISSING 401 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}" -H "Authorization: Bearer not-a-real-key")"; assert_status AUTH_BAD_BEARER 401 "${s}" "${b}"
owner_cookie="${LOG_DIR}/owner.cookies"
viewer_cookie="${LOG_DIR}/viewer.cookies"
outsider_cookie="${LOG_DIR}/outsider.cookies"
e2e_login_user_tag "$owner_email" "$owner_cookie"
e2e_login_user_tag "$viewer_email" "$viewer_cookie"
e2e_login_user_tag "$outsider_email" "$outsider_cookie"
owner_csrf="$(e2e_csrf_from_cookie_jar "$owner_cookie")"
viewer_csrf="$(e2e_csrf_from_cookie_jar "$viewer_cookie")"
outsider_csrf="$(e2e_csrf_from_cookie_jar "$outsider_cookie")"
owner_auth=(-b "$owner_cookie" -c "$owner_cookie" -H "X-CSRF-Token: ${owner_csrf}")
viewer_auth=(-b "$viewer_cookie" -c "$viewer_cookie" -H "X-CSRF-Token: ${viewer_csrf}")
outsider_auth=(-b "$outsider_cookie" -c "$outsider_cookie" -H "X-CSRF-Token: ${outsider_csrf}")

create_payload="{\"name\":\"${tenant_name}\",\"slug\":\"${tenant_slug}\"}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "${create_payload}" "${b}" "${owner_auth[@]}")"; assert_status TENANT_CREATE 200 "${s}" "${b}"
tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"
assert_json_equals TENANT_CREATE_SLUG "${b}" 'j["data"]["tenant"]["slug"]' "${tenant_slug}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me "" "${b}" "${owner_auth[@]}")"; assert_status ME_OWNER 200 "${s}" "${b}"; assert_json_equals ME_OWNER_ID "${b}" 'j["data"]["user"]["id"]' "${owner_id}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/me/tenants "" "${b}" "${owner_auth[@]}")"; assert_status ME_TENANTS 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${owner_auth[@]}")"; assert_status TENANT_MISSING 400 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${outsider_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_OUTSIDER 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ../etc/passwd")"; assert_status TENANT_MALFORMED_SELECTOR 403 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_CONTEXT_OWNER 200 "${s}" "${b}"; assert_json_equals TENANT_CONTEXT_ROLE "${b}" 'j["data"]["tenant_context"]["role"]' owner

add_member_payload="{\"user_id\":\"${viewer_id}\",\"role\":\"viewer\",\"status\":\"active\"}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/members" "${add_member_payload}" "${b}" "${owner_auth[@]}")"; assert_status MEMBER_ADD_VIEWER 200 "${s}" "${b}"; assert_json_equals MEMBER_ADD_ROLE "${b}" 'j["data"]["member"]["role"]' viewer

b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" "${viewer_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_CONTEXT_VIEWER 200 "${s}" "${b}"; assert_json_equals VIEWER_ROLE "${b}" 'j["data"]["tenant_context"]["role"]' viewer

b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/tenants/${tenant_id}" '{"name":"viewer attack"}' "${b}" "${viewer_auth[@]}")"; assert_status VIEWER_TENANT_UPDATE_FORBIDDEN 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/tenants/${tenant_id}" '{"name":"owner updated"}' "${b}" "${owner_auth[@]}")"; assert_status OWNER_TENANT_UPDATE 200 "${s}" "${b}"; assert_json_equals OWNER_TENANT_UPDATE_NAME "${b}" 'j["data"]["tenant"]["name"]' "owner updated"

b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/api-key-grants" "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status ORG_APIKEY_GRANTS_REMOVED 404 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" '{"name":"bad-scope","scopes":["unknown"]}' "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_BAD_SCOPE 400 "${s}" "${b}"
personal_key_payload="{\"name\":\"limited\",\"scopes\":[\"tenant:read\",\"member:read\"]}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" "${personal_key_payload}" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_CREATE_LIMITED 200 "${s}" "${b}"
raw_key="$(json_value "${b}" 'j["data"]["raw_key"]')"
api_key_id="$(json_value "${b}" 'j["data"]["api_key"]["id"]')"

b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/api-keys" "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_LIST_TENANT 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/api-keys" "" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_SELF_MANAGE_FORBIDDEN 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_TENANT_REQUIRED 400 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_AUTH_CONTEXT 200 "${s}" "${b}"; assert_json_equals APIKEY_AUTH_TYPE "${b}" 'j["data"]["tenant_context"]["auth_type"]' api_key
member_manage_payload="{\"user_id\":\"${outsider_id}\",\"role\":\"viewer\",\"status\":\"active\"}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/members" "${member_manage_payload}" "${b}" -H "Authorization: Bearer ${raw_key}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_SCOPE_FORBIDDEN 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/tenants/${tenant_id}/api-keys/${api_key_id}" "" "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status APIKEY_REVOKE 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "Authorization: Bearer ${raw_key}")"; assert_status APIKEY_REVOKED_REJECTED 401 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/api-keys" '{bad-json' "${b}" "${owner_auth[@]}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status MALFORMED_JSON 400 "${s}" "${b}"

log "[E2E] all scenarios passed tenant=${tenant_id} owner=${owner_id} viewer=${viewer_id} outsider=${outsider_id}"
