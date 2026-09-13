import assert from "node:assert/strict";
import test from "node:test";

import { slugify } from "../src/util.js";

test("slugify lowercases and hyphenates", () => {
  assert.equal(slugify("Hello, World!"), "hello-world");
});
