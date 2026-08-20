# BaseAdmin

Full-stack admin panel base project — Go (Gin + GORM) backend, React (Vite) frontend, Redis-backed sessions, PostgreSQL by default (MySQL-switchable).

## Prerequisites

- Go 1.26+
- Node.js 20+
- PostgreSQL 15 or MySQL 8 (or use Docker)
- Redis 7 (or use Docker)
- Docker + Docker Compose (optional, recommended for local dev)

## Quick start (Docker)

```bash
cp backend/.env.example backend/.env
docker compose up -d --build
docker compose exec backend ./baseadmin migrate
docker compose exec backend ./baseadmin seed
```

- Backend: http://localhost:8500
- Frontend: http://localhost:8501
- Default login: `admin@example.com` / `Admin@12345`

## Manual setup

### Backend

```bash
cd backend
cp .env.example .env   # edit DB/Redis credentials as needed
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g main.go -o docs   # generates docs/ — gitignored, main.go imports it, so this must run first on a fresh clone
go run main.go migrate
go run main.go seed
go run main.go          # starts server on APP_PORT (default 8500)
```

### Frontend

```bash
cd frontend
cp .env.example .env   # set VITE_API_URL if backend isn't on localhost:8500
npm install
npm run dev             # dev server on :8501 (see vite.config.js)
npm run build            # production build → dist/
```

## Switching database: PostgreSQL → MySQL

Edit `backend/.env`:

```
DB_DRIVER=mysql   # was: postgres
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password
DB_NAME=baseadmin
```

No code changes required — `config/database.go` selects the GORM dialector from `DB_DRIVER`. Re-run `go run main.go migrate` against the new database.

## Environment config

See `backend/.env.example` for the full list (app, DB, Redis, session, storage, rate limiting). Every var there is actually read by `config/config.go` — nothing speculative.

## Features

- **Auth**: email/password login, Redis-backed sessions (`HttpOnly`, `SameSite=Strict`, `Secure` in production), CSRF token, per-account password history (no reuse of last 5), account lockout after repeated failed logins (`RATE_LIMIT_LOGIN`, `RATE_LIMIT_LOGIN_LOCKOUT_MINUTES`).
- **Users**: CRUD, avatar upload, role assignment, status (active/inactive/suspended), per-device session list with remote revoke ("Sessions" action), CSV/Excel export (`GET /users/export/csv`, `/export/excel`, respects the active search/status/role filters). Role or status changes force-invalidate that user's other active sessions immediately (no waiting for token expiry).
- **Roles & permissions**: per-menu CRUD permission matrix (view/create/edit/delete), superadmin role is protected from edit/delete/permission changes.
- **Menus & menu sections**: drag-free reordering (up/down), used to build the sidebar tree and the permission matrix.
- **Settings**: app name, logo, favicon, base/primary color (live color picker, applied instantly via a CSS custom property so it doesn't require a rebuild), maintenance mode. Branding fields are also exposed unauthenticated at `GET /settings/public` so the login page renders correctly before a session exists.
- **Activity logs**: audit trail (actor, action, module, before/after diff, IP, user agent) with filtering; `DELETE /activity-logs/cleanup` prunes rows older than `LOG_CLEANUP_DAYS` (manual trigger — see Known TODOs).
- **UI toolbar**: a shadcn-style grouped menubar (`components/ui/Menubar.jsx`) is the standard per-page toolbar — title + actions (Add New, Filter, Others/export) live in one bar in the Navbar, registered per-page via the `usePageActions` hook + `pageActionsStore`. Filter/dropdown panels render through a React portal so they aren't clipped by the toolbar's horizontal scroll container.
- **Security headers**: CSP (stricter default policy; a slightly relaxed one scoped to `/swagger` for swagger-ui's inline scripts), `X-Frame-Options`, `X-Content-Type-Options`, HSTS, `Referrer-Policy` — see `middleware/security.go` (backend/API) and `frontend/nginx.conf` (the actual SPA HTML, more critical since it's what a browser executes).
- **Dependency scanning**: `.github/workflows/security-scan.yml` runs `govulncheck` (Go) and `npm audit --audit-level=high` on every push/PR to `main` plus a weekly schedule; `.github/dependabot.yml` opens weekly update PRs for Go modules, npm, and both Dockerfiles.

## Folder structure

```
/
├── backend/
│   ├── main.go              # entrypoint; supports `migrate` and `seed` subcommands
│   ├── config/               # env loading, DB dialector, Redis client
│   ├── database/
│   │   ├── migrations/       # GORM AutoMigrate runner
│   │   └── seeders/          # default roles/users/menus/settings
│   ├── middleware/            # CORS, security headers (incl. CSP), rate limit + login lockout, CSRF, session auth, permission guard, etc.
│   ├── modules/                # auth, users (incl. export.go), roles, menus, menu_sections, permissions, settings, activity_logs
│   ├── session/                # Redis session store
│   ├── storage/                # StorageInterface + local disk implementation
│   └── utils/                   # shared model helpers
└── frontend/
    ├── nginx.conf             # security headers for the built SPA in production
    └── src/
        ├── api/                # axios client + per-resource wrappers
        ├── components/         # ui/ (Button, Table, Modal, Menubar, ...) and layout/ (Sidebar, Navbar)
        ├── hooks/               # usePermission, usePageActions
        ├── pages/                # auth, dashboard, users, menus, menu_sections, roles, settings, activity_logs
        ├── router/               # route guards
        ├── store/                # Zustand: authStore, settingsStore, sidebarStore, pageActionsStore
        └── utils/                # color.js (hex ↔ CSS var helpers for the live base-color picker), url.js, validation.js
```

## API docs

Swagger annotations are on every handler and already wired up. `docs/` (the generated Go package `main.go` imports) is **gitignored** and rebuilt via `swag init -g main.go -o docs` — the Docker build does this automatically (see `backend/Dockerfile`); for manual/local runs you need to run it yourself first (see Manual setup above), and again any time you change handler doc-comments. View the UI at:

```
http://localhost:8500/swagger/index.html
```

Requires an authenticated **superadmin** session (`requireSuperadmin` guard on the route).

## Known TODOs / gaps

- **Redis has no password by default** (`docker-compose.yml` runs `redis-server` without `--requirepass`, and `REDIS_PASS` is empty in `.env.example`). Fine for isolated local dev; before any shared or production deployment, set `--requirepass` + `REDIS_PASS`, and avoid publishing port `6379` to the host.
- `LOG_CLEANUP_DAYS` cleanup is exposed as `DELETE /activity-logs/cleanup` but not auto-scheduled — add a cron trigger for nightly cleanup.
- Only local file storage is implemented; `StorageInterface` is ready for an S3 driver if needed later (no `STORAGE_DRIVER` switch exists yet — `main.go` always constructs `LocalStorage`).
- Frontend bundle is a single ~1.2MB chunk — consider route-based code-splitting (`React.lazy`) as the app grows.
- No 2FA/MFA.
- `npm audit` currently flags a moderate-severity dev-only advisory in `esbuild`/`vite` (SSRF against the local dev server, not the production build); fixing it means a Vite 8 major upgrade, deliberately not done automatically — see the CI job for status.

## Default credentials

```
Email:    admin@example.com
Password: Admin@12345
```

Change this immediately in any non-local environment.
