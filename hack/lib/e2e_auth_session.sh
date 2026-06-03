#!/usr/bin/env bash
# Shared HTTP-session helpers for e2e scripts. Requires caller-defined:
#   ROOT_DIR, BASE_URL, LOG_DIR, psql_query, http_request, bodyfile, json_value,
#   assert_status, and optionally appctl/compose/log.

E2E_PASSWORD="${E2E_PASSWORD:-TestPassword12345!}"
E2E_ENCRYPTION_KEY="${E2E_ENCRYPTION_KEY:-docker-dev-32-byte-encryption-key!!}"

_e2e_log() {
  if declare -F log >/dev/null 2>&1; then
    log "$*"
  else
    echo "$*"
  fi
}

e2e_appctl() {
  if declare -F appctl >/dev/null 2>&1; then
    appctl "$@"
  else
    (cd "${ROOT_DIR}" && go run . "$@")
  fi
}

e2e_clear_config_cache() {
  if [[ "$#" -eq 0 ]]; then
    return
  fi
  if declare -F compose >/dev/null 2>&1; then
    local redis_service="${REDIS_SERVICE:-redis}"
    if compose exec -T "$redis_service" redis-cli DEL "$@" >/dev/null 2>&1; then
      return
    fi
  fi
  if command -v redis-cli >/dev/null 2>&1; then
    if redis-cli -h 127.0.0.1 -p 6379 DEL "$@" >/dev/null 2>&1; then
      return
    fi
  fi
  if command -v docker >/dev/null 2>&1; then
    for container in repomind-redis multi-tenant-saas-redis; do
      if docker ps --format '{{.Names}}' | grep -qx "${container}"; then
        docker exec "${container}" redis-cli DEL "$@" >/dev/null 2>&1 || true
        return
      fi
    done
  fi
}

e2e_configure_http_cookies() {
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('auth.session.cookie.secure','false','bool','HTTP e2e cookie override',false,now(),now()) ON CONFLICT(key) DO UPDATE SET value='false', value_type='bool', description='HTTP e2e cookie override', is_encrypted=false, updated_at=now();" >/dev/null
  e2e_clear_config_cache config:auth.session.cookie.secure
}

e2e_write_local_config() {
  local target="${E2E_LOCAL_CONFIG_FILE:-${LOG_DIR}/e2e-local-config.yaml}"
  python3 - "${ROOT_DIR}/manifest/config/config.yaml" "${target}" "${E2E_ENCRYPTION_KEY}" <<'PY'
from pathlib import Path
import re, sys
src, dst, key = sys.argv[1:4]
text = Path(src).read_text()
text = re.sub(r'^encryptionKey: .*$', f'encryptionKey: "{key}"', text, flags=re.M)
Path(dst).write_text(text)
PY
  export GF_GCFG_FILE="${target}"
}

e2e_encrypt_config_value() {
  local plaintext="$1"
  local helper="${LOG_DIR}/e2e-encrypt-config.go"
  cat >"${helper}" <<'GO'
package main

import (
  "context"
  "fmt"
  "os"

  "multi-tenant-saas/utility/crypto"
)

func main() {
  if len(os.Args) != 2 {
    panic("usage: e2e-encrypt-config <plaintext>")
  }
  if err := crypto.InitEncryption(context.Background()); err != nil {
    panic(err)
  }
  encrypted, err := crypto.Encrypt(os.Args[1])
  if err != nil {
    panic(err)
  }
  fmt.Println("E2E_ENCRYPTED:" + encrypted)
}
GO
  (cd "${ROOT_DIR}" && go run "${helper}" "${plaintext}") | awk -F'E2E_ENCRYPTED:' '/^E2E_ENCRYPTED:/{print $2; exit}'
}

