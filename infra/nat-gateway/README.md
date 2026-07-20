# Dedicated K3s NAT gateway

`nat-gw-1` provides outbound IPv4 access for the private-only K3s nodes. It is not part of the ingress or Kubernetes API HA path.

Production cutover completed on 2026-07-20. All three nodes use `10.10.0.190` as their default gateway and `1.1.1.1` plus `1.0.0.1` as their resolver pair. Host and ClusterFirst Pod tests passed DNS, HTTPS, Docker Registry and database TCP/5432; the observed public source was `188.121.120.198`. The existing `migrate-main-db` Job was deliberately not rerun.

## Addressing and roles

- private interface: `eth0`, `10.10.0.190/24`, MAC `fa:16:3e:5a:d5:79`
- public interface: `eth1`, DHCP address `188.121.120.198/22`, MAC `fa:16:3e:5f:7f:a3`
- the former unused `10.10.0.181`/`eth2` port was detached on 2026-07-20; only the fixed `.190` private port remains
- allowed NAT clients: `10.10.0.21`, `10.10.0.22`, and `10.10.0.23`
- Kubernetes API VIP: `10.10.0.6:6443`; it is not an egress gateway after cutover

The subnet gateway and DHCP service remain enabled in the cloud control plane. No subnet-wide static route is used. The gateway has only one guest default route, learned through public `eth1`.

## Installed files

| Repository file | Target path |
| --- | --- |
| `01-nat-gw.yaml` | `/etc/netplan/01-nat-gw.yaml` |
| `70-nat-gateway.conf` | `/etc/sysctl.d/70-nat-gateway.conf` |
| `99-disable-network-config.cfg` | `/etc/cloud/cloud.cfg.d/99-disable-network-config.cfg` |
| `99-preserve-hostname.cfg` | `/etc/cloud/cloud.cfg.d/99-preserve-hostname.cfg` |
| `nftables.conf` | `/etc/nftables.conf` |
| `upgraded-network` | `/var/lib/cloud/data/upgraded-network` on each K3s node |
| `node-1.yaml` | `node-1:/etc/netplan/50-cloud-init.yaml` |
| `node-2.yaml` | `node-2:/etc/netplan/50-cloud-init.yaml` |
| `node-3.yaml` | `node-3:/etc/netplan/50-cloud-init.yaml` |

Before loading the firewall, validate it with `sudo nft -c -f /etc/nftables.conf`. The forward policy is drop, input/output remain accept, and only the three K3s node addresses can be masqueraded through `eth1`.

The Cloud-Init override also disables network hotplug updates. Arvan/OpenStack hotplug events otherwise regenerate `/etc/netplan/50-cloud-init.yaml` and can reintroduce DHCP and a private default route on `eth0`.

The three node templates preserve their fixed private addresses, use `10.10.0.190` as the only default gateway, and use the Cloudflare pair `1.1.1.1` plus `1.0.0.1` as upstream DNS. They intentionally define only `eth1`: the nodes no longer have the old public `eth0` device, and keeping a stale MAC match makes `netplan apply` report that it cannot find a unique interface. Install `99-disable-network-config.cfg`, install the `upgraded-network` marker, and remove `/etc/udev/rules.d/90-cloud-init-hook-hotplug.rules` on every node so cloud boot/hotplug cannot overwrite the node route.

`8.8.8.8` was removed from the node pair after one of five direct queries timed out through the new gateway while `1.1.1.1` passed five of five and `1.0.0.1` passed ten of ten. The database TCP path itself passed twenty of twenty direct probes.

Arvan/OpenStack metadata still advertises the nodes' removed public NICs. Cloud-Init 25.1.4 eagerly converts datasource network metadata before it checks `network: {config: disabled}`; the stale MAC therefore raises `Unable to find a system nic`. The `/var/lib/cloud/data/upgraded-network` marker is the early-disable path used by this installed Cloud-Init version and must remain present while the metadata is stale.

Do not run live `netplan apply` or `netplan try` on the K3s nodes during this rollout. Testing on `node-2` showed that restarting `systemd-networkd` tears down the only K3s interface long enough to make the node NotReady and hang SSH. Apply route and DNS at runtime, validate them, persist the template, and defer reboot validation to a controlled maintenance window.

## Verification

```bash
ip -br address
ip route show default
sysctl net.ipv4.ip_forward
sudo nft list ruleset
systemctl is-enabled nftables
systemctl is-active nftables
sudo conntrack -C
cat /proc/sys/net/netfilter/nf_conntrack_max
```

Alert when the conntrack count exceeds 70 percent of `nf_conntrack_max`. Do not tune the maximum without measured pressure.

The legacy NAT services on `lb-1` and `lb-2` remain enabled during the rollback window. Remove them only after at least 24 hours of stable gateway operation.

The final gateway reboot preserved its hostname, single public default route, static private address, sysctls and nftables rules. A `node-2` reboot used for persistence testing exposed shutdown-time I/O errors on transient Longhorn block devices and required a provider-panel power cycle; its subsequent boot was clean. Avoid further node reboots until that event is investigated in a maintenance window.
