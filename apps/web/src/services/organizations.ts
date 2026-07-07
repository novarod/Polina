import { apiFetch } from "@/services/api";
import type { Organization } from "@/types/organization";

export interface CreateOrganizationInput {
  name: string;
  slug: string;
}

export function createOrganization(
  input: CreateOrganizationInput
): Promise<Organization> {
  return apiFetch<Organization>("/organizations", {
    method: "POST",
    body: input,
  });
}

export function renameOrganization(
  orgId: string,
  name: string
): Promise<Organization> {
  return apiFetch<Organization>(`/organizations/${orgId}`, {
    method: "PATCH",
    body: { name },
  });
}

export function deleteOrganization(orgId: string): Promise<void> {
  return apiFetch<void>(`/organizations/${orgId}`, { method: "DELETE" });
}
