# 0009. Statement Cold Archive to Object Storage

## Status

Accepted

## Context

Statements older than roughly 24 months are rarely read but inflate hot
PostgreSQL size and backup cost. Deleting them loses compliance/history;
keeping them fully hot is expensive.

## Decision

We will run a background archival job that moves aged statements to S3 (or
compatible) cold storage, leave stubs in Postgres, and rehydrate on demand for
transparent reads. Feature-flag gating may control rollout. See
`docs/STATEMENT_COLD_ARCHIVE.md`, `internal/worker/statement_archive_job.go`,
`internal/storage/s3/`, migration `0010_add_statement_archival`.

## Consequences

### Positive

- Controls hot DB growth while retaining retrievability
- Aligns with tenant export/S3 patterns already in the codebase

### Negative

- Rehydrate latency and S3 failure modes on rare reads
- Cross-service IAM/credentials become a dependency

### Neutral / Follow-ups

- Coordinated with partitioning flag (ADR-0004) where applicable
