// Clean control: pure helper, no patterns any core rule hunts for. Must fire zero findings.

export function slugify(text) {
  return text.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
}
