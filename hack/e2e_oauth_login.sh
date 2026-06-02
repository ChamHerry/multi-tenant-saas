#!/usr/bin/env bash
# e2e_oauth_login.sh — Docker-backed end-to-end test for GitHub/Google OAuth login.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

APP_PORT="${APP_PORT:-8000}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${APP_PORT}}"
E2E_DOCKER_ACTION="${E2E_DOCKER_ACTION:-restart}"
READY_TIMEOUT="${READY_TIMEOUT:-120}"
READY_INTERVAL="${READY_INTERVAL:-2}"
FAKE_OAUTH_PORT="${FAKE_OAUTH_PORT:-19090}"
FAKE_OAUTH_HOST="${FAKE_OAUTH_HOST:-host.docker.internal}"
FAKE_PUBLIC_BASE="${FAKE_PUBLIC_BASE:-http://${FAKE_OAUTH_HOST}:${FAKE_OAUTH_PORT}}"
FAKE_LOCAL_BASE="${FAKE_LOCAL_BASE:-http://127.0.0.1:${FAKE_OAUTH_PORT}}"
LOG_DIR="${E2E_LOG_DIR:-${TMPDIR:-/tmp}/multi-tenant-saas-e2e-oauth}"
SCENARIO_LOG="${LOG_DIR}/oauth-login.log"
FAKE_LOG="${LOG_DIR}/fake-oauth.log"
FAILURE_LOG="${LOG_DIR}/compose-failure.log"
mkdir -p "$LOG_DIR"
: >"$SCENARIO_LOG"
: >"$FAKE_LOG"
: >"$FAILURE_LOG"

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }; }
need docker
need curl
need python3
need bash

log() { echo "$*" | tee -a "$SCENARIO_LOG" >&2; }

compose() { docker compose -f docker-compose.yml "$@"; }
manage_docker() { APP_PORT="$APP_PORT" READY_TIMEOUT="$READY_TIMEOUT" READY_INTERVAL="$READY_INTERVAL" ./scripts/manage-docker.sh "$@"; }

fake_pid=""

print_failure_context() {
  {
    echo "--- docker compose ps ---"
    compose ps || true
    echo
    echo "--- docker compose logs (tail=160) ---"
    compose logs --tail=160 app postgres redis || true
    echo
    echo "--- fake oauth log ---"
    cat "$FAKE_LOG" || true
  } >"$FAILURE_LOG" 2>&1
  cat "$FAILURE_LOG" >&2 || true
}

cleanup() {
  local status=$?
  if [[ -n "$fake_pid" ]] && kill -0 "$fake_pid" >/dev/null 2>&1; then
    kill "$fake_pid" >/dev/null 2>&1 || true
    wait "$fake_pid" >/dev/null 2>&1 || true
  fi
  if (( status != 0 )); then
    log "[E2E] failure detected, dumping compose/fake-provider context"
    print_failure_context
  fi
}
trap cleanup EXIT

ensure_stack_ready() {
  case "$E2E_DOCKER_ACTION" in
    skip) log "[E2E] reuse existing docker stack" ;;
    start|up|restart) log "[E2E] manage-docker action: $E2E_DOCKER_ACTION"; manage_docker "$E2E_DOCKER_ACTION" ;;
    *) echo "unsupported E2E_DOCKER_ACTION=$E2E_DOCKER_ACTION" >&2; exit 2 ;;
  esac
  log "[E2E] wait for /readyz"
  manage_docker ready >/dev/null
}

