/**
 * Reads the API key from the environment — the correct pattern.
 * This is a clean control: referencing a secret name must not fire the
 * hardcoded-secret rule; only the literal value may.
 */
const key = process.env.OPENAI_API_KEY;

if (!key) {
  throw new Error("OPENAI_API_KEY is not set");
}

export function hasKey(): boolean {
  return typeof key === "string" && key.length > 0;
}
