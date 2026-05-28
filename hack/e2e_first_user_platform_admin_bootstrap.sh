#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8000}"
PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-55432}"
PGUSER="${PGUSER:-saas_template}"
PGPASSWORD="${PGPASSWORD:-secret}"
PGDATABASE="${PGDATABASE:-saas_template}"
RESET_PUBLIC_DB="${RESET_PUBLIC_DB:-}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/repomind-e2e-first-admin-bootstrap}"
mkdir -p "${LOG_DIR}"
SERVER_LOG="${LOG_DIR}/repomind-server.log"
SCENARIO_LOG="${LOG_DIR}/first-admin-bootstrap.log"
: >"${SCENARIO_LOG}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }; }
need go
need curl
need python3

psql_query() {
  local sql="$1"
  local output
  if command -v psql >/dev/null 2>&1; then
    if output="$(psql -Atc "$sql" 2>&1)"; then
      printf '%s\n' "$output"
      return
    fi
  fi
  if command -v docker >/dev/null 2>&1; then
    for container in repomind-pg multi-tenant-saas-postgres; do
      if docker ps --format '{{.Names}}' | grep -qx "${container}"; then
        docker exec -e PGPASSWORD="${PGPASSWORD}" "${container}" psql -U "${PGUSER}" -d "${PGDATABASE}" -Atc "$sql"
        return
      fi
    done
  fi
  echo "missing usable database client: psql connection or docker container multi-tenant-saas-postgres/repomind-pg" >&2
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

log() { echo "$*" | tee -a "${SCENARIO_LOG}" >&2; }

start_server() {
  if lsof -tiTCP:8000 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "port 8000 is already in use; stop the existing local server before running e2e" >&2
    exit 1
  fi
  (cd "${ROOT_DIR}" && go run . >"${SERVER_LOG}" 2>&1 & echo $! >"${LOG_DIR}/server.pid")
  server_pid="$(cat "${LOG_DIR}/server.pid")"
  for _ in $(seq 1 80); do
    if curl -sS -f "${BASE_URL}/readyz" >/dev/null 2>&1; then return; fi
    if ! kill -0 "${server_pid}" >/dev/null 2>&1; then
      echo "server exited during startup" >&2
      tail -180 "${SERVER_LOG}" >&2 || true
      exit 1
    fi
    sleep 0.25
  done
  echo "server did not become ready" >&2
  tail -180 "${SERVER_LOG}" >&2 || true
  exit 1
}

reset_public_data() {
  if [[ "${RESET_PUBLIC_DB}" != "1" ]]; then
    cat >&2 <<'MSG'
This e2e verifies the empty-users bootstrap path and must reset public app data.
Re-run with RESET_PUBLIC_DB=1 against an isolated/local database.
MSG
    exit 1
  fi
  psql_query "TRUNCATE TABLE
    public.audit_logs,
    public.auth_login_attempts,
    public.auth_sessions,
    public.user_password_credentials,
    public.user_identities,
    public.platform_admins,
    public.api_key_tenant_grants,
    public.api_keys,
    public.tenant_invitations,
    public.tenant_lifecycle_jobs,
    public.tenant_usage_reservations,
    public.tenant_usage_counters,
    public.tenant_quotas,
    public.subscriptions,
    public.tenant_memberships,
    public.tenants,
    public.users
  RESTART IDENTITY CASCADE;" >/dev/null
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

register_user() {
  local tag="$1" jar="$2" body status
  body="$(bodyfile)"
  status="$(http_request POST /api/v1/auth/register "{\"email\":\"${tag}@example.test\",\"password\":\"FirstAdminPassword123!\",\"display_name\":\"${tag}\"}" "${body}" -c "${jar}")"
  assert_status "REGISTER_${tag}" 200 "${status}" "${body}"
  json_value "${body}" 'j["data"]["user"]["id"]'
}

log "[E2E] start server"
start_server

log "[E2E] serial bootstrap scenario"
reset_public_data
first_cookie="${LOG_DIR}/first.cookies"
second_cookie="${LOG_DIR}/second.cookies"
first_id="$(register_user first-admin-bootstrap "${first_cookie}")"
admin_count="$(psql_query "SELECT count(*) FROM public.platform_admins;")"
[[ "${admin_count}" == "1" ]] || { log "[FAIL] expected one platform admin, got ${admin_count}"; exit 1; }
role="$(psql_query "SELECT role FROM public.platform_admins WHERE user_id='${first_id}' AND status='active';")"
[[ "${role}" == "super_admin" ]] || { log "[FAIL] expected first user super_admin, got role=${role}"; exit 1; }
log "[PASS] first registered user bootstrapped as super_admin"

b="$(bodyfile)"; s="$(http_request GET /api/v1/me/access "" "${b}" -b "${first_cookie}")"; assert_status FIRST_ACCESS 200 "${s}" "${b}"
assert_json_equals FIRST_PLATFORM_ROLE "${b}" 'j["data"]["platform_admin"]["role"]' super_admin
assert_json_equals FIRST_HAS_ADMIN_MANAGE "${b}" '"platform:admin:manage" in j["data"]["platform_admin"]["permissions"]' true
b="$(bodyfile)"; s="$(http_request GET /api/v1/admin/session "" "${b}" -b "${first_cookie}")"; assert_status FIRST_ADMIN_SESSION 200 "${s}" "${b}"

second_id="$(register_user second-ordinary-bootstrap "${second_cookie}")"
admin_count="$(psql_query "SELECT count(*) FROM public.platform_admins;")"
[[ "${admin_count}" == "1" ]] || { log "[FAIL] second registration changed platform admin count to ${admin_count}"; exit 1; }
second_admin="$(psql_query "SELECT count(*) FROM public.platform_admins WHERE user_id='${second_id}';")"
[[ "${second_admin}" == "0" ]] || { log "[FAIL] second user unexpectedly became platform admin"; exit 1; }
log "[PASS] non-empty users table keeps later registrations ordinary"

log "[E2E] concurrent empty-users bootstrap scenario"
reset_public_data
for i in $(seq 1 10); do
  body="${LOG_DIR}/parallel-${i}.json"
  status_file="${LOG_DIR}/parallel-${i}.status"
  cookie_file="${LOG_DIR}/parallel-${i}.cookies"
  http_request POST /api/v1/auth/register "{\"email\":\"parallel-${i}@example.test\",\"password\":\"FirstAdminPassword123!\",\"display_name\":\"parallel ${i}\"}" "${body}" -c "${cookie_file}" >"${status_file}" &
done
wait
for i in $(seq 1 10); do
  status="$(cat "${LOG_DIR}/parallel-${i}.status")"
  assert_status "PARALLEL_REGISTER_${i}" 200 "${status}" "${LOG_DIR}/parallel-${i}.json"
done
user_count="$(psql_query "SELECT count(*) FROM public.users;")"
admin_count="$(psql_query "SELECT count(*) FROM public.platform_admins WHERE role='super_admin' AND status='active';")"
[[ "${user_count}" == "10" ]] || { log "[FAIL] expected 10 users after parallel registration, got ${user_count}"; exit 1; }
[[ "${admin_count}" == "1" ]] || { log "[FAIL] expected one active super_admin after parallel registration, got ${admin_count}"; exit 1; }
log "[PASS] parallel first registrations produce exactly one super_admin"

log "[E2E] first-user platform admin bootstrap scenarios passed"
