# 0007. Webhook HMAC Verification and Event Deduplication

## Status

Accepted

## Context

Inbound provider webhooks are spoofable and frequently retried. Unverified or
duplicated events enable fraud and double-processing of payments/subscription
lifecycle hooks.

## Decision

We will verify provider HMAC signatures (Stripe/PayPal/GitHub/Square/custom),
enforce timestamp tolerance for replay resistance, and deduplicate by
tenant-scoped provider event ID before business handling. See
`docs/webhook_security.md`, `docs/WEBHOOK_IDEMPOTENCY.md`,
`internal/middleware/webhook_verification.go`.

## Consequences

### Positive

- Authenticity and replay controls at the edge of the API
- Dedup prevents duplicate domain effects from provider retries

### Negative

- Clock skew and secret rotation require ops discipline
- Provider-specific signing variants increase middleware complexity

### Neutral / Follow-ups

- Outbound webhook delivery reliability remains owned by the outbox (ADR-0001)
