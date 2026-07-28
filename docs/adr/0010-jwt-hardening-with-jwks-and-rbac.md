# 0010. JWT Hardening with JWKS and RBAC

## Status

Accepted

## Context

A multi-tenant billing API is exposed to token confusion, cross-audience reuse,
and painful static-secret rotation. Weak JWT validation and flat “authenticated
or not” checks are insufficient for admin/merchant/customer separation.

## Decision

We will enforce strict JWT validation (algorithm allow-list, issuer, audience,
`nbf`, clock skew), cache JWKS with stampede protection and refresh-on-unknown
`kid`, and apply a role→permission RBAC matrix in middleware. See
`docs/JWT_HARDENING.md`, `docs/SECURITY.md`, `internal/auth/jwt.go`,
`internal/auth/jwks_cache.go`, `internal/auth/rbac_matrix.yaml`.

## Consequences

### Positive

- Stronger authn/authz defaults for public and admin APIs
- Key rotation without downtime via JWKS

### Negative

- JWKS/network dependency; cache bugs can cause auth outages
- RBAC matrix must stay synchronized with new endpoints

### Neutral / Follow-ups

- Admin HMAC request signing (`docs/admin-signing.md`) adds a second factor for privileged routes
