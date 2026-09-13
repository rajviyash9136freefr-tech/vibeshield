/**
 * Golden fixture: correct password hashing (argon2id via the argon2 package).
 * Clean control for the hashing family: must fire zero findings.
 */
const argon2 = require("argon2");

async function hashPassword(plain) {
  return argon2.hash(plain, { type: argon2.argon2id, memoryCost: 19456, timeCost: 2 });
}

async function verifyPassword(hashed, plain) {
  return argon2.verify(hashed, plain);
}

module.exports = { hashPassword, verifyPassword };
