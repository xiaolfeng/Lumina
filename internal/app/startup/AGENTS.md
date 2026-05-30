# STARTUP KNOWLEDGE BASE

## OVERVIEW
`internal/app/startup/` defines boot-time dependency registration and seed execution order.

## WHERE TO LOOK
| Task | File | Rule |
|---|---|---|
| Register new startup node | `startup.go` | Append `xRegNode.RegNodeList` in `Init()` |
| Add DB init behavior | `startup_database.go` | Keep DSN/env defaults and GORM config centralized |
| Add Redis init behavior | `startup_redis.go` | Use `xEnv.GetEnv*` defaults |
| Add business seed stage | `startup_prepare.go` + `prepare/*.go` | Seed logic goes to `prepare/` only |
| Add new migratable model | `startup_database.go` | Append to `migrateTables` in dependency order |

## EXECUTION ORDER (DO NOT BREAK)
1. `databaseInit` (`xCtx.DatabaseKey`)
2. `nosqlInit` (`xCtx.RedisClientKey`)
3. `businessDataPrepare` (`xCtx.Exec`)

`prepare` depends on DB in context (`xCtxUtil.MustGetDB`), so DB node must run first.

## CONVENTIONS
- Startup init funcs use signature `(ctx context.Context) (any, error)`.
- Startup logs use `xLog.NamedINIT`.
- Database naming strategy uses env-driven `DATABASE_PREFIX` and singular table naming.
- `prepare` methods must be idempotent (repeatable across restarts).
- Seed failures should be observable in logs with entity identifiers.

## ANTI-PATTERNS
- Do not initialize DB/Redis anywhere outside startup nodes.
- Do not call seed logic directly from `main.go` or route/handler logic.
- Do not reorder `migrateTables` casually; respect FK dependencies.
- Do not create env reads without defaults in startup code.

## PREPARE SUBDIR NOTES
`prepare/` is for startup seed data only (e.g., default roles). Keep each seed concern in `prepare_<domain>.go` and wire from `prepare.go`.

## MIGRATE ORDER
`migrateTables` order respects FK dependencies: `Role` precedes `User` because `User.RoleName` references `Role.Name`.
