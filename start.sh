#!/bin/bash
set -e

echo "=========================================="
echo "    VST Monster Full Ecosystem Startup    "
echo "=========================================="

export PGHOST="${PGHOST:-localhost}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-postgres}"
export PGPASSWORD="${PGPASSWORD:-postgres}"
export PGDATABASE="${PGDATABASE:-vst_monster}"

cleanup() {
    echo ""
    echo "Stopping VST Monster background processes..."
    if [ ! -z "$REGISTRY_PID" ]; then
        kill "$REGISTRY_PID" 2>/dev/null || true
    fi
    if [ ! -z "$CLIENT_PID" ]; then
        kill "$CLIENT_PID" 2>/dev/null || true
    fi
    echo "Processes terminated cleanly."
}

trap cleanup SIGINT SIGTERM EXIT

# Check services
if command -v service >/dev/null 2>&1; then
    sudo service postgresql status >/dev/null 2>&1 || sudo service postgresql start >/dev/null 2>&1 || true
    sudo service redis-server status >/dev/null 2>&1 || sudo service redis-server start >/dev/null 2>&1 || true
fi

# Initialize database schema if database exists
if command -v psql >/dev/null 2>&1; then
    PGPASSWORD="$PGPASSWORD" psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -f registry/schema.sql 2>/dev/null || true
fi

# 1. Start Registry API
echo "[+] Starting Registry API Server (port 3000)..."
cd registry
npm start &
REGISTRY_PID=$!
cd ..

# Wait briefly for Registry API to initialize
sleep 2

# 2. Start Desktop Client
echo "[+] Starting Tauri Desktop Client..."
cd client/src-tauri
cargo run &
CLIENT_PID=$!
cd ../..

echo "[✔] VST Monster Ecosystem running! Press Ctrl+C to terminate."
wait "$CLIENT_PID" "$REGISTRY_PID" 2>/dev/null || true
