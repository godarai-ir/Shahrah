# ADR 0001: Frontend Library Policy

- Status: accepted
- Scope: `apps/tenant-admin`

## Decision

### UI primitives

Feature code imports primitives only from `components/ui` and product patterns only from `components/app-ui`. Direct imports from `radix-ui`, `@radix-ui/*`, or `@base-ui/react` are restricted to the primitive wrapper layer. Existing wrappers may use different primitive implementations during migration, but that detail must not leak into features.

The configured shadcn components and the repository design system determine composition, styling, accessibility, icon, and RTL behavior.

### Charts

Recharts through `components/ui/chart` is the default chart system. Direct Recharts usage is allowed only in a feature-local chart module behind a dynamic client boundary when the shared wrapper cannot express the visualization. ECharts is not used by current source and must not be introduced without a new ADR demonstrating a concrete capability or performance requirement.

### Dates

Persist and transport timestamps as ISO 8601 strings in UTC. Use the shared display components for user-visible timestamps. `date-fns-jalali` is reserved for Jalali calendar calculations; features must not implement date formatting independently.

### Server state

Server Components own initial data where practical. React Query is the standard for repeatedly refreshed or mutated browser-owned server state. Local React state is reserved for ephemeral presentation state, and URL state owns shareable filters and pagination.

### Runtime validation

Zod schemas or generated control-plane schemas validate all untrusted payloads. Generic TypeScript parameters on `response.json()` do not constitute validation.

## Consequences

- Coding agents have one default choice for each concern.
- Primitive implementations can change without feature rewrites.
- Duplicate request helpers, chart systems, and date formatting are migration debt and must not be copied into new code.
- Exceptions require an ADR and performance evidence.
