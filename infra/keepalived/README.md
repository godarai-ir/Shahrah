# Keepalived private API runbook

These files define the intended private Kubernetes API VIP `10.10.0.6` on `lb-1` and `lb-2`.

Keepalived is part of the live HAProxy + Keepalived edge. Nginx is not the host-level HA layer.

## Nodes

- `lb-1`: `10.10.0.7`
- `lb-2`: `10.10.0.8`
- VIP: `10.10.0.6/24`
- current working API endpoint observed from bastion: `10.10.0.6:6443`

## What to replace

- `AUTH_PASS` with a shared password of at most `8` characters

## Important Keepalived note

For VRRP `auth_type PASS`, `auth_pass` is limited to 8 characters.

Use the same value on both nodes, for example:

- `K8sVip01`

## Confirmed interface

- private NIC on both `lb-1` and `lb-2`: `eth0`

## Expected behavior

- `lb-1` starts as `MASTER`
- `lb-2` starts as `BACKUP`
- if HAProxy or the node fails, VIP `10.10.0.6` moves to the healthy node

## Live validation from 2026-07-16

The private VIP behavior is proven end-to-end after disabling private-network Port Security on `lb-1` and `lb-2`:

- `10.10.0.6/24` is owned by `lb-1` in the normal state.
- HAProxy on both `lb-1` and `lb-2` binds `10.10.0.6:6443`.
- From bastion, `10.10.0.6:6443` is reachable.
- An unauthenticated request to `https://10.10.0.6:6443/version` returns Kubernetes API `401 Unauthorized`.
- During a controlled failover, the VIP moved from `lb-1` to `lb-2` and API access stayed reachable.
- After `lb-1` returned, the VIP moved back to `lb-1`.

Use `10.10.0.6:6443` as the operational private Kubernetes API endpoint. Keep Port Security disabled only for the load balancers that own the VIP; do not disable it for the K3s nodes unless another floating-address design explicitly needs that.

## Validation

1. `keepalived -t -f /etc/keepalived/keepalived.conf`
2. `ip addr show`
3. from bastion: `nc -vz 10.10.0.6 6443`
4. from bastion: `curl -k https://10.10.0.6:6443/version`
5. in a maintenance window, stop Keepalived or HAProxy on `lb-1` and verify the VIP moves to `lb-2`
6. restore `lb-1` and verify the VIP moves back to the higher-priority node

If validation says `Disabling track script ... due to insecure`, fix ownership and mode:

```bash
sudo chown root:root /etc/keepalived/check_haproxy.sh
sudo chmod 700 /etc/keepalived/check_haproxy.sh
sudo chown root:root /etc/keepalived
sudo chmod 755 /etc/keepalived
```
