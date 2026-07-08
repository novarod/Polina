import { describe, expect, it } from "vitest";

import { MAX_NODES, validateGraph } from "@/lib/validate-graph";
import type { EditorGraph } from "@/types/graph";

function graph(partial: Partial<EditorGraph>): EditorGraph {
  return { nodes: [], edges: [], ...partial };
}

const validLinear = graph({
  nodes: [
    { id: "a", type: "START" },
    { id: "b", type: "OBJECTIVE" },
    { id: "c", type: "END" },
  ],
  edges: [
    { id: "e1", source: "a", target: "b" },
    { id: "e2", source: "b", target: "c" },
  ],
});

describe("validateGraph", () => {
  it("accepts a valid linear graph", () => {
    expect(validateGraph(validLinear).errorCount).toBe(0);
  });

  it("accepts a valid branching graph", () => {
    const branching = graph({
      nodes: [
        { id: "s", type: "START" },
        { id: "d", type: "DIALOGUE" },
        { id: "k", type: "KILL" },
        { id: "c", type: "COLLECT" },
        { id: "e", type: "END" },
      ],
      edges: [
        { id: "e1", source: "s", target: "d" },
        { id: "e2", source: "d", target: "k" },
        { id: "e3", source: "d", target: "c" },
        { id: "e4", source: "k", target: "e" },
        { id: "e5", source: "c", target: "e" },
      ],
    });
    expect(validateGraph(branching).errorCount).toBe(0);
  });

  it("returns no errors for an empty graph", () => {
    expect(validateGraph(graph({})).errorCount).toBe(0);
  });

  it("flags a dead-end node", () => {
    const g = graph({
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
        { id: "c", type: "END" },
      ],
      edges: [{ id: "e1", source: "a", target: "b" }],
    });
    const result = validateGraph(g);
    expect(result.nodeErrors.get("b")).toContain(
      "Sem aresta de saída (beco sem saída)"
    );
  });

  it("flags a cycle on the involved node", () => {
    const g = graph({
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
        { id: "c", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "b" },
        { id: "e3", source: "b", target: "c" },
      ],
    });
    const result = validateGraph(g);
    expect(result.nodeErrors.get("b")).toContain(
      "Ciclo detectado envolvendo este nó"
    );
  });

  it("flags every extra START node", () => {
    const g = graph({
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "START" },
        { id: "c", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "c" },
        { id: "e2", source: "b", target: "c" },
      ],
    });
    const result = validateGraph(g);
    expect(result.nodeErrors.get("a")).toContain("Só pode existir um nó START");
    expect(result.nodeErrors.get("b")).toContain("Só pode existir um nó START");
  });

  it("flags a missing START as a graph error", () => {
    const g = graph({
      nodes: [{ id: "c", type: "END" }],
      edges: [],
    });
    expect(validateGraph(g).graphErrors).toContain(
      "O grafo precisa de exatamente um nó START"
    );
  });

  it("flags a missing END as a graph error", () => {
    const g = graph({
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
      ],
      edges: [{ id: "e1", source: "a", target: "b" }],
    });
    expect(validateGraph(g).graphErrors).toContain(
      "O grafo precisa de pelo menos um nó END"
    );
  });

  it("flags unreachable nodes", () => {
    const g = graph({
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "END" },
        { id: "orphan", type: "OBJECTIVE" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "orphan", target: "b" },
      ],
    });
    const result = validateGraph(g);
    expect(result.nodeErrors.get("orphan")).toContain(
      "Não é alcançável a partir do nó START"
    );
  });

  it("flags edges pointing to unknown nodes and stops there", () => {
    const g = graph({
      nodes: [{ id: "a", type: "START" }],
      edges: [{ id: "e1", source: "a", target: "ghost" }],
    });
    const result = validateGraph(g);
    expect(result.graphErrors).toEqual([
      'A aresta "e1" referencia o nó de destino inexistente "ghost"',
    ]);
    expect(result.nodeErrors.size).toBe(0);
  });

  it("flags the node limit and stops there", () => {
    const nodes = Array.from({ length: MAX_NODES + 1 }, (_, i) => ({
      id: `n${i}`,
      type: "OBJECTIVE",
    }));
    const result = validateGraph(graph({ nodes }));
    expect(result.graphErrors).toEqual([
      `O grafo excede o máximo de ${MAX_NODES} nós (tem ${MAX_NODES + 1})`,
    ]);
    expect(result.nodeErrors.size).toBe(0);
  });
});
