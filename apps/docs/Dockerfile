FROM node:20-alpine AS build
WORKDIR /app

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

FROM node:20-alpine AS runtime
WORKDIR /app

COPY package.json ./
RUN npm install --omit=dev

COPY --from=build /app/build ./build
EXPOSE 3000

CMD ["npm", "run", "serve", "--", "--host", "0.0.0.0", "--port", "3000"]
