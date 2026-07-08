"use client";

import { Button } from "@/components/ui/button";

interface SaveBarProps {
  dirty: boolean;
  saving: boolean;
  errorCount: number;
  onSave: () => void;
}

export function SaveBar({ dirty, saving, errorCount, onSave }: SaveBarProps) {
  return (
    <div
      data-testid="save-bar"
      className="flex items-center gap-3 rounded-sm border-2 border-foreground/70 bg-card px-3 py-2 shadow-[4px_4px_0_0] shadow-foreground/25"
    >
      {errorCount > 0 && (
        <p
          data-testid="graph-error-count"
          className="text-sm font-medium text-destructive"
        >
          {errorCount === 1
            ? "1 problema no grafo"
            : `${errorCount} problemas no grafo`}
        </p>
      )}
      {dirty && (
        <p
          data-testid="dirty-indicator"
          className="text-sm text-muted-foreground"
        >
          Mudanças não salvas
        </p>
      )}
      <Button size="sm" onClick={onSave} disabled={saving}>
        {saving ? "Salvando..." : "Salvar"}
      </Button>
    </div>
  );
}
