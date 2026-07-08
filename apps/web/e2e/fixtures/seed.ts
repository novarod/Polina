const apiUrl = process.env.API_URL ?? "http://localhost:8080";

interface SeededMission {
  orgId: string;
  workspaceId: string;
  missionId: string;
  path: string;
}

async function apiRequest<T>(
  sessionToken: string,
  method: string,
  path: string,
  body: unknown
): Promise<T> {
  const response = await fetch(`${apiUrl}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      cookie: `session=${sessionToken}`,
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw new Error(`Seed falhou em ${method} ${path}: ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function seedMission(
  sessionToken: string,
  graph: unknown
): Promise<SeededMission> {
  const stamp = `${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
  const org = await apiRequest<{ id: string }>(
    sessionToken,
    "POST",
    "/organizations",
    { name: `Canvas E2E ${stamp}`, slug: `canvas-e2e-${stamp}` }
  );
  const workspace = await apiRequest<{ id: string }>(
    sessionToken,
    "POST",
    `/organizations/${org.id}/workspaces`,
    { name: "Canvas WS", description: "" }
  );
  const mission = await apiRequest<{ id: string }>(
    sessionToken,
    "POST",
    `/organizations/${org.id}/workspaces/${workspace.id}/missions`,
    { name: "Missão Canvas", description: "Grafo de teste" }
  );
  if (graph !== null) {
    await apiRequest(
      sessionToken,
      "PUT",
      `/organizations/${org.id}/workspaces/${workspace.id}/missions/${mission.id}/graph`,
      graph
    );
  }
  return {
    orgId: org.id,
    workspaceId: workspace.id,
    missionId: mission.id,
    path: `/orgs/${org.id}/workspaces/${workspace.id}/missions/${mission.id}`,
  };
}
