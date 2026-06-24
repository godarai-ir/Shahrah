# Service-Scoped CI/CD and GitOps Versioning

This repository now uses service-scoped release behavior for staging and production promotions.

## Goals
- Build and push only changed service images.
- Tag images with both immutable and semver tags.
- Update only changed Helm values keys in GitOps.
- Reduce ArgoCD rollout blast radius.

## Workflow: Staging
Workflow file: `.github/workflows/gitops-staging-release.yml`

### What it does
1. Detects changed services by path (or manual `workflow_dispatch` input).
2. Computes immutable tag: `sha-<full-commit-sha>`.
3. Computes semver tag per service prefix (best-effort Conventional Commits bump).
4. Builds/pushes only selected images.
5. Opens a PR in `Delpak/shahrah-gitops` that updates only the related service tag keys in `helm/shahrah/values.staging.yaml`.

### Supported service IDs
- `control-plane`
- `model-discovery`
- `model-discovery-migrate`
- `model-discovery-web`
- `support-admin`
- `support-admin-migrate`
- `silk-road`
- `tunnel-client`
- `tunnel-server`
- `goose`
- `docs`
- `website`
- `blog`
- `tenant-admin`

## Workflow: Production
Workflow file: `.github/workflows/gitops-promote-production.yml`

Manual inputs:
- `service_id`
- `image_tag` (must start with `sha-`)

The workflow updates only one values key in `helm/shahrah/values.production.yaml` and opens a PR.

## Tagging policy
- Deployment source of truth: immutable `sha-*` tags.
- Additional discoverability tag: service semver tag (for example `v1.4.2`).
- Avoid `latest` for deployments.

## Git tag naming for semver history
Semver discovery uses tags in this repository in the format:
- `<service-id>/vMAJOR.MINOR.PATCH`

Examples:
- `model-discovery/v0.8.3`
- `model-discovery-web/v0.8.3`
- `support-admin/v1.2.0`

If no tag exists, the initial generated version starts at `v0.1.0`.

## Conventional Commits bump mapping (auto)
- `BREAKING CHANGE` or `type!:` -> major
- `feat:` -> minor
- otherwise -> patch

## Validation checklist
- A change under `apps/model-discovery/**` should only build:
  - `model-discovery`
  - `model-discovery-migrate`
- A change under `apps/support-admin/**` should only build:
  - `support-admin`
  - `support-admin-migrate`
- A change under `apps/model-discovery-web/**` should only build:
  - `model-discovery-web`
- GitOps PR should only change:
  - `modelDiscovery.image.tag`
  - `migrations.modelDiscovery.image.tag`
  - `modelDiscoveryWeb.image.tag` when the public web app is selected
- ArgoCD should only roll related workloads.

## Rollback
- Revert the GitOps PR or set the service tag back to a previous immutable `sha-*` value.
