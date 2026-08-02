# پروفایل‌های پیشنهادی Firewall پنل برای VPSهای Production

تاریخ آخرین بازبینی: 2026-07-20

این سند پروفایل‌های ریزدانه Firewall پنل را برای استقرار فعلی گدارAI تعریف می‌کند. فرض اصلی این است که ترافیک ورودی به‌صورت پیش‌فرض بسته است و فقط ruleهای لازم باز می‌شوند.

وضعیت فعلی K3s nodeها private-only است. `node-1`، `node-2` و `node-3` Public IP ندارند و خروجی اینترنت آن‌ها از Gateway اختصاصی `10.10.0.190` با public source برابر `188.121.120.198` عبور می‌کند. Port Security پورت خصوصی همین Gateway غیرفعال است. VIP خصوصی `10.10.0.6` فقط endpoint رسمی Kubernetes API است؛ NAT روی LBها صرفاً در پنجره rollback بیست‌وچهارساعته فعال مانده است.

## متغیرها

قبل از ساخت پروفایل‌ها، این مقدارها را با وضعیت واقعی پنل جایگزین کنید:

```text
ADMIN_IPS = IPهای ثابت ادمین‌ها یا VPN مدیریت
PRIVATE_CIDR = 10.10.0.0/24
BASTION = 10.10.0.5
LB_1 = 10.10.0.7
LB_2 = 10.10.0.8
LB_1_PUBLIC = 185.231.180.132
LB_2_PUBLIC = 37.152.187.55
LB_1_DETACHED_FLOATING_IP = 37.32.31.230
LB_2_DETACHED_FLOATING_IP = 37.32.31.139
NODE_1 = 10.10.0.21
NODE_2 = 10.10.0.22
NODE_3 = 10.10.0.23
NODES = 10.10.0.21, 10.10.0.22, 10.10.0.23
NAT_GW_PRIVATE = 10.10.0.190
NAT_GW_PUBLIC = 188.121.120.198
```

## سیاست پایه

- ورودی: پیش‌فرض بسته باشد.
- خروجی: در این مرحله باز بماند، مگر اینکه برای outbound نیز سیاست دقیق جداگانه تعریف شود.
- روی هر دو LB فایروال nftables میزبان فعال است و ورودی NIC عمومی را محدود می‌کند؛ policy پنل provider همچنان لایه مکمل است.
- پورت `6443` نباید از اینترنت عمومی باز باشد.
- برای nodeهای K3s، هیچ rule عمومی لازم نیست؛ دسترسی مدیریتی فقط از Bastion و شبکه خصوصی انجام شود.

## پروفایل‌های پایه

### `admin-ssh-public`

برای دسترسی SSH عمومی فقط به Bastion یا دسترسی اضطراری محدود.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `ADMIN_IPS` |

### `admin-ssh-private`

برای SSH از Bastion به ماشین‌های خصوصی.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `10.10.0.5` |

### `private-node-trust`

این پروفایل ساده‌ترین حالت برای اعتماد داخلی است. اگر پنل اجازه ساخت ruleهای زیاد را سخت می‌کند، می‌توان موقتاً از آن استفاده کرد؛ اما برای Production، پروفایل‌های ریزدانه پایین بهترند.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | All | All | `10.10.0.0/24` |

## پروفایل‌های Bastion

روی `bastion` اعمال شود.

### `bastion-public-admin`

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `ADMIN_IPS` |

### `bastion-private-services`

اگر Squid روی Bastion برای کلاستر لازم است، این پروفایل را اعمال کنید.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `3128` | `10.10.0.21` |
| Inbound | TCP | `3128` | `10.10.0.22` |
| Inbound | TCP | `3128` | `10.10.0.23` |

پورت‌های `10808` و `10809` روی Bastion local هستند و نباید از شبکه عمومی یا خصوصی باز شوند.

## پروفایل‌های Load Balancer

روی `lb-1` و `lb-2` اعمال شود.

پس از cutover تاریخ 2026-07-20، ترافیک عمومی از NIC و IPv4 مستقیم هر Load Balancer تحویل می‌شود:

- `185.231.180.132` روی `eth1` برای `lb-1`
- `37.152.187.55` روی `eth1` برای `lb-2`

