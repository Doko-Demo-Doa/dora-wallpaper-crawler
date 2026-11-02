# Dora wallpaper crawler

A Go service that serves the wallpaper API and crawls [Dora World's wallpaper page](https://dora-world.com/wallpaper) hourly. New images are copied into Box and recorded in PostgreSQL.

## Requirements

- Go 1.24+
- PostgreSQL
- `DATABASE_URL`
- Box credentials for uploads (the read API can run without them). Either:
  - `BOX_DEVELOPER_TOKEN` — simplest for local testing, but a developer token
    expires after about an hour and cannot be refreshed, or
  - `BOX_CLIENT_ID` + `BOX_CLIENT_SECRET` + `BOX_ENTERPRISE_ID` (or
    `BOX_USER_ID`) for the client credentials grant, which refreshes itself and
    is what you want in a deployment. A subject ID (`BOX_ENTERPRISE_ID` /
    `BOX_USER_ID`) is required: without one Box rejects the token request.

Generate the developer token and app credentials in the [Box developer console](https://app.box.com/developers/services). The token must be authorized for the account that owns the destination folder.

Start PostgreSQL for local development:

```sh
docker run --name dev-postgres -p 5432:5432 -e POSTGRES_PASSWORD=12345678 -d postgres
```

Set environment variables, then run:

```sh
export DATABASE_URL='postgresql://postgres:12345678@localhost:5432/postgres'
export BOX_DEVELOPER_TOKEN='YOUR_KEY_HERE'
go run .
```

The service listens on port `3000` by default. Set `PORT` to change it. It applies the wallpaper table migration at startup and starts a crawl immediately, then repeats hourly.

## API

- `GET /api/` returns desktop wallpaper source URLs that are not yet stored.
- `GET /api/all` returns stored wallpaper records.
- `GET /api/list-files` returns files under `./temp`.

The SQL migration history is in `drizzle/` for reference. Database setup on application startup is handled by the Go service.
