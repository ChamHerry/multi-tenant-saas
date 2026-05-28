# RepoMind PostgreSQL dev image

PostgreSQL 16 image with project-required extensions:

- Apache AGE (`postgresql-16-age`)
- pgvector (`postgresql-16-pgvector`)

Local container convention:

```bash
docker build -t repomind-postgres:16-age-vector manifest/docker/postgres
docker run -d --name repomind-pg \
  -e POSTGRES_DB=repomind \
  -e POSTGRES_USER=repomind \
  -e POSTGRES_PASSWORD=secret \
  -p 55432:5432 \
  -v repomind-pg-data:/var/lib/postgresql/data \
  repomind-postgres:16-age-vector
```
