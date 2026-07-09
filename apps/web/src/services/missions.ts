import { apiFetch } from "@/services/api";
import type { EditorGraph } from "@/types/graph";
import type {
  Mission,
  MissionVersionDetail,
  PublishResponse,
} from "@/types/mission";

export interface MissionInput {
  name: string;
  description: string;
}

function missionsPath(orgId: string, workspaceId: string): string {
  return `/organizations/${orgId}/workspaces/${workspaceId}/missions`;
}

export function createMission(
  orgId: string,
  workspaceId: string,
  input: MissionInput
): Promise<Mission> {
  return apiFetch<Mission>(missionsPath(orgId, workspaceId), {
    method: "POST",
    body: input,
  });
}

export function updateMission(
  orgId: string,
  workspaceId: string,
  missionId: string,
  input: MissionInput
): Promise<Mission> {
  return apiFetch<Mission>(`${missionsPath(orgId, workspaceId)}/${missionId}`, {
    method: "PATCH",
    body: input,
  });
}

export function updateMissionGraph(
  orgId: string,
  workspaceId: string,
  missionId: string,
  graph: EditorGraph
): Promise<Mission> {
  return apiFetch<Mission>(
    `${missionsPath(orgId, workspaceId)}/${missionId}/graph`,
    { method: "PUT", body: graph }
  );
}

export function publishMission(
  orgId: string,
  workspaceId: string,
  missionId: string
): Promise<PublishResponse> {
  return apiFetch<PublishResponse>(
    `${missionsPath(orgId, workspaceId)}/${missionId}/publish`,
    { method: "POST" }
  );
}

export function getMissionVersion(
  orgId: string,
  workspaceId: string,
  missionId: string,
  hash: string
): Promise<MissionVersionDetail> {
  return apiFetch<MissionVersionDetail>(
    `${missionsPath(orgId, workspaceId)}/${missionId}/versions/${hash}`
  );
}

export function deleteMission(
  orgId: string,
  workspaceId: string,
  missionId: string
): Promise<void> {
  return apiFetch<void>(`${missionsPath(orgId, workspaceId)}/${missionId}`, {
    method: "DELETE",
  });
}
