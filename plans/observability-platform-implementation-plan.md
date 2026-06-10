# Shahrah Observability Platform Implementation Plan

## Objective
Implement a production-grade observability platform for Shahrah on K3s using Argo CD and GitOps, with Longhorn as the in-cluster storage class for persistent observability workloads.

The target outcome is:
- centralized logs for cluster and application workloads
- cluster and app metrics in Prometheus
- operational dashboards in Grafana
- routed alerts in Alertmanager
- distributed tracing for `control-plane` through OTLP
- secure, retention-aware, GitOps-managed production operations

## Inputs Reviewed
- `D:\source\Shahrah\infra\README`
- `D:\source\Shahrah\infra\argocd\README.md`
- `D:\source\Shahrah\services\control-plane\cmd\server\main.go`
- `D:\source\Shahrah\services\control-plane\internal\server\server.go`
- `D:\source\Shahrah\services\control-plane\internal\observability\tracing\config.go`
- `D:\source\Shahrah\services\control-plane\internal\observability\metrics\config.go`
- `D:\source\Shahrah\plans\model-metrics-clickhouse-implementation-plan.md`
- `D:\source\Shahrah\plans\model-metrics-runbook.md`

## Confirmed Infrastructure Assumptions
- runtime platform: K3s
- deployment model: Argo CD + `shahrah-gitops`
- ingress strategy: `ingress-nginx` behind HAProxy
- storage class: Longhorn
- external PostgreSQL remains outside the cluster
- observability workloads should run in a dedicated namespace

## Scope Decisions
### Include now
- `kube-prometheus-stack` for:
  - Prometheus
  - Alertmanager
  - Grafana
  - node-exporter
  - kube-state-metrics
- `Loki` for centralized logging
- `Promtail` or `Grafana Alloy` for log shipping
- `OpenTelemetry Collector`
- `Tempo` for traces
- internal-only access model for observability control planes
- initial dashboards, alerts, retention, and runbooks

### Defer
- multi-region observability
- external object storage for Loki/Tempo
- SSO integration for Grafana
- advanced trace analytics and exemplars
- full cross-service business metrics rollout

### Exclude from this phase
- public Grafana exposure
- ELK/OpenSearch stack
- custom-built observability stack from scratch

## Architecture Decisions
### Preferred stack
- Metrics:
  - `kube-prometheus-stack`
- Logs:
  - `Loki`
  - `Promtail` for simpler initial rollout
- Traces:
  - `OpenTelemetry Collector`
  - `Tempo`

### Why this stack
- mature Helm packaging and Argo CD friendliness
- lower operational cost than ELK for current Shahrah scale
- native fit with Grafana for logs, metrics, and traces in one UI
- aligns with current `control-plane` OTLP-capable tracing implementation

### Namespace layout
- `observability`
- optional later split:
  - `observability-system`
  - `observability-data`

For the initial rollout, keep a single namespace to reduce operational complexity.

## Production Constraints
### Security
- Grafana, Prometheus, Alertmanager, Loki, Tempo must not be public by default
- access should be restricted to:
  - bastion
  - VPN
  - private ingress with authentication if needed

### Storage
- use Longhorn-backed PVCs
- do not keep observability data on ephemeral node storage
- define retention before go-live

### Data sensitivity
- raw request and response bodies must not be shipped centrally by default
- trace payload capture must default to metadata-only
- payload encryption key and PII hashing must be configured before wider trace rollout

### GitOps boundary
- all production observability manifests belong in `shahrah-gitops`
- this repo remains the architecture, implementation, and runbook source

## Recommended GitOps Layout
These files should be created in `D:\source\shahrah-gitops`:

```text
apps/
  observability-kube-prometheus-stack.yaml
  observability-loki.yaml
  observability-promtail.yaml
  observability-tempo.yaml
  observability-otel-collector.yaml

platform/observability/
  kube-prometheus-stack/
    values.yaml
  loki/
    values.yaml
  promtail/
    values.yaml
  tempo/
    values.yaml
  otel-collector/
    values.yaml
```

Rules:
- `apps/` contains Argo CD `Application` manifests only
- `platform/observability/` contains chart values and supporting manifests only
- observability apps should sync before `shahrah-production` tracing integrations depend on them

## Phase Plan
## Phase 0 - Preconditions and Guardrails
### Goals
- finalize architecture decisions
- freeze retention defaults
- define access policy
- identify dangerous logging behavior before centralization

### Actions
- confirm internal access path for Grafana:
  - VPN
  - bastion tunnel
  - authenticated private ingress
- confirm alert delivery destination:
  - email
  - Telegram
  - Slack
  - webhook
