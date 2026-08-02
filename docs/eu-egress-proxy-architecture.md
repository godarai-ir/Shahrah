# EU provider egress architecture

## Decision

All outbound provider API requests made by Silk Road, including non-billable credential probes, use the explicit HTTP CONNECT proxy hosted through `node-4-eu`. Control Plane never calls provider validation endpoints directly in production; it sends a typed `credential_probe` operation through Tunnel Client and Tunnel Server to Silk Road.

Standard `HTTP_PROXY` and `HTTPS_PROXY` variables are retained for compatibility, but `SILK_ROAD_EGRESS_PROXY_URL` is the application-owned source of truth. Production sets `SILK_ROAD_EGRESS_PROXY_REQUIRED=true`, so an absent or invalid explicit proxy prevents Silk Road from starting.

## Request paths

Inference:

```text
caller -> control-plane -> tunnel-client -> tunnel-server -> silk-road -> EU proxy -> provider
```

Credential validation:

```text
control-plane credential service -> credential_probe tunnel operation
  -> tunnel-client -> tunnel-server -> silk-road allowlisted probe builder
  -> EU proxy -> provider
```

Silk Road builds probe URLs from provider-specific contracts and enforces public HTTPS host allowlists before dispatch. Provider credentials remain inside the existing encrypted tunnel request and are never encoded into the proxy URL.

## Operational contract

- Application release must precede the GitOps rollout that enables required proxy mode.
- Host infrastructure must be healthy before the NetworkPolicy is enabled.
- The Silk Road NetworkPolicy permits DNS and the private gateway endpoint only.
- Direct provider egress is considered a configuration failure, not an automatic fallback.
- Proxy health is observable through pod readiness/startup, systemd state, Squid access logs, and an external egress-IP probe.

Host-level configuration and rollout commands live in [`../infra/eu-egress-proxy/README.md`](../infra/eu-egress-proxy/README.md).
