# AGENTS.md

## Layout

- **operator/**: Go operator (Kubebuilder). See operator/AGENTS.md
- **charts/byo-addons/**: Helm chart for the operator

## Never edit by hand

- `operator/api/**/zz_generated.deepcopy.go` → run `make -C operator generate`
- `operator/config/crd/bases/` → run `make -C operator manifests`
- `charts/byo-addons/crds/` → run `make sync-chart-crds`

## Done means

- `make -C operator test` passes
- `helm lint charts/byo-addons` passes
