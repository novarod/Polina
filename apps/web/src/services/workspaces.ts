import { apiFetch } from "@/services/api";
import type { Workspace } from "@/types/workspace";

export interface WorkspaceInput {
  name: string;
  description: string;
}

export function createWorkspace(
  orgId: string,
  input: WorkspaceInput
): Promise<Workspace> {
  return apiFetch<Workspace>(`/organizations/${orgId}/workspaces`, {
    method: "POST",
    body: input,
  });
}

export function updateWorkspace(
  orgId: string,
  workspaceId: string,
  input: WorkspaceInput
): Promise<Workspace> {
  return apiFetch<Workspace>(
    `/organizations/${orgId}/workspaces/${workspaceId}`,
    { method: "PATCH", body: input }
  );
}

export function deleteWorkspace(
  orgId: string,
  workspaceId: string
): Promise<void> {
  return apiFetch<void>(`/organizations/${orgId}/workspaces/${workspaceId}`, {
    method: "DELETE",
  });
}
