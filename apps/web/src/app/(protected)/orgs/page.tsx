import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { BreadcrumbNav } from "@/components/nav/breadcrumb-nav";
import { OrgList } from "@/components/orgs/org-list";
import { getMyOrganizations } from "@/services/organizations-server";

export const metadata: Metadata = {
  title: "Organizações — Polina",
};

export default async function OrgsPage() {
  const orgs = await getMyOrganizations();
  if (!orgs) {
    notFound();
  }

  return (
    <main className="grid content-start gap-6 p-6">
      <BreadcrumbNav items={[{ label: "Organizações" }]} />
      <OrgList orgs={orgs} />
    </main>
  );
}
