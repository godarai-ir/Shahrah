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
- Pod startup fails when required proxy configuration is absent or invalid, but Pod readiness does not test proxy reachability.
- End-to-end proxy health is established through systemd state, both gateway listeners, nftables counters, Squid CONNECT logs, and an external egress-IP probe from a NetworkPolicy-selected Pod.

## Production validation

The production rollout was validated on 2026-08-02. Argo CD remained `Synced/Healthy`, all transport replicas were Ready, direct HTTPS from the settled Silk Road test Pod was blocked, and proxy tests from all three K3s nodes exited as `130.185.121.142` in `DE`.

A synthetic Gemini `credential_probe`, injected at Tunnel Client with the same tunnel contract emitted by Control Plane, returned Gemini's `API_KEY_INVALID` application response. This proves the downstream path through Tunnel Client, Tunnel Server, Silk Road, the private gateway bridge, the reverse SSH tunnel, and Squid reached the provider. A production credential was not used or modified by this smoke test.

The K3s policy controller showed a short convergence interval for a newly created diagnostic Pod. Operational smoke tests therefore wait 15 seconds after the Pod becomes Ready before asserting that direct egress is denied. The host allowlist on `nat-gw-1:1443` remains an independent boundary and accepts only the three K3s node addresses.

When the private gateway listener was absent, Tunnel Client readiness still returned Silk Road health while provider dispatch failed with `proxyconnect tcp: dial tcp 10.10.0.190:1443: connect: connection refused`. Do not use `/ready` alone as an egress health signal.

Host-level configuration and rollout commands live in [`../infra/eu-egress-proxy/README.md`](../infra/eu-egress-proxy/README.md).
