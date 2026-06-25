-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
CREATE TYPE member_role AS ENUM ('VIEWER', 'DESIGNER', 'ADMIN');
CREATE TYPE invite_status AS ENUM ('PENDING', 'ACCEPTED', 'CANCELLED', 'EXPIRED');
CREATE TYPE mission_status AS ENUM ('DRAFT', 'APPROVED');

-- Users (global, not tenant-scoped)
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    password    TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Organizations (top-level tenant)
CREATE TABLE organizations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Members (User × Organization link)
CREATE TABLE members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    role            member_role NOT NULL DEFAULT 'VIEWER',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    UNIQUE (user_id, organization_id)
);

-- Invites
CREATE TABLE invites (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    invited_by_id   UUID NOT NULL REFERENCES members(id),
    email           TEXT NOT NULL,
    role            member_role NOT NULL DEFAULT 'VIEWER',
    status          invite_status NOT NULL DEFAULT 'PENDING',
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Workspaces (project inside an org)
CREATE TABLE workspaces (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Missions (quest graph)
CREATE TABLE missions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id),
    name            TEXT NOT NULL,
    description     TEXT,
    status          mission_status NOT NULL DEFAULT 'DRAFT',
    active_hash     TEXT,
    graph           JSONB NOT NULL DEFAULT '{"nodes":[],"edges":[]}',
    created_by_id   UUID NOT NULL REFERENCES members(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Mission Versions (immutable snapshots)
CREATE TABLE mission_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id      UUID NOT NULL REFERENCES missions(id),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    version_number  INT NOT NULL,
    hash            TEXT NOT NULL UNIQUE,
    graph           JSONB NOT NULL,
    mission_data    JSONB NOT NULL,
    published_by_id UUID NOT NULL REFERENCES members(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (mission_id, version_number)
);

-- Organization API Keys (for UE5 plugin)
CREATE TABLE organization_api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name            TEXT NOT NULL,
    key_hash        TEXT NOT NULL UNIQUE,
    last_used_at    TIMESTAMPTZ,
    created_by_id   UUID NOT NULL REFERENCES members(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at      TIMESTAMPTZ
);

-- Player Mission State (runtime progress)
CREATE TABLE player_mission_states (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id      UUID NOT NULL REFERENCES missions(id),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    player_id       TEXT NOT NULL,
    current_node_id TEXT NOT NULL,
    completed_nodes JSONB NOT NULL DEFAULT '[]',
    state_data      JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (mission_id, player_id)
);

-- Indexes
CREATE INDEX idx_members_org ON members(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_members_user ON members(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workspaces_org ON workspaces(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missions_workspace ON missions(workspace_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missions_org ON missions(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missions_active_hash ON missions(active_hash) WHERE active_hash IS NOT NULL;
CREATE INDEX idx_mission_versions_mission ON mission_versions(mission_id);
CREATE INDEX idx_mission_versions_hash ON mission_versions(hash);
CREATE INDEX idx_api_keys_org ON organization_api_keys(organization_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_player_states_mission ON player_mission_states(mission_id);
