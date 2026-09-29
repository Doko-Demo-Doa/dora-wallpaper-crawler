# Dora wallpaper crawler

A Go service that serves the wallpaper API and crawls [Dora World's wallpaper page](https://dora-world.com/wallpaper) hourly. New images are uploaded to UploadThing and recorded in PostgreSQL.

## Requirements

- Go 1.24+
- PostgreSQL
- `DATABASE_URL`
- `UPLOADTHING_TOKEN` for uploads (the read API can run without it)

Start PostgreSQL for local development:

```sh
docker run --name dev-postgres -p 5432:5432 -e POSTGRES_PASSWORD=12345678 -d postgres
```

Set environment variables, then run:

```sh
export DATABASE_URL='postgresql://postgres:12345678@localhost:5432/postgres'
export UPLOADTHING_TOKEN='YOUR_KEY_HERE'
go run .
```

The service listens on port `3000` by default. Set `PORT` to change it. It applies the wallpaper table migration at startup and starts a crawl immediately, then repeats hourly.

## API

- `GET /api/` returns desktop wallpaper source URLs that are not yet stored.
- `GET /api/all` returns stored wallpaper records.
- `GET /api/list-files` returns files under `./temp`.

The SQL migration history is in `drizzle/` for reference. Database setup on application startup is handled by the Go service.
