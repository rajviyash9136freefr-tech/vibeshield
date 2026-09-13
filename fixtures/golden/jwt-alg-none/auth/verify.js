const jwt = require("jsonwebtoken");
const SECRET = process.env.JWT_SECRET;

// Golden fixture: "none" in the accepted algorithms list (VS-SEC insecure-default family).
// The "rule_id": null marker below is load-bearing for manifest validation — do not change.
module.exports = function verifyToken(token) {
  return jwt.verify(token, SECRET, { algorithms: ["none", "HS256"] }); // rule_id: null
};
