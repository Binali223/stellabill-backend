# 0005. Compensating Saga for Multi-Step Billing Flows

## Status

Accepted

## Context

Flows such as cancel-subscription-with-refund touch multiple aggregates. A
partial failure (refund succeeds, status update fails, or the reverse) leaves
inconsistent billing state that is hard to repair ad hoc.

## Decision

We will implement an explicit compensating saga coordinator with
execute/compensate steps and persisted saga/step state in PostgreSQL (plus an
in-memory store for tests), starting with `cancel_subscription_with_refund`.
See `internal/saga/` and migration `0012_saga`.

## Consequences

### Positive

- Recoverable multi-step workflows with auditable step state
- Compensations are first-class rather than tribal knowledge

### Negative

- Saga frameworks add complexity vs single-transaction updates
- Compensations themselves can fail and need operator playbooks

### Neutral / Follow-ups

- New multi-step billing flows should register as saga definitions, not ad hoc retries