Floating IPهای قبلی detach و صرفاً برای rollback حفظ شده‌اند. در پنل provider، ruleهای public `80/443` باید برای NICهای عمومی مستقیم یا VPS مقصد آن‌ها اعمال شوند. روی Ubuntu نیز `lb-public-filter.service` فقط 80/443 و ترافیک ضروری شبکه را روی `eth1` می‌پذیرد و SSH عمومی را می‌بندد.

### `lb-public-web`

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `80` | `0.0.0.0/0` |
| Inbound | TCP | `443` | `0.0.0.0/0` |

اگر IPv6 فعال نیست، rule برای `::/0` نسازید. اگر IPv6 فعال است:

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `80` | `::/0` |
| Inbound | TCP | `443` | `::/0` |

### `lb-private-admin`

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `10.10.0.5` |

برای دسترسی اضطراری، فقط در صورت نیاز و با IPهای محدود:

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `ADMIN_IPS` |

### `lb-private-k8s-api`

در وضعیت فعلی، VIP خصوصی `10.10.0.6:6443` مسیر رسمی Kubernetes API است. برای کارکرد VRRP/VIP، Port Security شبکه خصوصی باید روی `lb-1` و `lb-2` غیرفعال بماند.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `6443` | `10.10.0.5` |
| Inbound | TCP | `6443` | `10.10.0.21` |
| Inbound | TCP | `6443` | `10.10.0.22` |
| Inbound | TCP | `6443` | `10.10.0.23` |

### `lb-private-node-egress-nat`

این پروفایل legacy فقط برای rollback موقت خروجی nodeها از مسیر Load Balancerهاست. پس از حداقل ۲۴ ساعت پایداری `nat-gw-1` باید همراه سرویس NAT قدیمی حذف شود.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | All | All | `10.10.0.21` |
| Inbound | All | All | `10.10.0.22` |
| Inbound | All | All | `10.10.0.23` |

## پروفایل NAT Gateway اختصاصی

روی `nat-gw-1` اعمال شود. Port Security فقط روی fixed port خصوصی `10.10.0.190` با MAC برابر `fa:16:3e:5a:d5:79` غیرفعال است؛ پورت عمومی دست‌نخورده مانده و پورت اضافی `.181` حذف شده است.

### `nat-gw-private-admin-and-forwarding`

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `10.10.0.5` |
| Inbound | All | All | `10.10.0.21` |
| Inbound | All | All | `10.10.0.22` |
| Inbound | All | All | `10.10.0.23` |

داخل Ubuntu، forward policy برابر drop است و فقط سه node بالا به سمت interface عمومی MASQUERADE می‌شوند. از 2026-08-02، input chain پورت TCP/1443 را فقط برای `10.10.0.21`، `10.10.0.22` و `10.10.0.23` می‌پذیرد و همان پورت را برای سایر sourceها drop می‌کند. این listener خصوصی ورودی Silk Road به پراکسی خروجی اروپا است؛ سایر ورودی‌ها همچنان policy برابر accept دارند.

## پروفایل‌های K3s Node

روی `node-1`، `node-2` و `node-3` اعمال شود.

### `k3s-private-admin`

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `22` | `10.10.0.5` |

### `k3s-api-from-lb-and-nodes`

برای دسترسی HAProxyها و control-planeها به Kubernetes API روی هر node.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `6443` | `10.10.0.7` |
| Inbound | TCP | `6443` | `10.10.0.8` |
| Inbound | TCP | `6443` | `10.10.0.21` |
| Inbound | TCP | `6443` | `10.10.0.22` |
| Inbound | TCP | `6443` | `10.10.0.23` |

### `k3s-etcd-internal`

برای embedded etcd بین control-plane nodeها.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `2379` | `10.10.0.21` |
| Inbound | TCP | `2379` | `10.10.0.22` |
| Inbound | TCP | `2379` | `10.10.0.23` |
| Inbound | TCP | `2380` | `10.10.0.21` |
| Inbound | TCP | `2380` | `10.10.0.22` |
| Inbound | TCP | `2380` | `10.10.0.23` |

### `k3s-flannel-vxlan`

