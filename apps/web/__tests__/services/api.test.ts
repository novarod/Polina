import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError, apiFetch } from "@/services/api";

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("apiFetch", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns the parsed body on success", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      jsonResponse(200, { user_id: "u1", name: "Alice" })
    );

    const result = await apiFetch<{ user_id: string; name: string }>(
      "/auth/me"
    );

    expect(result).toEqual({ user_id: "u1", name: "Alice" });
    expect(fetch).toHaveBeenCalledWith("/api/auth/me", {
      method: "GET",
      headers: undefined,
      body: undefined,
    });
  });

  it("redirects to /login on 401 by default", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      jsonResponse(401, { message: "invalid or expired session" })
    );
    const assign = vi
      .spyOn(window.location, "assign")
      .mockImplementation(() => {});

    await expect(apiFetch("/organizations")).rejects.toMatchObject({
      status: 401,
    });
    expect(assign).toHaveBeenCalledWith("/login");
  });

  it("throws instead of redirecting when redirectOn401 is false", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      jsonResponse(401, { message: "invalid credentials" })
    );
    const assign = vi
      .spyOn(window.location, "assign")
      .mockImplementation(() => {});

    await expect(
      apiFetch("/auth/login", {
        method: "POST",
        body: { email: "a@b.com", password: "x" },
        redirectOn401: false,
      })
    ).rejects.toMatchObject({ status: 401, message: "invalid credentials" });
    expect(assign).not.toHaveBeenCalled();
  });

  it("surfaces the API error message on non-401 failures", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      jsonResponse(422, { message: "email already in use" })
    );

    await expect(apiFetch("/auth/register", { method: "POST", body: {} }))
      .rejects.toMatchObject({ status: 422, message: "email already in use" });
  });

  it("wraps network failures in a readable ApiError", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new TypeError("fail"));

    const error = await apiFetch("/auth/login").catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      status: 0,
      message: "Não foi possível conectar ao servidor",
    });
  });
});
