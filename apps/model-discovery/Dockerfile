## syntax=docker/dockerfile:1.7
FROM node:22-alpine AS builder
WORKDIR /app

COPY package.json ./
RUN corepack enable && corepack prepare pnpm@11.1.3 --activate
RUN --mount=type=cache,id=pnpm-store-model-discovery,target=/root/.local/share/pnpm/store \
    pnpm install

COPY . .
RUN pnpm prisma:generate
RUN pnpm build
RUN --mount=type=cache,id=pnpm-store-model-discovery,target=/root/.local/share/pnpm/store \
    pnpm prune --prod

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

RUN addgroup -S nodejs && adduser -S nestjs -G nodejs

COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/prisma ./prisma

USER nestjs
EXPOSE 3002
CMD ["node", "dist/main"]
