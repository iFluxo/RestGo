# RestGo

A simple REST API built with **native Go (standard library only)** — no framework, no database. Runs locally and on **Vercel**.

## Structure

```
cmd/server/main.go   # server entrypoint (auto-detected by Vercel)
internal/api/        # router, handlers, middleware, tests
internal/store/      # Item model + in-memory store
vercel.json          # Go framework preset for Vercel
```

## Running

```bash
go run ./cmd/server      # http://localhost:8080
PORT=9000 go run ./cmd/server
go test ./...
```

Build a binary:

```bash
go build -o bin/server ./cmd/server
```

## Endpoints

| Method | Path          | Description           | Success |
|--------|---------------|-----------------------|---------|
| GET    | `/`           | Simple status         | 200     |
| GET    | `/health`     | Health check + uptime | 200     |
| GET    | `/items`      | List items            | 200     |
| POST   | `/items`      | Create item           | 201     |
| GET    | `/items/{id}` | Get item              | 200     |
| PUT    | `/items/{id}` | Update item           | 200     |
| DELETE | `/items/{id}` | Delete item           | 204     |

`GET /` and `GET /health` respond with `{"status":"ok"}`; `/health` also includes the process uptime in milliseconds, e.g. `{"status":"ok","uptime":1234}`.

`POST`/`PUT` payload:

```json
{ "name": "Milk Coffee", "price": 18000 }
```

Rules: `name` is required (max 100 characters), `price` >= 0, body max 1 MB, unknown fields are rejected. Errors always look like `{"error":"..."}`.

```bash
curl localhost:8080/
curl localhost:8080/health
curl localhost:8080/items
curl -X POST localhost:8080/items -H 'Content-Type: application/json' -d '{"name":"Fried Rice","price":25000}'
curl -X PUT localhost:8080/items/1 -H 'Content-Type: application/json' -d '{"name":"Special Fried Rice","price":30000}'
curl -X DELETE localhost:8080/items/1
```

## Deploying to Vercel

Vercel detects the root `go.mod` and the `cmd/server/main.go` entrypoint, then runs the Go server as a backend. `vercel.json` only pins the `go` preset.

```bash
npm i -g vercel
vercel --prod
```

Or import the repository in the Vercel dashboard. Do not set a `PORT` env var — Vercel provides it automatically and the server already reads it.

> **Note:** data is stored in memory, so it resets on every cold start / scale event and is not suitable for production. Replace `internal/store` with a database (Postgres, Redis, etc.) if you need persistence. Sample seed data is also re-created on start.
