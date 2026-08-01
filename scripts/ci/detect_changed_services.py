#!/usr/bin/env python3
"""Detect images that differ from the versions currently deployed in GitOps."""

from __future__ import annotations

import argparse
import functools
import json
import re
import subprocess
from collections.abc import Callable, Iterable
from pathlib import Path
from typing import Any

import yaml

SERVICES = [
    {
        "id": "control-plane",
        "context": ".",
        "dockerfile": "services/control-plane/Dockerfile",
        "target": "",
        "values_path": "controlPlane.image.tag",
        "additional_values_paths": ["metricsDlqReplayer.image.tag"],
        "semver_prefix": "control-plane",
        "paths": ["services/control-plane/"],
    },
    {
        "id": "model-discovery",
        "context": ".",
        "dockerfile": "apps/model-discovery/Dockerfile",
        "target": "",
        "values_path": "modelDiscovery.image.tag",
        "semver_prefix": "model-discovery",
        "paths": ["apps/model-discovery/"],
    },
    {
        "id": "model-discovery-migrate",
        "context": ".",
        "dockerfile": "apps/model-discovery/Dockerfile",
        "target": "migrate",
        "values_path": "migrations.modelDiscovery.image.tag",
        "additional_values_paths": ["migrations.modelDiscoveryCatalogImport.image.tag"],
        "semver_prefix": "model-discovery",
        "paths": ["apps/model-discovery/"],
    },
    {
        "id": "model-discovery-web",
        "context": ".",
        "dockerfile": "apps/model-discovery-web/Dockerfile",
        "target": "",
        "values_path": "modelDiscoveryWeb.image.tag",
        "semver_prefix": "model-discovery-web",
        "paths": ["apps/model-discovery-web/"],
    },
    {
        "id": "support-admin",
        "context": ".",
        "dockerfile": "apps/support-admin/Dockerfile",
        "target": "",
        "values_path": "supportAdmin.image.tag",
        "semver_prefix": "support-admin",
        "paths": ["apps/support-admin/"],
    },
    {
        "id": "support-admin-migrate",
        "context": ".",
        "dockerfile": "apps/support-admin/Dockerfile",
        "target": "migrate",
        "values_path": "migrations.supportAdmin.image.tag",
        "semver_prefix": "support-admin",
        "paths": ["apps/support-admin/"],
    },
    {
        "id": "silk-road",
        "context": ".",
        "dockerfile": "services/silk-road/Dockerfile",
        "target": "",
        "values_path": "silkRoad.image.tag",
        "semver_prefix": "silk-road",
        "paths": ["services/silk-road/"],
    },
    {
        "id": "tunnel-client",
        "context": ".",
        "dockerfile": "services/tunnel-client/Dockerfile",
        "target": "",
        "values_path": "tunnelClient.image.tag",
        "semver_prefix": "tunnel-client",
        "paths": ["services/tunnel-client/"],
    },
    {
        "id": "tunnel-server",
        "context": ".",
        "dockerfile": "services/tunnel-server/Dockerfile",
        "target": "",
        "values_path": "tunnelServer.image.tag",
        "semver_prefix": "tunnel-server",
        "paths": ["services/tunnel-server/"],
    },
    {
        "id": "goose",
        "context": "infra/compose/images/goose-ci",
        "dockerfile": "infra/compose/images/goose-ci/Dockerfile",
        "target": "",
        "values_path": "migrations.mainDb.image.tag",
        "semver_prefix": "goose",
        "paths": ["infra/compose/images/goose-ci/", "db/"],
    },
    {
        "id": "docs",
        "context": ".",
        "dockerfile": "apps/docs/Dockerfile",
        "target": "",
        "values_path": "docs.image.tag",
        "semver_prefix": "docs",
        "paths": ["apps/docs/"],
    },
    {
        "id": "website",
        "context": ".",
        "dockerfile": "apps/website/Dockerfile",
        "target": "",
        "values_path": "website.image.tag",
        "semver_prefix": "website",
        "paths": ["apps/website/"],
    },
    {
        "id": "blog",
        "context": ".",
        "dockerfile": "apps/blog/Dockerfile",
        "target": "",
        "values_path": "blog.image.tag",
        "semver_prefix": "blog",
        "paths": ["apps/blog/"],
    },
    {
        "id": "tenant-admin",
        "context": ".",
        "dockerfile": "apps/tenant-admin/Dockerfile",
        "target": "",
        "values_path": "tenantAdmin.image.tag",
        "semver_prefix": "tenant-admin",
        "paths": ["apps/tenant-admin/"],
    },
]

SHARED_PATHS = [
    "packages/",
    "pnpm-lock.yaml",
    "pnpm-workspace.yaml",
    "package.json",
    "turbo.json",
]

SEMVER_RE = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def run_git(args: list[str], *, check: bool = True) -> str:
    result = subprocess.run(
        ["git", *args],
        check=check,
        text=True,
        capture_output=True,
    )
    return result.stdout.strip()


