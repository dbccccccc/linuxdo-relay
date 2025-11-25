#!/bin/sh
set -euo pipefail

APP_BIN="/app/linuxdo-relay"
RUNTIME_PATH="${APP_RUNTIME_CONFIG_PATH:-/app/runtimeconfig/config.json}"
RUNTIME_DIR="$(dirname "$RUNTIME_PATH")"

if [ ! -d "$RUNTIME_DIR" ]; then
    mkdir -p "$RUNTIME_DIR"
fi

chown -R app:app "$RUNTIME_DIR"

exec su-exec app "$APP_BIN" "$@"
