# syntax=docker/dockerfile:1.7
FROM node:22-alpine@sha256:16e22a550f3863206a3f701448c45f7912c6896a62de43add43bb9c86130c3e2 AS deps
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/model-discovery-web/package.json ./apps/model-discovery-web/package.json
COPY packages/web/theme-sync/package.json ./packages/web/theme-sync/package.json
RUN --mount=type=cache,id=pnpm-cache-model-discovery-web,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter @shahrah/model-discovery-web...

FROM deps AS builder
COPY apps/model-discovery-web ./apps/model-discovery-web
COPY packages/web/theme-sync ./packages/web/theme-sync
RUN pnpm --filter @shahrah/model-discovery-web build

FROM node:22-alpine@sha256:16e22a550f3863206a3f701448c45f7912c6896a62de43add43bb9c86130c3e2 AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

# A workspace build places the standalone server under apps/model-discovery-web.
COPY --from=builder --chown=nextjs:nextjs /app/apps/model-discovery-web/.next/standalone ./
COPY --from=builder --chown=nextjs:nextjs /app/apps/model-discovery-web/.next/static ./apps/model-discovery-web/.next/static
COPY --from=builder --chown=nextjs:nextjs /app/apps/model-discovery-web/public ./apps/model-discovery-web/public

WORKDIR /app/apps/model-discovery-web
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