def nested_value(data: dict[str, Any], dotted_path: str) -> Any:
    current: Any = data
    for key in dotted_path.split("."):
        if not isinstance(current, dict) or key not in current:
            raise ValueError(f"Production values path is missing: {dotted_path}")
        current = current[key]
    return current


def resolve_deployed_commit(
    service: dict[str, Any],
    production_values: dict[str, Any],
    head_sha: str,
) -> tuple[str, str]:
    deployed_tag = nested_value(production_values, service["values_path"])
    if not isinstance(deployed_tag, str) or not SEMVER_RE.fullmatch(deployed_tag):
        raise ValueError(
            f"Invalid deployed tag at {service['values_path']}: {deployed_tag!r}"
        )

    git_tag = f"{service['semver_prefix']}/{deployed_tag}"
    try:
        deployed_sha = run_git(["rev-parse", "--verify", f"{git_tag}^{{}}"])
    except subprocess.CalledProcessError as exc:
        raise ValueError(f"Deployed Git tag does not exist: {git_tag}") from exc

    ancestor = subprocess.run(
        ["git", "merge-base", "--is-ancestor", deployed_sha, head_sha],
        check=False,
        capture_output=True,
    )
    if ancestor.returncode != 0:
        raise ValueError(f"Deployed tag {git_tag} is not an ancestor of {head_sha}")

    return deployed_tag, deployed_sha


def path_matches(path: str, prefixes: Iterable[str]) -> bool:
    return any(path == prefix or path.startswith(prefix) for prefix in prefixes)


def select_changed_services(
    services: list[dict[str, Any]],
    manual_services: str,
    resolve_base: Callable[[dict[str, Any]], tuple[str, str]],
    changed_files_for_base: Callable[[str], list[str]],
) -> tuple[list[str], dict[str, str], list[str]]:
    id_to_service = {service["id"]: service for service in services}
    all_service_ids = [service["id"] for service in services]
    selected: list[str] = []
    deployed_bases: dict[str, str] = {}
    unresolved_bases: set[str] = set()
    warnings: list[str] = []

    if manual_services:
        for raw in manual_services.split(","):
            service_id = raw.strip()
            if not service_id:
                continue
            if service_id == "all":
                selected = list(all_service_ids)
                break
            if service_id not in id_to_service:
                raise ValueError(f"Unknown service id: {service_id}")
            if service_id not in selected:
                selected.append(service_id)
    else:
        for service in services:
            service_id = service["id"]
            try:
                _deployed_tag, deployed_sha = resolve_base(service)
                deployed_bases[service_id] = deployed_sha
                changed_files = changed_files_for_base(deployed_sha)
            except (ValueError, subprocess.CalledProcessError) as exc:
                selected.append(service_id)
                unresolved_bases.add(service_id)
                warnings.append(f"{service_id}: {exc}; selecting conservatively")
                continue

            relevant_paths = [*service["paths"], *SHARED_PATHS]
            if any(path_matches(path, relevant_paths) for path in changed_files):
                selected.append(service_id)

    # Manual releases still resolve deployed bases so automatic bump detection uses
    # the production range rather than an arbitrary recent commit window.
    for service_id in selected:
        if service_id in deployed_bases or service_id in unresolved_bases:
            continue
        try:
            _deployed_tag, deployed_sha = resolve_base(id_to_service[service_id])
            deployed_bases[service_id] = deployed_sha
        except ValueError as exc:
            warnings.append(
                f"{service_id}: {exc}; bump detection will use recent history"
            )

    return selected, deployed_bases, warnings


def detect_bump_type(
    manual_bump: str,
    deployed_bases: Iterable[str],
    head_sha: str,
) -> str:
    if manual_bump and manual_bump != "auto":
        if manual_bump not in {"patch", "minor", "major"}:
            raise ValueError(f"Unknown release type: {manual_bump}")
        return manual_bump

    messages: list[str] = []
    for base_sha in dict.fromkeys(deployed_bases):
        try:
            messages.append(
                run_git(["log", "--format=%s%n%b", f"{base_sha}..{head_sha}"])
            )
        except subprocess.CalledProcessError:
            continue

    if not messages:
        messages.append(run_git(["log", "-n", "50", "--format=%s%n%b", head_sha]))
    message_text = "\n".join(messages)

    if "BREAKING CHANGE" in message_text or re.search(
        r"^[a-z]+(\([^\)]+\))?!:", message_text, re.MULTILINE
    ):
        return "major"
    if re.search(r"^feat(\([^\)]+\))?:", message_text, re.MULTILINE):
        return "minor"
    return "patch"


@functools.lru_cache(maxsize=1)
def semver_tags() -> dict[str, list[tuple[tuple[int, int, int], str]]]:
    """Load all service versions and peeled commit SHAs in one Git process."""
    tags_by_prefix: dict[str, list[tuple[tuple[int, int, int], str]]] = {}
    output = run_git(
        [
            "for-each-ref",
            "--format=%(refname:short)|%(*objectname)|%(objectname)",
            "refs/tags",
        ]
    )
    for line in output.splitlines():
        tag, peeled_sha, object_sha = line.split("|", 2)
        match = re.fullmatch(r"(.+)/v(\d+)\.(\d+)\.(\d+)", tag)
        if not match:
            continue
        prefix = match.group(1)
        version = tuple(int(match.group(index)) for index in (2, 3, 4))
        tags_by_prefix.setdefault(prefix, []).append(
            (version, peeled_sha or object_sha)
        )
    return tags_by_prefix


