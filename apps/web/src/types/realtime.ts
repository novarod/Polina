export interface PresenceUser {
  id: string;
  name: string;
}

export interface CursorMove {
  userId: string;
  x: number;
  y: number;
}

export interface MissionCountEntry {
  mission_id: string;
  count: number;
}

export interface RealtimeServerFrame {
  type: string;
  v?: number;
  topic?: string;
  user?: PresenceUser;
  users?: PresenceUser[];
  missions?: MissionCountEntry[];
  user_id?: string;
  mission_id?: string;
  count?: number;
  x?: number;
  y?: number;
  code?: string;
}

export interface MissionTopicParams {
  orgId: string;
  workspaceId: string;
  missionId: string;
}

export interface CanvasPresence {
  peers: PresenceUser[];
  subscribeCursor: (listener: (move: CursorMove) => void) => () => void;
  sendPos: (x: number, y: number) => void;
}
