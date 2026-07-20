# K3s Node Configs

These are per-node config templates for the current Shahrah cluster shape.

## Nodes

- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`
- private API endpoint: `10.10.0.6`
- private API hostname: `k3s-api.internal`

The nodes are private-only and do not have public IPv4 addresses. Their only default route points to the dedicated gateway `10.10.0.190`; the private VIP `10.10.0.6` is reserved for the Kubernetes API.

## Before use

Replace:

- `REPLACE_WITH_FULL_SERVER_NODE_TOKEN` on `node-2` and `node-3`

## Confirmed interface

- private NIC on `node-1`, `node-2`, and `node-3`: `eth1`

## Current egress model

Outbound internet access for all K3s nodes is provided by `nat-gw-1` (`10.10.0.190`) and is SNATed to `188.121.120.198`.

The cutover completed on 2026-07-20 in the order `node-2`, `node-3`, then `node-1`. All three nodes passed host and pinned-Pod checks for CoreDNS, HTTPS, Docker Registry and database TCP/5432. Their resolver pair is `1.1.1.1` and `1.0.0.1`; `8.8.8.8` was excluded after an intermittent direct-query timeout.

Expected node route:

```bash
default via 10.10.0.190 dev eth1
```

The legacy LB NAT service remains enabled only during the 24-hour rollback window and should have no production node traffic:

```bash
systemctl is-active k3s-egress-nat.service
```

Each node keeps `/var/lib/cloud/data/upgraded-network` and a Cloud-Init network-disable override because provider metadata still advertises removed public NIC MACs. Do not run live `netplan apply` or `netplan try` on a K3s node: testing on `node-2` tore down its only interface long enough to make the node NotReady. Apply runtime route/DNS changes first and validate any reboot only in a controlled maintenance window.

## Install flow

1. copy `node-1.config.yaml` to `/etc/rancher/k3s/config.yaml` on `node-1`
2. install K3s server on `node-1`
3. read the token from `/var/lib/rancher/k3s/server/node-token`
4. fill the full token into `node-2.config.yaml` and `node-3.config.yaml`
5. copy each config to `/etc/rancher/k3s/config.yaml`
6. install K3s server on `node-2` and `node-3`

Important:

- use the full contents of `/var/lib/rancher/k3s/server/node-token`
- do not truncate at the first `::`
- do not commit the real token into Git

## Built-in addons

These configs intentionally disable:

- `servicelb`
- `traefik`

Reason:

- ingress will be provided by Argo CD-managed `ingress-nginx`
- public traffic is already handled by dedicated HAProxy load balancers

## Operator runbooks

- `windows-kubectl-access.md`
  - Windows and PowerShell access to the cluster through the bastion and private API load balancer
