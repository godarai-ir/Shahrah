# Silk Road EU egress proxy

Production provider traffic leaves Europe through an explicit HTTP CONNECT proxy. This path is intentionally scoped to Silk Road; Control Plane probes use the same provider-egress dispatch path as inference requests.

```text
control-plane -> silk-road
  -> 10.10.0.190:1443 on nat-gw-1
  -> systemd-socket-proxyd -> HAProxy 127.0.0.1:15443
  -> mutually authenticated TLS over TCP/443
  -> stunnel on node-4-eu 130.185.121.142:443
  -> Squid 127.0.0.1:443 -> provider HTTPS API
```

SSH is not part of the production data path. The former reverse SSH tunnel is retained, disabled, as a rollback mechanism because WireGuard is not usable on the Iran-to-Europe path.

The design has three fail-closed boundaries:

- Silk Road requires `SILK_ROAD_EGRESS_PROXY_URL` when `SILK_ROAD_EGRESS_PROXY_REQUIRED=true`.
- the Kubernetes NetworkPolicy allows provider egress only to DNS and `10.10.0.190:1443`.
- `node-4-eu:443` accepts only `nat-gw-1` at the firewall and requires a client certificate signed by the private egress CA.

No provider credential, private key, certificate, or proxy credential belongs in Git. Squid remains bound only to loopback on `node-4-eu`; the gateway listener is private and its firewall accepts only K3s node source addresses.

## Production status

The mTLS path was cut over at `2026-08-10 20:46:37 UTC` without changing the Kubernetes endpoint:

- `haproxy.service` and `silk-road-egress-proxy.socket` are enabled/active on `nat-gw-1`; the socket-activated bridge service is active, and both processes have no restarts after cutover.
- `silk-road-egress-tls-server.service` and `squid.service` are enabled/active on `node-4-eu` with no restarts after cutover.
- both Silk Road replicas returned `HTTP/1.1 200 Connection established` for Cloudflare and Moonshot through `10.10.0.190:1443`.
- direct TCP/443 from both NetworkPolicy-selected Silk Road Pods was blocked.
- an external trace through the production endpoint reported public IP `130.185.121.142` and location `DE`.
- a non-billable Moonshot `/v1/models` request reached the provider and returned the expected unauthenticated `401` response.
- the legacy `nat-gw-reverse-proxy.service` is disabled/stopped after observation; its unit, restricted SSH key, and server-side SSH policy are retained for rollback.

The rollout backups are under `/var/backups/shahrah-egress-mtls/20260810T203200Z` on both hosts. The older gateway firewall backup remains `/etc/nftables.conf.pre-eu-egress-20260802T095559Z` on `nat-gw-1`.

## Installed files

