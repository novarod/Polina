## Description
<!-- Summarize what has been done and why. Link the related issue: Closes #123 -->

## Changes
<!-- This section is used to generate release notes. Keep it concise.
  Suggested format:
  [FEAT] Added organization module (tenant-scoped CRUD + admin membership)
  [FIX] Propagate the COUNT error in member.ListByOrg
  [REFACTOR] Introduced the Querier abstraction and WithinTx transaction manager
-->

## How to Test / Points to Consider
<!-- Help the reviewer know where to focus. -->
1. Step-by-step instructions (e.g., `POST /organizations` with `{name, slug}` returns 201 and a
   `GET /organizations` shows the org with role `ADMIN`).
2. Note any database changes (`make migrate`) or new env vars (keep `apps/api/.env.example` in sync).

## Checklist
- [ ] `go build ./...`, `go vet ./...` and `gofmt -l .` are clean.
- [ ] `make test` (and `make test-integration` when DB-touching) pass.
- [ ] I performed a self-review of my own code.
- [ ] No leftover debug code or commented-out blocks.
- [ ] **No framework/infra imports leaked into the pure domain layer**:
      `apps/api/internal/domain/**` must not import Echo, pgx or `net/http`.
- [ ] I updated the documentation affected by this change (README, `specs/`).
