FROM node:lts-alpine AS www-builder
COPY --from=oven/bun:1-alpine /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /app
COPY ./package.json ./bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM golang:alpine AS api-builder
WORKDIR /app
COPY api .
RUN go build -buildmode=exe -buildvcs=false -mod=readonly ./cmd/main.go

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=www-builder /app/build /app/build
COPY --from=api-builder /app/main /app/api
EXPOSE 8080
CMD ./api -docker
