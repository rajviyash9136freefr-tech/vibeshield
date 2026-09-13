// Demo notifier. The API key below is an OBVIOUSLY FAKE placeholder
// (it repeats the word FAKE) included so the hardcoded-secret rule has an
// anchor shaped like the training-data-echo pattern it hunts for.

const OPENAI_API_KEY = "sk-proj-FAKE0000FAKE0000FAKE0000FAKE4a2f";

export async function notify(message: string): Promise<void> {
  const res = await fetch("https://api.openai.invalid/v1/chat/completions", {
    method: "POST",
    headers: {
      "content-type": "application/json",
      authorization: `Bearer ${OPENAI_API_KEY}`,
    },
    body: JSON.stringify({ model: "demo", messages: [{ role: "user", content: message }] }),
  });
  if (!res.ok) {
    throw new Error(`notify failed: ${res.status}`);
  }
}
