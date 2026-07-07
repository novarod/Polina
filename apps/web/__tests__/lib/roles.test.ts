import { describe, expect, it } from "vitest";

import { roleAtLeast, type Role } from "@/lib/roles";

describe("roleAtLeast", () => {
  const cases: Array<[Role, Role, boolean]> = [
    ["VIEWER", "VIEWER", true],
    ["VIEWER", "DESIGNER", false],
    ["VIEWER", "ADMIN", false],
    ["DESIGNER", "VIEWER", true],
    ["DESIGNER", "DESIGNER", true],
    ["DESIGNER", "ADMIN", false],
    ["ADMIN", "VIEWER", true],
    ["ADMIN", "DESIGNER", true],
    ["ADMIN", "ADMIN", true],
  ];

  it.each(cases)("%s atLeast %s -> %s", (role, min, expected) => {
    expect(roleAtLeast(role, min)).toBe(expected);
  });
});
