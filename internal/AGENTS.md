# INTERNAL KNOWLEDGE BASE

## OVERVIEW
`internal/` implements the business runtime pipeline: route -> handler -> logic -> repository -> entity.

## STRUCTURE
```text
internal/
├── app/
│   ├── route/       # Route groups + middleware
│   └── startup/     # Infra init + prepare execution
├── handler/         # HTTP input/output boundary
├── logic/           # Business orchestration
├── repository/      # Data access boundary
├── entity/          # GORM models
└── constant/        # Shared constants (e.g., gene numbers)
```

## WHERE TO LOOK
| Task | File/Dir | Why |
|---|---|---|
| Add route group | `app/route/route.go` + `route_*.go` | Route wiring entry is `NewRoute` |
| Add new handler | `handler/handler.go`, `handler/*.go` | `NewHandler[T]` centralizes dependency wiring |
| Add new logic module | `logic/*.go` | Logic takes context-injected db/rdb |
| Add new repository | `repository/*.go` | Repository returns `*xError.Error` |
| Add new table/entity | `entity/*.go` + `app/startup/startup_database.go` | Migration list is explicit |
| Add startup seed data | `app/startup/prepare/*.go` | Seed runs in startup Exec phase |
| Add request/response DTO | `api/<domain>/` | DTOs are per-domain subpackages |

## CONVENTIONS
- Keep handlers thin: validate/bind, call logic, map result/err to response.
- Logic is orchestration only; persistence and SQL belong in repository layer.
- Repositories expose `(data, *xError.Error)` style APIs, not raw `error` only.
- Use `xLog.WithName` by layer: CONT/LOGC/REPO/INIT.
- Use request context for downstream calls (`ctx.Request.Context()` or injected context).

## ANTI-PATTERNS
- Do not access DB/Redis in handlers directly.
- Do not return HTTP JSON directly from logic/repository layers.
- Do not construct ad-hoc handlers bypassing `NewHandler[T]` pattern.
- Do not place business constants inside handler/logic files; keep in `constant/`.

## DEBUG PATH
1. Request not routed -> check `app/route/route.go` and feature `route_*.go`.
2. Request routed but wrong output -> check `handler/*.go` then `logic/*.go`.
3. Infra/resource failures -> check `repository/*.go` and startup init state.
