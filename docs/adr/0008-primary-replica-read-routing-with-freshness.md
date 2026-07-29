# 0008. Primary–Replica Read Routing with Freshness

## Status

Accepted

## Context

Plan and subscription reads are heavy relative to writes. Serving all traffic
from the primary limits scale. Naïve replica reads break read-your-writes after
mutations and can serve stale tenant data.

## Decision

We will introduce `db.ReadRouter`: route eligible reads to a replica when
healthy, fall back to primary on replica failure, and force primary when a
freshness token is present. This is soft read scaling—not full CQRS. See
`internal/db/router.go` and `docs/runbooks/multi-region-failover.md`.

## Consequences

### Positive

- Offloads read traffic while preserving correctness for fresh reads
- Automatic primary fallback improves availability

### Negative

- Replica lag remains a product concern for non-freshness paths
- Router configuration/ops complexity across environments

### Neutral / Follow-ups

- Pair with multi-region failover runbooks for DR drills
