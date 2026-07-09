import { validateGraph } from "@/lib/validate-graph";
import type { EditorGraph } from "@/types/graph";

import { expect, test } from "../fixtures/session";
import { seedMission } from "../fixtures/seed";

const apiUrl = process.env.API_URL ?? "http://localhost:8080";

const corpus: Array<{ name: string; graph: EditorGraph }> = [
  {
    name: "linear válido",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
        { id: "c", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "c" },
      ],
    },
  },
  {
    name: "ciclo",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
        { id: "c", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "b" },
        { id: "e3", source: "b", target: "c" },
      ],
    },
  },
  {
    name: "dois STARTs",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "START" },
        { id: "c", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "c" },
        { id: "e2", source: "b", target: "c" },
      ],
    },
  },
  {
    name: "sem END",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "b", target: "a" },
      ],
    },
  },
  {
    name: "nó inalcançável",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "END" },
        { id: "orfao", type: "OBJECTIVE" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "orfao", target: "b" },
      ],
    },
  },
  {
    name: "dead-end",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "OBJECTIVE" },
        { id: "c", type: "END" },
      ],
      edges: [{ id: "e1", source: "a", target: "b" }],
    },
  },
  {
    name: "aresta órfã",
    graph: {
      nodes: [
        { id: "a", type: "START" },
        { id: "b", type: "END" },
      ],
      edges: [
        { id: "e1", source: "a", target: "b" },
        { id: "e2", source: "a", target: "fantasma" },
      ],
    },
  },
];

test("espelho client-side e dag.Validate concordam em todo o corpus", async ({
  sessionToken,
}) => {
  const seeded = await seedMission(sessionToken, null);

  for (const { name, graph } of corpus) {
    const mirrorAccepts = validateGraph(graph).errorCount === 0;
    const response = await fetch(
      `${apiUrl}/organizations/${seeded.orgId}/workspaces/${seeded.workspaceId}/missions/${seeded.missionId}/graph`,
      {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          cookie: `session=${sessionToken}`,
        },
        body: JSON.stringify(graph),
      }
    );
    const apiAccepts = response.status === 200;

    expect(
      apiAccepts,
      `divergência em "${name}": espelho=${mirrorAccepts ? "aceita" : "rejeita"}, API=${response.status}`
    ).toBe(mirrorAccepts);
  }
});
