"use client";

import { XIcon } from "lucide-react";

import { NodeSprite } from "@/components/canvas/node-sprite";
import { Button } from "@/components/ui/button";

export interface SelectedNode {
  id: string;
  type: string;
  payload: unknown;
}

interface NodePanelProps {
  node: SelectedNode | null;
  onClose: () => void;
}

export function NodePanel({ node, onClose }: NodePanelProps) {
  if (!node) {
    return null;
  }
  return (
    <aside
      data-testid="node-panel"
      className="absolute top-3 right-3 z-10 w-72 rounded-sm border-2 border-foreground/70 bg-card p-4 text-card-foreground shadow-[4px_4px_0_0] shadow-foreground/25"
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2 text-primary">
          <NodeSprite type={node.type} />
          <div className="min-w-0">
            <p className="font-display text-[10px] leading-4">{node.type}</p>
            <p className="truncate text-sm font-medium text-card-foreground">
              {node.id}
            </p>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label="Fechar painel"
          onClick={onClose}
        >
          <XIcon />
        </Button>
      </div>
      <p className="mt-4 font-display text-[9px] text-muted-foreground">
        Dados
      </p>
      {node.payload == null ? (
        <p className="mt-1 text-sm text-muted-foreground">Sem dados</p>
      ) : (
        <pre
          data-testid="node-panel-data"
          className="mt-1 max-h-64 overflow-auto rounded-sm bg-muted p-2 text-xs"
        >
          {JSON.stringify(node.payload, null, 2)}
        </pre>
      )}
    </aside>
  );
}
