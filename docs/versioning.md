# Versioning

This project uses semantic versioning:

- `MAJOR`: incompatible CRD/API behavior changes.
- `MINOR`: backwards-compatible features, new providers, or new controllers.
- `PATCH`: bug fixes, documentation, and packaging updates.

The current version is stored in `VERSION`, the Helm chart `version`, and the default image tag.

Recommended branch model:

- `main`: stable, releasable code.
- `codex/feature-name` or `feature/feature-name`: new work.
- `release/vX.Y`: stabilization branches when needed.

Recommended tag format:

```bash
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

For CRDs, keep backwards compatibility within a minor line. Add a new API version such as `v1beta1` before removing or renaming fields.
