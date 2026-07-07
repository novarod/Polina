import { cache } from "react";

import { serverFetch } from "@/services/server-api";
import type { Role } from "@/lib/roles";
import type { OrganizationListItem } from "@/types/organization";

export const getMyOrganizations = cache(
  (): Promise<OrganizationListItem[] | null> => {
    return serverFetch<OrganizationListItem[]>("/organizations");
  }
);

export async function getOrgRole(orgId: string): Promise<Role | null> {
  const orgs = await getMyOrganizations();
  return orgs?.find((org) => org.id === orgId)?.role ?? null;
}
