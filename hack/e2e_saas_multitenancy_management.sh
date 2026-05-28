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
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e-saas}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/repomind-server.log"
SCENARIO_LOG="${LOG_DIR}/saas-multitenancy-management.log"
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
if isinstance(value, bool): print("true" if value else "false")
elif value is None: print("")
else: print(value)
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

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }

stamp="$(date +%Y%m%d%H%M%S)-$$"
owner_tag="saas-owner-${stamp}"
invitee_tag="saas-invitee-${stamp}"
outsider_tag="saas-outsider-${stamp}"
tenant_slug="saas-${stamp}"

log "[E2E] setup users"
owner_id="$(create_user "${owner_tag}")"
invitee_id="$(create_user "${invitee_tag}")"
outsider_id="$(create_user "${outsider_tag}")"
psql_query "INSERT INTO public.platform_admins(user_id, role, status, created_at, updated_at) VALUES ('${owner_id}', 'super_admin', 'active', now(), now()) ON CONFLICT (user_id) DO UPDATE SET role='super_admin', status='active', updated_at=now()" >/dev/null

log "[E2E] start server"
start_server

b="$(bodyfile)"; s="$(http_request GET /readyz "" "${b}")"; assert_status READYZ 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST /api/v1/tenants "{\"name\":\"SaaS ${stamp}\",\"slug\":\"${tenant_slug}\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status TENANT_CREATE 200 "${s}" "${b}"
tenant_id="$(json_value "${b}" 'j["data"]["tenant"]["id"]')"

invite_payload="{\"invitee_email\":\"${invitee_tag}@example.test\",\"role\":\"member\",\"message\":\"welcome\"}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/invitations" "${invite_payload}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITE_CREATE 200 "${s}" "${b}"
invitation_token="$(json_value "${b}" 'j["data"]["token"]')"
assert_json_equals INVITE_STATUS "${b}" 'j["data"]["invitation"]["status"]' pending

b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/invitations" "" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITE_LIST_TENANT 200 "${s}" "${b}"
assert_json_equals INVITE_LIST_COUNT "${b}" 'len(j["data"]["invitations"])' 1

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/invitations "" "${b}" -H "X-User-ID: ${invitee_id}")"; assert_status INVITE_LIST_MINE 200 "${s}" "${b}"
assert_json_equals INVITE_MINE_COUNT "${b}" 'len(j["data"]["invitations"])' 1

b="$(bodyfile)"; s="$(http_request POST /api/v1/invitations/accept "{\"token\":\"${invitation_token}\"}" "${b}" -H "X-User-ID: ${invitee_id}")"; assert_status INVITE_ACCEPT 200 "${s}" "${b}"
assert_json_equals INVITE_ACCEPT_ROLE "${b}" 'j["data"]["member"]["role"]' member

b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "X-User-ID: ${invitee_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITEE_TENANT_CONTEXT 200 "${s}" "${b}"
assert_json_equals INVITEE_ROLE "${b}" 'j["data"]["tenant_context"]["role"]' member

b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/audit-logs" "" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status AUDIT_GET 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "${b}" -H "X-User-ID: ${outsider_id}")"; assert_status ADMIN_FORBIDDEN 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/tenants "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_TENANTS 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/session "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_SESSION 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/plans "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_PLANS_REMOVED 404 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/admin/tenants/${tenant_id}/plan" "{\"plan\":\"pro\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_TENANT_PLAN_REMOVED 404 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/admin/tenants/${tenant_id}/quota" "{\"max_members\":7}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_TENANT_QUOTA_REMOVED 404 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET "/api/v1/tenants/${tenant_id}/quota" "" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status TENANT_QUOTA_REMOVED 404 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/admin/users/${outsider_id}/status" "{\"status\":\"disabled\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_USER_DISABLE 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request PATCH "/api/v1/admin/users/${outsider_id}/status" "{\"status\":\"active\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_USER_ENABLE 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/platform-admins "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_LIST_PLATFORM_ADMINS 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request POST /api/v1/admin/platform-admins "{\"user_id\":\"${outsider_id}\",\"role\":\"auditor\"}" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_GRANT_PLATFORM_ADMIN 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request DELETE "/api/v1/admin/platform-admins/${outsider_id}" "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_REVOKE_PLATFORM_ADMIN 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/admin/tenants/${tenant_id}/suspend" "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_TENANT_SUSPEND 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "X-User-ID: ${invitee_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status SUSPENDED_TENANT_BLOCKED 403 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/admin/tenants/${tenant_id}/restore" "" "${b}" -H "X-User-ID: ${owner_id}")"; assert_status ADMIN_TENANT_RESTORE 200 "${s}" "${b}"
b="$(bodyfile)"; s="$(http_request GET /api/v1/tenant-context "" "${b}" -H "X-User-ID: ${invitee_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status RESTORED_TENANT_CONTEXT 200 "${s}" "${b}"

second_payload="{\"invitee_email\":\"${outsider_tag}@example.test\",\"role\":\"viewer\"}"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/invitations" "${second_payload}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITE_CREATE_SECOND 200 "${s}" "${b}"
second_invitation_id="$(json_value "${b}" 'j["data"]["invitation"]["id"]')"
b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/invitations/${second_invitation_id}/revoke" "" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITE_REVOKE 200 "${s}" "${b}"

b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/audit-logs/export" "{\"action\":\"tenant.invitation.create\"}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status AUDIT_EXPORT 200 "${s}" "${b}"
assert_json_equals AUDIT_EXPORT_STATUS "${b}" 'j["data"]["job"]["status"]' succeeded

b="$(bodyfile)"; s="$(http_request POST "/api/v1/tenants/${tenant_id}/invitations" "{\"invitee_email\":\"extra-${outsider_tag}@example.test\",\"role\":\"viewer\"}" "${b}" -H "X-User-ID: ${owner_id}" -H "X-Tenant-ID: ${tenant_id}")"; assert_status INVITE_CREATE_WITHOUT_MEMBER_LIMIT 200 "${s}" "${b}"

log "[E2E] all SaaS multitenancy management scenarios passed tenant=${tenant_id} owner=${owner_id} invitee=${invitee_id}"
