# Architecture Decision Records

Nygard-style ADRs for Stellabill backend.

> This index is **auto-generated** by `make adr-index` / `go run ./cmd/adr-lint -write-index`.
> Do not edit by hand — update individual ADR files instead.

See [ADR_TOOLS.md](ADR_TOOLS.md) for authoring rules and [0000-template.md](0000-template.md) for the template.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-transactional-outbox-for-reliable-events.md) | Transactional Outbox for Reliable Events | Accepted |
| [0002](0002-outbox-jwe-encryption-for-sensitive-payloads.md) | Outbox JWE Encryption for Sensitive Payloads | Accepted |
| [0003](0003-application-scoped-multi-tenant-isolation.md) | Application-Scoped Multi-Tenant Isolation | Accepted |
| [0004](0004-in-process-feature-flags-for-progressive-rollout.md) | In-Process Feature Flags for Progressive Rollout | Accepted |
| [0005](0005-compensating-saga-for-multi-step-billing-flows.md) | Compensating Saga for Multi-Step Billing Flows | Accepted |
| [0006](0006-http-idempotency-keys-on-mutations.md) | HTTP Idempotency Keys on Mutations | Accepted |
| [0007](0007-webhook-hmac-verification-and-event-dedup.md) | Webhook HMAC Verification and Event Deduplication | Accepted |
| [0008](0008-primary-replica-read-routing-with-freshness.md) | Primary–Replica Read Routing with Freshness | Accepted |
| [0009](0009-statement-cold-archive-to-object-storage.md) | Statement Cold Archive to Object Storage | Accepted |
| [0010](0010-jwt-hardening-with-jwks-and-rbac.md) | JWT Hardening with JWKS and RBAC | Accepted |

## Template

| File | Purpose |
| --- | --- |
| [0000-template.md](0000-template.md) | Copy this file when adding a new ADR |
