# syntax=docker/dockerfile:1.7
FROM node:20-alpine AS deps
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/docs/package.json ./apps/docs/package.json
COPY packages/web/theme-sync/package.json ./packages/web/theme-sync/package.json
COPY packages/web/theme-sync ./packages/web/theme-sync
RUN --mount=type=cache,id=pnpm-cache-docs,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter @shahrah/docs...

FROM deps AS build
COPY apps/docs ./apps/docs
COPY packages/web/theme-sync ./packages/web/theme-sync
RUN pnpm --filter @shahrah/docs build

FROM node:20-alpine AS prod-deps
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/docs/package.json ./apps/docs/package.json
COPY packages/web/theme-sync/package.json ./packages/web/theme-sync/package.json
COPY packages/web/theme-sync ./packages/web/theme-sync
RUN --mount=type=cache,id=pnpm-cache-docs,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --prod --frozen-lockfile --filter @shahrah/docs...

FROM prod-deps AS runtime
WORKDIR /app/apps/docs

COPY --from=build /app/apps/docs/build ./build
COPY --from=build /app/apps/docs/docusaurus.config.ts ./docusaurus.config.ts
COPY --from=build /app/apps/docs/sidebars.ts ./sidebars.ts
COPY --from=build /app/apps/docs/src ./src
COPY --from=build /app/apps/docs/static ./static

EXPOSE 3000
CMD ["node", "./node_modules/@docusaurus/core/bin/docusaurus.mjs", "serve", "--host", "0.0.0.0", "--port", "3000"]
