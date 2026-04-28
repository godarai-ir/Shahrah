> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Using tfy apply

> preview and apply changes to your resources using tfy apply

## What is `tfy apply`?

`tfy apply` is a CLI command that lets you create or update resources, such as deployments, workspaces, clusters, and provider accounts, by applying a manifest file.

A manifest is a YAML file that describes the resource you want to create or update.

> **Note:** If you haven't set up the CLI yet, see the [CLI setup guide](/docs/setup-cli).

## Example: Deploying an OpenAI Provider Account

Here's a sample manifest for creating an OpenAI provider account.

```yaml lines theme={"dark"}
name: openai-account-cli
type: provider-account/openai
collaborators:
  - subject: user:gowtham.v@truefoundry.com
    role_id: provider-account-manager
  - subject: team:everyone
    role_id: provider-account-access
auth_data:
  api_key: tfy-secret://truefoundry:openai-group:API_KEY
  type: api-key
integrations:
  - type: integration/model/openai
    name: gpt-4
    model_id: gpt-4
    model_types:
      - chat
    cost:
      value:
        input: 0.03
        output: 0.06
      metric: per_1000_tokens
  - type: integration/model/openai
    name: gpt-4o
    model_id: gpt-4o
    model_types:
      - chat
    cost:
      value:
        input: 0.0025
        output: 0.01
      metric: per_1000_tokens
```

## What is Dry Run?

The `--dry-run` flag simulates the apply process without making any changes. This is useful to validate your manifest and see what would happen if you actually applied it.

## What is Show Diff?

The `--show-diff` flag, when used with `--dry-run`, displays the differences between your manifest and the currently deployed resources. This helps you review what changes will be made before applying them.

## Previewing Changes Before Applying

To see what changes will be made without actually applying them, you can use the following command:

```bash lines theme={"dark"}
tfy apply -f <your-manifest-file>.yaml --dry-run --show-diff
```

This will show you a diff of what will change, without updating anything.

<Frame caption="Diff output for dry run">
  <img src="https://mintcdn.com/truefoundry/2MMcllD7kMlpnaWX/images/tfy-apply-diff-example.png?fit=max&auto=format&n=2MMcllD7kMlpnaWX&q=85&s=846b60348ffc041ba7881a162c0fb230" width="2022" height="1164" data-path="images/tfy-apply-diff-example.png" />
</Frame>

## Applying the changes

To apply the changes in your manifest, run:

```bash lines theme={"dark"}
tfy apply -f <your-manifest-file>.yaml
```

This will actually create or update the resource as described in your YAML file.

## Directory Mode

Instead of specifying individual manifest files with `-f`, you can apply all manifests from a directory using the `-d` (or `--dir` / `--directory`) flag.

```bash lines theme={"dark"}
tfy apply --dir ./manifests
```

<Note>
  The `-f` and `--dir` flags are mutually exclusive. You must use exactly one of them.
</Note>

### How Directory Mode Works

* All `*.yaml` and `*.yml` files under the specified directory are discovered recursively
* Files are automatically sorted by manifest type to respect dependencies (e.g., team is applied before workspace is applied)

### Previewing Directory Changes

You can combine directory mode with dry run to preview all changes:

```bash lines theme={"dark"}
tfy apply --dir ./manifests --dry-run --show-diff
```

## Git-Based Diff Mode

When working with manifests stored in a git repository, you can use `--diffs-only` to apply only the files that have changed compared to a specific git reference.

```bash lines theme={"dark"}
tfy apply --dir ./manifests --diffs-only
```

This compares your working directory against `HEAD` and only applies YAML files that git considers modified or newly added.

### Specifying a Comparison Reference

Use `--ref` to compare against a specific commit, branch, or tag instead of `HEAD`:

```bash lines theme={"dark"}
tfy apply --dir ./manifests --diffs-only --ref main
```

```bash lines theme={"dark"}
tfy apply --dir ./manifests --diffs-only --ref v1.2.0
```

<Note>
  The `--ref` flag is only valid when used with `--diffs-only`.
</Note>

### What Gets Included

* **Modified files**: YAML files that have changes compared to the reference
* **New files**: Untracked YAML files under the directory are treated as new and included
* **Deleted files**: Tracked in a separate set (used with `--sync`)

## Sync Mode

The `--sync` flag enables cleanup of resources whose manifest files have been deleted from git. After applying changed manifests, it will delete platform resources that correspond to removed YAML files.

```bash lines theme={"dark"}
tfy apply --dir ./manifests --diffs-only --sync
```

<Warning>
  The `--sync` flag requires both `--dir` and `--diffs-only` to be set.
</Warning>

<Note>
  If a deleted file's content cannot be parsed from the git history, that entry may be skipped during deletion.
</Note>

## Quick Reference

| Flags                                    | Effect                                                                                              |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `-f manifest.yaml`                       | Apply a single manifest file                                                                        |
| `-f manifest1.yaml -f manifest2.yaml`    | Apply multiple manifest files                                                                       |
| `--dir ./manifests`                      | Apply all YAML files under the directory, sorted by dependency order                                |
| `--dir ./manifests --dry-run`            | Apply all YAML files under the directory, sorted by dependency order, but simulate without applying |
| `--dir ./manifests --diffs-only`         | Apply only git-changed YAML files under the directory (vs HEAD)                                     |
| `--dir ./manifests --diffs-only --ref v` | Apply only git-changed YAML files (vs ref v)                                                        |
| `--dir ./manifests --diffs-only --sync`  | Apply changed files, then delete resources for YAML files removed in git                            |
