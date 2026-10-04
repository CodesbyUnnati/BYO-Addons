# AGENTS.md (operator)

## Build

```bash
make build
```

## Test

```bash
make test
```

## After changing anything in api/

```bash
make generate manifests
```

From the repo root:

```bash
make sync-chart-crds
```
