"use client";

import { useCallback, useMemo, useState, useSyncExternalStore } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  ReactFlow,
  type NodeMouseHandler,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { NodePanel, type SelectedNode } from "@/components/canvas/node-panel";
import { QuestNode } from "@/components/canvas/quest-node";
import { layoutGraph, type QuestFlowNode } from "@/lib/graph-layout";
import type { EditorGraph } from "@/types/graph";

const nodeTypes = { quest: QuestNode };

const emptySubscribe = () => () => {};

function useMounted(): boolean {
  return useSyncExternalStore(
    emptySubscribe,
    () => true,
    () => false
  );
}

export function MissionCanvas({ graph }: { graph: EditorGraph }) {
  const { nodes, edges } = useMemo(() => layoutGraph(graph), [graph]);
  const [selected, setSelected] = useState<SelectedNode | null>(null);
  const mounted = useMounted();

  const onNodeClick: NodeMouseHandler<QuestFlowNode> = useCallback(
    (_event, node) => {
      setSelected({
        id: node.id,
        type: node.data.nodeType,
        payload: node.data.payload,
      });
    },
    []
  );
  const onPaneClick = useCallback(() => setSelected(null), []);

  if (graph.nodes.length === 0) {
    return (
      <div
        data-testid="canvas-empty"
        className="flex h-[60vh] items-center justify-center rounded-sm border-2 border-dashed border-foreground/40 bg-card/50"
      >
        <p className="text-muted-foreground">
          Grafo vazio — a edição de missões chega no próximo módulo.
        </p>
      </div>
    );
  }

  if (!mounted) {
    return (
      <div className="h-[60vh] animate-pulse rounded-sm border-2 border-foreground/40 bg-card/50" />
    );
  }

  return (
    <div
      data-testid="mission-canvas"
      className="relative h-[60vh] overflow-hidden rounded-sm border-2 border-foreground/70"
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        nodeTypes={nodeTypes}
        fitView
        minZoom={0.2}
        nodesDraggable={false}
        nodesConnectable={false}
        edgesFocusable={false}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={16}
          color="var(--border)"
        />
        <MiniMap
          pannable
          bgColor="var(--card)"
          maskColor="color-mix(in oklch, var(--background) 60%, transparent)"
        />
        <Controls showInteractive={false} />
      </ReactFlow>
      <NodePanel node={selected} onClose={onPaneClick} />
    </div>
  );
}
