# Silk Road EU egress proxy

Production provider traffic leaves Europe through an explicit HTTP CONNECT proxy. This path is intentionally scoped to Silk Road; Control Plane probes use the same direct provider-egress dispatch path as inference requests.

```text
control-plane -> silk-road
  -> 10.10.0.190:1443 -> nat-gw-1 loopback:1443
  -> reverse SSH tunnel -> node-4-eu loopback:443 -> Squid -> provider HTTPS API
```

The design has two fail-closed boundaries:

- Silk Road requires `SILK_ROAD_EGRESS_PROXY_URL` when `SILK_ROAD_EGRESS_PROXY_REQUIRED=true`.
- the Kubernetes NetworkPolicy allows provider egress only to DNS and `10.10.0.190:1443`.

No provider credential, private key, or proxy credential belongs in Git. Squid is bound only to loopback on `node-4-eu`; the gateway listener is private and its firewall accepts only K3s node source addresses.

## Production status

The path was installed and validated on 2026-08-02:

- `squid.service` and `nat-gw-reverse-proxy.service` are enabled and active on `node-4-eu`.
- `silk-road-egress-proxy.socket` is enabled and active on `nat-gw-1`; its socket-activated service became active after the first connection.
- both `127.0.0.1:1443` and `10.10.0.190:1443` listen on `nat-gw-1`.
- proxy smoke tests pinned to `node-1`, `node-2`, and `node-3` all reported public IP `130.185.121.142` and Cloudflare location `DE`.
- direct HTTPS from a settled Pod selected by the Silk Road NetworkPolicy was blocked.
- before the internal relay removal, a synthetic `credential_probe` traversed the legacy relays and Silk Road, then reached Gemini through Squid and returned the expected invalid-key application response. This validates the Silk Road-to-provider segment, not the new direct in-cluster hop.

The pre-change gateway firewall backup is `/etc/nftables.conf.pre-eu-egress-20260802T095559Z` on `nat-gw-1`. No reboot validation was performed during this rollout.

## Installed files

| Repository file | Host | Target path |
| --- | --- | --- |
| `node-4-eu/squid.conf` | `node-4-eu` | `/etc/squid/squid.conf` |
| `node-4-eu/nat-gw-reverse-proxy.service` | `node-4-eu` | `/etc/systemd/system/nat-gw-reverse-proxy.service` |
| `nat-gw-1/silk-road-egress-proxy.socket` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.socket` |
| `nat-gw-1/silk-road-egress-proxy.service` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.service` |
| `nat-gw-1/60-reverse-tunnel-liveness.conf` | `nat-gw-1` | `/etc/ssh/sshd_config.d/60-reverse-tunnel-liveness.conf` |
| `../nat-gateway/nftables.conf` | `nat-gw-1` | `/etc/nftables.conf` |

The reverse tunnel runs as `ubuntu`. Its private key remains at `/home/ubuntu/.ssh/node4_to_nat_proxy` on `node-4-eu`; the public key is installed in `/home/ubuntu/.ssh/authorized_keys` on `nat-gw-1`. The key is restricted to source `130.185.121.142` and listener `127.0.0.1:1443`; see `nat-gw-1/authorized_keys.example`.

## Rollout order

1. Install and validate Squid on `node-4-eu`.
2. Install the SSH liveness drop-in and restricted public key on `nat-gw-1`. Validate the SSH configuration with `sshd -t`, reload `ssh.service`, install the private key for `ubuntu` on `node-4-eu`, then enable `nat-gw-reverse-proxy.service`.
3. Install both gateway units and enable only `silk-road-egress-proxy.socket`; the service is socket-activated.
4. Validate and atomically load the tracked nftables rules.
5. From a NetworkPolicy-selected Pod pinned to each K3s node, verify direct HTTPS is blocked and `curl --proxy http://10.10.0.190:1443 https://www.cloudflare.com/cdn-cgi/trace` reports `ip=130.185.121.142` and `loc=DE`. Allow the K3s policy controller to converge before interpreting the result; the rollout test used 15 seconds.
6. Publish the direct-egress Control Plane image, then roll it out with the GitOps change. Keep the temporary `TUNNEL_CLIENT_URL=http://silk-road:8083` alias for one release so the previous Control Plane image can also operate while relay workloads are pruned.
7. Probe a provider credential and verify Silk Road logs report `egress proxy configured=true required=true` without logging the proxy URL or credentials.

## Host verification

```bash
# node-4-eu
sudo squid -k parse
systemctl is-active squid nat-gw-reverse-proxy
ss -lntp | grep ':443\|:1443'

# nat-gw-1
sudo sshd -t
sudo sshd -T | grep 'clientaliveinterval\|clientalivecountmax\|tcpkeepalive'
systemctl is-active silk-road-egress-proxy.socket
systemctl is-active silk-road-egress-proxy.service
ss -lntp | grep '10.10.0.190:1443\|127.0.0.1:1443'
sudo nft -c -f /etc/nftables.conf
sudo nft list chain inet filter input
```

Silk Road readiness confirms only the local process, not live reachability of the external proxy. Verify the egress IP and inspect both `journalctl -u silk-road-egress-proxy.service` and `/var/log/squid/access.log` after changes. A missing gateway listener produces a `proxyconnect ... connection refused` dispatch error while the Pods can remain Ready.

## Local development

The local Compose stack also requires the EU proxy. Before starting the stack on
Windows, connect the VPN and keep this forward running:

```powershell
putty.exe -load node-4-eu -l ubuntu -N -L 1443:127.0.0.1:443
docker compose -f infra/compose/local.yml up -d silk-road control-plane
```

Compose gives Silk Road these defaults:

```text
SILK_ROAD_EGRESS_PROXY_URL=http://host.docker.internal:1443
SILK_ROAD_EGRESS_PROXY_REQUIRED=true
```

The expected startup log is `egress proxy configured=true required=true`. This
proxies provider dispatches and credential probes, but `/ready` still does not
test external proxy reachability. Validate with a provider probe and confirm its
CONNECT entry in `/var/log/squid/access.log` on `node-4-eu`.

Rollback the GitOps configuration first. Removing the required proxy environment variables restores the previous application egress behavior; only then disable the gateway socket or reverse tunnel. Restore the timestamped nftables backup only after validating it with `sudo nft -c -f`.
