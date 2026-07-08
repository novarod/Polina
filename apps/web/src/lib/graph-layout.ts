import dagre from "@dagrejs/dagre";
import type { Edge, Node } from "@xyflow/react";

import type { EditorGraph, EditorNodePosition } from "@/types/graph";

const NODE_WIDTH = 208;
const NODE_HEIGHT = 76;

export interface QuestNodeData extends Record<string, unknown> {
  nodeType: string;
  payload: unknown;
}

export type QuestFlowNode = Node<QuestNodeData, "quest">;

function autoLayout(graph: EditorGraph): Map<string, EditorNodePosition> {
  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: "TB", nodesep: 60, ranksep: 90 });
  g.setDefaultEdgeLabel(() => ({}));
  for (const node of graph.nodes) {
    g.setNode(node.id, { width: NODE_WIDTH, height: NODE_HEIGHT });
  }
  for (const edge of graph.edges) {
    g.setEdge(edge.source, edge.target);
  }
  dagre.layout(g);

  const positions = new Map<string, EditorNodePosition>();
  for (const node of graph.nodes) {
    const placed = g.node(node.id);
    const x = placed ? placed.x - NODE_WIDTH / 2 : 0;
    const y = placed ? placed.y - NODE_HEIGHT / 2 : 0;
    positions.set(node.id, {
      x: Number.isFinite(x) ? x : 0,
      y: Number.isFinite(y) ? y : 0,
    });
  }
  return positions;
}

export function toEditorGraph(value: unknown): EditorGraph {
  if (typeof value === "object" && value !== null) {
    const candidate = value as Partial<EditorGraph>;
    return {
      nodes: Array.isArray(candidate.nodes) ? candidate.nodes : [],
      edges: Array.isArray(candidate.edges) ? candidate.edges : [],
    };
  }
  return { nodes: [], edges: [] };
}

export function layoutGraph(graph: EditorGraph): {
  nodes: QuestFlowNode[];
  edges: Edge[];
} {
  const hasUnpositioned = graph.nodes.some((node) => !node.position);
  const autoPositions = hasUnpositioned
    ? autoLayout(graph)
    : new Map<string, EditorNodePosition>();

  const nodes = graph.nodes.map<QuestFlowNode>((node) => ({
    id: node.id,
    type: "quest",
    position: node.position ??
      autoPositions.get(node.id) ?? { x: 0, y: 0 },
    width: NODE_WIDTH,
    height: NODE_HEIGHT,
    data: { nodeType: node.type, payload: node.data ?? null },
  }));
  const edges = graph.edges.map<Edge>((edge) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
  }));
  return { nodes, edges };
}
