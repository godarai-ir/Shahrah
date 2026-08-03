# syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/blog/package.json ./apps/blog/package.json
COPY packages/web/theme-sync/package.json ./packages/web/theme-sync/package.json
RUN --mount=type=cache,id=pnpm-cache-blog,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter @shahrah/blog...

FROM deps AS build
COPY apps/blog ./apps/blog
COPY packages/web/theme-sync ./packages/web/theme-sync
RUN pnpm --filter @shahrah/blog build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV ASTRO_TELEMETRY_DISABLED=1

COPY --from=build /app/apps/blog/dist ./dist
COPY --from=build /app/apps/blog/server.mjs ./server.mjs

EXPOSE 3000
CMD ["node", "server.mjs"]