e2e_configure_local_runtime() {
  local session_secret="${E2E_SESSION_SECRET:-e2e-local-session-secret-0123456789abcdef}"
  local api_key_secret="${E2E_API_KEY_SECRET:-e2e-local-api-key-secret-0123456789abcdef}"
  local web_base_url="${E2E_WEB_BASE_URL:-${BASE_URL:-http://127.0.0.1:8000}}"

  e2e_write_local_config

  local encrypted_session_secret encrypted_api_key_secret
  encrypted_session_secret="$(e2e_encrypt_config_value "${session_secret}")"
  encrypted_api_key_secret="$(e2e_encrypt_config_value "${api_key_secret}")"
  if [[ -z "${encrypted_session_secret}" || -z "${encrypted_api_key_secret}" ]]; then
    echo "failed to encrypt local e2e runtime secrets" >&2
    exit 1
  fi

  psql_query "INSERT INTO public.system_setup(id,status,version,initialized_at,metadata) VALUES(1,'initialized','e2e-local',now(),'{\"source\":\"e2e_local_runtime\"}'::jsonb) ON CONFLICT(id) DO UPDATE SET status='initialized', version='e2e-local', initialized_at=COALESCE(public.system_setup.initialized_at, now()), metadata=public.system_setup.metadata || '{\"source\":\"e2e_local_runtime\"}'::jsonb;" >/dev/null
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('auth.session.secret','${encrypted_session_secret}','secret','Local e2e runtime session secret',true,now(),now()) ON CONFLICT(key) DO UPDATE SET value='${encrypted_session_secret}', value_type='secret', description='Local e2e runtime session secret', is_encrypted=true, updated_at=now();" >/dev/null
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('auth.apiKey.secret','${encrypted_api_key_secret}','secret','Local e2e runtime API key secret',true,now(),now()) ON CONFLICT(key) DO UPDATE SET value='${encrypted_api_key_secret}', value_type='secret', description='Local e2e runtime API key secret', is_encrypted=true, updated_at=now();" >/dev/null
  psql_query "INSERT INTO public.system_config(key,value,value_type,description,is_encrypted,created_at,updated_at) VALUES('web.baseUrl','${web_base_url}','string','Local e2e web base URL',false,now(),now()) ON CONFLICT(key) DO UPDATE SET value='${web_base_url}', value_type='string', description='Local e2e web base URL', is_encrypted=false, updated_at=now();" >/dev/null
  e2e_configure_http_cookies
  e2e_clear_config_cache config:auth.session.secret config:auth.apiKey.secret config:web.baseUrl config:auth.session.cookie.secure
}

e2e_ensure_setup_completed() {
  e2e_configure_http_cookies
  local b s requires setup_email setup_password payload
  b="$(bodyfile)"; s="$(http_request GET /api/v1/setup/state "" "$b")"; assert_status SETUP_STATE_CHECK 200 "$s" "$b"
  requires="$(json_value "$b" 'j["data"]["requires_setup"]')"
  if [[ "$requires" != "true" ]]; then
    _e2e_log "[E2E] setup already initialized"
    return
  fi

  _e2e_log "[E2E] complete setup bootstrap"
  setup_email="e2e-bootstrap-${stamp:-$(date +%s)}@example.test"
  setup_password="SetupPassword12345!"
  payload="$(printf '{"admin":{"email":"%s","password":"%s","display_name":"E2E Bootstrap"},"runtime":{"web_base_url":"%s","generate_session_secret":true,"generate_api_key_secret":true}}' "$setup_email" "$setup_password" "$BASE_URL")"
  b="$(bodyfile)"; s="$(http_request POST /api/v1/setup/complete "$payload" "$b")"; assert_status SETUP_BOOTSTRAP_COMPLETE 200 "$s" "$b"
  e2e_configure_http_cookies
}

e2e_create_password_user() {
  local tag="$1"
  local email="${tag}@example.test"
  e2e_appctl user-password-create --email "$email" --password "$E2E_PASSWORD" --name "$tag" --email-verified true >/dev/null
  psql_query "SELECT id FROM public.users WHERE email='${email}' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1"
}

e2e_login_user_email() {
  local email="$1" jar="$2"
  local b s
  b="$(bodyfile)"
  s="$(http_request POST /api/v1/auth/login "$(printf '{"email":"%s","password":"%s"}' "$email" "$E2E_PASSWORD")" "$b" -c "$jar")"
  assert_status "LOGIN_${email}" 200 "$s" "$b"
}

e2e_login_user_tag() {
  local tag="$1" jar="$2"
  e2e_login_user_email "${tag}@example.test" "$jar"
}

e2e_csrf_from_cookie_jar() {
  local jar="$1"
  awk '$0 !~ /^#/ && $6 ~ /csrf/ {print $7}' "$jar" | tail -1
}
