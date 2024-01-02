FROM node:lts-alpine AS www-builder
WORKDIR /app
COPY ./package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM golang:alpine AS api-builder
WORKDIR /app
COPY api /app/api
COPY go.* /app/
COPY *.go /app/
RUN go build -buildmode=exe -buildvcs=false -mod=readonly -v

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=www-builder /app/build /app/build
COPY --from=api-builder /app/dev-website /app/api
EXPOSE 8080
CMD ./api -docker