برای شبکه overlay فعلی K3s با Flannel VXLAN.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | UDP | `8472` | `10.10.0.21` |
| Inbound | UDP | `8472` | `10.10.0.22` |
| Inbound | UDP | `8472` | `10.10.0.23` |

### `k3s-kubelet-internal`

برای kubelet، metrics-server و دسترسی‌های داخلی Kubernetes.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `10250` | `10.10.0.21` |
| Inbound | TCP | `10250` | `10.10.0.22` |
| Inbound | TCP | `10250` | `10.10.0.23` |

### `k3s-ingress-nodeports-from-lb`

برای رسیدن HAProxy به `ingress-nginx` روی NodePortها.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `30081` | `10.10.0.7` |
| Inbound | TCP | `30081` | `10.10.0.8` |
| Inbound | TCP | `30444` | `10.10.0.7` |
| Inbound | TCP | `30444` | `10.10.0.8` |

### `node-exporter-private`

اگر Prometheus داخل کلاستر node exporter را از طریق node IPها scrape می‌کند، این پروفایل را محدود به شبکه خصوصی اعمال کنید.

| Direction | Protocol | Port | Source |
| --- | --- | --- | --- |
| Inbound | TCP | `9100` | `10.10.0.0/24` |

## ماتریس اعمال روی VPSها

### `bastion`

```text
bastion-public-admin
bastion-private-services
```

### `lb-1` و `lb-2`

```text
lb-public-web
lb-private-admin
lb-private-k8s-api
lb-private-node-egress-nat
```

### `node-1`، `node-2` و `node-3`

```text
k3s-private-admin
k3s-api-from-lb-and-nodes
k3s-etcd-internal
k3s-flannel-vxlan
k3s-kubelet-internal
k3s-ingress-nodeports-from-lb
node-exporter-private
```

## پورت‌هایی که نباید عمومی باز شوند

اگر پنل default deny دارد، ساخت ruleهای deny لازم نیست. در غیر این صورت، این پورت‌ها را از اینترنت عمومی ببندید:

| Protocol | Port | Public source |
| --- | --- | --- |
| TCP | `6443` | `0.0.0.0/0` |
| TCP | `2379-2380` | `0.0.0.0/0` |
| UDP | `8472` | `0.0.0.0/0` |
| TCP | `10250` | `0.0.0.0/0` |
| TCP | `9100` | `0.0.0.0/0` |
| TCP | `30081` | `0.0.0.0/0` |
| TCP | `30444` | `0.0.0.0/0` |

## ترتیب اجرای امن

1. ابتدا پروفایل‌ها را روی `lb-2` یا یک node کم‌ریسک‌تر اعمال کنید.
2. از Bastion اتصال SSH را تست کنید.
3. از LBها به nodeها پورت‌های `30081` و `30444` را تست کنید.
4. از Bastion، پورت `6443` روی VIP `10.10.0.6` را تست کنید.
5. از هر node، دسترسی خروجی به Docker Registry را از مسیر `10.10.0.190` تست کنید و public source برابر `188.121.120.198` را تأیید کنید.
6. `kubectl get nodes` را از مسیر tunnel فعلی اجرا کنید.
7. پس از موفقیت تست‌ها، پروفایل‌ها را مرحله‌ای روی بقیه ماشین‌ها اعمال کنید.

فرمان‌های نمونه برای بررسی:

```bash
nc -vz 10.10.0.6 6443
nc -vz 10.10.0.21 30081
nc -vz 10.10.0.21 30444
curl -4 https://registry-1.docker.io/v2/
kubectl get nodes -o wide
```

## بدهی باقی‌مانده

- اگر مالکیت VIP، اعضای LB یا سیاست Port Security تغییر کرد، پروفایل‌های مربوط به `10.10.0.6` را دوباره بازبینی کنید.
- اگر outbound policy نیز محدود شود، باید برای DNS، registry، GitHub/Argo CD، PostgreSQL خارجی، object storage و providerهای مدل rule جداگانه تعریف شود.
- اگر CNI از Flannel VXLAN به WireGuard یا CNI دیگری تغییر کند، پورت‌های شبکه nodeها باید به‌روزرسانی شوند.
