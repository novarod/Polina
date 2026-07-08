import type { EditorGraph } from "@/types/graph";

export const MAX_NODES = 1000;
export const MAX_EDGES = 2000;

export interface GraphValidation {
  nodeErrors: Map<string, string[]>;
  graphErrors: string[];
  errorCount: number;
}

function addNodeError(
  nodeErrors: Map<string, string[]>,
  nodeId: string,
  message: string
): void {
  const existing = nodeErrors.get(nodeId) ?? [];
  existing.push(message);
  nodeErrors.set(nodeId, existing);
}

function detectCycle(
  adjacency: Map<string, string[]>,
  nodeIds: string[]
): string | null {
  const state = new Map<string, number>();
  const visit = (id: string): string | null => {
    state.set(id, 1);
    for (const next of adjacency.get(id) ?? []) {
      if (state.get(next) === 1) {
        return next;
      }
      if (!state.has(next)) {
        const found = visit(next);
        if (found !== null) {
          return found;
        }
      }
    }
    state.set(id, 2);
    return null;
  };
  for (const id of nodeIds) {
    if (!state.has(id)) {
      const found = visit(id);
      if (found !== null) {
        return found;
      }
    }
  }
  return null;
}

function reachableFrom(
  adjacency: Map<string, string[]>,
  start: string
): Set<string> {
  const visited = new Set<string>([start]);
  const queue = [start];
  while (queue.length > 0) {
    const current = queue.pop() as string;
    for (const next of adjacency.get(current) ?? []) {
      if (!visited.has(next)) {
        visited.add(next);
        queue.push(next);
      }
    }
  }
  return visited;
}

export function validateGraph(graph: EditorGraph): GraphValidation {
  const nodeErrors = new Map<string, string[]>();
  const graphErrors: string[] = [];
  const result = () => ({
    nodeErrors,
    graphErrors,
    errorCount:
      graphErrors.length +
      [...nodeErrors.values()].reduce((sum, list) => sum + list.length, 0),
  });

  if (graph.nodes.length === 0 && graph.edges.length === 0) {
    return result();
  }

  if (graph.nodes.length > MAX_NODES) {
    graphErrors.push(
      `O grafo excede o máximo de ${MAX_NODES} nós (tem ${graph.nodes.length})`
    );
  }
  if (graph.edges.length > MAX_EDGES) {
    graphErrors.push(
      `O grafo excede o máximo de ${MAX_EDGES} arestas (tem ${graph.edges.length})`
    );
  }
  if (graphErrors.length > 0) {
    return result();
  }

  const nodeIds = graph.nodes.map((node) => node.id);
  const nodeSet = new Set(nodeIds);
  const adjacency = new Map<string, string[]>(
    nodeIds.map((id) => [id, [] as string[]])
  );
  let edgeRefError = false;
  for (const edge of graph.edges) {
    if (!nodeSet.has(edge.source)) {
      graphErrors.push(
        `A aresta "${edge.id}" referencia o nó de origem inexistente "${edge.source}"`
      );
      edgeRefError = true;
      continue;
    }
    if (!nodeSet.has(edge.target)) {
      graphErrors.push(
        `A aresta "${edge.id}" referencia o nó de destino inexistente "${edge.target}"`
      );
      edgeRefError = true;
      continue;
    }
    adjacency.get(edge.source)?.push(edge.target);
  }
  if (edgeRefError) {
    return result();
  }

  const cycleNode = detectCycle(adjacency, nodeIds);
  if (cycleNode !== null) {
    addNodeError(nodeErrors, cycleNode, "Ciclo detectado envolvendo este nó");
  }

  const starts = graph.nodes.filter((node) => node.type === "START");
  if (starts.length === 0) {
    graphErrors.push("O grafo precisa de exatamente um nó START");
  } else if (starts.length > 1) {
    for (const start of starts) {
      addNodeError(nodeErrors, start.id, "Só pode existir um nó START");
    }
  }

  if (!graph.nodes.some((node) => node.type === "END")) {
    graphErrors.push("O grafo precisa de pelo menos um nó END");
  }

  if (starts.length === 1) {
    const reachable = reachableFrom(adjacency, starts[0].id);
    for (const node of graph.nodes) {
      if (!reachable.has(node.id)) {
        addNodeError(
          nodeErrors,
          node.id,
          "Não é alcançável a partir do nó START"
        );
      }
    }
  }

  for (const node of graph.nodes) {
    if (node.type === "END") {
      continue;
    }
    if ((adjacency.get(node.id) ?? []).length === 0) {
      addNodeError(
        nodeErrors,
        node.id,
        "Sem aresta de saída (beco sem saída)"
      );
    }
  }

  return result();
}
