# SaaS template PostgreSQL dev image

Plain PostgreSQL 16 image for local development. The multi-tenant SaaS template
baseline does not require project-specific PostgreSQL extensions such as Apache
AGE, pgvector, pg_trgm, or pgcrypto.

Local container convention:

```bash
docker build -t saas-template-postgres:16 manifest/docker/postgres
docker run -d --name saas-template-pg \
  -e POSTGRES_DB=saas_template \
  -e POSTGRES_USER=saas_template \
  -e POSTGRES_PASSWORD=secret \
  -p 55432:5432 \
  -v saas-template-pg-data:/var/lib/postgresql/data \
  saas-template-postgres:16
```
