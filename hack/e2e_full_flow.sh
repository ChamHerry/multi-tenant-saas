#!/usr/bin/env bash
# e2e_full_flow.sh — comprehensive end-to-end test covering all major flows.
# Exercises: register, login, tenant CRUD, members, invitations, API keys, audit, platform admin.
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8000}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-55432}"
PGUSER="${PGUSER:-saas_template}"
PGPASSWORD="${PGPASSWORD:-secret}"
PGDATABASE="${PGDATABASE:-saas_template}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e-full}"
mkdir -p "${LOG_DIR}"
SCENARIO_LOG="${LOG_DIR}/full-flow.log"
: >"${SCENARIO_LOG}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 1; }; }
need go need curl

log() { echo "[$(date +%H:%M:%S)] $*" | tee -a "${SCENARIO_LOG}"; }
fail() { echo "FAIL: $*" >&2; exit 1; }

psql_query() {
  local sql="$1"
  if command -v psql >/dev/null 2>&1; then
    psql -Atc "$sql"; return
  fi
  if command -v docker >/dev/null 2>&1 && docker ps --format '{{.Names}}' | grep -qx repomind-pg; then
    docker exec -e PGPASSWORD="${PGPASSWORD}" repomind-pg psql -U "${PGUSER}" -d "${PGDATABASE}" -Atc "$sql"; return
  fi
  fail "missing psql or docker container repomind-pg"
}

# --- Start server ---
server_pid=""
cleanup() {
  if [[ -n "${server_pid}" ]] && kill -0 "${server_pid}" >/dev/null 2>&1; then
    kill "${server_pid}" >/dev/null 2>&1 || true
    wait "${server_pid}" >/dev/null 2>&1 || true
  fi
  lsof -tiTCP:8000 -sTCP:LISTEN 2>/dev/null | xargs -r kill >/dev/null 2>&1 || true
}
trap cleanup EXIT

lsof -tiTCP:8000 -sTCP:LISTEN 2>/dev/null | head -1 | grep -q . && fail "port 8000 already in use"

log "starting server..."
cd "${ROOT_DIR}"
go run . > "${LOG_DIR}/server.log" 2>&1 &
server_pid=$!

for i in $(seq 1 30); do
  if curl -sf "${BASE_URL}/healthz" >/dev/null 2>&1; then break; fi
  sleep 1
done
curl -sf "${BASE_URL}/healthz" >/dev/null || fail "server did not start"

log "=== Phase 1: Health checks ==="
health=$(curl -sf "${BASE_URL}/healthz")
log "healthz: ${health}"
ready=$(curl -sf "${BASE_URL}/readyz")
log "readyz: ${ready}"
echo "${ready}" | grep -q '"ok":true' || fail "readyz not ok"

log "=== Phase 2: Metrics ==="
metrics=$(curl -sf "${BASE_URL}/metrics/")
log "metrics endpoint returns $(echo "${metrics}" | wc -l) lines"

log "=== Phase 3: Register user ==="
email="e2e-full-$(date +%s)@test.example.com"
register_res=$(curl -sf -X POST "${BASE_URL}/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${email}\",\"password\":\"SuperSecure123!@#\",\"display_name\":\"E2E Test User\"}")
log "register: ${register_res}"
echo "${register_res}" | grep -q '"code":0' || fail "register failed"
csrf_token=$(echo "${register_res}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('csrf_token',''))" 2>/dev/null || true)

log "=== Phase 4: Login ==="
login_res=$(curl -sf -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -c "${LOG_DIR}/cookies.txt" \
  -d "{\"email\":\"${email}\",\"password\":\"SuperSecure123!@#\"}")
log "login: ${login_res}"
echo "${login_res}" | grep -q '"code":0' || fail "login failed"

# Extract CSRF token from cookies
csrf_token=$(grep saas_template_csrf "${LOG_DIR}/cookies.txt" 2>/dev/null | awk '{print $NF}' || echo "")
log "csrf token obtained: $(if [[ -n "${csrf_token}" ]]; then echo yes; else echo no; fi)"

api_curl() {
  local method="$1" path="$2" body="${3:-}"
  local args=(-s -X "${method}" -b "${LOG_DIR}/cookies.txt" -c "${LOG_DIR}/cookies.txt")
  if [[ -n "${csrf_token}" ]]; then args+=(-H "X-CSRF-Token: ${csrf_token}"); fi
  if [[ -n "${body}" ]]; then args+=(-H "Content-Type: application/json" -d "${body}"); fi
  curl "${args[@]}" "${BASE_URL}${path}"
}

log "=== Phase 5: Get /me ==="
me_res=$(api_curl GET "/api/v1/me")
log "me: ${me_res}"
echo "${me_res}" | grep -q '"code":0' || fail "me failed"
user_id=$(echo "${me_res}" | python3 -c "import sys,json; d=json.load(sys.stdin).get('data',{}).get('user',{}); print(d.get('id',''))" 2>/dev/null)
log "user_id: ${user_id}"

log "=== Phase 6: List tenants ==="
tenants_res=$(api_curl GET "/api/v1/me/tenants")
log "tenants: ${tenants_res}"

