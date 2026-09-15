# Interview Project

## Requirements

Install these tools before you start:

- Go 1.26.3 or later
- Node.js 20.19 or later in the 20.x release, or Node.js 22.12 or later
- npm

## Start the backend

Open a terminal in the repository root.

```sh
go run .
```

The backend listens at `http://localhost:8000`.

## Start the frontend

Open a second terminal in the repository root.

Install the frontend dependencies from the lock file:

```sh
cd frontend
npm ci
```

Start the frontend server:

```sh
npm run dev
```

Open `http://localhost:5173` in a browser. Enter a name to open the Weekly
Performance Report. Open the application in another tab with a different name to
see the viewer list update in real time.

Press `Ctrl+C` in each terminal to stop the servers.

## Architecture

The browser keeps a Server-Sent Events connection open while a user views a
report. The Go server stores each connection in a concurrency-safe in-memory hub
and broadcasts a new snapshot when someone connects or disconnects. Viewer names
are deduplicated, so the same user can have multiple tabs open safely.

Presence is scoped by report ID. The current UI uses `weekly-performance`, while
the endpoint supports additional reports at:

```text
GET /api/reports/{reportID}/presence?name={viewerName}
```

State is intentionally ephemeral and local to one server process. Production
authentication and a shared presence store such as Redis are outside this
exercise's scope.

## Checks

```sh
GOCACHE=/tmp/hype10-go-cache go test -race .
cd frontend
npm run build
npm run lint
```
