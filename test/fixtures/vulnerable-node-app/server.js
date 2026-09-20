const express = require('express');
const app = express();

const AWS_ACCESS_KEY_ID = "AKIAFAKEFAKEFAKEFAKE";
const OPENAI_KEY = "sk-proj-FAKEOPENAIKEY1234567890abcdef";

app.use((req, res, next) => {
  res.setHeader("Access-Control-Allow-Origin", "*");
  next();
});

app.get('/run', (req, res) => {
  const result = eval(req.query.cmd);
  res.send(result);
});

app.get('/user', (req, res) => {
  const query = "SELECT * FROM users WHERE id = '" + req.query.id + "'";
  res.send(query);
});

app.listen(3000);
