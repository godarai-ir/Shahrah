# K3s Node Configs

These are per-node config templates for the current Shahrah cluster shape.

## Nodes

- `node-1`: `10.10.0.21`
- `node-2`: `10.10.0.22`
- `node-3`: `10.10.0.23`
- private API endpoint for current bootstrap: `10.10.0.7`
- private API hostname: `k3s-api.internal`

## Before use

Replace:

- `REPLACE_WITH_FULL_SERVER_NODE_TOKEN` on `node-2` and `node-3`

## Confirmed interface

- private NIC on `node-1`, `node-2`, and `node-3`: `eth1`

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
