const PRESENCE_TOKENS = [
  "var(--presence-1)",
  "var(--presence-2)",
  "var(--presence-3)",
  "var(--presence-4)",
  "var(--presence-5)",
  "var(--presence-6)",
] as const;

export function presenceColor(userId: string): string {
  let hash = 0;
  for (const char of userId) {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }
  return PRESENCE_TOKENS[hash % PRESENCE_TOKENS.length];
}

export function presenceInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const first = parts[0]?.charAt(0) ?? "";
  const last = parts.length > 1 ? (parts.at(-1)?.charAt(0) ?? "") : "";
  const joined = (first + last).toUpperCase();
  return joined === "" ? "?" : joined;
}
