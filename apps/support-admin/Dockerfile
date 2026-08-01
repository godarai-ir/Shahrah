# syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=1

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/support-admin/package.json ./apps/support-admin/package.json
RUN --mount=type=cache,id=pnpm-cache-support-admin,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter @shahrah/support-admin...

FROM deps AS builder
COPY apps/support-admin ./apps/support-admin
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
ENV DATABASE_URL=postgresql://support_admin:placeholder@localhost:5432/support_admin?schema=public
RUN pnpm --filter @shahrah/support-admin prisma:generate
RUN pnpm --filter @shahrah/support-admin build

FROM deps AS migrate
COPY apps/support-admin ./apps/support-admin
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
ENV DATABASE_URL=postgresql://support_admin:placeholder@localhost:5432/support_admin?schema=public
WORKDIR /app/apps/support-admin
RUN pnpm prisma:generate

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

# A workspace build places the standalone server under apps/support-admin.
COPY --from=builder --chown=nextjs:nextjs /app/apps/support-admin/.next/standalone ./
COPY --from=builder --chown=nextjs:nextjs /app/apps/support-admin/.next/static ./apps/support-admin/.next/static
COPY --from=builder --chown=nextjs:nextjs /app/apps/support-admin/public ./apps/support-admin/public

WORKDIR /app/apps/support-admin
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
