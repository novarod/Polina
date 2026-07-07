import { afterEach, describe, expect, it, vi } from "vitest";

async function loadEnv() {
  vi.resetModules();
  const mod = await import("@/lib/env");
  return mod.env;
}

describe("env", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("defaults API_URL to the local backend when absent", async () => {
    vi.stubEnv("API_URL", undefined);
    const env = await loadEnv();
    expect(env.apiUrl).toBe("http://localhost:8080");
  });

  it("strips a trailing slash", async () => {
    vi.stubEnv("API_URL", "https://api.polina.gg/");
    const env = await loadEnv();
    expect(env.apiUrl).toBe("https://api.polina.gg");
  });

  it("fails fast on a malformed URL", async () => {
    vi.stubEnv("API_URL", "not-a-url");
    await expect(loadEnv()).rejects.toThrow(
      'API_URL must be an absolute URL like "http://localhost:8080", got "not-a-url"'
    );
  });

  it("rejects non-http protocols", async () => {
    vi.stubEnv("API_URL", "ftp://api.polina.gg");
    await expect(loadEnv()).rejects.toThrow("API_URL must use http or https");
  });
});
