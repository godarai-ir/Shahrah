# Silk Road EU egress proxy

Production provider traffic leaves Europe through an explicit HTTP CONNECT proxy. This path is intentionally scoped to Silk Road; control-plane probes use the same tunnel dispatch path as inference requests.

```text
control-plane -> tunnel-client -> tunnel-server -> silk-road
  -> 10.10.0.190:1443 -> nat-gw-1 loopback:1443
  -> reverse SSH tunnel -> node-4-eu loopback:443 -> Squid -> provider HTTPS API
```

The design has two fail-closed boundaries:

- Silk Road requires `SILK_ROAD_EGRESS_PROXY_URL` when `SILK_ROAD_EGRESS_PROXY_REQUIRED=true`.
- the Kubernetes NetworkPolicy allows provider egress only to DNS and `10.10.0.190:1443`.

No provider credential, private key, or proxy credential belongs in Git. Squid is bound only to loopback on `node-4-eu`; the gateway listener is private and its firewall accepts only K3s node source addresses.

## Installed files

| Repository file | Host | Target path |
| --- | --- | --- |
| `node-4-eu/squid.conf` | `node-4-eu` | `/etc/squid/squid.conf` |
| `node-4-eu/nat-gw-reverse-proxy.service` | `node-4-eu` | `/etc/systemd/system/nat-gw-reverse-proxy.service` |
| `nat-gw-1/silk-road-egress-proxy.socket` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.socket` |
| `nat-gw-1/silk-road-egress-proxy.service` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.service` |
| `../nat-gateway/nftables.conf` | `nat-gw-1` | `/etc/nftables.conf` |

The SSH private key remains at `/root/.ssh/node4_to_nat_proxy` on `node-4-eu`. Its public key on `nat-gw-1` must be restricted to the reverse listener and source address; see `nat-gw-1/authorized_keys.example`.

## Rollout order

1. Install and validate Squid on `node-4-eu`.
2. Install the restricted public key on `nat-gw-1`, then enable `nat-gw-reverse-proxy.service` on `node-4-eu`.
3. Install both gateway units and enable only `silk-road-egress-proxy.socket`; the service is socket-activated.
4. Validate and atomically load the tracked nftables rules.
5. From each K3s node, verify `curl --proxy http://10.10.0.190:1443 https://www.cloudflare.com/cdn-cgi/trace` reports the EU address.
6. Merge/deploy the application change before the GitOps change. The application understands the explicit proxy and tunneled probe operation; the GitOps rollout makes both mandatory.
7. Probe a provider credential and verify Silk Road logs report `egress proxy configured=true required=true` without logging the proxy URL or credentials.

## Host verification

```bash
# node-4-eu
sudo squid -k parse
systemctl is-active squid nat-gw-reverse-proxy
ss -lntp | grep ':443\|:1443'

# nat-gw-1
systemctl is-active silk-road-egress-proxy.socket
ss -lntp | grep '10.10.0.190:1443\|127.0.0.1:1443'
sudo nft -c -f /etc/nftables.conf
```

Rollback the GitOps configuration first. Removing the required proxy environment variables restores the previous application egress behavior; only then disable the gateway socket or reverse tunnel.
