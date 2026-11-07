"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.server = void 0;
const express_1 = __importDefault(require("express"));
const cors_1 = __importDefault(require("cors"));
const pg_1 = require("pg");
const child_process_1 = require("child_process");
const path_1 = __importDefault(require("path"));
const app = (0, express_1.default)();
const port = process.env.PORT || 3000;
app.use((0, cors_1.default)());
app.use(express_1.default.json());
// Database connection setup
const pool = new pg_1.Pool({
    user: process.env.PGUSER || 'postgres',
    host: process.env.PGHOST || 'localhost',
    database: process.env.PGDATABASE || 'vst_monster',
    password: process.env.PGPASSWORD || 'postgres',
    port: parseInt(process.env.PGPORT || '5432'),
});
let isCrawling = false;
let crawlerLogs = [];
// Health check endpoint
app.get('/health', async (_req, res) => {
    let dbOk = false;
    try {
        const dbRes = await pool.query('SELECT 1');
        dbOk = dbRes.rowCount ? dbRes.rowCount > 0 : false;
    }
    catch (err) {
        dbOk = false;
    }
    res.json({
        status: 'ok',
        timestamp: new Date().toISOString(),
        database: dbOk ? 'connected' : 'disconnected',
        crawlerRunning: isCrawling,
        version: '0.2.0'
    });
});
// Full joined plugins with distributions and releases
app.get('/plugins', async (req, res) => {
    try {
        const query = `
      SELECT
        p.id,
        p.name,
        p.developer,
        p.license_model,
        p.created_at,
        p.updated_at,
        r.version,
        r.release_date,
        d.platform,
        d.architecture,
        d.download_url,
        d.sha256_hash,
        d.strategy,
        d.extraction_rules as metadata
      FROM plugins p
      LEFT JOIN LATERAL (
        SELECT id, version, release_date
        FROM plugin_releases
        WHERE plugin_id = p.id
        ORDER BY release_date DESC NULLS LAST
        LIMIT 1
      ) r ON true
      LEFT JOIN LATERAL (
        SELECT platform, architecture, download_url, sha256_hash, strategy, extraction_rules
        FROM plugin_distributions
        WHERE release_id = r.id
        LIMIT 1
      ) d ON true
      ORDER BY p.name ASC;
    `;
        const result = await pool.query(query);
        res.json(result.rows);
    }
    catch (err) {
        console.error('Error fetching plugins', err);
        res.status(500).json({ error: 'Internal server error' });
    }
});
// Manual plugin registration endpoint
app.post('/plugins', async (req, res) => {
    const { name, developer, license_model, version, platform, architecture, download_url, sha256_hash, description } = req.body;
    if (!name || !developer) {
        return res.status(400).json({ error: 'Name and developer are required fields.' });
    }
    try {
        const pluginRes = await pool.query(`
      INSERT INTO plugins (name, developer, license_model, updated_at)
      VALUES ($1, $2, $3, NOW())
      ON CONFLICT (name, developer) DO UPDATE SET
        license_model = EXCLUDED.license_model,
        updated_at = NOW()
      RETURNING id;
    `, [name, developer, license_model || 'free']);
        const pluginId = pluginRes.rows[0].id;
        const relVersion = version || '1.0.0';
        const relRes = await pool.query(`
      INSERT INTO plugin_releases (plugin_id, version, release_date)
      VALUES ($1, $2, NOW())
      ON CONFLICT (plugin_id, version) DO UPDATE SET
        release_date = EXCLUDED.release_date
      RETURNING id;
    `, [pluginId, relVersion]);
        const releaseId = relRes.rows[0].id;
        if (download_url) {
            await pool.query(`
        INSERT INTO plugin_distributions (release_id, platform, architecture, download_url, sha256_hash, strategy, extraction_rules)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
      `, [
                releaseId,
                platform || 'linux',
                architecture || 'x86_64',
                download_url,
                sha256_hash || '',
                'extract_binaries',
                JSON.stringify({ description: description || '', source: 'manual' })
            ]);
        }
        return res.status(201).json({ message: 'Plugin registered successfully', id: pluginId });
    }
    catch (err) {
        console.error('Error inserting plugin', err);
        return res.status(500).json({ error: err.message || 'Failed to insert plugin' });
    }
});
// Statistics endpoint for the high-value dashboard metrics
app.get('/stats', async (_req, res) => {
    try {
        const totalPlugins = await pool.query('SELECT count(*) FROM plugins');
        const licenseStats = await pool.query(`
      SELECT license_model, count(*)
      FROM plugins
      GROUP BY license_model
    `);
        const distributions = await pool.query('SELECT count(*) FROM plugin_distributions');
        res.json({
            totalPlugins: parseInt(totalPlugins.rows[0]?.count || '0', 10),
            totalDistributions: parseInt(distributions.rows[0]?.count || '0', 10),
            licenses: licenseStats.rows,
            crawlerRunning: isCrawling
        });
    }
    catch (err) {
        console.error('Error fetching stats', err);
        res.status(500).json({ error: 'Internal server error' });
    }
});
// Trigger crawler execution from UI
app.post('/crawler/run', (req, res) => {
    if (isCrawling) {
        return res.status(409).json({ message: 'Crawler is already actively running.' });
    }
    isCrawling = true;
    crawlerLogs = [`[${new Date().toLocaleTimeString()}] Crawler started`];
    const crawlerDir = path_1.default.resolve(__dirname, '../../crawler');
    const crawlerProcess = (0, child_process_1.spawn)('go', ['run', '.'], {
        cwd: crawlerDir,
        env: {
            ...process.env,
            PGHOST: process.env.PGHOST || 'localhost',
            PGPORT: process.env.PGPORT || '5432',
            PGUSER: process.env.PGUSER || 'postgres',
            PGPASSWORD: process.env.PGPASSWORD || 'postgres',
            PGDATABASE: process.env.PGDATABASE || 'vst_monster',
        }
    });
    crawlerProcess.stdout.on('data', (data) => {
        const lines = data.toString().trim().split('\n');
        crawlerLogs.push(...lines);
        if (crawlerLogs.length > 200)
            crawlerLogs = crawlerLogs.slice(-200);
    });
    crawlerProcess.stderr.on('data', (data) => {
        const lines = data.toString().trim().split('\n');
        crawlerLogs.push(...lines);
        if (crawlerLogs.length > 200)
            crawlerLogs = crawlerLogs.slice(-200);
    });
    crawlerProcess.on('close', (code) => {
        isCrawling = false;
        crawlerLogs.push(`[${new Date().toLocaleTimeString()}] Crawler finished with exit code ${code}`);
    });
    return res.json({ message: 'Crawler initiated successfully', status: 'started' });
});
// Retrieve crawler logs
app.get('/crawler/logs', (_req, res) => {
    res.json({
        running: isCrawling,
        logs: crawlerLogs
    });
});
// System shutdown / quit endpoint
app.post('/system/shutdown', (_req, res) => {
    res.json({ message: 'Ecosystem shutdown initiated' });
    setTimeout(() => {
        process.exit(0);
    }, 500);
});
exports.server = app.listen(port, () => {
    console.log(`Registry API server is running on port ${port}`);
});
exports.default = app;
//# sourceMappingURL=index.js.map