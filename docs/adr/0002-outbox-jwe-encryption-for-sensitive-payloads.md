# 0002. Outbox JWE Encryption for Sensitive Payloads

## Status

Accepted

## Context

Outbox rows and downstream logs can contain payment or PII-bearing webhook
payloads. TLS protects data in transit but not at rest in PostgreSQL or in
operator-accessible backups. Plaintext event bodies increase blast radius on
DB compromise.

## Decision

We will encrypt sensitive outbox event types with subscriber-registered JWKs
using JWE compact serialization (`RSA-OAEP-256` / `A256GCM`), publish as
`application/jose+json`, and permanently fail to the dead-letter queue when no
active key exists. See `docs/outbox-jwe.md`, `internal/outbox/jwe.go`, and
migration `0008_create_subscriber_keys`.

## Consequences

### Positive

- Payload confidentiality at rest for high-sensitivity event types
- Per-subscriber key rotation without rewriting the outbox schema

### Negative

- Key management complexity; misconfigured keys stall delivery (by design)
- Debugging requires authorized decryption tooling

### Neutral / Follow-ups

- Aligns with PII policy in `internal/docs/PII_POLICY.md`
