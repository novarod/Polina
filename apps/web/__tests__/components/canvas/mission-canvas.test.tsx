import { fireEvent, render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it } from "vitest";

import { MissionCanvas } from "@/components/canvas/mission-canvas";
import type { EditorGraph } from "@/types/graph";

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
    { id: "start", type: "START", position: { x: 0, y: 0 } },
    {
      id: "talk",
      type: "DIALOGUE",
      data: { npc: "Aldeão" },
      position: { x: 0, y: 120 },
    },
    { id: "end", type: "END", position: { x: 0, y: 240 } },
  ],
  edges: [
    { id: "e1", source: "start", target: "talk" },
    { id: "e2", source: "talk", target: "end" },
  ],
};

describe("MissionCanvas", () => {
  it("shows the empty state for an empty graph", () => {
    render(<MissionCanvas graph={{ nodes: [], edges: [] }} />);

    expect(screen.getByTestId("canvas-empty")).toBeVisible();
    expect(screen.queryByTestId("mission-canvas")).not.toBeInTheDocument();
  });

  it("renders one custom node per graph node", () => {
    render(<MissionCanvas graph={graph} />);

    const nodes = screen.getAllByTestId("quest-node");
    expect(nodes).toHaveLength(3);
    expect(
      nodes.map((node) => node.getAttribute("data-node-type"))
    ).toEqual(expect.arrayContaining(["START", "DIALOGUE", "END"]));
  });

  it("opens the read-only panel when a node is clicked", async () => {
    render(<MissionCanvas graph={graph} />);

    fireEvent.click(screen.getByText("talk"));

    expect(await screen.findByTestId("node-panel")).toBeVisible();
    expect(screen.getByTestId("node-panel-data")).toHaveTextContent("Aldeão");
  });
});
