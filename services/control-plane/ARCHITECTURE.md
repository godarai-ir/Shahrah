# Control-plane architecture

This document is the operational map for humans and coding agents changing the
Go control-plane. It describes the target modular-monolith boundaries and the
rules enforced while legacy code is migrated.

## Dependency direction

```text
cmd/server
  -> internal/config            process configuration snapshot
  -> internal/bootstrap         composition and lifecycle only
      -> HTTP adapters          Fiber parsing, authorization, response mapping
          -> application        use-case orchestration and application ports
              -> domain         policies, transitions, canonical errors
              -> repositories   PostgreSQL, ClickHouse, cache, and remote adapters
```

Dependencies point inward. Domain and application packages do not import Fiber,
PostgreSQL, ClickHouse, Viper, or process environment APIs. Persistence adapters
translate driver-specific failures before returning through application ports.

The gateway follows a parallel boundary:

```text
endpoint adapter -> gateway executor -> router/tunnel
                                  |---> billing/reservations
                                  `---> logs/traces/metrics
```

Endpoint adapters own protocol-specific decoding and validation. The executor
owns admission, dispatch, settlement, and observability exactly once for each
sync or streaming lifecycle.

## Package ownership

| Area | Owns | Must not own |
| --- | --- | --- |
| `internal/config` | Reading environment/YAML and producing immutable typed settings | Runtime clients, database connections, feature decisions |
| `internal/bootstrap` | Construction order, dependency wiring, lifecycle registration | Business rules or environment parsing helpers |
| `internal/<feature>` | Domain policy, application service, consumer-owned ports | Fiber or database-driver errors |
| `internal/management/<feature>` | HTTP routes and transport mapping | SQL and transaction orchestration |
| `internal/storage` | Legacy persistence adapters during migration | New business state on `Store` |
| `internal/proxy` | Gateway protocol adapters and shared execution pipeline | Provider HTTP adapters owned by Silk Road |
| `packages/contracts` | Shared public API contracts | Control-plane implementation details |

## Store strangler rule

`storage.Store` is a compatibility facade, not a feature repository. Every
migration moves receiver methods from `*Store` to a narrow repository, wires an
application service in bootstrap, and changes production consumers to use that
service. Promoted methods may remain temporarily for old tests or commands, but
new code must not depend on them.

The prompt slice is the reference implementation:

- `internal/prompts.Service` owns prompt use cases and the repository port;
- `storage.PromptRepository` implements persistence;
- management and gateway wiring consume `prompts.Service`;
- PostgreSQL errors are translated before reaching HTTP.

IAM follows the same boundary incrementally: `internal/iam.Service` owns the
application port, `storage.IAMRepository` owns persistence, and production
management wiring consumes the service rather than the repository or Store.

The same rule now applies to tenant administration and gateway policies:
`internal/tenantadmin.Service` and `internal/policies.Service` own application
ports, while their storage repositories remain replaceable adapters. Canonical
tenant-admin DTOs and errors live with the feature instead of in storage.

Virtual-model and deployment management follows the same path through
`internal/virtualmodels.Service` and `storage.ModelRepository`. Observability,
provider-native resources, platform credentials, runtime-profile persistence,
and database health each have dedicated repositories. `storage.Store` now owns
only construction-time compatibility and resource shutdown; feature code must
not accept it.

## Configuration rule

Environment and YAML are read at the `internal/config` edge. Constructed services
receive typed values. A package must not call `os.Getenv` to change behavior at
request time. Secret values remain redacted and are never included in logs or
error messages.

## Transactions and detached work

Application services define the use-case transaction boundary. Repositories may
implement that boundary with a database transaction, but transport handlers do
not begin transactions. Work that must survive a client disconnect uses an
explicit bounded detached-work helper and records failure; arbitrary
`context.Background()` calls in request handlers are not allowed.

## Change recipes

### Add or change a management feature

1. Add domain rules and canonical errors to the feature package.
2. Add an application method and the smallest repository port required by that method.
3. Implement or update the persistence adapter and translate driver errors there.
4. Wire the application service in bootstrap.
5. Keep the Fiber handler limited to auth, parsing, calling the use case, and mapping canonical errors.
6. Add domain, application, adapter, and HTTP contract tests at their respective boundaries.

### Add a gateway endpoint

1. Add or reuse shared request-executor lifecycle behavior.
2. Keep provider capability and identity in `providerregistry`.
3. Implement endpoint-specific decode, validation, usage extraction, and error framing.
4. Add sync, streaming, settlement, disconnect, and redaction contract tests.
5. Add provider transport behavior in Silk Road, not in control-plane.

### Lower an architecture ratchet

`internal/architecture/ratchet_test.go` records current maximum debt. A change
may lower a baseline when it removes debt. Raising a baseline requires an
explicit architecture decision and must not be used to make CI pass.

## Verification tiers

1. Focused package tests while iterating.
2. `go test ./internal/...` for the modular-monolith boundary.
3. `go vet ./...` and the repository `golangci-lint` command.
4. Unit, contract, and integration suites from the root `Makefile`.

Integration tests that compose Silk Road and tunnel services live in the
separate `tests/integration` Go module. This keeps system-test dependencies out
of the production module and its default feedback loop.
