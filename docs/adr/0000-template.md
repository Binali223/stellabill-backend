# Architecture Decision Record Template

<!--
Nygard-style ADR template. Copy this file when recording a new decision:

  cp docs/adr/0000-template.md docs/adr/NNNN-short-title.md

Rules enforced by `make docs-lint` / `go run ./cmd/adr-lint`:
- Filename: NNNN-kebab-case.md (NNNN is a unique zero-padded integer; 0000 is reserved for this template)
- Required sections (exact H2 headings): Status, Context, Decision, Consequences
- Status must be one of: Proposed | Accepted | Deprecated | Superseded | Rejected
- Title H1 must match `# NNNN. Title` where NNNN matches the filename number
-->

# 0000. Title of the Decision

## Status

Proposed

<!-- Optional: Superseded by [ADR-NNNN](NNNN-other.md) -->

## Context

Describe the forces at play: business requirements, technical constraints,
security/compliance needs, and alternatives that were considered. Keep this
factual and time-scoped so future readers understand *why* a decision was needed.

## Decision

State the chosen approach in one or two clear sentences. Prefer "We will…"
language. Link to implementing packages, migrations, or docs when helpful.

## Consequences

### Positive

- Benefit of the decision

### Negative

- Trade-off or operational cost

### Neutral / Follow-ups

- Related work, monitoring, or future ADRs
