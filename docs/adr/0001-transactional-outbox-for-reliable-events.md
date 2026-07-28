# 0001. Transactional Outbox for Reliable Events

## Status

Accepted

## Context

Billing and subscription mutations must notify external systems (webhooks,
subscribers) after a successful commit. Publishing to an external broker inside
the request path creates dual-write risk: the database can commit while the
publish fails (or the reverse), leaving silent data loss. A separate message
bus was not yet part of the operational stack.

## Decision

We will persist domain events in `outbox_events` in the **same PostgreSQL
transaction** as the business write, then dispatch asynchronously with
`FOR UPDATE SKIP LOCKED`, exponential backoff, and a dead-letter path after max
retries. See `docs/outbox-pattern.md`, `internal/outbox/`, and migrations
`0002_create_outbox`, `0007_outbox_dead_letter_view`, `0009_add_outbox_deduplication`.

## Consequences

### Positive

- Exactly-once *intent* at the DB boundary; retries are safe and observable
- No hard dependency on Kafka/RabbitMQ for core reliability

### Negative

- Polling workers add DB load; delivery lag under backlog
- Consumers must still be idempotent

### Neutral / Follow-ups

- Outbox payloads for sensitive types are encrypted per ADR-0002
- Chaos/runbook coverage under `docs/runbooks/`
