export interface ApiKey {
  id: string;
  name: string;
  last_used_at: string | null;
  created_at: string;
  revoked_at: string | null;
}

export interface CreatedApiKey {
  id: string;
  name: string;
  key: string;
  created_at: string;
}
