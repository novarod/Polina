export type MissionStatus = "DRAFT" | "APPROVED";

export interface MissionVersion {
  id: string;
  version_number: number;
  hash: string;
  published_by_id: string;
  created_at: string;
}

export interface MissionVersionDetail extends MissionVersion {
  mission_data: unknown;
}

export interface PublishResponse {
  mission_id: string;
  version: number;
  hash: string;
  status: MissionStatus;
  active_hash: string | null;
}

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
