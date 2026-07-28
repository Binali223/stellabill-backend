# adr-tools configuration for Stellabill

This repository uses [Nygard-style Architecture Decision Records](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
under `docs/adr/`.

## Layout

| Path | Purpose |
| --- | --- |
| `.adr-dir` | Points adr-tools (and our linter) at `docs/adr` |
| `docs/adr/0000-template.md` | Required template — copy for new ADRs |
| `docs/adr/NNNN-*.md` | Accepted / proposed decisions |
| `docs/adr/README.md` | Auto-generated index (`make adr-index`) |

## Creating a new ADR

```bash
# With adr-tools (optional):
export ADR_TEMPLATE="$PWD/docs/adr/0000-template.md"
adr new "Short title of the decision"

# Or manually:
NEXT=$(printf '%04d' $(( $(ls docs/adr/[0-9][0-9][0-9][0-9]-*.md 2>/dev/null | wc -l) )))
cp docs/adr/0000-template.md "docs/adr/${NEXT}-short-title.md"
# Edit Status/Context/Decision/Consequences, then:
make adr-index
make docs-lint
```

## Validation

```bash
make docs-lint   # template + unique numbers + index freshness
go test ./internal/adr/... -cover
```

CI runs the same checks so duplicate ADR numbers cannot merge.
