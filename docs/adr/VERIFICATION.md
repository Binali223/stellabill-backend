# ADR index — verification notes

Issue: Stellabill/stellabill-backend #473  
Branch: `docs/adr-index`  
Commit: `docs: introduce ADR index and backfill decisions`

## Security / correctness

- Tenancy ADR (**0003**) documents **application-scoped** isolation and explicitly states Postgres RLS is **not** in use (matches codebase).
- Progressive rollout ADR (**0004**) documents in-process feature flags as the implemented “canary” mechanism (not edge traffic splitting).
- Template + linter reject malformed Status values and duplicate `NNNN` numbers before merge (CI `make docs-lint`).

## Test output

```text
$ make docs-lint
go run ./cmd/adr-lint -check-index
ADR lint OK (10 decisions, template present, unique numbers).
go test ./internal/adr/... -count=1 -cover
ok  	stellarbill-backend/internal/adr	0.400s	coverage: 99.3% of statements
```

Coverage for `internal/adr`: **99.3%** statements (≥95% guideline).

## Edge cases covered by tests

- Duplicate ADR numbers
- Missing template `0000-template.md`
- Stale `README.md` index
- Invalid filename / title number mismatch / missing sections / empty or illegal Status
- `.adr-dir` empty, relative, absolute, and unreadable