start_fake_oauth() {
  if lsof -tiTCP:"$FAKE_OAUTH_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "fake OAuth port ${FAKE_OAUTH_PORT} is already in use" >&2
    exit 1
  fi
  python3 -u - "$FAKE_OAUTH_PORT" >"$FAKE_LOG" 2>&1 <<'PY' &
import json
import sys
import threading
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, urlencode, urlparse, urlunparse

PORT = int(sys.argv[1])
CODES = {}
LOCK = threading.Lock()

def first(qs, name, default=""):
    values = qs.get(name)
    if not values:
        return default
    return values[0]

class Handler(BaseHTTPRequestHandler):
    def log_message(self, fmt, *args):
        sys.stdout.write(fmt % args + "\n")
        sys.stdout.flush()

    def send_json(self, status, payload):
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        parsed = urlparse(self.path)
        qs = parse_qs(parsed.query)
        if parsed.path == "/health":
            self.send_json(200, {"ok": True})
            return
        if parsed.path == "/authorize":
            redirect_uri = first(qs, "redirect_uri")
            state = first(qs, "state")
            if not redirect_uri or not state:
                self.send_response(400)
                self.end_headers()
                self.wfile.write(b"missing redirect_uri/state")
                return
            profile = {
                "provider": first(qs, "provider", "github"),
                "id": first(qs, "id", "fake-id"),
                "email": first(qs, "email", "fake@example.test"),
                "name": first(qs, "name", "Fake OAuth User"),
                "avatar": first(qs, "avatar", "https://example.test/avatar.png"),
                "verified": first(qs, "verified", "true").lower() == "true",
            }
            with LOCK:
                code = f"code-{len(CODES)+1}-{int(time.time()*1000)}"
                CODES[code] = profile
            dest = urlparse(redirect_uri)
            params = parse_qs(dest.query)
            params["code"] = [code]
            params["state"] = [state]
            location = urlunparse(dest._replace(query=urlencode(params, doseq=True)))
            self.send_response(302)
            self.send_header("Location", location)
            self.end_headers()
            return
        if parsed.path == "/github/user":
            profile = self.profile_for_request()
            try:
                github_id = int(profile["id"])
            except ValueError:
                github_id = abs(hash(profile["id"])) % 100000000
            self.send_json(200, {
                "id": github_id,
                "login": profile["name"].replace(" ", "-").lower(),
                "name": profile["name"],
                "avatar_url": profile["avatar"],
            })
            return
        if parsed.path == "/github/emails":
            profile = self.profile_for_request()
            self.send_json(200, [{
                "email": profile["email"],
                "primary": True,
                "verified": profile["verified"],
            }])
            return
        if parsed.path == "/google/userinfo":
            profile = self.profile_for_request()
            self.send_json(200, {
                "sub": profile["id"],
                "email": profile["email"],
                "email_verified": profile["verified"],
                "name": profile["name"],
                "picture": profile["avatar"],
            })
            return
        self.send_response(404)
        self.end_headers()
        self.wfile.write(b"not found")

    def do_POST(self):
        parsed = urlparse(self.path)
        if parsed.path != "/token":
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b"not found")
            return
        length = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(length).decode("utf-8")
        form = parse_qs(body)
        code = first(form, "code")
        with LOCK:
            exists = code in CODES
        if not exists:
            self.send_response(400)
            self.end_headers()
            self.wfile.write(b"invalid code")
            return
        self.send_json(200, {
            "access_token": f"token-{code}",
            "token_type": "Bearer",
            "expires_in": 3600,
        })

    def profile_for_request(self):
        auth = self.headers.get("Authorization", "")
        code = ""
        if auth.startswith("Bearer token-"):
            code = auth[len("Bearer token-"):]
        with LOCK:
            return CODES.get(code, {
                "provider": "github",
                "id": "fallback",
                "email": "fallback@example.test",
                "name": "Fallback User",
                "avatar": "https://example.test/avatar.png",
                "verified": True,
            })

server = ThreadingHTTPServer(("0.0.0.0", PORT), Handler)
print(f"fake oauth listening on :{PORT}", flush=True)
server.serve_forever()
PY
  fake_pid=$!
  for _ in $(seq 1 80); do
    if curl -fsS "${FAKE_LOCAL_BASE}/health" >/dev/null 2>&1; then
      log "[E2E] fake OAuth provider ready: local=${FAKE_LOCAL_BASE} app-visible=${FAKE_PUBLIC_BASE}"
      return
    fi
    if ! kill -0 "$fake_pid" >/dev/null 2>&1; then
      echo "fake OAuth provider exited during startup" >&2
      cat "$FAKE_LOG" >&2 || true
      exit 1
    fi
    sleep 0.25
  done
  echo "fake OAuth provider did not become ready" >&2
  cat "$FAKE_LOG" >&2 || true
  exit 1
}

