import { describe, expect, it } from "vitest";

import { layoutGraph } from "@/lib/graph-layout";
import { serializeGraph } from "@/lib/serialize-graph";
import type { EditorGraph } from "@/types/graph";

describe("serializeGraph", () => {
  it("round-trips a graph keeping positions and omitting dimensions", () => {
    const source: EditorGraph = {
      nodes: [
        { id: "a", type: "START", position: { x: 10, y: 20 } },
        {
          id: "b",
          type: "DIALOGUE",
          data: { npc: "Aldeão" },
          position: { x: 30, y: 40 },
        },
      ],
      edges: [{ id: "e1", source: "a", target: "b" }],
    };
    const { nodes, edges } = layoutGraph(source);

    const serialized = serializeGraph(nodes, edges);

    expect(serialized).toEqual({
      nodes: [
        { id: "a", type: "START", position: { x: 10, y: 20 } },
        {
          id: "b",
          type: "DIALOGUE",
          data: { npc: "Aldeão" },
          position: { x: 30, y: 40 },
        },
      ],
      edges: [{ id: "e1", source: "a", target: "b" }],
    });
    expect(serialized.nodes[0]).not.toHaveProperty("width");
    expect(serialized.nodes[0]).not.toHaveProperty("data");
  });
});
