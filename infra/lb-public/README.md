# Direct public ingress for load balancers

Production ingress no longer depends on the provider Floating IP path. Each load balancer owns a direct public NIC while Keepalived continues to manage only the private Kubernetes API VIP.

## Live topology

| Host | Private NIC | Public NIC | Public gateway | Role |
| --- | --- | --- | --- | --- |
| `lb-1` | `eth0`, `10.10.0.7/24`, MAC `fa:16:3e:d8:b2:72` | `eth1`, `185.231.180.132/22`, MAC `fa:16:3e:76:51:44` | `185.231.180.1` | Keepalived MASTER |
| `lb-2` | `eth0`, `10.10.0.8/24`, MAC `fa:16:3e:cc:41:ab` | `eth1`, `37.152.187.55/23`, MAC `fa:16:3e:50:d1:17` | `37.152.186.1` | Keepalived BACKUP |

The private VIP `10.10.0.6/24` remains the Kubernetes API endpoint and normally resides on `lb-1`. It is not a public ingress VIP and is not a node default gateway.

DNS records `@` and `*` contain both direct public IPs with TTL two minutes. The provider's authoritative DNS may select one value per client/query path; registering two values alone must not be treated as proof of health-checked failover.

## Host configuration

- `lb-1.netplan.yaml` and `lb-2.netplan.yaml` are the versioned copies of `/etc/netplan/01-lb-public.yaml`.
- Private `eth0` is static and has no default route. Existing static resolvers `1.1.1.1` and `8.8.8.8` remain declared on this link.
- Public `eth1` uses DHCP and is the only source of a default route and resolver configuration.
- Cloud-Init network configuration is disabled to prevent provider hotplug metadata from replacing Netplan at boot.
- `lb-public-filter.service` loads `lb-public-filter.nft` and allows only HTTP/HTTPS, DHCP, essential ICMP and established traffic on public `eth1`.
- Public SSH is intentionally blocked. Management continues through Bastion and the private addresses.

## Validation

```bash
ip -4 route show default
systemctl is-active haproxy keepalived lb-public-filter.service
systemctl is-enabled lb-public-filter.service
nft list table inet lb_public_filter
ss -lnt '( sport = :80 or sport = :443 or sport = :6443 )'
```

From an external system, validate each origin independently with SNI:

```bash
curl --resolve godarai.ir:443:185.231.180.132 https://godarai.ir/
curl --resolve godarai.ir:443:37.152.187.55 https://godarai.ir/
```

## Rollback

The old Floating IPs `37.32.31.230` and `37.32.31.139` are detached but retained. Do not release them until the direct-public design has completed its stability window. A rollback requires restoring the backed-up Netplan/provider associations and then validating SSH through Bastion, HAProxy, the private VIP and DNS before removing the direct NIC path.

Remote pre-cutover backups are under `/root/lb-public-cutover-*` on each load balancer. The database migration Job was not executed as part of this ingress cutover.
