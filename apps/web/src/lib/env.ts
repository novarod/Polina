const DEFAULT_API_URL = "http://localhost:8080";

function readApiUrl(): string {
  const raw = process.env.API_URL ?? DEFAULT_API_URL;
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    throw new Error(
      `API_URL must be an absolute URL like "http://localhost:8080", got "${raw}"`
    );
  }
  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error(
      `API_URL must use http or https, got "${url.protocol}//"`
    );
  }
  return url.toString().replace(/\/$/, "");
}

export const env = {
  apiUrl: readApiUrl(),
};
