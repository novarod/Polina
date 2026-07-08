import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { NodePanel } from "@/components/canvas/node-panel";

describe("NodePanel", () => {
  it("renders nothing without a selected node", () => {
    render(<NodePanel node={null} onClose={() => {}} />);
    expect(screen.queryByTestId("node-panel")).not.toBeInTheDocument();
  });

  it("shows id, type and formatted data", () => {
    render(
      <NodePanel
        node={{
          id: "talk",
          type: "DIALOGUE",
          payload: { npc: "Aldeão", lines: [1, 2] },
        }}
        onClose={() => {}}
      />
    );

    expect(screen.getByText("talk")).toBeVisible();
    expect(screen.getByText("DIALOGUE")).toBeVisible();
    expect(screen.getByTestId("node-panel-data").textContent).toBe(
      JSON.stringify({ npc: "Aldeão", lines: [1, 2] }, null, 2)
    );
  });

  it("indicates when the node has no data", () => {
    render(
      <NodePanel
        node={{ id: "start", type: "START", payload: null }}
        onClose={() => {}}
      />
    );

    expect(screen.getByText("Sem dados")).toBeVisible();
    expect(screen.queryByTestId("node-panel-data")).not.toBeInTheDocument();
  });

  it("calls onClose from the close button", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <NodePanel
        node={{ id: "start", type: "START", payload: null }}
        onClose={onClose}
      />
    );

    await user.click(screen.getByLabelText("Fechar painel"));
    expect(onClose).toHaveBeenCalled();
  });
});
