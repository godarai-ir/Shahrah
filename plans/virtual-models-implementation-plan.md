# Virtual Models Implementation Plan (MVP)

## Documented Facts
- Control-plane already has deployment CRUD, model listing, and router/proxy/tunnel chain.
- Storage uses PostgreSQL migrations under `db/main/migrations`.
- Admin APIs are mounted in `internal/server/server.go`.

## Recommendations
1. Add schema migration `013_virtual_models.sql`.
2. Add platform contracts for `VirtualModel*` types.
3. Implement storage repository methods for create/list/get/update and alias resolution.
4. Add admin APIs:
   - `POST /api/v1/virtual-models`
   - `GET /api/v1/virtual-models`
   - `GET /api/v1/virtual-models/:id`
   - `PATCH /api/v1/virtual-models/:id`
5. Add gateway virtual-model resolve + fallback executor in sync chat path.
6. Extend `/v1/models` to include virtual aliases.
7. Add unit tests for resolver and fallback executor.

## Assumptions
- Priority routing only.
- Stream fallback remains deferred.

## Testing Strategy
- Run unit tests for `internal/...` and `tests/unit/...`.
- Add virtualmodels unit tests for resolution/fallback behavior.
- Integration tests use in-memory store doubles and keep existing direct path tests intact.

## Sources Used
- `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
- `.portkey-docs/markdown/product/docs__product__ai-gateway__fallbacks.md`

## Impact of `plans/README.md`
- Preserved compatibility and incremental delivery approach.
