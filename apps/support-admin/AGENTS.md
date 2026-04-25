# Support Admin Agent Notes

- This app uses modern Next.js App Router.
- Check local Next.js docs in `node_modules/next/dist/docs/` before assuming older APIs still apply.
- Run commands from the repository root with `pnpm --filter @shahrah/support-admin ...`.
- Preserve the internal-ops purpose of this app: support queues, customer review, billing operations, and synced tenant data.
- Avoid introducing local workspace files or separate lockfiles; this app belongs to the root monorepo workspace.
