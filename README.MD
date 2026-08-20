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

Open `http://localhost:5173` in a browser. Select **Ping backend** to display the message from the backend.

Press `Ctrl+C` in each terminal to stop the servers.
