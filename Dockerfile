FROM node:lts-alpine AS builder
WORKDIR /app
COPY ./package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:lts-alpine AS runner
WORKDIR /app
COPY --from=builder /app/build .
COPY --from=builder /app/package.json .
COPY --from=builder /app/package-lock.json .
RUN npm ci --omit dev
EXPOSE 3000
CMD ["node", "."]
