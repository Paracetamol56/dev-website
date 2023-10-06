FROM oven/bun:latest AS builder

WORKDIR /app
COPY . .
RUN bun install
RUN bun run build

FROM oven/bun:latest AS runner

WORKDIR /app
COPY --from=builder /app/build .
ENTRYPOINT ["bun", "./index.js"]
EXPOSE 3000
