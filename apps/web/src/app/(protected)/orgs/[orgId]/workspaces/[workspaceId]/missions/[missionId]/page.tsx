import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { MissionView } from "@/components/missions/mission-view";
import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { roleAtLeast } from "@/lib/roles";
import { getOrgRole } from "@/services/organizations-server";
import { serverFetch } from "@/services/server-api";
import type { Mission, MissionVersion } from "@/types/mission";
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
  const [org, workspace, mission, versions, role] = await Promise.all([
    serverFetch<Organization>(`/organizations/${orgId}`),
    serverFetch<Workspace>(basePath),
    serverFetch<Mission>(`${basePath}/missions/${missionId}`),
    serverFetch<MissionVersion[]>(
      `${basePath}/missions/${missionId}/versions`
    ),
    getOrgRole(orgId),
  ]);
  if (!org || !workspace || !mission || !versions || !role) {
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
      <MissionView
        mission={mission}
        versions={versions}
        canEdit={roleAtLeast(role, "DESIGNER")}
        orgId={orgId}
        workspaceId={workspaceId}
      />
    </main>
  );
}
