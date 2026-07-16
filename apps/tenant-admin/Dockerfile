FROM node:22-alpine@sha256:16e22a550f3863206a3f701448c45f7912c6896a62de43add43bb9c86130c3e2 AS deps
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/tenant-admin/package.json ./apps/tenant-admin/package.json
RUN pnpm install --frozen-lockfile --filter @shahrah/tenant-admin...

FROM deps AS build
ARG APP_VERSION=0.0.0-dev
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_APP_VERSION=$APP_VERSION

COPY apps/tenant-admin ./apps/tenant-admin
RUN pnpm --filter @shahrah/tenant-admin build

FROM node:22-alpine@sha256:16e22a550f3863206a3f701448c45f7912c6896a62de43add43bb9c86130c3e2 AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

# A workspace build places the standalone server under apps/tenant-admin.
COPY --from=build /app/apps/tenant-admin/.next/standalone ./
COPY --from=build /app/apps/tenant-admin/.next/static ./apps/tenant-admin/.next/static
COPY --from=build /app/apps/tenant-admin/public ./apps/tenant-admin/public
COPY --from=build /app/apps/tenant-admin/messages ./apps/tenant-admin/messages
RUN chown -R nextjs:nextjs /app

WORKDIR /app/apps/tenant-admin
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
