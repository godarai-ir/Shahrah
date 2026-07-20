# HAProxy + Keepalived edge runbook

These files document Shahrah's live K3s edge pattern: HAProxy handles public TCP passthrough and Keepalived is used for the private Kubernetes API failover design.

Nginx is not the host-level HA layer on `lb-1` or `lb-2`. TLS is passed through HAProxy and terminates in Kubernetes through `ingress-nginx`.

The K3s nodes are private-only. Their outbound internet access is provided by the dedicated `nat-gw-1`; `lb-1` and `lb-2` retain only public ingress and private Kubernetes API HA responsibilities. Their legacy NAT services remain enabled temporarily for rollback, without production node traffic.

## Topology

- `lb-1`: `10.10.0.7`
- `lb-2`: `10.10.0.8`
- `lb-1` direct public IPv4 on `eth1`: `185.231.180.132/22` (gateway `185.231.180.1`)
- `lb-2` direct public IPv4 on `eth1`: `37.152.187.55/23` (gateway `37.152.186.1`)
- detached rollback-only Floating IPs: `37.32.31.230` and `37.32.31.139`
- private K8s API VIP: `10.10.0.6`
- current working private K8s API endpoint: `10.10.0.6:6443`
- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`

## Responsibilities

- public `80/tcp` -> ingress-nginx NodePort `30081`
- public `443/tcp` -> ingress-nginx NodePort `30444`
- private `6443/tcp` -> K3s servers `:6443`
- K3s node egress -> `nat-gw-1` (`10.10.0.190`), outside this HAProxy layer

After the dedicated-gateway cutover on 2026-07-20, HAProxy and Keepalived keep the public ingress and private Kubernetes API VIP roles; they no longer carry node egress. Keep the legacy NAT service enabled, without traffic, for at least 24 hours after the successful cutover.

## Files

- `haproxy.cfg`: current `lb-1` config, with private API listener on `10.10.0.6:6443`.
- `lb-2.haproxy.cfg`: current `lb-2` config, with private API listener on `10.10.0.6:6443`.
- `60-haproxy-nonlocal-bind.conf`: required so the backup node can bind the VIP before it owns the address.
- `k3s-egress-nat.sh`: installed as `/usr/local/sbin/k3s-egress-nat.sh` on both load balancers.
- `70-k3s-egress-nat.conf`: installed as `/etc/sysctl.d/70-k3s-egress-nat.conf` on both load balancers.
- `k3s-egress-nat.service`: installed as `/etc/systemd/system/k3s-egress-nat.service` on both load balancers.

## Important

- `6443` must not be reachable from the public internet
- if HAProxy binds to a floating VIP on both nodes, enable `net.ipv4.ip_nonlocal_bind=1`
- private-network Port Security must be disabled for `lb-1` and `lb-2` so VRRP/VIP ARP can work
- private NIC on both load balancers is confirmed as `eth0`
- public ingress is delivered directly to `eth1`; each LB has exactly one default route through its public gateway
- keep private `eth0` free of default routes; the existing static resolver entries remain defined there
- keep `/etc/netplan/01-lb-public.yaml`, Cloud-Init network lock and `lb-public-filter.service` aligned with the templates under `infra/lb-public/`
- keep `k3s-egress-nat.service` active on both load balancers only through the 24-hour rollback window
- do not disable the legacy NAT services until all nodes use `10.10.0.190` and the 24-hour rollback window has passed
- provider panel rules remain relevant, and nftables on each LB additionally restricts public `eth1` input to HTTP/HTTPS, DHCP, essential ICMP and established traffic
- describe this layer as HAProxy + Keepalived; the host-level edge is not Nginx

## Live notes from 2026-07-16 audit

- `haproxy.service` is active and enabled on both load balancers.
- `keepalived.service` is active and enabled on both load balancers.
- `nginx` is not installed on either load balancer.
- public fixed IPs were removed from the load balancers and replaced with floating IPs on 2026-07-16.
- at the time of this historical audit, `lb-1` used default route `10.10.0.1` on `eth0` for `37.32.31.230`; the 2026-07-20 direct-public cutover replaced it.
- at the time of this historical audit, `lb-2` used default route `10.10.0.1` on `eth0` for `37.32.31.139`; the 2026-07-20 direct-public cutover replaced it.
- after disabling private-network Port Security on `lb-1` and `lb-2`, VIP `10.10.0.6` became reachable from bastion.
- HAProxy on both load balancers was aligned to `bind 10.10.0.6:6443`.
- from bastion, `10.10.0.6:6443` is reachable and Kubernetes API returns `401 Unauthorized` for an unauthenticated `/version` request.
- a controlled Keepalived failover from `lb-1` to `lb-2` was tested successfully; the VIP moved to `lb-2` and API access stayed reachable.
- after restoring `lb-1`, the VIP moved back to `lb-1` because it has the higher priority.
- public IPv4 addresses were removed from `node-1`, `node-2`, and `node-3`.
- K3s node default routes pointed to `10.10.0.6` during the 2026-07-16 design; they were moved to `10.10.0.190` on 2026-07-20.
- egress through `nat-gw-1` was validated from every node and pinned Pod, with Docker Registry returning `401`, GitHub returning `200` and public source `188.121.120.198`.

## Direct-public ingress cutover on 2026-07-20

- the Floating IP path timed out while private HAProxy listeners stayed healthy
- direct public NICs were attached to both load balancers and their Netplan files were consolidated
- both direct origins return HTTPS `200` with SNI `godarai.ir`; the live domain also returns `200`
- DNS `@` and `*` contain both direct IPs with TTL two minutes; observed authoritative responses currently select `lb-2`, so provider-side DNS failover remains a separate validation item
- public SSH to both direct IPs is blocked by the host nftables filter
- old Floating IPs are detached but retained for rollback

On 2026-07-20 Port Security was found enabled again on both private LB ports. It was disabled with the provider confirmation flow and private SSH, HTTP, HTTPS, and `10.10.0.6:6443` connectivity from Bastion immediately recovered.

Treat `10.10.0.6:6443` as the official private Kubernetes API endpoint. Keep direct listeners on `10.10.0.7:6443` or `10.10.0.8:6443` out of the operational path unless a break-glass rollback is needed.

## Validation

Run on each load balancer:

```bash
systemctl is-active haproxy keepalived
systemctl is-active k3s-egress-nat.service
haproxy -c -f /etc/haproxy/haproxy.cfg
ss -tulpen | grep -E ':(80|443|6443)\b'
iptables -t nat -S POSTROUTING | grep '10.10.0.2'
```

Run from bastion:

```bash
nc -vz 10.10.0.7 6443
nc -vz 10.10.0.8 6443
nc -vz 10.10.0.6 6443
```

Expected current result:

- both direct public IPs return successful HTTP/HTTPS responses.
- `10.10.0.6:6443` succeeds from bastion.
- an unauthenticated request to `https://10.10.0.6:6443/version` returns Kubernetes API `401 Unauthorized`.
- `10.10.0.6` appears on exactly one load balancer at a time.
- K3s nodes can reach Docker Registry and GitHub through `10.10.0.190`; the VIP is not their default gateway.

## Rollout checklist

1. apply `sysctl.d/60-haproxy-nonlocal-bind.conf`
2. install HAProxy on both `lb-1` and `lb-2`
3. place `haproxy.cfg`
4. validate config with `haproxy -c -f /etc/haproxy/haproxy.cfg`
5. start HAProxy
6. configure Keepalived for VIP `10.10.0.6`
