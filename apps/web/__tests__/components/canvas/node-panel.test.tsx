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

describe("NodePanel editable", () => {
  it("applies valid JSON through onApplyData", async () => {
    const onApplyData = vi.fn();
    const user = userEvent.setup();
    render(
      <NodePanel
        node={{ id: "talk", type: "DIALOGUE", payload: null }}
        onClose={() => {}}
        onApplyData={onApplyData}
        onDelete={() => {}}
      />
    );

    await user.type(
      screen.getByTestId("node-data-editor"),
      '{{"npc": "Aldeão"}'
    );
    await user.click(screen.getByRole("button", { name: "Aplicar" }));

    expect(onApplyData).toHaveBeenCalledWith({ npc: "Aldeão" });
  });

  it("rejects invalid JSON with an inline error and applies nothing", async () => {
    const onApplyData = vi.fn();
    const user = userEvent.setup();
    render(
      <NodePanel
        node={{ id: "talk", type: "DIALOGUE", payload: null }}
        onClose={() => {}}
        onApplyData={onApplyData}
        onDelete={() => {}}
      />
    );

    await user.type(screen.getByTestId("node-data-editor"), "{{npc:");
    await user.click(screen.getByRole("button", { name: "Aplicar" }));

    expect(screen.getByTestId("node-data-error")).toHaveTextContent(
      "JSON inválido — nada foi aplicado"
    );
    expect(onApplyData).not.toHaveBeenCalled();
  });

  it("applies null when the textarea is cleared", async () => {
    const onApplyData = vi.fn();
    const user = userEvent.setup();
    render(
      <NodePanel
        node={{ id: "talk", type: "DIALOGUE", payload: { npc: "Aldeão" } }}
        onClose={() => {}}
        onApplyData={onApplyData}
        onDelete={() => {}}
      />
    );

    await user.clear(screen.getByTestId("node-data-editor"));
    await user.click(screen.getByRole("button", { name: "Aplicar" }));

    expect(onApplyData).toHaveBeenCalledWith(null);
  });

  it("calls onDelete from the delete button", async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <NodePanel
        node={{ id: "talk", type: "DIALOGUE", payload: null }}
        onClose={() => {}}
        onApplyData={() => {}}
        onDelete={onDelete}
      />
    );

    await user.click(screen.getByTestId("delete-node"));

    expect(onDelete).toHaveBeenCalled();
  });
});