def latest_semver(prefix: str) -> tuple[int, int, int] | None:
    versions = [version for version, _sha in semver_tags().get(prefix, [])]
    return max(versions, default=None)


def latest_semver_for_commit(
    prefix: str, commit_sha: str
) -> tuple[int, int, int] | None:
    versions = [
        version
        for version, tag_sha in semver_tags().get(prefix, [])
        if tag_sha == commit_sha
    ]
    return max(versions, default=None)


def bump_semver(
    current: tuple[int, int, int] | None, bump: str
) -> tuple[int, int, int]:
    if current is None:
        if bump == "major":
            return (1, 0, 0)
        return (0, 1, 0)

    major, minor, patch = current
    if bump == "major":
        return (major + 1, 0, 0)
    if bump == "minor":
        return (major, minor + 1, 0)
    return (major, minor, patch + 1)


def build_matrix(
    selected: list[str], head_sha: str, bump_type: str
) -> list[dict[str, Any]]:
    id_to_service = {service["id"]: service for service in SERVICES}
    version_cache: dict[str, str] = {}
    matrix: list[dict[str, Any]] = []

    for service_id in selected:
        service = id_to_service[service_id]
        prefix = service["semver_prefix"]
        if prefix not in version_cache:
            existing = latest_semver_for_commit(prefix, head_sha)
            if existing is None:
                existing = bump_semver(latest_semver(prefix), bump_type)
            version_cache[prefix] = f"v{existing[0]}.{existing[1]}.{existing[2]}"

        matrix.append(
            {
                "id": service["id"],
                "context": service["context"],
                "dockerfile": service["dockerfile"],
                "target": service["target"],
                "values_path": service["values_path"],
                "additional_values_paths": service.get("additional_values_paths", []),
                "semver_prefix": prefix,
                "semver_tag": version_cache[prefix],
            }
        )

    return matrix


def detect(
    head_sha: str,
    production_values_path: Path,
    manual_services: str = "",
    manual_bump: str = "auto",
) -> dict[str, Any]:
    with production_values_path.open(encoding="utf-8") as production_file:
        production_values = yaml.safe_load(production_file)
    if not isinstance(production_values, dict):
        raise TypeError("Production values must contain a YAML mapping")

    diff_cache: dict[str, list[str]] = {}

    def resolve_base(service: dict[str, Any]) -> tuple[str, str]:
        return resolve_deployed_commit(service, production_values, head_sha)

    def changed_files_for_base(base_sha: str) -> list[str]:
        if base_sha not in diff_cache:
            output = run_git(["diff", "--name-only", base_sha, head_sha])
            diff_cache[base_sha] = [line for line in output.splitlines() if line]
        return diff_cache[base_sha]

    selected, deployed_bases, warnings = select_changed_services(
        SERVICES,
        manual_services.strip(),
        resolve_base,
        changed_files_for_base,
    )
    bump_type = detect_bump_type(
        manual_bump.strip() or "auto",
        (
            deployed_bases[service_id]
            for service_id in selected
            if service_id in deployed_bases
        ),
        head_sha,
    )
    matrix = build_matrix(selected, head_sha, bump_type)
    release_tag = f"sha-{head_sha}"
    summary = {
        "release_tag": release_tag,
        "bump_type": bump_type,
        "services": [item["id"] for item in matrix],
        "semver_tags": {item["id"]: item["semver_tag"] for item in matrix},
    }
    return {
        "release_tag": release_tag,
        "matrix": matrix,
        "changed_services": ",".join(item["id"] for item in matrix),
        "build_summary": summary,
        "warnings": warnings,
    }


def write_github_outputs(path: Path, result: dict[str, Any]) -> None:
    compact = lambda value: json.dumps(value, separators=(",", ":"))
    with path.open("a", encoding="utf-8") as output_file:
        output_file.write(f"release_tag={result['release_tag']}\n")
        output_file.write(f"matrix={compact(result['matrix'])}\n")
        output_file.write(f"changed_services={result['changed_services']}\n")
        output_file.write(f"build_summary={compact(result['build_summary'])}\n")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--head-sha", required=True)
    parser.add_argument("--production-values", required=True, type=Path)
    parser.add_argument("--manual-services", default="")
    parser.add_argument("--manual-bump", default="auto")
    parser.add_argument("--github-output", type=Path)
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    result = detect(
        args.head_sha,
        args.production_values,
        args.manual_services,
        args.manual_bump,
    )
    for warning in result["warnings"]:
        print(f"::warning::{warning}")
    if args.github_output:
        write_github_outputs(args.github_output, result)
    print(json.dumps(result["build_summary"], separators=(",", ":")))


if __name__ == "__main__":
    main()
