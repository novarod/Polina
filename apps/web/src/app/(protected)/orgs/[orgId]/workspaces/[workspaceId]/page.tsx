import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { MissionList } from "@/components/missions/mission-list";
import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { getOrgRole } from "@/services/organizations-server";
import { serverFetch } from "@/services/server-api";
import type { Mission } from "@/types/mission";
import type { Organization } from "@/types/organization";
import type { Workspace } from "@/types/workspace";

export const metadata: Metadata = {
  title: "Missões — Polina",
};

export default async function WorkspacePage({
  params,
}: {
  params: Promise<{ orgId: string; workspaceId: string }>;
}) {
  const { orgId, workspaceId } = await params;
  const [org, workspace, missions, role] = await Promise.all([
    serverFetch<Organization>(`/organizations/${orgId}`),
    serverFetch<Workspace>(
      `/organizations/${orgId}/workspaces/${workspaceId}`
    ),
    serverFetch<Mission[]>(
      `/organizations/${orgId}/workspaces/${workspaceId}/missions`
    ),
    getOrgRole(orgId),
  ]);
  if (!org || !workspace || !missions || !role) {
    notFound();
  }

  return (
    <main className="grid content-start gap-6 p-6">
      <BreadcrumbNav
        items={[
          { label: "Organizações", href: "/orgs" },
          { label: org.name, href: `/orgs/${orgId}` },
          { label: workspace.name },
        ]}
      />
      <MissionList
        orgId={orgId}
        workspaceId={workspaceId}
        role={role}
        missions={missions}
      />
    </main>
  );
}
