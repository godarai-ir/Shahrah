## syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=1

COPY package.json ./
RUN --mount=type=cache,id=npm-cache-support-admin,target=/root/.npm \
    npm install

FROM deps AS builder
COPY . ./
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
ENV DATABASE_URL=postgresql://support_admin:placeholder@localhost:5432/support_admin?schema=public
RUN npm run prisma:generate
RUN npm run build

FROM deps AS migrate
COPY . ./
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
ENV DATABASE_URL=postgresql://support_admin:placeholder@localhost:5432/support_admin?schema=public
RUN npm run prisma:generate

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

COPY --from=builder --chown=nextjs:nextjs /app/public ./public
COPY --from=builder --chown=nextjs:nextjs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nextjs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
