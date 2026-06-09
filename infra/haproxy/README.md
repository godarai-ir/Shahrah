# HAProxy Plan

These files are the starting templates for Shahrah's K3s edge and private API access.

## Topology

- `lb-1`: `10.10.0.7`
- `lb-2`: `10.10.0.8`
- private K8s API VIP: `10.10.0.6`
- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`

## Responsibilities

- public `80/tcp` -> ingress-nginx NodePort `30081`
- public `443/tcp` -> ingress-nginx NodePort `30444`
- private `6443/tcp` on VIP `10.10.0.6` -> K3s servers `:6443`

## Important

- `6443` must not be reachable from the public internet
- if HAProxy binds to the VIP on both nodes, enable `net.ipv4.ip_nonlocal_bind=1`
- private NIC on both load balancers is confirmed as `eth0`

## Suggested rollout

1. apply `sysctl.d/60-haproxy-nonlocal-bind.conf`
2. install HAProxy on both `lb-1` and `lb-2`
3. place `haproxy.cfg`
4. validate config with `haproxy -c -f /etc/haproxy/haproxy.cfg`
5. start HAProxy
6. configure Keepalived for VIP `10.10.0.6`
