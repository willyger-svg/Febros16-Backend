# Febros16 Backend Guidelines

## Development Rules
1. **Database Migrations:**
   Whenever you create a new SQL migration file in the `migrations/` directory, you MUST add the filename to the `migrationFiles` array in `config/db.go`. If you forget this, the migration will not execute on deployment.

2. **Preventing Ghost Deletes (AuthZ Enforcement):**
   When writing `PUT` or `DELETE` endpoints, never return a `200 OK` without confirming that rows were modified. You must use `db.Exec()` and explicitly check `result.RowsAffected()`. If `RowsAffected == 0`, return a `404 Not Found` to prevent the frontend from experiencing "Ghost Deletes" (Optimistic UI failures).

3. **Compilation Safety:**
   Always double-check that custom variables and middleware context keys (like `middleware.UserRoleKey`) exist before pushing code to Render.

## Project Tracking
- Always update `PROJECT_STATUS.md` at the root of the project to track the current phase, completed work, and pending features. This ensures continuity across different sessions.
