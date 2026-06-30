import express, { Request, Response } from 'express';
import cors from 'cors';
import { Pool } from 'pg';

const app = express();
const port = process.env.PORT || 3000;

app.use(cors());
app.use(express.json());

// Database connection setup
const pool = new Pool({
  user: process.env.PGUSER || 'postgres',
  host: process.env.PGHOST || 'localhost',
  database: process.env.PGDATABASE || 'vst_monster',
  password: process.env.PGPASSWORD || 'postgres',
  port: parseInt(process.env.PGPORT || '5432'),
});

app.get('/plugins', async (req: Request, res: Response) => {
  try {
    const result = await pool.query('SELECT * FROM plugins ORDER BY name ASC');
    res.json(result.rows);
  } catch (err) {
    console.error('Error fetching plugins', err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

app.listen(port, () => {
  console.log(`Registry API server is running on port ${port}`);
});
