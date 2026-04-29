# backup-system

Backup management system for Mikrotik and database nodes, with a React dashboard.

---

## Running the Go backend

```bash
# From project root
cp .env.example .env   # edit as needed
go run ./cmd/web
# Listens on WEB_LISTEN (default: 127.0.0.1:8080)
```

---

## Running the React frontend (dev)

```bash
cd frontend
npm install
npm run dev
# Opens at http://localhost:1909
# /api/* and /storage/* are proxied to http://localhost:8080
```

To point the proxy at a different backend port, create `frontend/.env`:

```
VITE_API_TARGET=http://localhost:8080
```

## Building for production

```bash
cd frontend
npm run build
# Output: frontend/dist/
```

Serve `dist/` from any static file server, with `/api/*` reverse-proxied to the Go backend.

---

## First-time setup

1. Start the Go backend.
2. Open `http://localhost:1909` — you'll be redirected to `/setup`.
3. Create the initial admin account.
4. Log in and start monitoring.
