import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useGraphEditor } from "@/hooks/use-graph-editor";
import { ApiError } from "@/services/api";
import { updateMissionGraph } from "@/services/missions";
import type { EditorGraph } from "@/types/graph";

vi.mock("@/services/missions");

const toastSuccess = vi.hoisted(() => vi.fn());
vi.mock("sonner", () => ({
  toast: { success: toastSuccess },
}));

const initialGraph: EditorGraph = {
  nodes: [
    { id: "start-1", type: "START", position: { x: 0, y: 0 } },
    { id: "end-1", type: "END", position: { x: 0, y: 200 } },
  ],
  edges: [{ id: "e-start-1-end-1", source: "start-1", target: "end-1" }],
};

function setup(graph: EditorGraph = initialGraph) {
  return renderHook(() =>
    useGraphEditor({
      orgId: "o1",
      workspaceId: "w1",
      missionId: "m1",
      initialGraph: graph,
    })
  );
}

describe("useGraphEditor", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("starts clean and valid", () => {
    const { result } = setup();
    expect(result.current.dirty).toBe(false);
    expect(result.current.validation.errorCount).toBe(0);
    expect(result.current.hasStart).toBe(true);
  });

  it("adds a node with a generated readable id and marks dirty", () => {
    const { result } = setup();
    act(() => result.current.addNode("DIALOGUE", { x: 100, y: 100 }));

    expect(
      result.current.nodes.some((node) => node.id === "dialogue-1")
    ).toBe(true);
    expect(result.current.dirty).toBe(true);

    act(() => result.current.addNode("DIALOGUE", { x: 100, y: 100 }));
    expect(
      result.current.nodes.some((node) => node.id === "dialogue-2")
    ).toBe(true);
  });

  it("ignores an exact duplicate connection", () => {
    const { result } = setup();
    act(() =>
      result.current.onConnect({
        source: "start-1",
        target: "end-1",
        sourceHandle: null,
        targetHandle: null,
      })
    );
    expect(result.current.edges).toHaveLength(1);
  });

  it("deletes a node together with its edges", () => {
    const { result } = setup();
    act(() => result.current.deleteNode("end-1"));

    expect(result.current.nodes).toHaveLength(1);
    expect(result.current.edges).toHaveLength(0);
    expect(result.current.validation.errorCount).toBeGreaterThan(0);
  });

  it("updates node data through the editor", () => {
    const { result } = setup();
    act(() =>
      result.current.updateNodeData("start-1", { musica: "tema-da-vila" })
    );

    expect(
      result.current.nodes.find((node) => node.id === "start-1")?.data.payload
    ).toEqual({ musica: "tema-da-vila" });
    expect(result.current.dirty).toBe(true);
  });

  it("saves, resets dirty and keeps the baseline", async () => {
    vi.mocked(updateMissionGraph).mockResolvedValue(
      {} as Awaited<ReturnType<typeof updateMissionGraph>>
    );
    const { result } = setup();
    act(() => result.current.updateNodeData("start-1", { x: 1 }));
    expect(result.current.dirty).toBe(true);

    await act(() => result.current.save());

    expect(updateMissionGraph).toHaveBeenCalledWith(
      "o1",
      "w1",
      "m1",
      expect.objectContaining({
        nodes: expect.arrayContaining([
          expect.objectContaining({ id: "start-1", data: { x: 1 } }),
        ]),
      })
    );
    await waitFor(() => {
      expect(result.current.dirty).toBe(false);
      expect(toastSuccess).toHaveBeenCalledWith("Grafo salvo");
    });
  });

  it("surfaces a 422 as apiError and stays dirty", async () => {
    vi.mocked(updateMissionGraph).mockRejectedValue(
      new ApiError(422, 'dag validation failed: [node "x" ...]')
    );
    const { result } = setup();
    act(() => result.current.deleteNode("end-1"));

    await act(() => result.current.save());

    expect(result.current.apiError).toContain("dag validation failed");
    expect(result.current.dirty).toBe(true);

    act(() => result.current.dismissApiError());
    expect(result.current.apiError).toBeNull();
  });
});