- set initial retention:
  - Prometheus: `15d`
  - Loki: `7d`
  - Tempo: `3d`
- confirm Longhorn capacity budget for observability PVCs
- review `services/control-plane/internal/server/server.go` logging format and disable body logging for production

### Exit criteria
- retention values approved
- access model approved
- alert receiver chosen
- production logging policy approved

## Phase 1 - Metrics Platform
### Goals
- establish cluster-level visibility first
- make node, pod, namespace, and ingress health visible

### Deliverables
- Argo CD application for `kube-prometheus-stack`
- Longhorn-backed PVC for Prometheus
- Longhorn-backed PVC for Grafana if persistence enabled
- base dashboards:
  - cluster overview
  - node health
  - namespace usage
  - pod restarts
  - ingress-nginx overview

### Recommended defaults
- Prometheus:
  - replicas: `1`
  - retention: `15d`
  - storage: Longhorn PVC
- Grafana:
  - replicas: `1`
  - persistence: enabled
- Alertmanager:
  - replicas: `1`
  - persistence: enabled if local state retention matters

### Alerts to enable now
- node not ready
- memory pressure
- disk pressure
- pod crashloop
- high restart count
- PVC usage high
- ingress controller unavailable

### Exit criteria
- Prometheus targets healthy
- dashboards render real cluster data
- alert test notification reaches the chosen receiver

## Phase 2 - Centralized Logging
### Goals
- collect container logs from all K3s nodes
- make operational troubleshooting possible without direct node access

### Deliverables
- Argo CD application for `Loki`
- Argo CD application for `Promtail`
- Longhorn-backed PVC for Loki
- Grafana datasource for Loki

### Log labeling requirements
Attach at least:
- `namespace`
- `pod`
- `container`
- `node`
- `app`

### Hard requirements before go-live
- disable or redact request and response body logging in production
- redact:
  - `Authorization`
  - `Cookie`
  - `Set-Cookie`
  - API keys
  - bearer tokens
- confirm prompt and model response payloads are not centrally logged by default

### Exit criteria
- `control-plane`, `tenant-admin`, `support-admin`, and `ingress-nginx` logs visible in Grafana
- log volume remains within expected ingest budget
- sensitive fields are confirmed redacted

## Phase 3 - Alerting Hardening
### Goals
- move from platform-default alerts to Shahrah-operational alerts

### Deliverables
- alert routing policy
- severity labels:
  - `critical`
  - `warning`
  - `info`
- runbook links in alert annotations

### Shahrah-specific alerts
- `control-plane` not ready
- `control-plane` restart spike
- high 5xx rate on public ingress
- p95 latency increase
- Redis unavailable
- ClickHouse unavailable
- Typesense unavailable
- external PostgreSQL connectivity degraded

### Exit criteria
- test alerts are delivered
- alert noise level is acceptable
- every enabled alert has an owner and a response action

## Phase 4 - Tracing Pipeline
### Goals
- provide production tracing for `control-plane`
- keep app code decoupled from specific trace backend endpoints

### Deliverables
- Argo CD application for `Tempo`
- Argo CD application for `OpenTelemetry Collector`
- Tempo datasource in Grafana
- OTLP pipeline:
  - `control-plane -> OTel Collector -> Tempo`

### Required `control-plane` configuration
These environment variables should be carried via Helm values into production:
- `TRACE_OTLP_ENDPOINT`
- `TRACE_OTLP_INSECURE`
- `TRACE_OTLP_SERVICE_NAME`
- `TRACE_DEFAULT_SAMPLE_RATE`
- `TRACE_SLOW_REQUEST_MS`
- `TRACE_CAPTURE_POLICY`
- `TRACE_PAYLOAD_ENCRYPTION_KEY`
- `TRACE_PII_HASH_SALT`

### Initial production defaults
- `TRACE_DEFAULT_SAMPLE_RATE=0.01`
- slow requests always sampled
- `TRACE_CAPTURE_POLICY=metadata-only`
- payload capture enabled only for explicit debugging workflows

### Exit criteria
- traces appear in Tempo
- Grafana links traces from operational views
- trace privacy settings are enforced

## Phase 5 - Application Metrics Instrumentation
### Current gap
`control-plane` currently exposes:
- `/health`
- `/ready`
- internal admin query endpoints for metrics and traces

It does not currently expose a standard Prometheus scrape endpoint such as `/metrics`.

### Goals
- add first-class app instrumentation
- expose scrape-compatible metrics for Prometheus

### Deliverables
- Prometheus instrumentation in `services/control-plane`
- `/metrics` endpoint
- `ServiceMonitor` for `control-plane`

