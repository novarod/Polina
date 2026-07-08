"use client";

import { useState } from "react";
import { XIcon } from "lucide-react";

import { NodeSprite } from "@/components/canvas/node-sprite";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

export interface SelectedNode {
  id: string;
  type: string;
  payload: unknown;
}

interface NodePanelProps {
  node: SelectedNode | null;
  onClose: () => void;
  onApplyData?: (payload: unknown) => void;
  onDelete?: () => void;
}

function toDraft(payload: unknown): string {
  return payload == null ? "" : JSON.stringify(payload, null, 2);
}

function DataEditor({
  node,
  onApplyData,
}: {
  node: SelectedNode;
  onApplyData: (payload: unknown) => void;
}) {
  const [draft, setDraft] = useState(() => toDraft(node.payload));
  const [syntaxError, setSyntaxError] = useState<string | null>(null);

  function apply() {
    const trimmed = draft.trim();
    if (trimmed === "") {
      setSyntaxError(null);
      onApplyData(null);
      return;
    }
    try {
      const parsed: unknown = JSON.parse(trimmed);
      setSyntaxError(null);
      onApplyData(parsed);
    } catch {
      setSyntaxError("JSON inválido — nada foi aplicado");
    }
  }

  function cancel() {
    setDraft(toDraft(node.payload));
    setSyntaxError(null);
  }

  return (
    <div className="mt-1 grid gap-2">
      <Textarea
        data-testid="node-data-editor"
        aria-label="Dados do nó (JSON)"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        className="max-h-64 min-h-24 font-mono text-xs"
      />
      {syntaxError && (
        <p
          role="alert"
          data-testid="node-data-error"
          className="text-sm font-medium text-destructive"
        >
          {syntaxError}
        </p>
      )}
      <div className="flex gap-2">
        <Button size="xs" onClick={apply}>
          Aplicar
        </Button>
        <Button size="xs" variant="ghost" onClick={cancel}>
          Cancelar
        </Button>
      </div>
    </div>
  );
}

export function NodePanel({
  node,
  onClose,
  onApplyData,
  onDelete,
}: NodePanelProps) {
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
      {onApplyData ? (
        <DataEditor key={node.id} node={node} onApplyData={onApplyData} />
      ) : node.payload == null ? (
        <p className="mt-1 text-sm text-muted-foreground">Sem dados</p>
      ) : (
        <pre
          data-testid="node-panel-data"
          className="mt-1 max-h-64 overflow-auto rounded-sm bg-muted p-2 text-xs"
        >
          {JSON.stringify(node.payload, null, 2)}
        </pre>
      )}
      {onDelete && (
        <Button
          variant="destructive"
          size="xs"
          className="mt-4"
          data-testid="delete-node"
          onClick={onDelete}
        >
          Deletar nó
        </Button>
      )}
    </aside>
  );
}
