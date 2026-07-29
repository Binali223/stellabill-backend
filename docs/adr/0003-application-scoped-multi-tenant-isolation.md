# 0003. Application-Scoped Multi-Tenant Isolation

## Status

Accepted

## Context

Stellabill hosts many merchants in one deployment. Cross-tenant reads/writes
(IDOR via IDs or pagination cursors) are a critical security failure mode.
PostgreSQL Row Level Security (`ENABLE ROW LEVEL SECURITY` / `CREATE POLICY`)
is **not** used in current migrations; isolation must still be enforceable and
testable.

## Decision

We will enforce tenancy in the application layer: JWT `tenant_id` claims,
repository queries always scoped by `tenant_id`, HMAC-signed pagination cursors
that embed tenant, RBAC for admin vs merchant scope, and fuzz/isolation tests.
We explicitly **do not** rely on Postgres RLS at this time. Evidence:
`internal/auth/claims.go`, `internal/pagination/scoped_cursor.go`,
`internal/tests/tenant_isolation_fuzz_test.go`.

## Consequences

### Positive

- Works with current ORM/SQL style and testcontainers setup
- Failures are visible in application code review and fuzz tests

### Negative

- A missing `WHERE tenant_id = …` is a footgun; RLS would provide defense in depth
- DB admins/superuser paths can still see all rows

### Neutral / Follow-ups

- Revisit Postgres RLS as a hardening layer in a future ADR if threat model requires it
