"use client";

import { useReactFlow, useStore } from "@xyflow/react";

import { NodeSprite } from "@/components/canvas/node-sprite";
import { Button } from "@/components/ui/button";
import type { EditorNodePosition } from "@/types/graph";

const paletteTypes = [
  "START",
  "END",
  "DIALOGUE",
  "KILL",
  "COLLECT",
  "OBJECTIVE",
];

interface NodePaletteProps {
  hasStart: boolean;
  onAdd: (type: string, position: EditorNodePosition) => void;
}

export function NodePalette({ hasStart, onAdd }: NodePaletteProps) {
  const { screenToFlowPosition } = useReactFlow();
  const domNode = useStore((state) => state.domNode);

  function centerPosition(): EditorNodePosition {
    const rect = domNode?.getBoundingClientRect();
    if (!rect) {
      return { x: 0, y: 0 };
    }
    return screenToFlowPosition({
      x: rect.x + rect.width / 2,
      y: rect.y + rect.height / 2,
    });
  }

  return (
    <div
      data-testid="node-palette"
      className="flex flex-col gap-1 rounded-sm border-2 border-foreground/70 bg-card p-2 shadow-[4px_4px_0_0] shadow-foreground/25"
    >
      <p className="font-display text-[9px] text-muted-foreground">Nós</p>
      {paletteTypes.map((type) => {
        const startBlocked = type === "START" && hasStart;
        return (
          <Button
            key={type}
            variant="ghost"
            size="xs"
            className="justify-start text-primary"
            disabled={startBlocked}
            title={startBlocked ? "O grafo já tem um nó START" : undefined}
            data-testid={`palette-${type}`}
            onClick={() => onAdd(type, centerPosition())}
          >
            <NodeSprite type={type} />
            <span className="text-foreground">{type}</span>
          </Button>
        );
      })}
    </div>
  );
}
