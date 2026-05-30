# GitOps CI/CD Flow

This repository builds and pushes immutable container images, then proposes deployment changes via pull request in `Delpak/shahrah-gitops`.

## Workflows

- `.github/workflows/gitops-staging-release.yml`
  - Trigger: push to `main`/`master` or manual run.
  - Creates one shared `release_tag` (`sha-<full-commit-sha>`) per run.
  - Builds and pushes all images with that exact shared tag.
  - Opens a PR in GitOps repo updating `helm/shahrah/values.staging.yaml` global `imageTag`.

- `.github/workflows/gitops-promote-production.yml`
  - Trigger: manual.
  - Input: `image_tag` (must start with `sha-`).
  - Opens a PR in GitOps repo updating `helm/shahrah/values.production.yaml` global `imageTag`.

## Required GitHub Secrets (app repo)

- `REGISTRY_URL`  
  Example: `registry.distr.sh`
- `REGISTRY_NAMESPACE`  
  Example: `godar`
- `REGISTRY_USERNAME`
- `REGISTRY_PASSWORD`
- `GITOPS_REPO_TOKEN`  
  PAT with `contents:write` + `pull_requests:write` on `Delpak/shahrah-gitops`.

## Images built

- `shahrah/control-plane`
- `shahrah/model-discovery`
- `shahrah/model-discovery-migrate` (`migrate` target)
- `shahrah/support-admin`
- `shahrah/silk-road`
- `shahrah/tunnel-client`
- `shahrah/tunnel-server`
- `shahrah/goose`

All are pushed as:

`<REGISTRY_URL>/<REGISTRY_NAMESPACE>/<image>:sha-<commit>`

## Promotion policy

1. Merge app code.
2. CI builds/pushes immutable SHA images.
3. CI opens staging PR in GitOps repo.
4. Validate staging.
5. Manually run production promotion workflow with same SHA tag.
