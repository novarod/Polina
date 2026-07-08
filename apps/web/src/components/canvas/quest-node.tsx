"use client";

import { Handle, Position, type NodeProps } from "@xyflow/react";

import { NodeSprite } from "@/components/canvas/node-sprite";
import type { QuestFlowNode } from "@/lib/graph-layout";
import { cn } from "@/lib/utils";

export function QuestNode({ id, data, selected }: NodeProps<QuestFlowNode>) {
  return (
    <div
      data-testid="quest-node"
      data-node-type={data.nodeType}
      className={cn(
        "flex h-full w-full items-center rounded-sm border-2 border-foreground/70 bg-card px-3 py-2 text-card-foreground shadow-[4px_4px_0_0] shadow-foreground/25",
        selected && "border-primary shadow-primary/40"
      )}
    >
      <div className="flex min-w-0 items-center gap-2 text-primary">
        <NodeSprite type={data.nodeType} />
        <div className="min-w-0">
          <p className="font-display text-[10px] leading-4">{data.nodeType}</p>
          <p className="truncate text-sm font-medium text-card-foreground">
            {id}
          </p>
        </div>
      </div>
      {data.nodeType !== "START" && (
        <Handle type="target" position={Position.Top} />
      )}
      {data.nodeType !== "END" && (
        <Handle type="source" position={Position.Bottom} />
      )}
    </div>
  );
}
