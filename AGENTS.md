# Repository Guidelines

## Project Structure & Module Organization
This repository is a two-app monorepo:
- `backend/`: Go API and ingestion workflow. Entry point is `cmd/server/main.go`; DB utility command is `cmd/testdb/main.go`.
- `backend/internal/`: layered backend modules: `config/`, `handler/`, `service/`, `repository/`, `model/`.
- `backend/sql/init.sql`: PostgreSQL/pgvector bootstrap SQL.
- `frontend/`: React + TypeScript app (Webpack). Main app code lives in `src/` with `components/`, `hooks/`, and API client code in `src/api.ts`.
- `docker-compose.yml`: full local stack (Postgres, NetEase API bridge, backend, frontend).

## Build, Test, and Development Commands
- `docker-compose up -d`: start full local stack.
- `cd backend && go run ./cmd/server`: run backend locally on `:8080`.
- `cd backend && go test ./...`: run backend tests (and package compile checks if tests are missing).
- `cd frontend && npm install && npm run start`: start frontend dev server on `:3001`.
- `cd frontend && npm run build`: production frontend bundle to `frontend/dist`.

## Coding Style & Naming Conventions
- Go: format with `gofmt` (or `go fmt ./...`) before commit; keep package names short and lowercase.
- TypeScript/React: `strict` mode is enabled (`frontend/tsconfig.json`); prefer typed props/hooks and avoid `any`.
- Naming: React components in PascalCase (e.g., `SongCard.tsx`), hooks as `useXxx.ts`, backend files lowercase by domain (`search.go`, `netease.go`).
- Use 4-space indentation in Markdown/JSON examples and follow existing file style for code.

## Testing Guidelines
- Backend tests should be colocated as `*_test.go` next to implementation files.
- Frontend tests are not scaffolded yet; if added, place under `frontend/src/` using `*.test.ts(x)` naming.
- Minimum pre-PR check: run `go test ./...` and `npm run build` to catch compile/runtime integration issues.

## Commit & Pull Request Guidelines
- Follow Conventional Commit style already used in history: `feat: ...`, `fix: ...`, `chore: ...`.
- Keep commits scoped by layer (`backend/service`, `frontend/components`, etc.) and avoid mixing unrelated changes.
- PRs should include: purpose, key changes, local verification commands run, and any env/config updates.
- For UI/API behavior changes, include screenshots or example request/response snippets.

## Security & Configuration Tips
- Keep secrets in environment variables (`GEMINI_API_KEY`, `NETEASE_PHONE`); never commit real credentials.
- Use `.env` only for local development; ensure defaults in `backend/internal/config/config.go` remain safe.
