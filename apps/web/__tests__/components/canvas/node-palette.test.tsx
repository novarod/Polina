import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReactFlowProvider } from "@xyflow/react";
import { describe, expect, it, vi } from "vitest";

import { NodePalette } from "@/components/canvas/node-palette";

function renderPalette(hasStart: boolean, onAdd = vi.fn()) {
  render(
    <ReactFlowProvider>
      <NodePalette hasStart={hasStart} onAdd={onAdd} />
    </ReactFlowProvider>
  );
  return onAdd;
}

describe("NodePalette", () => {
  it("lists the six node types", () => {
    renderPalette(false);
    for (const type of [
      "START",
      "END",
      "DIALOGUE",
      "KILL",
      "COLLECT",
      "OBJECTIVE",
    ]) {
      expect(screen.getByTestId(`palette-${type}`)).toBeInTheDocument();
    }
  });

  it("adds a node with the clicked type", async () => {
    const user = userEvent.setup();
    const onAdd = renderPalette(false);

    await user.click(screen.getByTestId("palette-DIALOGUE"));

    expect(onAdd).toHaveBeenCalledWith(
      "DIALOGUE",
      expect.objectContaining({ x: expect.any(Number) })
    );
  });

  it("disables START when the graph already has one", () => {
    renderPalette(true);
    expect(screen.getByTestId("palette-START")).toBeDisabled();
    expect(screen.getByTestId("palette-END")).toBeEnabled();
  });
});
