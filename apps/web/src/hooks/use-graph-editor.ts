import { useCallback, useEffect, useMemo, useState } from "react";
import {
  addEdge,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
} from "@xyflow/react";
import { toast } from "sonner";

import { layoutGraph, type QuestFlowNode } from "@/lib/graph-layout";
import { serializeGraph } from "@/lib/serialize-graph";
import { validateGraph } from "@/lib/validate-graph";
import { ApiError } from "@/services/api";
import { updateMissionGraph } from "@/services/missions";
import type { EditorGraph, EditorNodePosition } from "@/types/graph";

const NODE_WIDTH = 208;
const NODE_HEIGHT = 76;

interface UseGraphEditorArgs {
  orgId: string;
  workspaceId: string;
  missionId: string;
  initialGraph: EditorGraph;
}

function nextNodeId(type: string, nodes: QuestFlowNode[]): string {
  const prefix = type.toLowerCase();
  let max = 0;
  for (const node of nodes) {
    const match = new RegExp(`^${prefix}-(\\d+)$`).exec(node.id);
    if (match) {
      max = Math.max(max, Number(match[1]));
    }
  }
  return `${prefix}-${max + 1}`;
}

export function useGraphEditor({
  orgId,
  workspaceId,
  missionId,
  initialGraph,
}: UseGraphEditorArgs) {
  const initial = useMemo(() => layoutGraph(initialGraph), [initialGraph]);
  const [nodes, setNodes, onNodesChange] = useNodesState<QuestFlowNode>(
    initial.nodes
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>(initial.edges);
  const [baseline, setBaseline] = useState(() =>
    JSON.stringify(serializeGraph(initial.nodes, initial.edges))
  );
  const [saving, setSaving] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const serialized = useMemo(
    () => serializeGraph(nodes, edges),
    [nodes, edges]
  );
  const dirty = useMemo(
    () => JSON.stringify(serialized) !== baseline,
    [serialized, baseline]
  );
  const validation = useMemo(() => validateGraph(serialized), [serialized]);
  const hasStart = useMemo(
    () => nodes.some((node) => node.data.nodeType === "START"),
    [nodes]
  );

  useEffect(() => {
    if (!dirty) {
      return;
    }
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const addNode = useCallback(
    (type: string, position: EditorNodePosition) => {
      setNodes((current) => {
        const offset = (current.length % 4) * 28;
        return [
          ...current,
          {
            id: nextNodeId(type, current),
            type: "quest",
            position: {
              x: position.x - NODE_WIDTH / 2 + offset,
              y: position.y - NODE_HEIGHT / 2 + offset,
            },
            width: NODE_WIDTH,
            height: NODE_HEIGHT,
            data: { nodeType: type, payload: null },
          },
        ];
      });
    },
    [setNodes]
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      if (!connection.source || !connection.target) {
        return;
      }
      setEdges((current) => {
        const duplicate = current.some(
          (edge) =>
            edge.source === connection.source &&
            edge.target === connection.target
        );
        if (duplicate) {
          return current;
        }
        return addEdge(
          {
            ...connection,
            id: `e-${connection.source}-${connection.target}`,
          },
          current
        );
      });
    },
    [setEdges]
  );

  const deleteNode = useCallback(
    (nodeId: string) => {
      setNodes((current) => current.filter((node) => node.id !== nodeId));
      setEdges((current) =>
        current.filter(
          (edge) => edge.source !== nodeId && edge.target !== nodeId
        )
      );
    },
    [setNodes, setEdges]
  );

  const updateNodeData = useCallback(
    (nodeId: string, payload: unknown) => {
      setNodes((current) =>
        current.map((node) =>
          node.id === nodeId
            ? { ...node, data: { ...node.data, payload } }
            : node
        )
      );
    },
    [setNodes]
  );

  const save = useCallback(async () => {
    setSaving(true);
    setApiError(null);
    try {
      await updateMissionGraph(orgId, workspaceId, missionId, serialized);
      setBaseline(JSON.stringify(serialized));
      toast.success("Grafo salvo");
    } catch (error) {
      setApiError(
        error instanceof ApiError
          ? error.message
          : "Algo deu errado ao salvar, tente novamente"
      );
    } finally {
      setSaving(false);
    }
  }, [orgId, workspaceId, missionId, serialized]);

  const dismissApiError = useCallback(() => setApiError(null), []);

  return {
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    onConnect,
    addNode,
    deleteNode,
    updateNodeData,
    dirty,
    saving,
    apiError,
    dismissApiError,
    validation,
    hasStart,
    save,
  };
}
