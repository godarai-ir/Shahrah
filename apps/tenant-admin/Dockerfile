FROM node:22-alpine AS deps
WORKDIR /app

COPY package.json ./
RUN npm install --no-audit --no-fund --package-lock=false

FROM node:22-alpine AS build
WORKDIR /app
ARG APP_VERSION=0.0.0-dev
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_APP_VERSION=$APP_VERSION

COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -S nextjs && adduser -S nextjs -G nextjs

# standalone output contains the minimal traced runtime dependencies.
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
COPY --from=build /app/messages ./messages
RUN chown -R nextjs:nextjs /app

USER nextjs
EXPOSE 3000
CMD ["node", "server.js"]
