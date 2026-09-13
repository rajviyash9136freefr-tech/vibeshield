// Golden fixture: expression evaluator runs request input (VS-SEC insecure-api family).
// The "rule_id": null marker below is load-bearing for manifest validation — do not change.

export function compute(req, res) {
  const expression = req.query.expression;
  const result = eval(expression); // rule_id: null
  res.json({ result });
}
