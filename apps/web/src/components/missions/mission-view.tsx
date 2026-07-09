"use client";

import { useMemo, useState } from "react";

import { MissionCanvas } from "@/components/canvas/mission-canvas";
import { MissionStatusBadge } from "@/components/missions/mission-status-badge";
import { PublishButton } from "@/components/missions/publish-button";
import { VersionList } from "@/components/missions/version-list";
import { toEditorGraph } from "@/lib/graph-layout";
import type { Mission, MissionVersion } from "@/types/mission";

interface MissionViewProps {
  mission: Mission;
  versions: MissionVersion[];
  canEdit: boolean;
  orgId: string;
  workspaceId: string;
}

export function MissionView({
  mission,
  versions,
  canEdit,
  orgId,
  workspaceId,
}: MissionViewProps) {
  const [dirty, setDirty] = useState(false);
  const graph = useMemo(() => toEditorGraph(mission.graph), [mission.graph]);

  return (
    <>
      <div className="grid gap-2">
        <div className="flex items-center gap-3">
          <h1 className="font-display text-sm">{mission.name}</h1>
          <MissionStatusBadge status={mission.status} />
          {canEdit && (
            <div className="ml-auto">
              <PublishButton
                orgId={orgId}
                workspaceId={workspaceId}
                missionId={mission.id}
                activeHash={mission.active_hash}
                dirty={dirty}
              />
            </div>
          )}
        </div>
        {mission.description && (
          <p className="text-muted-foreground">{mission.description}</p>
        )}
        {mission.active_hash && (
          <p className="text-xs text-muted-foreground">
            Versão ativa: <code>{mission.active_hash.slice(0, 10)}</code>
          </p>
        )}
      </div>
      {canEdit ? (
        <MissionCanvas
          editable
          graph={graph}
          orgId={orgId}
          workspaceId={workspaceId}
          missionId={mission.id}
          onDirtyChange={setDirty}
        />
      ) : (
        <MissionCanvas graph={graph} />
      )}
      <VersionList
        orgId={orgId}
        workspaceId={workspaceId}
        missionId={mission.id}
        versions={versions}
        activeHash={mission.active_hash}
      />
    </>
  );
}
