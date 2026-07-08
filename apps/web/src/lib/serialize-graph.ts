import type { Edge } from "@xyflow/react";

import type { QuestFlowNode } from "@/lib/graph-layout";
import type { EditorGraph, EditorNode } from "@/types/graph";

export function serializeGraph(
  nodes: QuestFlowNode[],
  edges: Edge[]
): EditorGraph {
  return {
    nodes: nodes.map((node) => {
      const serialized: EditorNode = {
        id: node.id,
        type: node.data.nodeType,
        position: { x: node.position.x, y: node.position.y },
      };
      if (node.data.payload != null) {
        serialized.data = node.data.payload;
      }
      return serialized;
    }),
    edges: edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
    })),
  };
}
