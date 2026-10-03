# Febros16 - Backend Project Status

## Architecture
- **Language**: Go 1.22+
- **Database**: PostgreSQL (Supabase/Render)
- **Deployment**: Render
- **Router**: standard `net/http` (Go 1.22 wildcard routing)

## Completed Phases
- **Phase 1: Setup & Auth**: Users table, Register, Login (JWT), Profile updating.
- **Phase 2: Content Foundation**: Content tables structure, basic setup.
- **Phase 3: Articles Engine**: Full CRUD for Articles, `author_id` AuthZ checks on PUT/DELETE, Optimistic UI (Ghost Delete) fixes (`db.Exec` + `RowsAffected`), and Dashboard Stats (`/api/v1/users/me`).
- **Phase 4: Research Projects**: Migrations added to `config/db.go`, Full CRUD endpoints registered and implemented.

## Current Status
- Stabilizing Phase 4 (Research Projects).
- Ensuring code safety and deployment stability on Render.

## Next Steps
- Implementation of Phase 5 (Opportunities / Campaigns logic) when requested.
