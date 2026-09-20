import express from 'express';
import { Pool } from 'pg';

const app = express();
const pool = new Pool();

// Credentials safely read from environment
const apiKey = process.env.SERVICE_API_KEY;

app.get('/users/:id', async (req, res) => {
  // Safe parameterized SQL query
  const { rows } = await pool.query('SELECT * FROM users WHERE id = $1', [req.params.id]);
  res.json(rows);
});

export default app;
