# 0004. In-Process Feature Flags for Progressive Rollout

## Status

Accepted

## Context

Risky schema and behavior changes (for example statement partitioning) need
gradual enablement without waiting on a full traffic-splitting canary
infrastructure. No external feature-flag SaaS or edge canary router is in the
stack today. Docs mention blue/green/canary deploys operationally; runtime
gating still needed inside the process.

## Decision

We will use an **in-process feature-flag manager** (defaults + env JSON + admin
GET/PATCH APIs) with Gin middleware gates. Flags such as
`statements_partitioning` enable progressive rollout. This is the implemented
“canary” mechanism for behavior, not HTTP traffic splitting. See
`internal/featureflags/`, `internal/middleware/featureflags.go`.

## Consequences

### Positive

- Fast, low-ops gating for dangerous paths
- Admin visibility/control without redeploy for flag flips (env/admin)

### Negative

- Flag state is per-process unless synchronized externally
- Not a substitute for percentage-based edge canaries

### Neutral / Follow-ups

- Pair with deploy runbooks under `docs/runbooks/` for release safety