| Repository file | Host | Target path |
| --- | --- | --- |
| `node-4-eu/squid.conf` | `node-4-eu` | `/etc/squid/squid.conf` |
| `node-4-eu/silk-road-egress-tls-server.conf` | `node-4-eu` | `/etc/stunnel/silk-road-egress-tls-server.conf` |
| `node-4-eu/silk-road-egress-tls-server.service` | `node-4-eu` | `/etc/systemd/system/silk-road-egress-tls-server.service` |
| `nat-gw-1/haproxy.cfg` | `nat-gw-1` | `/etc/haproxy/haproxy.cfg` |
| `nat-gw-1/silk-road-egress-proxy.socket` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.socket` |
| `nat-gw-1/silk-road-egress-proxy.service` | `nat-gw-1` | `/etc/systemd/system/silk-road-egress-proxy.service` |

The following legacy files remain rollback-only:

| Repository file | Host | Target path |
| --- | --- | --- |
| `node-4-eu/nat-gw-reverse-proxy.service` | `node-4-eu` | `/etc/systemd/system/nat-gw-reverse-proxy.service` |
| `nat-gw-1/60-reverse-tunnel-liveness.conf` | `nat-gw-1` | `/etc/ssh/sshd_config.d/60-reverse-tunnel-liveness.conf` |
| `nat-gw-1/authorized_keys.example` | `nat-gw-1` | documentation only |

TLS material is host-local:

- CA and client material: `/etc/silk-road-egress-tls` on `nat-gw-1`.
- HAProxy CA and combined client PEM: `/etc/haproxy/certs` on `nat-gw-1`.
- CA and server material: `/etc/silk-road-egress-tls` on `node-4-eu`.

Source CA/client keys on `nat-gw-1` are root-only (`0600`); runtime private-key files are readable only by the relevant service group (`0640`). The leaf certificates expire on `2028-11-12`; rotate them well before that date and verify the new chain before reloading either endpoint. Move the CA private key to an offline or managed secret store when one is available.

## Firewall contract

`node-4-eu` has one UFW ingress exception for this service:

```bash
sudo ufw allow proto tcp from 188.121.120.198 to 130.185.121.142 port 443 comment 'Shahrah mTLS egress from nat-gw-1'
```

Do not broaden this rule. mTLS is still mandatory even though the source IP is restricted. The existing `nat-gw-1` nftables rule continues to restrict `10.10.0.190:1443` to the K3s node addresses.

## Rollout order

1. Back up systemd, Squid, firewall, and TLS-related configuration on both hosts.
2. Install `stunnel4` on `node-4-eu` and HAProxy on `nat-gw-1` from the distribution packages.
3. Create the private CA and client key on `nat-gw-1`; create the server key on `node-4-eu`. Transfer only CSRs and certificates, never private keys.
4. Install and start the stunnel server, then add the source-restricted UFW TCP/443 rule.
5. Install and start HAProxy on loopback port `15443`. Validate its mTLS connection and HTTP CONNECT health check before cutover.
6. Change only `silk-road-egress-proxy.service` to bridge the unchanged private socket to `127.0.0.1:15443`, then restart the bridge service.
7. Validate both Silk Road replicas, provider connectivity, direct-egress denial, service restart counters, and Squid/HAProxy logs.
8. After the observation period, disable/stop `nat-gw-reverse-proxy.service` but retain its files and restricted keys for rollback.

HAProxy is used as the client-side TLS wrapper because the tested stunnel client handshake was dropped on the Iran-to-Europe path while HAProxy's TLS handshake over the same TCP/443 path succeeded. The server side remains stunnel and requires mutual authentication.

## Host verification

```bash
# node-4-eu
sudo squid -k parse
sudo systemctl is-active squid silk-road-egress-tls-server
sudo systemctl is-enabled squid silk-road-egress-tls-server
sudo ss -lntp | grep '130.185.121.142:443\|127.0.0.1:443'
sudo journalctl -u silk-road-egress-tls-server --since '-15 min'
sudo tail -n 100 /var/log/squid/access.log
sudo ufw status numbered

# nat-gw-1
sudo haproxy -c -f /etc/haproxy/haproxy.cfg
sudo systemctl is-active haproxy silk-road-egress-proxy.socket silk-road-egress-proxy.service
sudo ss -lntp | grep '10.10.0.190:1443\|127.0.0.1:15443'
sudo journalctl -u haproxy -u silk-road-egress-proxy.service --since '-15 min'
sudo nft -c -f /etc/nftables.conf
```

HAProxy performs a CONNECT health check to `www.cloudflare.com:443` every 30 seconds. Silk Road readiness confirms only the local process, not live reachability of the external proxy, so alert on backend-down messages, service restarts, certificate expiry, and failed provider dispatches as well.

## Rollback

Rollback changes the bridge target only; the Kubernetes endpoint stays `10.10.0.190:1443`:

1. On `node-4-eu`, enable/start `nat-gw-reverse-proxy.service`.
2. On `nat-gw-1`, confirm the SSH process owns `127.0.0.1:1443`.
3. Restore the backed-up `silk-road-egress-proxy.service` whose target is `127.0.0.1:1443`, or make that single target change explicitly.
4. Run `systemctl daemon-reload` and restart only `silk-road-egress-proxy.service`; do not stop its socket.
5. From both Silk Road replicas, verify CONNECT through `10.10.0.190:1443` before stopping any mTLS component.

The legacy private key remains `/home/ubuntu/.ssh/node4_to_nat_proxy` on `node-4-eu`. Its public key on `nat-gw-1` is restricted to source `130.185.121.142` and listener `127.0.0.1:1443`. Never delete or broaden these restrictions during rollback.

## Local development

Local development may still use an operator-owned SSH forward; it is separate from the production data path. Before starting the stack on Windows, connect the VPN and keep this forward running:

```powershell
putty.exe -load node-4-eu -l ubuntu -N -L 1443:127.0.0.1:443
docker compose -f infra/compose/local.yml up -d silk-road control-plane
```

Compose gives Silk Road these defaults:

```text
SILK_ROAD_EGRESS_PROXY_URL=http://host.docker.internal:1443
SILK_ROAD_EGRESS_PROXY_REQUIRED=true
```

The expected startup log is `egress proxy configured=true required=true`. Validate with a provider probe and confirm its CONNECT entry in `/var/log/squid/access.log` on `node-4-eu`.
