import { describe, expect, it } from "vitest";

import { layoutGraph, toEditorGraph } from "@/lib/graph-layout";
import type { EditorGraph } from "@/types/graph";

const branchingGraph: EditorGraph = {
  nodes: [
    { id: "start", type: "START" },
    { id: "talk", type: "DIALOGUE", data: { npc: "Aldeão" } },
    { id: "kill", type: "KILL" },
    { id: "collect", type: "COLLECT" },
    { id: "end", type: "END" },
  ],
  edges: [
    { id: "e1", source: "start", target: "talk" },
    { id: "e2", source: "talk", target: "kill" },
    { id: "e3", source: "talk", target: "collect" },
    { id: "e4", source: "kill", target: "end" },
    { id: "e5", source: "collect", target: "end" },
  ],
};

describe("toEditorGraph", () => {
  it("passes a valid graph through", () => {
    expect(toEditorGraph(branchingGraph)).toEqual(branchingGraph);
  });

  it.each([null, undefined, "x", 42, { nodes: "no", edges: null }])(
    "normalizes invalid input %s to an empty graph",
    (value) => {
      expect(toEditorGraph(value)).toEqual({ nodes: [], edges: [] });
    }
  );
});

describe("layoutGraph", () => {
  it("returns empty arrays for an empty graph", () => {
    expect(layoutGraph({ nodes: [], edges: [] })).toEqual({
      nodes: [],
      edges: [],
    });
  });

  it("keeps saved positions untouched", () => {
    const graph: EditorGraph = {
      nodes: [{ id: "start", type: "START", position: { x: 123, y: 456 } }],
      edges: [],
    };
    const { nodes } = layoutGraph(graph);
    expect(nodes[0].position).toEqual({ x: 123, y: 456 });
  });

  it("auto-positions nodes without position, with finite coordinates", () => {
    const { nodes, edges } = layoutGraph(branchingGraph);

    expect(nodes).toHaveLength(5);
    expect(edges).toHaveLength(5);
    for (const node of nodes) {
      expect(Number.isFinite(node.position.x)).toBe(true);
      expect(Number.isFinite(node.position.y)).toBe(true);
    }
  });

  it("gives siblings distinct positions", () => {
    const { nodes } = layoutGraph(branchingGraph);
    const kill = nodes.find((n) => n.id === "kill");
    const collect = nodes.find((n) => n.id === "collect");

    expect(kill?.position).not.toEqual(collect?.position);
  });

  it("mixes saved and auto positions", () => {
    const graph: EditorGraph = {
      nodes: [
        { id: "start", type: "START", position: { x: 10, y: 20 } },
        { id: "end", type: "END" },
      ],
      edges: [{ id: "e1", source: "start", target: "end" }],
    };
    const { nodes } = layoutGraph(graph);

    expect(nodes[0].position).toEqual({ x: 10, y: 20 });
    expect(nodes[1].position).not.toEqual({ x: 10, y: 20 });
    expect(Number.isFinite(nodes[1].position.x)).toBe(true);
  });

  it("maps type and data into the flow node payload", () => {
    const { nodes } = layoutGraph(branchingGraph);
    const talk = nodes.find((n) => n.id === "talk");

    expect(talk?.data).toEqual({
      nodeType: "DIALOGUE",
      payload: { npc: "Aldeão" },
    });
    expect(nodes.find((n) => n.id === "start")?.data.payload).toBeNull();
  });
});
