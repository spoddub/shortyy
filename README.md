### Hexlet tests and linter status:
[![Actions Status](https://github.com/spoddub/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/spoddub/go-project-278/actions)
[![CI](https://github.com/spoddub/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/spoddub/go-project-278/actions/workflows/ci.yml)

[Live demo](https://shortyy.onrender.com)

## Shortyy

`Shortyy` is a URL shortener service written in Go (Gin).  
It turns long URLs into short codes and redirects users to the original address.

---

## Features

- Create short links from long URLs (custom `short_name` is optional)
- Redirect by short code: `GET /r/:code`
- Store links and visits in PostgreSQL
- Visits analytics: IP, user agent, referer, redirect status, created_at
- Pagination for collections via `Range` header or `range` query parameter
- Validation with a consistent API error format
- Optional Sentry integration
- Docker deploy with Caddy (serves UI and reverse proxies API)

---

## Tools used

| Tool | What it is used for |
| --- | --- |
| [Go](https://go.dev/) | Language and toolchain. |
| [Gin](https://github.com/gin-gonic/gin) | HTTP router and middleware. |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool. |
| [PostgreSQL](https://www.postgresql.org/) | Persistent storage. |
| [sqlc](https://sqlc.dev/) | Type-safe SQL code generation. |
| [goose](https://github.com/pressly/goose) | Database migrations. |
| [go-playground/validator v10](https://github.com/go-playground/validator) | Request validation (Gin binding). |
| [golangci-lint](https://golangci-lint.run/) | All-in-one Go linter. |
| [GitHub Actions](https://docs.github.com/actions) | CI for linting, tests and builds. |
| [Sentry](https://sentry.io/) | Error monitoring (optional). |
| [Caddy](https://caddyserver.com/) | Static UI + reverse proxy in Docker/Render runtime. |
| [Render](https://render.com/) | Deployment platform. |

---

## API

### Links

- `GET /api/links` - list links (supports pagination)
- `POST /api/links` - create a link
- `GET /api/links/:id` - get link by id
- `PUT /api/links/:id` - update a link
- `DELETE /api/links/:id` - delete a link

Example request:

```bash
curl -s -X POST http://localhost:8080/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com/long-url","short_name":"exmpl"}'
```

Example response:

```json
{
  "id": 1,
  "original_url": "https://example.com/long-url",
  "short_name": "exmpl",
  "short_url": "http://localhost:8080/r/exmpl"
}
```

### Redirect

- `GET /r/:code` - redirects to `original_url` and creates a visit record

### Visits

- `GET /api/link_visits` - list visits (supports pagination)

---

## Pagination

Collections support pagination using either:

- `Range` request header: `Range: [0,10]`
- or query param: `?range=[0,10]`

The response includes:

- `Content-Range: <resource> <from>-<to>/<total>`

---

## Validation and errors

- Invalid JSON: `400 Bad Request` with `{ "error": "invalid request" }`
- Validation errors: `422 Unprocessable Entity` with `{ "errors": { "<field>": "<message>" } }`
- Unique `short_name` conflict: `422 Unprocessable Entity` with `{ "errors": { "short_name": "short name already in use" } }`

---

## Installation and local development

### Requirements

- Go (modern version, recommended 1.25+)
- PostgreSQL
- `make`
- `golangci-lint` installed locally (or via `go install`)
- Node.js (20+ recommended) if you run the UI locally

### Environment variables

- `DATABASE_URL` (required)
- `BASE_URL` (recommended, used to build `short_url`)
- `PORT` (defaults to `8080`)
- `SENTRY_DSN` (optional)

Example:

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/appdb?sslmode=disable"
export BASE_URL="http://localhost:8080"
export PORT="8080"
```

### Migrations

```bash
goose -dir db/migrations postgres "$DATABASE_URL" up
```

### Run backend

```bash
go run main.go
# app listens on :8080
```

### Run tests and linters

```bash
make test
make lint
```

---

## Deployment

The project is deployed on Render using Docker and is available here:  
https://shortyy.onrender.com
