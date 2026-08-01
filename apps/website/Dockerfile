# syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@10.0.0 --activate

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/website/package.json ./apps/website/package.json
RUN --mount=type=cache,id=pnpm-cache-website,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --frozen-lockfile --filter @shahrah/website...

FROM deps AS build
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

COPY apps/website ./apps/website
RUN pnpm --filter @shahrah/website build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME=0.0.0.0
ENV PORT=3000

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

# A workspace build places the standalone server under apps/website.
COPY --from=build --chown=nextjs:nextjs /app/apps/website/.next/standalone ./
COPY --from=build --chown=nextjs:nextjs /app/apps/website/.next/static ./apps/website/.next/static
COPY --from=build --chown=nextjs:nextjs /app/apps/website/public ./apps/website/public

RUN mkdir -p /app/apps/website/data && chown nextjs:nextjs /app/apps/website/data

WORKDIR /app/apps/website
USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
