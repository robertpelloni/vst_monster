# DEPLOY: VST Monster

## Environment Requirements
- **Node.js**: Node 20+ (Node 22 LTS verified).
- **Go**: Go 1.21+ (Go 1.25 verified).
- **Rust / Tauri v2**: Cargo 1.80+ and Linux GTK/WebKit system libraries:
  ```bash
  sudo apt-get install -y libglib2.0-dev libwebkit2gtk-4.1-dev build-essential curl wget file libxdo-dev libssl-dev libayatana-appindicator3-dev librsvg2-dev
  ```
- **Databases**: PostgreSQL 15+ and Redis.

## Local Startup
Launch the full ecosystem using the unified startup orchestrator:
```bash
./start.sh
```
This automatically initializes local PostgreSQL/Redis services, loads `registry/schema.sql`, launches the Registry API on port 3000, and boots the Tauri desktop application.

## Production Builds
- **Registry API**:
  ```bash
  cd registry
  npm install
  npx tsc
  npm start
  ```
- **Tauri Desktop Client**:
  ```bash
  cd client
  npm install
  npm run build
  cd src-tauri
  cargo build --release
  ```
