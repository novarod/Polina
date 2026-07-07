import { afterEach, describe, expect, it, vi } from "vitest";

import { serverFetch } from "@/services/server-api";

const getCookie = vi.hoisted(() => vi.fn());
vi.mock("next/headers", () => ({
  cookies: async () => ({ get: getCookie }),
}));

function withSession() {
  getCookie.mockReturnValue({ name: "session", value: "token-123" });
}

describe("serverFetch", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    getCookie.mockReset();
  });

  it("returns null without a session cookie", async () => {
    getCookie.mockReturnValue(undefined);
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    expect(await serverFetch("/auth/me")).toBeNull();
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("forwards the cookie and returns parsed data on 200", async () => {
    withSession();
    const fetchSpy = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(JSON.stringify({ id: "o1" }), { status: 200 })
    );

    const result = await serverFetch<{ id: string }>("/organizations/o1");

    expect(result).toEqual({ id: "o1" });
    expect(fetchSpy).toHaveBeenCalledWith(
      "http://localhost:8080/organizations/o1",
      {
        headers: { cookie: "session=token-123" },
        cache: "no-store",
      }
    );
  });

  it.each([401, 403, 404, 500])("returns null on %s", async (status) => {
    withSession();
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response("{}", { status })
    );

    expect(await serverFetch("/organizations/o1")).toBeNull();
  });

  it("returns null on network failure", async () => {
    withSession();
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new TypeError("fail"));

    expect(await serverFetch("/organizations/o1")).toBeNull();
  });
});
