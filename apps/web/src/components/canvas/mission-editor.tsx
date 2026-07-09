"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  type NodeMouseHandler,
} from "@xyflow/react";
import { XIcon } from "lucide-react";

import { NodePalette } from "@/components/canvas/node-palette";
import { NodePanel } from "@/components/canvas/node-panel";
import { QuestNode } from "@/components/canvas/quest-node";
import { SaveBar } from "@/components/canvas/save-bar";
import { Button } from "@/components/ui/button";
import { useGraphEditor } from "@/hooks/use-graph-editor";
import { useMounted } from "@/hooks/use-mounted";
import type { QuestFlowNode } from "@/lib/graph-layout";
import type { EditorGraph } from "@/types/graph";

const nodeTypes = { quest: QuestNode };

export interface MissionEditorProps {
  graph: EditorGraph;
  orgId: string;
  workspaceId: string;
  missionId: string;
  onDirtyChange?: (dirty: boolean) => void;
}

export function MissionEditor(props: MissionEditorProps) {
  const mounted = useMounted();
  if (!mounted) {
    return (
      <div className="h-[60vh] animate-pulse rounded-sm border-2 border-foreground/40 bg-card/50" />
    );
  }
  return (
    <ReactFlowProvider>
      <EditorCanvas {...props} />
    </ReactFlowProvider>
  );
}

function EditorCanvas({
  graph,
  orgId,
  workspaceId,
  missionId,
  onDirtyChange,
}: MissionEditorProps) {
  const editor = useGraphEditor({
    orgId,
    workspaceId,
    missionId,
    initialGraph: graph,
  });
  const [selectedId, setSelectedId] = useState<string | null>(null);

  useEffect(() => {
    onDirtyChange?.(editor.dirty);
  }, [editor.dirty, onDirtyChange]);

  const nodesWithErrors = useMemo(
    () =>
      editor.nodes.map((node) => {
        const errors = editor.validation.nodeErrors.get(node.id);
        return errors
          ? { ...node, data: { ...node.data, errors } }
          : node;
      }),
    [editor.nodes, editor.validation]
  );

  const selectedNode = useMemo(() => {
    if (!selectedId) {
      return null;
    }
    const node = editor.nodes.find((n) => n.id === selectedId);
    if (!node) {
      return null;
    }
    return {
      id: node.id,
      type: node.data.nodeType,
      payload: node.data.payload,
    };
  }, [selectedId, editor.nodes]);

  const onNodeClick: NodeMouseHandler<QuestFlowNode> = useCallback(
    (_event, node) => setSelectedId(node.id),
    []
  );
  const onPaneClick = useCallback(() => setSelectedId(null), []);
  const onDeleteSelected = useCallback(() => {
    if (selectedId) {
      editor.deleteNode(selectedId);
      setSelectedId(null);
    }
  }, [selectedId, editor]);

  return (
    <div
      data-testid="mission-editor"
      className="relative h-[60vh] overflow-hidden rounded-sm border-2 border-foreground/70"
    >
      <ReactFlow
        nodes={nodesWithErrors}
        edges={editor.edges}
        nodeTypes={nodeTypes}
        fitView
        minZoom={0.2}
        nodesConnectable
        deleteKeyCode={["Backspace", "Delete"]}
        onNodesChange={editor.onNodesChange}
        onEdgesChange={editor.onEdgesChange}
        onConnect={editor.onConnect}
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
        <Panel position="top-left">
          <NodePalette hasStart={editor.hasStart} onAdd={editor.addNode} />
        </Panel>
        <Panel position="bottom-center">
          <SaveBar
            dirty={editor.dirty}
            saving={editor.saving}
            errorCount={editor.validation.errorCount}
            onSave={editor.save}
          />
        </Panel>
      </ReactFlow>
      {editor.apiError && (
        <div
          data-testid="graph-api-error"
          role="alert"
          className="absolute top-3 left-1/2 z-20 flex w-max max-w-[80%] -translate-x-1/2 items-start gap-2 rounded-sm border-2 border-destructive bg-card px-3 py-2 text-sm text-destructive shadow-[4px_4px_0_0] shadow-destructive/40"
        >
          <span className="min-w-0">{editor.apiError}</span>
          <Button
            variant="ghost"
            size="icon-xs"
            aria-label="Fechar erro"
            onClick={editor.dismissApiError}
          >
            <XIcon />
          </Button>
        </div>
      )}
      <NodePanel
        node={selectedNode}
        onClose={onPaneClick}
        onApplyData={(payload) =>
          selectedId && editor.updateNodeData(selectedId, payload)
        }
        onDelete={onDeleteSelected}
      />
    </div>
  );
}
