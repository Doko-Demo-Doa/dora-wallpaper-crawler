# Bun with Elysia & Drizzle

This is a Bun server using Elysia for the API and Drizzle ORM for a Postgres DB. It crawls data from https://dora-world.com/wallpaper and upload it to my personal Uploadthing directory.

## Development

1. Install dependencies

```
bun install
```

2. Start a Docker container for the Postgres database

```
docker run --name dev-postgres -p 5432:5432 -e POSTGRES_PASSWORD=12345678 -d postgres
```

3. Add the `DATABASE_URL` env variable to a _.env_ file

```
DATABASE_URL="postgresql://postgres:12345678@localhost:5432/postgres?schema=public"
UPLOADTHING_TOKEN="YOUR_KEY_HERE"
```

4. Add your tables to `src/db/schema/index.ts`.

5. Generate migrations

```shell
bun run generate
```

6. Run migrations

```shell
bun run migrate
```

7. To start the development server run:

```bash
bun run dev
```

Open http://localhost:3000/ with your browser to see the result.
