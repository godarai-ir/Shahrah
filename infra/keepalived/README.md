# Keepalived Plan

These files provide the private Kubernetes API VIP `10.10.0.6` on `lb-1` and `lb-2`.

## Nodes

- `lb-1`: `10.10.0.7`
- `lb-2`: `10.10.0.8`
- VIP: `10.10.0.6/24`

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

## Validation

1. `keepalived -t -f /etc/keepalived/keepalived.conf`
2. `ip addr show`
3. from bastion: `nc -vz 10.10.0.6 6443`
4. stop HAProxy on `lb-1` and verify the VIP moves to `lb-2`

If validation says `Disabling track script ... due to insecure`, fix ownership and mode:

```bash
sudo chown root:root /etc/keepalived/check_haproxy.sh
sudo chmod 700 /etc/keepalived/check_haproxy.sh
sudo chown root:root /etc/keepalived
sudo chmod 755 /etc/keepalived
```
