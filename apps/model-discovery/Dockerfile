# syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=1

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/model-discovery/package.json ./apps/model-discovery/package.json
RUN --mount=type=cache,id=pnpm-cache-model-discovery,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter model-discovery...

FROM deps AS builder
COPY apps/model-discovery ./apps/model-discovery
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
ENV DATABASE_URL=postgresql://postgres:postgres@localhost:5432/model_discovery?schema=public
RUN pnpm --filter model-discovery prisma:generate
RUN pnpm --filter model-discovery build

FROM deps AS migrate
COPY apps/model-discovery ./apps/model-discovery
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
WORKDIR /app/apps/model-discovery
RUN pnpm prisma:generate

FROM node:22-alpine AS prod-deps
WORKDIR /app
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=1

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/model-discovery/package.json ./apps/model-discovery/package.json
RUN --mount=type=cache,id=pnpm-cache-model-discovery,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --prod --frozen-lockfile --filter model-discovery...

FROM prod-deps AS runner
WORKDIR /app/apps/model-discovery
ENV NODE_ENV=production

RUN addgroup -S nodejs && adduser -S nestjs -G nodejs

COPY --from=builder /app/apps/model-discovery/dist ./dist
COPY --from=builder /app/apps/model-discovery/prisma ./prisma
RUN mkdir -p /app/apps/model-discovery/uploads && chown -R nestjs:nodejs /app/apps/model-discovery/uploads

USER nestjs
EXPOSE 3002
CMD ["node", "dist/main"]
