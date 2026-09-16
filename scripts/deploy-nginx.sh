#!/usr/bin/env bash
# ============================================================
# Deploy konfigurasi nginx findit-api ke folder nginx cashmate.
#
# Karena HTTPS findit-api dilayani oleh container cashmate-nginx,
# file source-of-truth ini (milik repo findit-api) harus di-copy
# ke path yang ter-mount ke cashmate-nginx.
# ============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC="$SCRIPT_DIR/../nginx/findit-ssl.conf"
DST="$HOME/api/cashmate-api/nginx/conf.d/findit-ssl.conf"

if [ ! -f "$SRC" ]; then
  echo "ERROR: source config tidak ditemukan: $SRC" >&2
  exit 1
fi

echo "==> Copy $SRC -> $DST"
cp "$SRC" "$DST"

echo "==> Validasi nginx -t ..."
if ! docker exec cashmate-nginx nginx -t; then
  echo "ERROR: nginx -t GAGAL, konfigurasi TIDAK di-reload." >&2
  exit 1
fi

echo "==> Reload cashmate-nginx ..."
docker exec cashmate-nginx nginx -s reload

echo "Selesai! findit-api: https://139-190-96-203.sslip.io/findit/api/health"