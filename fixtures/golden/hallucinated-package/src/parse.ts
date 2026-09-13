// Golden fixture: parses an event payload with a hallucinated package (VS-PKG-001).
import { parseEvent } from "fast-parse-utils-v3";

export function summarize(raw: string): string {
  const event = parseEvent(raw);
  return `${event.type}:${event.id}`;
}
