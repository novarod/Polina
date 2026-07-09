import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { CreateKeyDialog } from "@/components/api-keys/create-key-dialog";
import { KeyList } from "@/components/api-keys/key-list";
import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { serverFetch } from "@/services/server-api";
import type { ApiKey } from "@/types/api-key";
import type { Organization } from "@/types/organization";

export const metadata: Metadata = {
  title: "API keys — Polina",
};

export default async function ApiKeysPage({
  params,
}: {
  params: Promise<{ orgId: string }>;
}) {
  const { orgId } = await params;
  const [org, keys] = await Promise.all([
    serverFetch<Organization>(`/organizations/${orgId}`),
    serverFetch<ApiKey[]>(`/organizations/${orgId}/api-keys`),
  ]);
  if (!org || !keys) {
    notFound();
  }

  return (
    <main className="grid content-start gap-6 p-6">
      <BreadcrumbNav
        items={[
          { label: "Organizações", href: "/orgs" },
          { label: org.name, href: `/orgs/${orgId}` },
          { label: "API keys" },
        ]}
      />
      <div className="flex items-center justify-between">
        <h1 className="font-display text-sm">API keys</h1>
        <CreateKeyDialog orgId={orgId} />
      </div>
      <KeyList orgId={orgId} keys={keys} />
    </main>
  );
}
