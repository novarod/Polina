import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { NodeSprite } from "@/components/canvas/node-sprite";

describe("NodeSprite", () => {
  it.each(["START", "END", "DIALOGUE", "KILL", "COLLECT", "OBJECTIVE"])(
    "renders the dedicated sprite for %s",
    (type) => {
      render(<NodeSprite type={type} />);
      expect(screen.getByTestId("node-sprite")).toHaveAttribute(
        "data-sprite",
        type
      );
    }
  );

  it("falls back to the generic sprite for unknown types", () => {
    render(<NodeSprite type="TELEPORT" />);
    expect(screen.getByTestId("node-sprite")).toHaveAttribute(
      "data-sprite",
      "UNKNOWN"
    );
  });
});
