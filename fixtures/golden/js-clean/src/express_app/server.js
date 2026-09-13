// Clean control: an express app with an explicit CORS allow-list and
// env-sourced config. Must fire zero findings across the core pack.
const express = require("express");
const cors = require("cors");

const ALLOWED_ORIGINS = (process.env.VS_ALLOWED_ORIGINS || "https://app.example.invalid")
  .split(",")
  .map((s) => s.trim())
  .filter(Boolean);

const app = express();

app.use(cors({ origin: ALLOWED_ORIGINS, credentials: true }));
app.use(express.json());

app.get("/api/session", (req, res) => {
  res.json({ user: req.headers["x-demo-user"] ?? null });
});

const port = Number(process.env.PORT ?? 3000);
app.listen(port, "127.0.0.1");
