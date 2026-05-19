## syntax=docker/dockerfile:1.7
FROM node:22-alpine AS deps
WORKDIR /app
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=1

COPY package.json ./
RUN --mount=type=cache,id=npm-cache-model-discovery,target=/root/.npm \
    npm install

FROM deps AS builder
COPY . .
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=
RUN npm run prisma:generate
RUN npm run build
RUN npm prune --omit=dev

FROM deps AS migrate
COPY . .
ENV PRISMA_SKIP_POSTINSTALL_GENERATE=

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

RUN addgroup -S nodejs && adduser -S nestjs -G nodejs

COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/prisma ./prisma
RUN mkdir -p /app/uploads && chown -R nestjs:nodejs /app/uploads

USER nestjs
EXPOSE 3002
CMD ["node", "dist/main"]
