export type MissionStatus = "DRAFT" | "APPROVED";

export interface Mission {
  id: string;
  organization_id: string;
  workspace_id: string;
  name: string;
  description: string;
  status: MissionStatus;
  active_hash: string | null;
  graph: unknown;
  created_by_id: string;
  created_at: string;
}
