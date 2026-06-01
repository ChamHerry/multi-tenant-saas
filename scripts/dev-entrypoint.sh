#!/bin/sh
set -e

# ── Build SPA if not already built ──
if [ ! -f /app/web/dist/index.html ]; then
  echo "[dev] Building SPA (one-time)..."
  cd /app/web && npm install --prefer-offline && npm run build
else
  echo "[dev] SPA already built, skip build."
  echo "[dev] To rebuild SPA: docker compose -f docker-compose.yml -f docker-compose.dev.yml exec dev npm run build --prefix /app/web"
fi

# ── Start Go backend (serves API + SPA on single port) ──
echo "[dev] Starting Go backend (API + SPA on :8000)..."
exec go run .
