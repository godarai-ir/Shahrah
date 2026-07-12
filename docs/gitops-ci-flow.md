# GitOps CI/CD Flow

This repository builds and pushes container images, creates per-service release tags, and proposes production deployment changes through pull requests in `Delpak/shahrah-gitops`.

## Versioning policy

Production GitOps values use per-service SemVer tags as the deployed image tag.

Each released service gets:

- A Git tag in this repository: `<service>/vMAJOR.MINOR.PATCH`
- A service-qualified SemVer image tag: `docker.io/delpak/godar-ai:<service>-vMAJOR.MINOR.PATCH`
- A service-qualified traceability image tag: `docker.io/delpak/godar-ai:<service>-sha-<commit>`

The SemVer tag is the human-facing deployment version. The SHA tag is retained for auditability and exact source-code traceability.

Examples:

```text
Git tag:    control-plane/v1.4.2
Image tags: docker.io/delpak/godar-ai:control-plane-v1.4.2
            docker.io/delpak/godar-ai:control-plane-sha-012345...
GitOps:     controlPlane.image.tag: v1.4.2
```

## Workflows

### `.github/workflows/gitops-production-release.yml`

Trigger:

- Push to `main` or `master`
- Manual run with optional service selection and SemVer bump type

Flow:

1. Detect changed services.
2. Compute the next SemVer tag per changed service prefix from existing Git tags.
3. Build and push each changed image with both tags when the SemVer tag is new:
   - `vMAJOR.MINOR.PATCH`
   - `sha-<full-commit-sha>`
4. Create annotated Git tags for the SemVer releases.
5. Open a GitOps PR updating only changed service tags in `helm/shahrah/values.production.yaml`.

The workflow keeps service versions independent. For example, `control-plane` and `docs` can move to different versions in the same run.

### `.github/workflows/gitops-promote-production.yml`

Trigger:

- Manual run only

Inputs:

- `service_id`: service to promote
- `image_tag`: SemVer image tag to deploy, for example `v1.2.3`

Flow:

1. Validate that `image_tag` matches `vMAJOR.MINOR.PATCH`.
2. Verify the matching service Git tag exists in this repository.
3. Open a GitOps PR updating only that service's image tag in `helm/shahrah/values.production.yaml`.

Manual promotion intentionally updates one service at a time. A single SemVer tag should not be applied to all services in this independently versioned monorepo.

## Required GitHub secrets

- `REGISTRY_URL`
  Required value: `docker.io`
- `REGISTRY_USERNAME`
  Docker Hub username: `delpak`
- `REGISTRY_PASSWORD`
  Docker Hub personal access token with read/write access to `delpak/godar-ai`.
- `GITOPS_REPO_TOKEN`
  PAT with `contents:write` and `pull_requests:write` on `Delpak/shahrah-gitops`.

## Image tags built in `delpak/godar-ai`

- `control-plane-*`
- `model-discovery-*`
- `model-discovery-migrate-*` (`migrate` target)
- `model-discovery-web-*`
- `support-admin-*`
- `support-admin-migrate-*` (`migrate` target)
- `silk-road-*`
- `tunnel-client-*`
- `tunnel-server-*`
- `goose-*`
- `docs-*`
- `website-*`
- `blog-*`
- `tenant-admin-*`

## SemVer bumping

The production release workflow derives the bump type from Conventional Commit messages:

- `BREAKING CHANGE` or `type!:` -> major
- `feat:` -> minor
- anything else -> patch

Manual workflow runs can override this with `release_type`.

If no previous tag exists for a service prefix, the first default release is `v0.1.0`.

## Release safety rules

- SemVer Git tags are immutable. If a tag already exists for another commit, the workflow fails instead of reusing it.
- Re-running the workflow for the same commit reuses the existing Git tag and safely republishes the same SemVer image tag to the configured registry. This supports registry migrations and recovery from partial pushes.
- GitOps deploys SemVer tags, while SHA tags remain available for audit and emergency diagnosis.
- Rollbacks should be GitOps PRs that restore a service's previous SemVer tag.
- Container registry tag immutability should be enabled for SemVer tags if the registry supports it.

## Production release policy

1. Merge app code to `main` or `master`.
2. CI detects changed services.
3. CI builds and pushes changed images with SemVer and SHA tags.
4. CI creates per-service Git tags.
5. CI opens a production GitOps PR using SemVer image tags.
6. Review and merge the GitOps PR to deploy.
