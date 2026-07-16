import { readFileSync } from "node:fs";
import { resolve } from "node:path";

import { describe, expect, it } from "vitest";

const appRoot = resolve(__dirname);
const repositoryRoot = resolve(appRoot, "../..");

describe("tenant-admin reproducible container build", () => {
  it("pins the base image and installs the frozen pnpm dependency graph", () => {
    const dockerfile = readFileSync(resolve(appRoot, "Dockerfile"), "utf8");

    expect(dockerfile).toMatch(
      /^FROM node:22-alpine@sha256:[a-f0-9]{64} AS deps/m,
    );
    expect(dockerfile).toContain(
      "pnpm install --frozen-lockfile --filter @shahrah/tenant-admin...",
    );
    expect(dockerfile).toContain("pnpm --filter @shahrah/tenant-admin build");
    expect(dockerfile).not.toMatch(/\bnpm (?:install|run)\b/);
    expect(dockerfile).toContain("WORKDIR /app/apps/tenant-admin");
    expect(dockerfile).toContain("USER nextjs");
  });

  it("builds from the repository root with a tenant-admin-specific context", () => {
    const workflow = readFileSync(
      resolve(repositoryRoot, ".github/workflows/gitops-production-release.yml"),
      "utf8",
    );
    const dockerignore = readFileSync(
      resolve(appRoot, "Dockerfile.dockerignore"),
      "utf8",
    );

    expect(workflow).toMatch(
      /"id": "tenant-admin",[\s\S]*?"context": "\."/,
    );
    expect(dockerignore).toContain("!pnpm-lock.yaml");
    expect(dockerignore).toContain("!pnpm-workspace.yaml");
    expect(dockerignore).toContain("!apps/tenant-admin/**");
    expect(dockerignore).toContain("apps/tenant-admin/.env.*");
    expect(workflow).toContain("sbom: ${{ matrix.id == 'tenant-admin' }}");
    expect(workflow).toContain(
      "provenance: ${{ matrix.id == 'tenant-admin' && 'mode=max' || 'true' }}",
    );
  });
});
