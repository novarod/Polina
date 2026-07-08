import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { MissionCanvas } from "@/components/canvas/mission-canvas";
import { MissionStatusBadge } from "@/components/missions/mission-status-badge";
import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { toEditorGraph } from "@/lib/graph-layout";
import { getOrgRole } from "@/services/organizations-server";
import { serverFetch } from "@/services/server-api";
import type { Mission } from "@/types/mission";
import type { Organization } from "@/types/organization";
import type { Workspace } from "@/types/workspace";

export const metadata: Metadata = {
  title: "Missão — Polina",
};

export default async function MissionPage({
  params,
}: {
  params: Promise<{ orgId: string; workspaceId: string; missionId: string }>;
}) {
  const { orgId, workspaceId, missionId } = await params;
  const basePath = `/organizations/${orgId}/workspaces/${workspaceId}`;
  const [org, workspace, mission, role] = await Promise.all([
    serverFetch<Organization>(`/organizations/${orgId}`),
    serverFetch<Workspace>(basePath),
    serverFetch<Mission>(`${basePath}/missions/${missionId}`),
    getOrgRole(orgId),
  ]);
  if (!org || !workspace || !mission || !role) {
    notFound();
  }

  return (
    <main className="grid content-start gap-6 p-6">
      <BreadcrumbNav
        items={[
          { label: "Organizações", href: "/orgs" },
          { label: org.name, href: `/orgs/${orgId}` },
          {
            label: workspace.name,
            href: `/orgs/${orgId}/workspaces/${workspaceId}`,
          },
          { label: mission.name },
        ]}
      />
      <div className="grid gap-2">
        <div className="flex items-center gap-3">
          <h1 className="font-display text-sm">{mission.name}</h1>
          <MissionStatusBadge status={mission.status} />
        </div>
        {mission.description && (
          <p className="text-muted-foreground">{mission.description}</p>
        )}
        {mission.active_hash && (
          <p className="text-xs text-muted-foreground">
            Versão ativa: <code>{mission.active_hash}</code>
          </p>
        )}
      </div>
      <MissionCanvas graph={toEditorGraph(mission.graph)} />
    </main>
  );
}
