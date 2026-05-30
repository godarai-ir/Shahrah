FROM node:22-alpine AS build
WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM node:22-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

COPY package.json ./
RUN npm install --omit=dev

COPY --from=build /app/dist ./dist
COPY --from=build /app/public ./public
COPY --from=build /app/astro.config.mjs ./astro.config.mjs

EXPOSE 3000
CMD ["npm", "run", "preview", "--", "--host", "0.0.0.0", "--port", "3000"]
