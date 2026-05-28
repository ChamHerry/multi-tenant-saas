#!/usr/bin/env bash
set -euo pipefail

PGHOST="${PGHOST:-127.0.0.1}"
PGPORT="${PGPORT:-55432}"
PGUSER="${PGUSER:-saas_template}"
PGPASSWORD="${PGPASSWORD:-secret}"
PGDATABASE="${PGDATABASE:-saas_template}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing dependency: $1" >&2; exit 1; }
}
need go

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

stamp="$(date +%Y%m%d%H%M%S)-$$"
email="mt-${stamp}@example.test"
slug="mt-${stamp}"
name="MultiTenant ${stamp}"

echo "[multitenancy] automigrate + user-upsert"
go run . user-upsert --provider password --auth-id "${email}" --email "${email}" --name "MT ${stamp}" --email-verified
user_id="$(psql_query "SELECT id FROM public.users WHERE email='${email}' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 1")"
if [[ -z "${user_id}" ]]; then
  echo "user not created" >&2
  exit 1
fi

echo "[multitenancy] tenant-create"
go run . tenant-create --name "${name}" --slug "${slug}" --owner-user "${user_id}"
tenant_id="$(psql_query "SELECT id FROM public.tenants WHERE slug='${slug}' AND deleted_at IS NULL LIMIT 1")"
if [[ -z "${tenant_id}" ]]; then
  echo "tenant not created" >&2
  exit 1
fi

echo "[multitenancy] list/resolve/context"
go run . tenant-member-list --tenant "${tenant_id}" >/dev/null
go run . user-tenant-list --user "${user_id}" >/dev/null
go run . tenant-context-resolve --user "${user_id}" --tenant "${tenant_id}" >/dev/null

read -r version dirty < <(psql_query "SELECT version, dirty FROM public.schema_migrations" | tr '|' ' ')
if [[ "${dirty}" != "f" ]]; then
  echo "public schema_migrations dirty=${dirty} version=${version}" >&2
  exit 1
fi

invalid_constraints="$(psql_query "SELECT conname FROM pg_constraint WHERE conname IN ('api_keys_membership_fk') AND NOT convalidated")"
if [[ -n "${invalid_constraints}" ]]; then
  echo "constraints not validated: ${invalid_constraints}" >&2
  exit 1
fi

echo "[multitenancy] ok user=${user_id} tenant=${tenant_id} version=${version}"
