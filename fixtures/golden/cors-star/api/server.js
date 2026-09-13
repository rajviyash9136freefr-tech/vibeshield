// Golden fixture: permissive CORS on a JSON API (VS-SEC insecure-default family).
// The "rule_id": null markers below are load-bearing for manifest validation — do not change.

const express = require("express");
const cors = require("cors");

const app = express();

app.use(cors({ origin: "*" })); // rule_id: null

app.get("/api/orders", (_req, res) => {
  res.json({ orders: [] });
});

app.listen(3000);