bodyfile() { mktemp "${LOG_DIR}/body.XXXXXX"; }
headersfile() { mktemp "${LOG_DIR}/headers.XXXXXX"; }

http_request() {
  local method="$1" url="$2" body="$3" outfile="$4" headers="$5"
  shift 5
  local args=(-sS -D "$headers" -o "$outfile" -w "%{http_code}" -X "$method")
  if [[ -n "$body" ]]; then args+=(-H "Content-Type: application/json" -d "$body"); fi
  args+=("$@" "$url")
  curl "${args[@]}"
}

api_request() {
  local method="$1" path="$2" body="$3" outfile="$4" headers="$5"
  shift 5
  http_request "$method" "${BASE_URL}${path}" "$body" "$outfile" "$headers" "$@"
}

location_header() {
  awk 'BEGIN{IGNORECASE=1} /^Location:/ {sub(/\r$/,""); sub(/^Location:[[:space:]]*/,""); print; exit}' "$1"
}

localize_fake_url() {
  python3 - "$1" "$FAKE_OAUTH_HOST" "127.0.0.1" "$FAKE_OAUTH_PORT" <<'PY'
import sys
from urllib.parse import urlparse, urlunparse
url, fake_host, local_host, port = sys.argv[1:5]
p = urlparse(url)
if p.hostname == fake_host:
    netloc = local_host
    if p.port:
        netloc += f":{p.port}"
    elif port:
        netloc += f":{port}"
    print(urlunparse(p._replace(netloc=netloc)))
else:
    print(url)
PY
}

urlencode() {
  python3 - "$1" <<'PY'
import sys
from urllib.parse import quote
print(quote(sys.argv[1], safe=""))
PY
}

json_value() {
  local file="$1" expr="$2"
  python3 - "$file" "$expr" <<'PY'
import json, sys
with open(sys.argv[1], 'r', encoding='utf-8') as f:
    j = json.load(f)
value = eval(sys.argv[2], {"__builtins__": {}}, {"j": j, "len": len})
if isinstance(value, bool):
    print("true" if value else "false")
elif value is None:
    print("")
else:
    print(value)
PY
}

sql_escape() {
  python3 - "$1" <<'PY'
import sys
print(sys.argv[1].replace("'", "''"))
PY
}

psql_query() {
  local sql="$1"
  compose exec -T postgres psql -U saas_template -d saas_template -Atc "$sql"
}

redis_del_config() {
  local key="$1"
  compose exec -T redis redis-cli DEL "config:${key}" >/dev/null
}

set_config() {
  local key="$1" value="$2" value_type="$3"
  local e_key e_value e_type
  e_key="$(sql_escape "$key")"
  e_value="$(sql_escape "$value")"
  e_type="$(sql_escape "$value_type")"
  psql_query "INSERT INTO public.system_config (key, value, value_type, description, is_encrypted, updated_at)
VALUES ('${e_key}', '${e_value}', '${e_type}', 'oauth e2e override', false, now())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    value_type = EXCLUDED.value_type,
    description = EXCLUDED.description,
    is_encrypted = false,
    updated_at = now();" >/dev/null
  redis_del_config "$key"
}

set_oauth_provider() {
  local provider="$1" enabled="$2" auth_url="$3" token_url="$4"
  set_config "web.baseUrl" "$BASE_URL" "string"
  set_config "oauth.${provider}.enabled" "$enabled" "bool"
  set_config "oauth.${provider}.clientId" "e2e-client-${provider}" "string"
  # Use string value type in E2E because direct SQL cannot call Config.Set encryption.
  set_config "oauth.${provider}.clientSecret" "e2e-secret-${provider}" "string"
  set_config "oauth.${provider}.redirectUri" "" "string"
  set_config "oauth.${provider}.authUrl" "$auth_url" "string"
  set_config "oauth.${provider}.tokenUrl" "$token_url" "string"
  if [[ "$provider" == "github" ]]; then
    set_config "oauth.github.scopes" "user:email" "string"
    set_config "oauth.github.userUrl" "${FAKE_PUBLIC_BASE}/github/user" "string"
    set_config "oauth.github.emailsUrl" "${FAKE_PUBLIC_BASE}/github/emails" "string"
  else
    set_config "oauth.google.scopes" "openid email profile" "string"
    set_config "oauth.google.userInfoUrl" "${FAKE_PUBLIC_BASE}/google/userinfo" "string"
  fi
}

