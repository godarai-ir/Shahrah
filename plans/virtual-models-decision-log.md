# Virtual Models Decision Log

## Documented Facts
- Virtual alias + provider target list pattern is proven in both TrueFoundry and Portkey.
- Status-code based fallback is common and compatible with OpenAI-like proxying.

## Capability Classification
| Capability | Decision | Rationale |
|---|---|---|
| Tenant-scoped virtual alias | include now | Core feature requirement |
| Priority routing | include now | Smallest reliable routing primitive |
| Retry + status-based fallback | include now | Required resilience behavior |
| Admin CRUD | include now | Needed operability |
| Virtual alias in `/v1/models` | include now | Discoverability + compatibility |
| Weighted routing | defer | Higher complexity, not P0 |
| Latency routing | defer | Requires metrics feedback loop |
| Sticky routing | defer | Session-state behavior not required for P0 |
| Metadata filtering | defer | Additional policy surface |
| Header overrides | defer | Provider-specific complexity |

## Recommendations
- Keep strict MVP boundary with priority routing and explicit retry/fallback policy.
- Normalize final fallback-exhausted errors to stable OpenAI-like envelope.

## Assumptions
- `fallback_candidate` false semantics are accepted; default behavior can be refined in P1 with pointer fields.
- Provider/model validity checks rely on current provider registry and request-time execution.

## Sources Used
- `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model.md`
- `.truefoundry-docs/markdown/docs__ai-gateway__virtual-model-advanced.md`
- `.portkey-docs/markdown/product/docs__product__ai-gateway__configs.md`
- `.portkey-docs/markdown/product/docs__product__ai-gateway__fallbacks.md`

## Impact of `plans/README.md`
- Confirmed MVP-first delivery and OpenAI-compatible API preservation.
