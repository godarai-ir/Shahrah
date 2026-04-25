# Tenant Admin Agent Notes

- This app uses modern Next.js App Router.
- Check local Next.js docs in `node_modules/next/dist/docs/` before assuming older APIs still apply.
- Run commands from the repository root with `pnpm --filter @shahrah/tenant-admin ...`.
- Preserve the tenant-facing purpose of this app: usage, keys, billing, routing, and settings.
- Avoid introducing local workspace files or separate lockfiles; this app belongs to the root monorepo workspace.