assert_status() {
  local id="$1" expected="$2" actual="$3" bodyfile="$4"
  if [[ "$actual" != "$expected" ]]; then
    log "[FAIL] ${id}: expected HTTP ${expected}, got ${actual}; body=$(cat "$bodyfile")"
    exit 1
  fi
  log "[PASS] ${id}: HTTP ${actual}"
}

assert_contains() {
  local id="$1" haystack="$2" needle="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    log "[FAIL] ${id}: expected ${haystack} to contain ${needle}"
    exit 1
  fi
  log "[PASS] ${id}: contains ${needle}"
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

csrf_from_cookie_jar() {
  local jar="$1"
  awk '$0 !~ /^#/ && $6 == "saas_template_csrf" {print $7}' "$jar" | tail -1
}

totp_code() {
  local secret="$1"
  python3 - "$secret" <<'PY'
import base64, hashlib, hmac, struct, sys, time
secret = sys.argv[1].strip().replace(' ', '').upper()
secret += '=' * ((8 - len(secret) % 8) % 8)
key = base64.b32decode(secret)
counter = int(time.time()) // 30
msg = struct.pack('>Q', counter)
digest = hmac.new(key, msg, hashlib.sha1).digest()
offset = digest[-1] & 0x0F
code = (struct.unpack('>I', digest[offset:offset+4])[0] & 0x7fffffff) % 1000000
print(f'{code:06d}')
PY
}

register_user() {
  local email="$1" password="$2" name="$3" cookie="$4" body headers status
  body="$(bodyfile)"; headers="$(headersfile)"
  status="$(api_request POST /api/v1/auth/register "{\"email\":\"${email}\",\"password\":\"${password}\",\"display_name\":\"${name}\"}" "$body" "$headers" -c "$cookie")"
  assert_status "REGISTER_${name// /_}" 200 "$status" "$body"
}

enable_totp_for_user() {
  local email="$1" password="$2" cookie="$3"
  local body headers status csrf secret code
  csrf="$(csrf_from_cookie_jar "$cookie")"
  [[ -n "$csrf" ]] || { log "[FAIL] missing CSRF for ${email}"; exit 1; }
  body="$(bodyfile)"; headers="$(headersfile)"
  status="$(api_request POST /api/v1/me/totp/setup "{\"password\":\"${password}\"}" "$body" "$headers" -b "$cookie" -c "$cookie" -H "X-CSRF-Token: ${csrf}")"
  assert_status "TOTP_SETUP_${email}" 200 "$status" "$body"
  secret="$(json_value "$body" 'j["data"]["secret"]')"
  code="$(totp_code "$secret")"
  body="$(bodyfile)"; headers="$(headersfile)"
  status="$(api_request POST /api/v1/me/totp/enable "{\"code\":\"${code}\"}" "$body" "$headers" -b "$cookie" -c "$cookie" -H "X-CSRF-Token: ${csrf}")"
  assert_status "TOTP_ENABLE_${email}" 200 "$status" "$body"
  printf '%s\n' "$secret"
}

fake_auth_url() {
  local provider="$1" id="$2" email="$3" name="$4"
  printf '%s/authorize?provider=%s&id=%s&email=%s&name=%s&verified=true' \
    "$FAKE_PUBLIC_BASE" \
    "$(urlencode "$provider")" \
    "$(urlencode "$id")" \
    "$(urlencode "$email")" \
    "$(urlencode "$name")"
}

run_oauth_flow() {
  local provider="$1" redirect_path="$2" cookie="$3"
  local body headers status start_loc auth_loc auth_local callback_status
  body="$(bodyfile)"; headers="$(headersfile)"
  status="$(api_request GET "/api/v1/auth/oauth/${provider}?redirect=$(urlencode "$redirect_path")" "" "$body" "$headers" -b "$cookie" -c "$cookie")"
  assert_status "OAUTH_START_${provider}" 302 "$status" "$body"
  start_loc="$(location_header "$headers")"
  [[ -n "$start_loc" ]] || { log "[FAIL] OAUTH_START_${provider}: missing provider Location"; exit 1; }
  auth_local="$(localize_fake_url "$start_loc")"
  body="$(bodyfile)"; headers="$(headersfile)"
  status="$(http_request GET "$auth_local" "" "$body" "$headers")"
  assert_status "FAKE_AUTHORIZE_${provider}" 302 "$status" "$body"
  auth_loc="$(location_header "$headers")"
  [[ -n "$auth_loc" ]] || { log "[FAIL] FAKE_AUTHORIZE_${provider}: missing callback Location"; exit 1; }
  body="$(bodyfile)"; headers="$(headersfile)"
  callback_status="$(http_request GET "$auth_loc" "" "$body" "$headers" -b "$cookie" -c "$cookie")"
  printf '%s\t%s\t%s\n' "$callback_status" "$(location_header "$headers")" "$auth_loc"
}

stamp="$(date +%Y%m%d%H%M%S)-$$"
password="OAuthPassword123!"

ensure_stack_ready
start_fake_oauth

log "[E2E] OAuth provider disabled is rejected"
set_oauth_provider "github" "false" "" ""
set_oauth_provider "google" "false" "" ""
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/oauth/providers "" "$b" "$h")"
assert_status PROVIDERS_DISABLED_LIST 200 "$s" "$b"
assert_json_equals PROVIDERS_COUNT "$b" 'len(j["data"]["providers"])' 2
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/oauth/github?redirect=/tenants "" "$b" "$h")"
assert_status DISABLED_GITHUB_START 302 "$s" "$b"
assert_contains DISABLED_GITHUB_LOCATION "$(location_header "$h")" "oauth_error=oauth_provider_disabled"

