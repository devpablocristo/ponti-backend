# App
FROM node:20.17.0 AS ui-builder

WORKDIR /app

COPY ui/package*.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY ui/ ./
RUN yarn build

# Server
FROM node:20.17.0 AS api-builder

WORKDIR /server

COPY api/package*.json api/yarn.lock ./
RUN yarn install --frozen-lockfile

COPY api/ ./
RUN yarn build

FROM node:20.17.0 AS final

WORKDIR /app

COPY --from=api-builder /server/dist ./dist
COPY --from=api-builder /server/node_modules ./node_modules

COPY --from=ui-builder /app/dist ./dist/public

ENV NODE_ENV=development
ENV PORT=3000

EXPOSE 3000

CMD ["node", "dist/index.js"]
