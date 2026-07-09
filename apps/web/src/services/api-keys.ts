import { apiFetch } from "@/services/api";
import type { CreatedApiKey } from "@/types/api-key";

export function createApiKey(
  orgId: string,
  name: string
): Promise<CreatedApiKey> {
  return apiFetch<CreatedApiKey>(`/organizations/${orgId}/api-keys`, {
    method: "POST",
    body: { name },
  });
}

export function revokeApiKey(orgId: string, keyId: string): Promise<void> {
  return apiFetch<void>(`/organizations/${orgId}/api-keys/${keyId}`, {
    method: "DELETE",
  });
}