### Required metric families
- HTTP request count
- HTTP request duration histogram
- HTTP response status counts
- upstream/provider latency
- provider error counts
- cache hit/miss
- rate-limit rejections
- budget-limit rejections
- trace exporter queue depth
- metrics exporter queue depth

### Exit criteria
- Prometheus scrapes `control-plane`
- Grafana dashboards show app-level latency and error rate
- alerts can trigger from app metrics instead of infrastructure proxies only

## Phase 6 - Dashboards and SLOs
### Goals
- expose platform and product health in one place

### Required dashboards
- cluster overview
- node health
- ingress overview
- `control-plane` service overview
- dependency health
- logs and traces troubleshooting dashboard

### Initial SLOs
- `control-plane` availability
- HTTP success rate
- p95 latency
- alert delivery health

### Exit criteria
- dashboards are stored as code
- SLO queries are stable
- operational handoff can use Grafana as the primary troubleshooting surface

## Phase 7 - Security and Hardening
### Required controls
- `NetworkPolicy` for the `observability` namespace
- no anonymous Grafana access
- rotated Grafana admin credentials
- secrets delivered through a safe Kubernetes secret workflow
- explicit resource requests and limits for all observability components

### Recommended future controls
- Grafana SSO
- object storage backends for Loki and Tempo
- mTLS where supported between internal telemetry components

### Exit criteria
- observability plane is not publicly exposed without auth
- all secrets are out of plaintext git manifests
- resource starvation tests do not destabilize the cluster

## Initial Sizing Guidance
These are starting points only and must be verified under real load.

### `kube-prometheus-stack`
- Prometheus:
  - CPU request: `500m`
  - memory request: `2Gi`
  - PVC: `50Gi`
- Grafana:
  - CPU request: `100m`
  - memory request: `512Mi`
  - PVC: `10Gi`
- Alertmanager:
  - CPU request: `100m`
  - memory request: `256Mi`
  - PVC: `5Gi`

### Logging
- Loki:
  - CPU request: `500m`
  - memory request: `1Gi`
  - PVC: `50Gi`
- Promtail:
  - CPU request: `100m`
  - memory request: `128Mi`

### Tracing
- Tempo:
  - CPU request: `250m`
  - memory request: `1Gi`
  - PVC: `20Gi`
- OTel Collector:
  - CPU request: `200m`
  - memory request: `256Mi`

## Risks
### Immediate risks
- current `control-plane` HTTP logger includes request and response bodies
- no standard Prometheus scrape endpoint exists yet
- observability apps are not yet present in `shahrah-gitops`
- unrestricted log ingestion can exhaust Longhorn capacity quickly

### Operational risks
- exposing Grafana publicly without auth
- enabling verbose trace payload capture by default
- enabling alerts without response ownership
- keeping retention high before real ingest baselines are known

## Acceptance Criteria
The observability platform is production-ready only when:
- Prometheus targets are healthy
- Grafana is reachable only through the approved internal access path
- Loki contains logs for key Shahrah workloads
- Alertmanager successfully routes a test alert
- Tempo stores traces from `control-plane`
- request and response payloads are not centrally logged by default
- trace capture uses metadata-only by default
- PVC utilization is monitored and within budget

## Execution Order
1. approve logging and trace privacy policy
2. add observability applications and values structure to `shahrah-gitops`
3. deploy `kube-prometheus-stack`
4. deploy `Loki` and `Promtail`
5. harden application logging
6. deploy `Tempo` and `OpenTelemetry Collector`
7. wire tracing env vars into `control-plane`
8. add Prometheus `/metrics` endpoint and `ServiceMonitor`
9. build Shahrah-specific dashboards and alerts
10. validate failover, storage growth, and incident response workflows

## Recommended Next Changes
### In `Shahrah`
- change production HTTP logging in `services/control-plane/internal/server/server.go`
- add Prometheus instrumentation to `services/control-plane`
- document observability runbooks and alert ownership

### In `shahrah-gitops`
- create Argo CD applications for observability platform components
- add Longhorn-backed Helm values for each component
- provision observability secrets safely
- wire internal-only ingress or access controls for Grafana

## Rollback Strategy
- observability platform rollout must not block Shahrah traffic serving
- if Loki or Tempo fail, app traffic should continue
- if trace export backend fails, `control-plane` should continue with local fallback behavior
- disable app-level instrumentation features behind config flags where possible

## Deliverables Checklist
- [ ] approved observability architecture
- [ ] GitOps app layout for observability
- [ ] base metrics stack
- [ ] centralized logging
- [ ] alert routing and runbooks
- [ ] trace pipeline
- [ ] app-level Prometheus metrics
- [ ] security hardening
- [ ] dashboards and SLOs
- [ ] production validation report