log "[E2E] GitHub new-user OAuth happy path and state replay"
github_email="oauth-github-${stamp}@example.test"
set_oauth_provider "github" "true" "$(fake_auth_url github "1001001" "$github_email" "OAuth GitHub E2E")" "${FAKE_PUBLIC_BASE}/token"
github_cookie="${LOG_DIR}/github-${stamp}.cookies"
IFS=$'\t' read -r status final_loc callback_url < <(run_oauth_flow github /tenants "$github_cookie")
assert_status GITHUB_CALLBACK 302 "$status" "$(bodyfile)"
assert_contains GITHUB_FINAL_REDIRECT "$final_loc" "/tenants"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/session "" "$b" "$h" -b "$github_cookie")"
assert_status GITHUB_SESSION 200 "$s" "$b"
assert_json_equals GITHUB_SESSION_EMAIL "$b" 'j["data"]["user"]["email"]' "$github_email"
identity_count="$(psql_query "SELECT count(*) FROM public.user_identities WHERE provider='github' AND email='${github_email}';")"
[[ "$identity_count" == "1" ]] || { log "[FAIL] expected one GitHub identity, got ${identity_count}"; exit 1; }
log "[PASS] GITHUB_IDENTITY_INSERTED"
b="$(bodyfile)"; h="$(headersfile)"
s="$(http_request GET "$callback_url" "" "$b" "$h" -b "$github_cookie" -c "$github_cookie")"
assert_status GITHUB_STATE_REPLAY 302 "$s" "$b"
assert_contains GITHUB_STATE_REPLAY_LOCATION "$(location_header "$h")" "oauth_error=oauth_state_invalid"

