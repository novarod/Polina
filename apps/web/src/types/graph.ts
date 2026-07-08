export interface EditorNodePosition {
  x: number;
  y: number;
}

export interface EditorNode {
  id: string;
  type: string;
  data?: unknown;
  position?: EditorNodePosition;
}

export interface EditorEdge {
  id: string;
  source: string;
  target: string;
}

export interface EditorGraph {
  nodes: EditorNode[];
  edges: EditorEdge[];
}
