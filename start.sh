#!/bin/bash
echo "Starting VST Monster Ecosystem..."

# Function to clean up background processes
cleanup() {
    echo "Stopping background processes..."
    if [ ! -z "$REGISTRY_PID" ]; then
        kill $REGISTRY_PID 2>/dev/null
    fi
    if [ ! -z "$CLIENT_PID" ]; then
        kill $CLIENT_PID 2>/dev/null
    fi
}

# Trap SIGINT (Ctrl+C) and SIGTERM
trap cleanup SIGINT SIGTERM

# Start Registry API
cd registry
npm start &
REGISTRY_PID=$!
cd ..

# Start Client UI
cd client/src-tauri
cargo run &
CLIENT_PID=$!
cd ../..

echo "Ecosystem started. Press Ctrl+C to stop."
wait
