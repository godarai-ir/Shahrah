FROM node:20-alpine
WORKDIR /app
COPY . .
RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter model-discovery prisma:generate
RUN pnpm --filter model-discovery build
EXPOSE 3002
CMD ["pnpm", "--filter", "model-discovery", "start:prod"]
