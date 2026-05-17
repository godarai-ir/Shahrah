FROM node:20-alpine
WORKDIR /app
COPY . .
RUN corepack enable && pnpm install --frozen-lockfile
RUN pnpm --filter @shahrah/support-admin build
EXPOSE 3000
CMD ["pnpm", "--filter", "@shahrah/support-admin", "start"]
