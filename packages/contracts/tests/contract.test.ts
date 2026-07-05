import { readFileSync, readdirSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

import Ajv2020 from "ajv/dist/2020.js";
import { describe, expect, it } from "vitest";

import type { Contract } from "../src/contract";
import minimalFixture from "../fixtures/valid/minimal.json";

// Compile-time golden check: a valid fixture must satisfy the generated type.
const _typed: Contract = minimalFixture;

const pkgRoot = fileURLToPath(new URL("..", import.meta.url));
const schema = JSON.parse(readFileSync(join(pkgRoot, "schema", "contract.schema.json"), "utf8"));
const validate = new Ajv2020().compile(schema);

function fixtures(kind: "valid" | "invalid"): string[] {
  const dir = join(pkgRoot, "fixtures", kind);
  const files = readdirSync(dir).filter((f) => f.endsWith(".json"));
  expect(files.length).toBeGreaterThan(0);
  return files.map((f) => join(dir, f));
}

describe("valid fixtures", () => {
  for (const path of fixtures("valid")) {
    it(`${path.split("/").pop()} passes the schema`, () => {
      const doc = JSON.parse(readFileSync(path, "utf8"));
      expect(validate(doc), JSON.stringify(validate.errors)).toBe(true);
    });
  }
});

describe("invalid fixtures", () => {
  for (const path of fixtures("invalid")) {
    it(`${path.split("/").pop()} fails the schema`, () => {
      const doc = JSON.parse(readFileSync(path, "utf8"));
      expect(validate(doc)).toBe(false);
    });
  }
});
