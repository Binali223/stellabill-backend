# 0006. HTTP Idempotency Keys on Mutations

## Status

Accepted

## Context

Clients retry POSTs under timeouts and flaky networks. Duplicate charges or
status transitions are unacceptable for a billing API. Relying only on
client-side discipline is insufficient.

## Decision

We will honor `Idempotency-Key` on `/api/v1` mutations, scoped by
`tenantID + callerID`, bind method/path/payload hash, cache successful
responses (24h TTL), and return `Idempotency-Replayed: true` on replay. Keys
persist in `idempotency_keys`. See `docs/idempotency.md`,
`internal/middleware/idempotency.go`, migration `0005_create_idempotency_keys`.

## Consequences

### Positive

- Safe client retries without double side effects
- Tenant-scoped keys prevent cross-tenant key collisions

### Negative

- Storage growth and TTL housekeeping
- Mismatched payload with same key must be rejected carefully

### Neutral / Follow-ups

- Complements webhook event dedup (ADR-0007) and outbox dedup (ADR-0001)
