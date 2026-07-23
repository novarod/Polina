import { describe, expect, it } from "vitest";

import { presenceColor, presenceInitials } from "@/lib/presence";

describe("presenceColor", () => {
  it("é determinística por user id", () => {
    expect(presenceColor("abc")).toBe(presenceColor("abc"));
  });

  it("sempre resolve para um token da paleta", () => {
    for (const id of ["a", "b", "c", "d", "e", "f", "g", "h"]) {
      expect(presenceColor(id)).toMatch(/^var\(--presence-[1-6]\)$/);
    }
  });
});

describe("presenceInitials", () => {
  it("usa primeira e última palavra do nome", () => {
    expect(presenceInitials("Ana Beatriz Souza")).toBe("AS");
  });

  it("usa uma letra para nome único", () => {
    expect(presenceInitials("ana")).toBe("A");
  });

  it("cai para ? quando vazio", () => {
    expect(presenceInitials("  ")).toBe("?");
  });
});
