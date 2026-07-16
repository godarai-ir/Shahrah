# HAProxy + Keepalived edge runbook

These files document Shahrah's live K3s edge pattern: HAProxy handles public TCP passthrough and Keepalived is used for the private Kubernetes API failover design.

Nginx is not the host-level HA layer on `lb-1` or `lb-2`. TLS is passed through HAProxy and terminates in Kubernetes through `ingress-nginx`.

The K3s nodes are private-only. Their outbound internet access is provided by NAT on `lb-1` and `lb-2`.

## Topology

- `lb-1`: `10.10.0.7`
- `lb-2`: `10.10.0.8`
- `lb-1` public floating IP: `37.32.31.230`
- `lb-2` public floating IP: `37.32.31.139`
- private K8s API VIP: `10.10.0.6`
- current working private K8s API endpoint: `10.10.0.6:6443`
- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`

## Responsibilities

- public `80/tcp` -> ingress-nginx NodePort `30081`
- public `443/tcp` -> ingress-nginx NodePort `30444`
- private `6443/tcp` -> K3s servers `:6443`
- K3s node egress -> NAT on `lb-1` / `lb-2`

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
- public floating IP traffic is delivered to `eth0`; keep the default route via `10.10.0.1`
- do not restore the old fixed-public-IP `eth1` netplan block after the floating IP cutover
- keep `k3s-egress-nat.service` active on both load balancers; K3s nodes no longer have public IPv4 addresses
- provider panel firewall rules are the active perimeter control; OS-level UFW is not currently used
- describe this layer as HAProxy + Keepalived; the host-level edge is not Nginx

## Live notes from 2026-07-16 audit

- `haproxy.service` is active and enabled on both load balancers.
- `keepalived.service` is active and enabled on both load balancers.
- `nginx` is not installed on either load balancer.
- public fixed IPs were removed from the load balancers and replaced with floating IPs on 2026-07-16.
- `lb-1` uses default route `10.10.0.1` on `eth0` for `37.32.31.230`.
- `lb-2` uses default route `10.10.0.1` on `eth0` for `37.32.31.139`.
- after disabling private-network Port Security on `lb-1` and `lb-2`, VIP `10.10.0.6` became reachable from bastion.
- HAProxy on both load balancers was aligned to `bind 10.10.0.6:6443`.
- from bastion, `10.10.0.6:6443` is reachable and Kubernetes API returns `401 Unauthorized` for an unauthenticated `/version` request.
- a controlled Keepalived failover from `lb-1` to `lb-2` was tested successfully; the VIP moved to `lb-2` and API access stayed reachable.
- after restoring `lb-1`, the VIP moved back to `lb-1` because it has the higher priority.
- public IPv4 addresses were removed from `node-1`, `node-2`, and `node-3`.
- K3s node default routes now point to `10.10.0.6`.
- egress from the K3s nodes was validated through VIP/NAT with Docker Registry returning `401` and GitHub returning `200`.

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

- `37.32.31.230:80/443` succeeds.
- `37.32.31.139:80/443` succeeds.
- `10.10.0.6:6443` succeeds from bastion.
- an unauthenticated request to `https://10.10.0.6:6443/version` returns Kubernetes API `401 Unauthorized`.
- `10.10.0.6` appears on exactly one load balancer at a time.
- K3s nodes can reach Docker Registry and GitHub through VIP/NAT.

## Rollout checklist

1. apply `sysctl.d/60-haproxy-nonlocal-bind.conf`
2. install HAProxy on both `lb-1` and `lb-2`
3. place `haproxy.cfg`
4. validate config with `haproxy -c -f /etc/haproxy/haproxy.cfg`
5. start HAProxy
6. configure Keepalived for VIP `10.10.0.6`
