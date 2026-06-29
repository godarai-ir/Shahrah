FROM node:22-alpine AS build
WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production
ENV ASTRO_TELEMETRY_DISABLED=1

COPY --from=build /app/dist ./dist
COPY server.mjs ./server.mjs

EXPOSE 3000
CMD ["node", "server.mjs"]
