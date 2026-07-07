import type { Role } from "@/lib/roles";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  created_at: string;
}

export interface OrganizationListItem {
  id: string;
  name: string;
  slug: string;
  role: Role;
}
