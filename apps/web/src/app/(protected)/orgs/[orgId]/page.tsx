import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { Button } from "@/components/ui/button";
import { WorkspaceList } from "@/components/workspaces/workspace-list";
import { roleAtLeast } from "@/lib/roles";
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
      <div className="flex items-center justify-between gap-2">
        <BreadcrumbNav
          items={[
            { label: "Organizações", href: "/orgs" },
            { label: org.name },
          ]}
        />
        {roleAtLeast(role, "ADMIN") && (
          <Button asChild variant="outline" size="xs" data-testid="api-keys-link">
            <Link href={`/orgs/${orgId}/api-keys`}>API keys</Link>
          </Button>
        )}
      </div>
      <WorkspaceList orgId={orgId} role={role} workspaces={workspaces} />
    </main>
  );
}
