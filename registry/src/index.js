"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const cors_1 = __importDefault(require("cors"));
const pg_1 = require("pg");
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
app.get('/plugins', async (req, res) => {
    try {
        const result = await pool.query('SELECT * FROM plugins ORDER BY name ASC');
        res.json(result.rows);
    }
    catch (err) {
        console.error('Error fetching plugins', err);
        res.status(500).json({ error: 'Internal server error' });
    }
});
app.listen(port, () => {
    console.log(`Registry API server is running on port ${port}`);
});
//# sourceMappingURL=index.js.map