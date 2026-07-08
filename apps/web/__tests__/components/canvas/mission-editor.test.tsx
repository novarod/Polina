import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { MissionCanvas } from "@/components/canvas/mission-canvas";
import { ApiError } from "@/services/api";
import { updateMissionGraph } from "@/services/missions";
import type { EditorGraph } from "@/types/graph";

vi.mock("@/services/missions");

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess },
}));

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

beforeAll(() => {
  globalThis.ResizeObserver =
    globalThis.ResizeObserver ?? ResizeObserverStub;
});

const graph: EditorGraph = {
  nodes: [
    { id: "start-1", type: "START", position: { x: 0, y: 0 } },
    {
      id: "dialogue-1",
      type: "DIALOGUE",
      data: { npc: "Aldeão" },
      position: { x: 0, y: 120 },
    },
    { id: "end-1", type: "END", position: { x: 0, y: 240 } },
  ],
  edges: [
    { id: "e1", source: "start-1", target: "dialogue-1" },
    { id: "e2", source: "dialogue-1", target: "end-1" },
  ],
};

function renderEditor(initial: EditorGraph = graph) {
  return render(
    <MissionCanvas
      editable
      graph={initial}
      orgId="o1"
      workspaceId="w1"
      missionId="m1"
    />
  );
}

describe("MissionCanvas editable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows palette and save bar, opens with no dirty state", () => {
    renderEditor();

    expect(screen.getByTestId("node-palette")).toBeVisible();
    expect(screen.getByTestId("save-bar")).toBeVisible();
    expect(screen.queryByTestId("dirty-indicator")).not.toBeInTheDocument();
    expect(screen.queryByTestId("graph-error-count")).not.toBeInTheDocument();
  });

  it("adds a node from the palette and marks the graph dirty", async () => {
    const user = userEvent.setup();
    renderEditor();

    await user.click(screen.getByTestId("palette-OBJECTIVE"));

    expect(screen.getAllByTestId("quest-node")).toHaveLength(4);
    expect(screen.getByTestId("dirty-indicator")).toBeVisible();
    expect(screen.getByTestId("graph-error-count")).toBeVisible();
  });

  it("shows live dead-end badges after deleting a node via the panel", async () => {
    const user = userEvent.setup();
    renderEditor();

    fireEvent.click(screen.getByText("end-1"));
    await user.click(screen.getByTestId("delete-node"));

    expect(screen.getAllByTestId("quest-node")).toHaveLength(2);
    await waitFor(() => {
      expect(
        screen.getAllByTestId("node-error-badge").length
      ).toBeGreaterThan(0);
    });
    expect(screen.queryByTestId("node-panel")).not.toBeInTheDocument();
  });

  it("saves through the service and clears the dirty state", async () => {
    vi.mocked(updateMissionGraph).mockResolvedValue(
      {} as Awaited<ReturnType<typeof updateMissionGraph>>
    );
    const user = userEvent.setup();
    renderEditor();

    await user.click(screen.getByTestId("palette-OBJECTIVE"));
    await user.click(screen.getByRole("button", { name: "Salvar" }));

    await waitFor(() => {
      expect(updateMissionGraph).toHaveBeenCalled();
      expect(toastSuccess).toHaveBeenCalledWith("Grafo salvo");
    });
    expect(screen.queryByTestId("dirty-indicator")).not.toBeInTheDocument();
  });

  it("shows the API 422 banner and keeps the dirty state", async () => {
    vi.mocked(updateMissionGraph).mockRejectedValue(
      new ApiError(422, "dag validation failed: dead-end")
    );
    const user = userEvent.setup();
    renderEditor();

    await user.click(screen.getByTestId("palette-OBJECTIVE"));
    await user.click(screen.getByRole("button", { name: "Salvar" }));

    expect(await screen.findByTestId("graph-api-error")).toHaveTextContent(
      "dag validation failed: dead-end"
    );
    expect(screen.getByTestId("dirty-indicator")).toBeVisible();

    await user.click(screen.getByLabelText("Fechar erro"));
    expect(screen.queryByTestId("graph-api-error")).not.toBeInTheDocument();
  });

  it("closes the node panel when the pane is clicked", async () => {
    renderEditor();

    fireEvent.click(screen.getByText("dialogue-1"));
    expect(await screen.findByTestId("node-panel")).toBeVisible();

    const pane = document.querySelector(".react-flow__pane");
    expect(pane).not.toBeNull();
    fireEvent.click(pane as Element);

    await waitFor(() => {
      expect(screen.queryByTestId("node-panel")).not.toBeInTheDocument();
    });
  });
});
