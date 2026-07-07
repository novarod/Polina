import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { WorkspaceList } from "@/components/workspaces/workspace-list";
import { getOrgRole } from "@/services/organizations-server";
import { serverFetch } from "@/services/server-api";
import type { Organization } from "@/types/organization";
import type { Workspace } from "@/types/workspace";

export const metadata: Metadata = {
  title: "Workspaces — Polina",
};

export default async function OrgPage({
  params,
}: {
  params: Promise<{ orgId: string }>;
}) {
  const { orgId } = await params;
  const [org, workspaces, role] = await Promise.all([
    serverFetch<Organization>(`/organizations/${orgId}`),
    serverFetch<Workspace[]>(`/organizations/${orgId}/workspaces`),
    getOrgRole(orgId),
  ]);
  if (!org || !workspaces || !role) {
    notFound();
  }

  return (
    <main className="grid content-start gap-6 p-6">
      <BreadcrumbNav
        items={[
          { label: "Organizações", href: "/orgs" },
          { label: org.name },
        ]}
      />
      <WorkspaceList orgId={orgId} role={role} workspaces={workspaces} />
    </main>
  );
}
