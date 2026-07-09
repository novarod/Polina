"use client";

import { useState } from "react";
import { CopyIcon } from "lucide-react";
import { toast } from "sonner";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/services/api";
import { getMissionVersion } from "@/services/missions";
import type { MissionVersion } from "@/types/mission";

interface VersionListProps {
  orgId: string;
  workspaceId: string;
  missionId: string;
  versions: MissionVersion[];
  activeHash: string | null;
}

type VersionDetailState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "loaded"; data: unknown };

export function VersionList({
  orgId,
  workspaceId,
  missionId,
  versions,
  activeHash,
}: VersionListProps) {
  const [details, setDetails] = useState<Record<string, VersionDetailState>>(
    {}
  );

  async function loadDetail(hash: string) {
    setDetails((current) => ({ ...current, [hash]: { status: "loading" } }));
    try {
      const detail = await getMissionVersion(
        orgId,
        workspaceId,
        missionId,
        hash
      );
      setDetails((current) => ({
        ...current,
        [hash]: { status: "loaded", data: detail.mission_data },
      }));
    } catch (error) {
      setDetails((current) => ({
        ...current,
        [hash]: {
          status: "error",
          message:
            error instanceof ApiError
              ? error.message
              : "Não foi possível carregar o contrato",
        },
      }));
    }
  }

  async function copyHash(hash: string) {
    try {
      await navigator.clipboard.writeText(hash);
      toast.success("Hash copiado");
    } catch {
      toast.error("Não foi possível copiar");
    }
  }

  return (
    <section data-testid="version-list" className="grid gap-3">
      <h2 className="font-display text-xs">Versões</h2>
      {versions.length === 0 ? (
        <p className="text-muted-foreground">
          Nenhuma versão publicada ainda.
        </p>
      ) : (
        <Accordion
          type="single"
          collapsible
          onValueChange={(value) => {
            if (value && !details[value]) {
              loadDetail(value);
            }
          }}
        >
          {versions.map((version) => {
            const detail = details[version.hash];
            return (
              <AccordionItem
                key={version.id}
                value={version.hash}
                data-testid="version-item"
              >
                <div className="flex items-center gap-2">
                  <AccordionTrigger className="flex-1">
                    <span className="flex items-center gap-2">
                      v{version.version_number}
                      <span className="text-xs text-muted-foreground">
                        {version.created_at.slice(0, 10)}
                      </span>
                      {version.hash === activeHash && (
                        <span
                          data-testid="version-active-badge"
                          className="rounded-sm bg-primary/15 px-2 py-0.5 text-xs font-medium text-primary"
                        >
                          Ativa
                        </span>
                      )}
                    </span>
                  </AccordionTrigger>
                  <code className="text-xs text-muted-foreground">
                    {version.hash.slice(0, 10)}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    aria-label={`Copiar hash da v${version.version_number}`}
                    onClick={() => copyHash(version.hash)}
                  >
                    <CopyIcon />
                  </Button>
                </div>
                <AccordionContent>
                  {!detail || detail.status === "loading" ? (
                    <p className="text-sm text-muted-foreground">
                      Carregando contrato...
                    </p>
                  ) : detail.status === "error" ? (
                    <p className="text-sm font-medium text-destructive">
                      {detail.message}
                    </p>
                  ) : (
                    <pre
                      data-testid="version-data"
                      className="max-h-72 overflow-auto rounded-sm bg-muted p-2 text-xs"
                    >
                      {JSON.stringify(detail.data, null, 2)}
                    </pre>
                  )}
                </AccordionContent>
              </AccordionItem>
            );
          })}
        </Accordion>
      )}
    </section>
  );
}