log "[E2E] Google existing password user without TOTP can sign in"
google_email="oauth-google-existing-${stamp}@example.test"
google_cookie="${LOG_DIR}/google-existing-${stamp}.cookies"
register_user "$google_email" "$password" "OAuth Google Existing" "$google_cookie"
set_oauth_provider "google" "true" "$(fake_auth_url google "google-existing-${stamp}" "$google_email" "OAuth Google Existing")" "${FAKE_PUBLIC_BASE}/token"
google_login_cookie="${LOG_DIR}/google-login-${stamp}.cookies"
IFS=$'\t' read -r status final_loc _ < <(run_oauth_flow google /security "$google_login_cookie")
assert_status GOOGLE_CALLBACK 302 "$status" "$(bodyfile)"
assert_contains GOOGLE_FINAL_REDIRECT "$final_loc" "/security"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/session "" "$b" "$h" -b "$google_login_cookie")"
assert_status GOOGLE_SESSION 200 "$s" "$b"
assert_json_equals GOOGLE_SESSION_EMAIL "$b" 'j["data"]["user"]["email"]' "$google_email"
google_identity_count="$(psql_query "SELECT count(*) FROM public.user_identities WHERE provider='google' AND email='${google_email}';")"
[[ "$google_identity_count" == "1" ]] || { log "[FAIL] expected one Google identity, got ${google_identity_count}"; exit 1; }
log "[PASS] GOOGLE_IDENTITY_LINKED"

log "[E2E] Google OAuth for TOTP-enabled user requires challenge and no pre-verify session"
totp_email="oauth-google-totp-${stamp}@example.test"
totp_setup_cookie="${LOG_DIR}/google-totp-setup-${stamp}.cookies"
register_user "$totp_email" "$password" "OAuth Google TOTP" "$totp_setup_cookie"
totp_secret="$(enable_totp_for_user "$totp_email" "$password" "$totp_setup_cookie")"
set_oauth_provider "google" "true" "$(fake_auth_url google "google-totp-${stamp}" "$totp_email" "OAuth Google TOTP")" "${FAKE_PUBLIC_BASE}/token"
totp_login_cookie="${LOG_DIR}/google-totp-login-${stamp}.cookies"
IFS=$'\t' read -r status verify_loc _ < <(run_oauth_flow google /security "$totp_login_cookie")
assert_status GOOGLE_TOTP_CALLBACK 302 "$status" "$(bodyfile)"
assert_contains GOOGLE_TOTP_VERIFY_REDIRECT "$verify_loc" "/verify-totp"
assert_contains GOOGLE_TOTP_CHALLENGE "$verify_loc" "challenge_token="
challenge_token="$(python3 - "$verify_loc" <<'PY'
import sys
from urllib.parse import parse_qs, urlparse
print(parse_qs(urlparse(sys.argv[1]).query).get("challenge_token", [""])[0])
PY
)"
[[ -n "$challenge_token" ]] || { log "[FAIL] missing challenge token"; exit 1; }
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/session "" "$b" "$h" -b "$totp_login_cookie")"
assert_status GOOGLE_TOTP_PRE_SESSION_REJECTED 401 "$s" "$b"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request POST /api/v1/auth/verify-totp "{\"challenge_token\":\"${challenge_token}\",\"code\":\"000000\"}" "$b" "$h" -b "$totp_login_cookie" -c "$totp_login_cookie")"
assert_status GOOGLE_TOTP_INVALID_REJECTED 403 "$s" "$b"
valid_code="$(totp_code "$totp_secret")"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request POST /api/v1/auth/verify-totp "{\"challenge_token\":\"${challenge_token}\",\"code\":\"${valid_code}\"}" "$b" "$h" -b "$totp_login_cookie" -c "$totp_login_cookie")"
assert_status GOOGLE_TOTP_VERIFY 200 "$s" "$b"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request GET /api/v1/auth/session "" "$b" "$h" -b "$totp_login_cookie")"
assert_status GOOGLE_TOTP_POST_SESSION 200 "$s" "$b"
assert_json_equals GOOGLE_TOTP_SESSION_EMAIL "$b" 'j["data"]["user"]["email"]' "$totp_email"
b="$(bodyfile)"; h="$(headersfile)"
s="$(api_request POST /api/v1/auth/verify-totp "{\"challenge_token\":\"${challenge_token}\",\"code\":\"${valid_code}\"}" "$b" "$h" -b "$totp_login_cookie" -c "$totp_login_cookie")"
assert_status GOOGLE_TOTP_REPLAY_REJECTED 403 "$s" "$b"

log "[E2E] OAuth login scenarios passed stamp=${stamp}"
