# EU provider egress architecture

## Decision

All outbound provider API requests made by Silk Road, including non-billable credential probes, use the explicit HTTP CONNECT proxy hosted through `node-4-eu`. Control Plane never calls provider validation endpoints directly in production; it sends a typed `credential_probe` operation directly to Silk Road.

Standard `HTTP_PROXY` and `HTTPS_PROXY` variables are retained for compatibility, but `SILK_ROAD_EGRESS_PROXY_URL` is the application-owned source of truth. Production sets `SILK_ROAD_EGRESS_PROXY_REQUIRED=true`, so an absent or invalid explicit proxy prevents Silk Road from starting.

## Request paths

Inference:

```text
caller -> control-plane -> silk-road -> EU proxy -> provider
```

Credential validation:

```text
control-plane credential service -> credential_probe provider-egress operation
  -> silk-road allowlisted probe builder
  -> EU proxy -> provider
```

Silk Road builds probe URLs from provider-specific contracts and enforces public HTTPS host allowlists before dispatch. Provider credentials remain inside the internal Control Plane-to-Silk Road request and are never encoded into the proxy URL. The current in-cluster hop is HTTP, so namespace access and NetworkPolicy restrictions are part of the credential boundary.

## Operational contract

- Application release must precede the GitOps rollout that enables required proxy mode.
- Host infrastructure must be healthy before the NetworkPolicy is enabled.
- The Silk Road NetworkPolicy permits DNS and the private gateway endpoint only.
- Direct provider egress is considered a configuration failure, not an automatic fallback.
- Pod startup fails when required proxy configuration is absent or invalid, but Pod readiness does not test proxy reachability.
- End-to-end proxy health is established through systemd state, the private gateway and loopback listeners, HAProxy backend health, mTLS server logs, firewall counters, Squid CONNECT logs, and an external egress-IP probe from a NetworkPolicy-selected Pod.

## Production validation

The original EU egress segment was validated on 2026-08-02. Its reverse SSH transport was replaced on 2026-08-10 with a mutually authenticated TLS connection over TCP/443. The private Kubernetes-facing endpoint remained `10.10.0.190:1443`, so no application or NetworkPolicy change was required for the transport cutover.

The production path is now:

```text
silk-road -> nat-gw-1:1443 -> HAProxy loopback:15443
  -> mTLS TCP/443 -> node-4-eu stunnel -> Squid loopback:443 -> provider
```

HAProxy authenticates `node-4-eu` against a private CA and presents a dedicated `nat-gw-1` client certificate. The server firewall permits TCP/443 only from the public address of `nat-gw-1`, and stunnel rejects clients without a trusted certificate. SSH is not in the active data path; the restricted reverse-tunnel configuration is retained, disabled, for rollback.

After cutover, both Silk Road replicas completed HTTP CONNECT requests to Cloudflare and Moonshot, direct TCP/443 remained blocked, the observed egress was `130.185.121.142` in `DE`, and all new-path services reported zero restarts. A non-billable Moonshot request returned the expected unauthenticated `401`, proving provider-level reachability without using or modifying a production credential.

The K3s policy controller showed a short convergence interval for a newly created diagnostic Pod. Operational smoke tests therefore wait 15 seconds after the Pod becomes Ready before asserting that direct egress is denied. The host allowlist on `nat-gw-1:1443` remains an independent boundary and accepts only the three K3s node addresses. The public TCP/443 listener on `node-4-eu` is independently restricted to `nat-gw-1` and requires mTLS.

When the private gateway listener is absent, Silk Road readiness can remain healthy while provider dispatch fails with `proxyconnect tcp: dial tcp 10.10.0.190:1443: connect: connection refused`. Do not use `/ready` alone as an egress health signal.

Host-level configuration and rollout commands live in [`../infra/eu-egress-proxy/README.md`](../infra/eu-egress-proxy/README.md).
