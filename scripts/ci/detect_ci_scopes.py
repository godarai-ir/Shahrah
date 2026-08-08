#!/usr/bin/env python3
"""Select the expensive CI scopes affected by a Git diff."""

from __future__ import annotations

import argparse
import json
import subprocess
from collections.abc import Iterable
from pathlib import Path

ALL_SCOPES = ("go", "tenant", "model", "javascript", "tenant_dom")

ALWAYS_ALL_PATHS = (
    ".github/workflows/ci.yml",
    ".github/actions/",
    "scripts/ci/detect_ci_scopes.py",
    "scripts/ci/test_detect_ci_scopes.py",
)

NODE_SHARED_PATHS = (
    "package.json",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
)

GO_PATHS = (
    "services/control-plane/",
    "services/silk-road/",
    "packages/contracts/",
    "db/",
    ".golangci.yml",
    "Makefile",
)

TENANT_PATHS = (
    "apps/tenant-admin/",
    "packages/web/",
    *NODE_SHARED_PATHS,
)

MODEL_PATHS = (
    "apps/model-discovery/",
    *NODE_SHARED_PATHS,
)


def path_matches(path: str, patterns: Iterable[str]) -> bool:
    normalized = path.replace("\\", "/")
    return any(
        normalized == pattern or normalized.startswith(pattern)
        for pattern in patterns
    )


def all_scopes(enabled: bool) -> dict[str, bool]:
    return {scope: enabled for scope in ALL_SCOPES}


def detect_scopes(changed_files: Iterable[str], *, force_all: bool = False) -> dict[str, bool]:
    if force_all:
        return all_scopes(True)

    paths = [path.replace("\\", "/") for path in changed_files]
    if any(path_matches(path, ALWAYS_ALL_PATHS) for path in paths):
        return all_scopes(True)

    go_changed = any(path_matches(path, GO_PATHS) for path in paths)
    tenant_changed = any(path_matches(path, TENANT_PATHS) for path in paths)
    model_changed = any(path_matches(path, MODEL_PATHS) for path in paths)

    return {
        "go": go_changed,
        "tenant": tenant_changed,
        "model": model_changed,
        "javascript": tenant_changed or model_changed,
        "tenant_dom": tenant_changed,
    }


def git_changed_files(base_sha: str, head_sha: str) -> list[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", base_sha, head_sha],
        check=True,
        text=True,
        capture_output=True,
    )
    return [line for line in result.stdout.splitlines() if line]


def is_missing_base(base_sha: str) -> bool:
    return not base_sha or set(base_sha) == {"0"}


def write_github_outputs(path: Path, scopes: dict[str, bool]) -> None:
    with path.open("a", encoding="utf-8") as output_file:
        for name in ALL_SCOPES:
            output_file.write(f"{name}={str(scopes[name]).lower()}\n")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-sha", default="")
    parser.add_argument("--head-sha", required=True)
    parser.add_argument("--github-output", type=Path)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    force_all = is_missing_base(args.base_sha)
    if force_all:
        print("::notice::No usable base SHA; selecting every CI scope")
        changed_files: list[str] = []
    else:
        try:
            changed_files = git_changed_files(args.base_sha, args.head_sha)
        except subprocess.CalledProcessError as exc:
            print(
                "::warning::Could not compute the Git diff; "
                "selecting every CI scope conservatively"
            )
            print(exc.stderr.strip())
            force_all = True
            changed_files = []

    scopes = detect_scopes(changed_files, force_all=force_all)
    if args.github_output:
        write_github_outputs(args.github_output, scopes)
    print(json.dumps({"changed_files": changed_files, "scopes": scopes}))


if __name__ == "__main__":
    main()