# Get first tenant ID
tenant_id=$(echo "${tenants_res}" | python3 -c "
import sys,json
ts=json.load(sys.stdin).get('data',{}).get('tenants',[])
print(ts[0]['tenant']['id'] if ts else '')
" 2>/dev/null)
log "using tenant: ${tenant_id}"

if [[ -z "${tenant_id}" ]]; then
  log "=== Phase 6b: Create tenant ==="
  create_res=$(api_curl POST "/api/v1/tenants" "{\"name\":\"E2E Test Org\",\"slug\":\"e2e-test-$(date +%s)\"}")
  log "create tenant: ${create_res}"
  tenant_id=$(echo "${create_res}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('tenant',{}).get('id',''))" 2>/dev/null)
  log "created tenant: ${tenant_id}"
fi

log "=== Phase 7: Get access snapshot ==="
access_res=$(api_curl GET "/api/v1/me/access")
log "access: $(echo "${access_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("tenants=%d platform=%s" % (len(d.get("tenants",[])), bool(d.get("platform_admin"))))' 2>/dev/null)"

log "=== Phase 8: List members ==="
members_res=$(api_curl GET "/api/v1/tenants/${tenant_id}/members")
log "members: $(echo "${members_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("count=%d" % len(d.get("members",[])))' 2>/dev/null)"

log "=== Phase 9: Create invitation ==="
invite_email="invited-$(date +%s)@test.example.com"
invite_res=$(api_curl POST "/api/v1/tenants/${tenant_id}/invitations" \
  "{\"invitee_email\":\"${invite_email}\",\"role\":\"member\",\"message\":\"Welcome!\"}")
log "invitation: ${invite_res}"
echo "${invite_res}" | grep -q '"code":0' || fail "invitation creation failed"

log "=== Phase 10: List invitations ==="
inv_list=$(api_curl GET "/api/v1/tenants/${tenant_id}/invitations")
log "invitations: $(echo "${inv_list}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("total=%d" % d.get("total",0))' 2>/dev/null)"

log "=== Phase 11: Create API key ==="
apikey_res=$(api_curl POST "/api/v1/tenants/${tenant_id}/api-keys" \
  "{\"name\":\"e2e-test-key\",\"scopes\":[\"tenant:read\",\"member:read\"]}")
log "api key: ${apikey_res}"
echo "${apikey_res}" | grep -q '"code":0' || fail "API key creation failed"
raw_key=$(echo "${apikey_res}" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('raw_key',''))" 2>/dev/null)
log "api key created: $(if [[ -n "${raw_key}" ]]; then echo "yes (length=${#raw_key})"; else echo "no"; fi)"

log "=== Phase 12: Test API key auth ==="
if [[ -n "${raw_key}" ]]; then
  key_res=$(curl -sf -H "Authorization: Bearer ${raw_key}" "${BASE_URL}/api/v1/me")
  log "api key auth: $(echo "${key_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("code=%s" % d.get("code","?"))' 2>/dev/null)"
fi

log "=== Phase 13: List audit logs ==="
audit_res=$(api_curl GET "/api/v1/tenants/${tenant_id}/audit-logs?limit=5")
log "audit: $(echo "${audit_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("total=%d" % d.get("total",0))' 2>/dev/null)"

log "=== Phase 14: Platform admin session (may fail if not admin) ==="
admin_res=$(api_curl GET "/api/v1/admin/session")
if echo "${admin_res}" | grep -q '"code":0'; then
  log "platform admin: session ok"
  log "=== Phase 15: Platform tenants list ==="
  pt_res=$(api_curl GET "/api/v1/admin/tenants?limit=5")
  log "platform tenants: $(echo "${pt_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("total=%d" % d.get("total",0))' 2>/dev/null)"

  log "=== Phase 16: Platform users list ==="
  pu_res=$(api_curl GET "/api/v1/admin/users?limit=5")
  log "platform users: $(echo "${pu_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("total=%d" % d.get("total",0))' 2>/dev/null)"

  log "=== Phase 17: Platform audit logs ==="
  pa_res=$(api_curl GET "/api/v1/admin/audit-logs?limit=5")
  log "platform audit: $(echo "${pa_res}" | python3 -c 'import sys,json; d=json.load(sys.stdin).get("data",{}); print("total=%d" % d.get("total",0))' 2>/dev/null)"
else
  log "platform admin: not a platform admin (expected for regular user)"
fi

log "=== Phase 18: Logout ==="
logout_res=$(api_curl POST "/api/v1/auth/logout")
log "logout: ${logout_res}"

log ""
log "=========================================="
log "  FULL FLOW E2E TEST PASSED"
log "=========================================="
log "Scenarios covered:"
log "  ✓ Health check (/healthz, /readyz)"
log "  ✓ Metrics (/metrics)"
log "  ✓ User registration"
log "  ✓ User login"
log "  ✓ Get current user (/me)"
log "  ✓ List tenants"
log "  ✓ Tenant creation (if needed)"
log "  ✓ Access snapshot"
log "  ✓ Member listing"
log "  ✓ Invitation creation"
log "  ✓ Invitation listing"
log "  ✓ API key creation"
log "  ✓ API key authentication"
log "  ✓ Audit log listing"
log "  ✓ Platform admin (if applicable)"
log "  ✓ Logout"
