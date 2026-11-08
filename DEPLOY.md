# DEPLOY: VST Monster Deployment & Operations Guide

This guide details comprehensive, production-ready environment setup, deployment architectures, build optimizations, and operational instructions for both the **Registry API Backend** and the **Tauri Desktop Client**.

---

## 1. Environment & Prerequisite Matrix

| Component | Minimum Version | Verified Production Version | Purpose |
| :--- | :--- | :--- | :--- |
| **Node.js** | `>= 20.0.0` | `22.x LTS` | Registry Express API & SvelteKit Client Toolchain |
| **Rust / Cargo** | `>= 1.80.0` | `1.85.0 Stable` | Native Tauri Desktop Installer & Scanner Runtime |
| **Go** | `>= 1.21.0` | `1.25 Stable` | High-throughput Scraper Engine & Worker Pipeline |
| **PostgreSQL** | `>= 15.0` | `16.x` | Primary Relational Datastore (Plugins & Releases) |
| **Redis** | `>= 7.0` | `7.2.x` | Cache Layer & Distributed Job Queue Coordination |

### Host System Dependencies (Linux Production / CI Runners)
For compiling the native Tauri v2 desktop application on Ubuntu/Debian:
```bash
sudo apt-get update && sudo apt-get install -y \
  build-essential \
  curl \
  wget \
  file \
  libssl-dev \
  libglib2.0-dev \
  libwebkit2gtk-4.1-dev \
  libayatana-appindicator3-dev \
  librsvg2-dev \
  libxdo-dev \
  patchelf
```

---

## 2. Registry API: Production Setup & Deployment

The Registry API is an Express Node.js/TypeScript service providing plugin metadata distribution, stats, and crawler trigger endpoints.

### 2.1 Production Environment Variables (`.env`)
Create a production `.env` file in `/app/registry/.env` or supply via container secrets:

```env
# Application Runtime
NODE_ENV=production
PORT=3000
HOST=0.0.0.0
CORS_ORIGIN=https://desktop.vstmonster.internal,tauri://localhost

# PostgreSQL Production Connection Pool
PGHOST=postgres.vstmonster.internal
PGPORT=5432
PGUSER=vst_user
PGPASSWORD=STRONG_PRODUCTION_SECRET_KEY
PGDATABASE=vst_monster
PGSSLMODE=require
PGPOOL_MAX=20
PG_IDLE_TIMEOUT=30000
PG_CONN_TIMEOUT=5000

# Redis Cache Cluster
REDIS_URL=rediss://default:STRONG_REDIS_SECRET@redis.vstmonster.internal:6379
REDIS_TTL_SECONDS=3600

# Crawler Security & Authentication
CRAWLER_AUTH_TOKEN=STRONG_CRAWLER_HMAC_SECRET
GITHUB_TOKEN=ghp_ProductionPATForAPIRateLimits
```

### 2.2 Database Initialization & Migrations
Prior to launching the Registry API in production, ensure PostgreSQL schema and unique constraints are populated:
```bash
# Run schema migration directly via psql
PGPASSWORD=$PGPASSWORD psql -h $PGHOST -U $PGUSER -d $PGDATABASE -f /app/registry/schema.sql
```

### 2.3 Production Build & PM2 / Systemd Configuration

#### Build Artifacts
```bash
cd /app/registry
npm ci --omit=dev
npx tsc
```

#### Production Systemd Unit (`/etc/systemd/system/vst-registry.service`)
```ini
[Unit]
Description=VST Monster Registry API Service
After=network.target postgresql.service redis.service
Requires=postgresql.service

[Service]
Type=simple
User=vstmonster
WorkingDirectory=/opt/vst-monster/registry
EnvironmentFile=/etc/vst-monster/registry.env
ExecStart=/usr/bin/node dist/index.js
Restart=always
RestartSec=5s
LimitNOFILE=65536
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### Nginx Reverse Proxy with SSL Termination
```nginx
server {
    listen 443 ssl http2;
    server_name api.vstmonster.audio;

    ssl_certificate /etc/letsencrypt/live/api.vstmonster.audio/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.vstmonster.audio/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for streaming crawler logs
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }

    location /health {
        proxy_pass http://127.0.0.1:3000/health;
        access_log off;
    }
}
```

---

## 3. Tauri Desktop Client: Production Build & Distribution

The desktop client bundles a native Rust binary with a single-page Svelte 5 frontend compiled to static web assets via `@sveltejs/adapter-static`.

### 3.1 Pre-Compilation Checks
Ensure `tsconfig.json` and Svelte diagnostics report zero errors:
```bash
cd /app/client
npm ci
npm run check
```

### 3.2 Production Frontend Asset Generation
Compile SvelteKit to `/app/client/build`:
```bash
cd /app/client
npm run build
```

### 3.3 Production Rust Binary Packaging & Code Signing

Run the Tauri v2 bundler to produce cross-platform installer packages:
```bash
cd /app/client
npx tauri build
```

#### Output Artifact Matrix:
- **Linux**:
  - AppImage: `client/src-tauri/target/release/bundle/appimage/vst-monster_0.5.7_amd64.AppImage`
  - Debian package: `client/src-tauri/target/release/bundle/deb/vst-monster_0.5.7_amd64.deb`
- **macOS**:
  - DMG bundle: `client/src-tauri/target/release/bundle/dmg/vst-monster_0.5.7_x64.dmg` / `aarch64.dmg`
  - Signed `.app`: `client/src-tauri/target/release/bundle/macos/VST Monster.app`
- **Windows**:
  - NSIS installer: `client/src-tauri/target/release/bundle/nsis/vst-monster_0.5.7_x64-setup.exe`
  - MSI installer: `client/src-tauri/target/release/bundle/msi/vst-monster_0.5.7_x64_en-US.msi`

#### Code Signing & Notarization
- **macOS**: Provide environment variables `APPLE_CERTIFICATE`, `APPLE_CERTIFICATE_PASSWORD`, `APPLE_SIGNING_IDENTITY`, and `APPLE_ID` / `APPLE_PASSWORD` (App-Specific Password) for automated `codesign` and Apple Notary API verification.
- **Windows**: Provide `TAURI_SIGNING_PRIVATE_KEY` and Authenticode `.pfx` via `signtool.exe`.

---

## 4. Operational Monitoring & Health Verification

- **API Liveness & Version Check**:
  ```bash
  curl -s http://localhost:3000/health | jq .
  # Expected: {"status":"healthy","service":"vst-monster-registry","version":"0.5.7","timestamp":"..."}
  ```
- **Catalog Statistics & Metrics**:
  ```bash
  curl -s http://localhost:3000/stats | jq .
  ```
- **Local Desktop Diagnostics**:
  Launch the Tauri client with Rust backtrace enabled for inspection:
  ```bash
  RUST_BACKTRACE=1 RUST_LOG=info ./client/src-tauri/target/release/client
  ```
