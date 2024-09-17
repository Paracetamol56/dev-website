# Dev

A website to showcase my small side projects and blog posts.

## Deployment

### Tech stack

- [Svelte](https://svelte.dev/)
- [SvelteKit](https://kit.svelte.dev/)
- [MeltUI](https://melt-ui.com/)
- [MongoDB](https://www.mongodb.com/)
- [Go](https://golang.org/)
- [Docker](https://www.docker.com/)

### Local development

#### Requirements
- [Node.js](https://nodejs.org/)
- [Go](https://golang.org/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)

#### Steps

1. Clone the repository
```sh
git clone git@github.com:Paracetamol56/dev-website.git
```

2. Install dependencies
```sh
# Go
go mod download
# Js
npm install
```

3. Start the DB in a container
```sh
docker-compose up -d db
```

3. Start the development server
```sh
npm run dev
```
in another terminal
```sh
go run main.go
```

4. Open the browser at [http://localhost:3000](http://localhost:3000)

### Production

Docker images are automatically built and pushed to [GitHub Container Registry](https://github.com/Paracetamol56/dev-website/pkgs/container/dev-website) on every push to the master branch.

