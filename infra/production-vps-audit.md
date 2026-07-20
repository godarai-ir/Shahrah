# گزارش بررسی استقرار VPSهای Production

تاریخ آخرین بررسی زنده: 2026-07-20

دامنه بررسی: اتصال SSH به هفت VPS، تطبیق interface/MAC/IP داخل guest، بررسی route و Netplan، وضعیت سرویس‌های LB، ruleهای NAT، تست تفکیک‌شده egress، وضعیت K3s و تست‌های DNS و TCP/HTTP/HTTPS در سطح host و Pod. در 2026-07-20 آماده‌سازی و Cutover کامل `nat-gw-1` روی هر سه node و سپس بازیابی رخداد DNS/etcd/Longhorn نیز انجام شد. وضعیت Fixed IP و فایروال دیتابیس از پنل ابرآروان مشاهده شد؛ دسترسی مستقیم به Neutron API یا control plane داخلی ارائه‌دهنده وجود ندارد.

در این گزارش:

- **مشاهده‌شده** یعنی در بررسی زنده 2026-07-19 مستقیماً اندازه‌گیری شده است.
- **اعلام‌شده از پنل** یعنی از اطلاعات یا تصویر پنل ابرآروان که مالک زیرساخت ارائه کرده استفاده شده است.
- **استنباط** یعنی نتیجه فنی حاصل از چند مشاهده است، نه مشاهده مستقیم control plane ارائه‌دهنده.
- **نامشخص** یعنی برای نتیجه قطعی به Neutron/OpenStack API یا تیم ارائه‌دهنده نیاز است.

## خلاصه اجرایی

استقرار فعلی شامل یک کلاستر سه‌نودی K3s، دو Load Balancer، یک Bastion و Gateway اختصاصی production با hostname برابر `nat-gw-1` است. مسیر داخلی HAProxy + Keepalived و VIP خصوصی `10.10.0.6:6443` برای Kubernetes API فعال مانده است، اما دیگر مسیر خروجی nodeها نیست. default route هر سه node اکنون `10.10.0.190` است و خروجی آن‌ها روی اینترنت با source IP برابر `188.121.120.198` دیده می‌شود.

در 2026-07-20 Gateway با یک default route عمومی، forwarding، nftables و MASQUERADE محدود به سه node آماده شد و persistence آن پس از reboot تأیید شد. Port Security دقیقاً روی پورت خصوصی `10.10.0.190` غیرفعال و پورت اضافی `.181` حذف شد. Rollout به ترتیب `node-2`، `node-3` و `node-1` انجام شد؛ هر سه node و Podهای pinned روی آن‌ها DNS، HTTPS، Docker Registry و TCP دیتابیس را با موفقیت گذراندند.

نتیجه نهایی egress:

- **مشاهده‌شده:** هر سه node فقط default route برابر `10.10.0.190` روی `eth1` دارند.
- **مشاهده‌شده:** resolverهای هر سه node برابر `1.1.1.1` و `1.0.0.1` هستند و hostname دیتابیس را به `185.231.181.201` resolve می‌کنند.
- **مشاهده‌شده:** HTTPS، Docker Registry با پاسخ مورد انتظار `401` و TCP دیتابیس روی `5432` از host و Podهای pinned موفق است.
- **مشاهده‌شده:** CoreDNS پس از rollout بدون upstream timeout جدید فعال است و `kubernetes.default` و hostname دیتابیس را resolve می‌کند.
- **مشاهده‌شده:** LBها دیگر ترافیک production egress nodeها را حمل نمی‌کنند؛ سرویس NAT قدیمی آن‌ها فقط برای پنجره rollback فعال مانده است.
- **استنباط:** ایجاد Gateway دارای NIC عمومی، محدودیت private egress/SNAT شبکه ارائه‌دهنده را بدون تغییر subnet-wide route دور زده است.

خرابی public edge مربوط به Floating IPها در 2026-07-20 با افزودن NIC و IPv4 عمومی مستقیم به هر دو LB دور زده شد. `godarai.ir` اکنون روی HTTPS پاسخ `200` می‌دهد و تست مستقیم با SNI روی هر دو IP جدید موفق است. Floating IPهای قبلی detach شده‌اند، ولی برای rollback حذف یا release نشده‌اند.

در رخداد بعدی همان روز، توقف heartbeat روی `node-2` باعث از دسترس خارج‌شدن تنها replicaی CoreDNS، شکست DNS داخل Pod و پاسخ `503` برای tenantها شد. هم‌زمان فشار دیسک و I/O مانع rebuild شدن replicaهای Longhorn و پایداری etcd بود. پس از snapshot تازه etcd، بازیابی مرحله‌ای `node-2` و `node-3`، پاک‌سازی imageهای بلااستفاده و journalها، همه nodeها، volumeها و workloadها سالم شدند. CoreDNS اکنون با حداقل دو replica و PDB تحت مدیریت GitOps اجرا می‌شود.

## تاریخچه تغییرات مهم

- **2026-07-16:** Port Security روی پورت‌های خصوصی `lb-1` و `lb-2` از پنل غیرفعال شد؛ HAProxy روی VIP `10.10.0.6:6443` قرار گرفت و failover کنترل‌شده Keepalived موفق بود.
- **2026-07-16:** Public IPهای nodeهای K3s حذف شدند و default route آن‌ها به VIP `10.10.0.6` منتقل شد. در همان روز egress از مسیر NAT روی LBها سالم تست شده بود.
- **2026-07-17:** Job مربوط به `migrate-main-db` با خطای DNS/عدم resolve دیتابیس شکست خورد.
- **2026-07-18:** DHCP شبکه خصوصی در پنل فعال شد. قفل Cloud-Init شبکه روی `lb-2` برداشته شد، ولی پورت قدیمی `lb-2` DHCP Offer نگرفت؛ static IP آن حفظ شد.
- **2026-07-19:** VPS جدید `test` به شبکه خصوصی افزوده شد و DHCP روی پورت جدید با موفقیت `10.10.0.190` را اختصاص داد. مقایسه forced-interface روی این VPS نشان داد private egress خراب و public egress سالم است.
- **2026-07-20:** نام guest به `nat-gw-1` تغییر کرد؛ `10.10.0.190/24` در Netplan استاتیک شد، تنها default route عمومی حفظ شد و Cloud-Init network/hostname قفل شد.
- **2026-07-20:** forwarding، nftables با forward policy برابر drop، MASQUERADE محدود به سه node و `conntrack` فعال شد. reboot و public source برابر `188.121.120.198` موفق بود.
- **2026-07-20:** آزمایش runtime روی `node-2` به‌دلیل نرسیدن packetهای routed به guest Gateway rollback شد؛ route پایدار nodeها تغییر نکرد.
- **2026-07-20:** پورت اضافی `.181` حذف و Port Security روی پورت‌های خصوصی `nat-gw-1`، `lb-1` و `lb-2` با confirmation پنل غیرفعال شد. listenerهای خصوصی LB/VIP دوباره در دسترس قرار گرفتند.
- **2026-07-20:** hotplug ابرآروان یک `50-cloud-init.yaml` جدید ساخت و DHCP خصوصی را برگرداند؛ hotplug network در override نهایی Cloud-Init غیرفعال، rule udev حذف و Netplan واحد دوباره اعمال شد.
- **2026-07-20:** آزمایش دوم `node-2` مسیر NAT را اثبات کرد، ولی DNS دیتابیس با resolverهای فعلی شکست خورد. query مستقیم به `1.1.1.1` و `8.8.8.8` موفق بود و node rollback شد.
- **2026-07-20:** `netplan try` روی `node-2` تنها interface کلاستر را موقتاً پایین آورد و node را NotReady کرد؛ از این نقطه route و DNS به‌صورت runtime اعمال و فایل‌ها بدون restart شبکه persist شدند.
- **2026-07-20:** metadata ابرآروان همچنان MAC مربوط به public NIC حذف‌شده nodeها را اعلام می‌کرد. override شبکه Cloud-Init، حذف hotplug rule و marker برابر `/var/lib/cloud/data/upgraded-network` روی هر سه node نصب شد.
- **2026-07-20:** در reboot اعتبارسنجی `node-2`، shutdown با خطاهای I/O روی block deviceهای transient Longhorn متوقف شد و یک Stop/Start از پنل لازم بود. boot بعدی بدون خطای I/O انجام شد و K3s/Longhorn بازیابی شدند.
- **2026-07-20:** `8.8.8.8` پس از یک timeout در پنج query مستقیم کنار گذاشته شد؛ زوج نهایی `1.1.1.1` و `1.0.0.1` در تست‌های تکراری پایدار بود.
- **2026-07-20:** Cutover به ترتیب `node-2`، `node-3` و `node-1` کامل شد. هر node حداقل ده دقیقه Ready و زیر بار مشاهده شد و rollback timerها لغو شدند.
- **2026-07-20:** CoreDNS rollout شد و تست‌های واقعی ClusterFirst روی هر سه node، DNS دیتابیس، TCP/5432 و public source برابر `188.121.120.198` را تأیید کردند.
- **2026-07-20:** reboot نهایی `nat-gw-1` hostname، Netplan واحد، default route عمومی، nftables و sysctlها را حفظ کرد؛ هیچ unit ناموفق مشاهده نشد.
- **2026-07-20:** برای رفع timeout دامنه، به `lb-1` و `lb-2` NIC عمومی مستقیم اضافه شد؛ Netplan هر دو ماشین یکپارچه، Cloud-Init network قفل و default route فقط روی NIC عمومی قرار گرفت. فایروال nftables ورودی عمومی را به HTTP/HTTPS محدود کرد.
- **2026-07-20:** Floating IPهای `37.32.31.230` و `37.32.31.139` detach ولی برای rollback حفظ شدند. رکوردهای A مربوط به `@` و `*` در پنل به IPهای مستقیم `185.231.180.132` و `37.152.187.55` تغییر کردند؛ سایت و هر دو origin پاسخ HTTPS `200` دادند.
- **2026-07-20:** از دست‌رفتن heartbeat `node-2` تنها replicaی CoreDNS را از دسترس خارج کرد؛ `tenant-admin` به‌دلیل `EAI_AGAIN` آماده نشد و ingress برای tenantها `503` داد.
- **2026-07-20:** snapshot پیش از بازیابی با نام `pre-node2-recovery-20260720T083529Z-node-1-1784536531` ذخیره شد. `node-2` و سپس `node-3` به‌صورت مرحله‌ای reboot و پس از تأیید quorum uncordon شدند.
- **2026-07-20:** imageهای بلااستفاده و journalهای حجیم روی هر سه node پاک‌سازی شد. Longhorn همه پنج volume را rebuild کرد و وضعیت نهایی همه آن‌ها `attached/healthy` است.
- **2026-07-20:** CoreDNS با HPA دارای `minReplicas: 2` و PDB دارای `minAvailable: 1` پایدار شد. مقدار cache حافظه Loki از پیش‌فرض غیرقابل‌زمان‌بندی ۸ GiB به ۱ GiB کاهش یافت و تمام Podها Running شدند.
- **2026-07-20:** drift پیکربندی Prometheus و Grafana در GitOps اصلاح و تمام Applicationهای Argo CD به `Synced/Healthy` رسیدند.

## فهرست ماشین‌ها و آدرس‌ها

| نقش | Hostname | interface / MAC خصوصی | IP خصوصی فعال | public address | وضعیت |
| --- | --- | --- | --- | --- | --- |
| Bastion | `bastion` | `private0` / `fa:16:3e:e4:0c:1b` | `10.10.0.5/24` | `public0`: `37.152.191.92/22` | فعال |
| Load balancer | `lb-1` | `eth0` / `fa:16:3e:d8:b2:72` | `10.10.0.7/24` + VIP `10.10.0.6/24` | `eth1`: `185.231.180.132/22` | MASTER فعلی؛ public ingress فعال |
| Load balancer | `lb-2` | `eth0` / `fa:16:3e:cc:41:ab` | `10.10.0.8/24` | `eth1`: `37.152.187.55/23` | BACKUP فعلی؛ public ingress فعال |
| K3s node | `node-1` | `eth1` / `fa:16:3e:3a:7d:ce` | `10.10.0.21/24` | ندارد | Ready |
| K3s node | `node-2` | `eth1` / `fa:16:3e:a7:91:15` | `10.10.0.22/24` | ندارد | Ready |
| K3s node | `node-3` | `eth1` / `fa:16:3e:1a:14:bd` | `10.10.0.23/24` | ندارد | Ready |
| NAT gateway | `nat-gw-1` | `eth0` / `fa:16:3e:5a:d5:79` | static: `10.10.0.190/24` | `eth1`: `188.121.120.198/22` | production egress فعال |

آدرس‌های اولیه پنل برای VPS جدید `10.10.0.181`، `10.10.0.190` و `188.121.120.198` بودند. وضعیت نهایی:

- `10.10.0.190` روی `eth0` با MAC برابر `fa:16:3e:5a:d5:79` فعال است. Fixed IP بودن این پورت در پنل تأیید و آدرس داخل guest استاتیک شده است.
- `188.121.120.198` روی `eth1` با MAC برابر `fa:16:3e:5f:7f:a3` فعال و از DHCP شبکه عمومی دریافت شده است.
- پورت اضافی `10.10.0.181` با MAC برابر `fa:16:3e:fd:65:58` در 2026-07-20 detach شد.

Floating IPهای قبلی LBها داخل Netplan یا interfaceهای guest تنظیم نشده‌اند و در پنل detached مانده‌اند. مسیر فعال ingress از NIC عمومی مستقیم هر LB است؛ این تصمیم dependency خراب Floating-IP mapping را از مسیر production حذف کرده و امکان rollback را حفظ کرده است.

## نسخه سیستم‌عامل و منابع

| ماشین | Ubuntu | Kernel | vCPU | RAM | root disk | Swap |
| --- | --- | --- | ---: | ---: | ---: | --- |
| `bastion` | 24.04.1 LTS | `6.8.0-136-generic` | 1 | 961 MiB | 23 GiB، مصرف 23% | ندارد |
| `lb-1` | 24.04.1 LTS | `6.8.0-134-generic` | 1 | 961 MiB | 23 GiB، مصرف 8% | ندارد |
| `lb-2` | 24.04.4 LTS | `6.8.0-134-generic` | 1 | 961 MiB | 23 GiB، مصرف 9% | ندارد |
| `node-1` | 24.04.1 LTS | `6.8.0-134-generic` | 4 | 11 GiB | 70 GiB، مصرف 53% | ندارد |
| `node-2` | 24.04.1 LTS | `6.8.0-134-generic` | 2 | 3.8 GiB | 70 GiB، مصرف 52% | ندارد |
| `node-3` | 24.04.4 LTS | `6.8.0-134-generic` | 2 | 3.8 GiB | 70 GiB، مصرف 45% | ندارد |
| `nat-gw-1` | 24.04.1 LTS | `6.8.0-45-generic` | 1 | 961 MiB | 23 GiB، مصرف 6% | ندارد |

## شبکه خصوصی و وضعیت OpenStack قابل مشاهده

تنظیمات فعلی شبکه خصوصی که از پنل اعلام شده است:

- CIDR: `10.10.0.0/24`
- Gateway: فعال، `10.10.0.1`
- DHCP: فعال
- DHCP allocation pool: `10.10.0.100` تا `10.10.0.199`
- Static Route پنل: خالی
- Port Security روی پورت‌های `lb-1` و `lb-2`: غیرفعال، طبق اعلام مالک پنل

در baseline، DHCP جدید روی ماشینی که آن زمان `test` نام داشت به‌صورت عملی کار کرد: `systemd-networkd` دریافت `10.10.0.190/24`، gateway برابر `10.10.0.1` و DNSهای `1.1.1.1` و `8.8.8.8` را از `10.10.0.1` ثبت کرد. بنابراین DHCP subnet برای پورت‌های جدید فعال است.

کارکرد DHCP به‌تنهایی به معنی وجود Internet Gateway یا SNAT سالم نیست. تست interface-specific پیش از آماده‌سازی `nat-gw-1` نتیجه زیر را داد:

| مسیر تست | نتیجه HTTPS به `1.1.1.1` |
| --- | --- |
| `nat-gw-1` از `eth0` خصوصی / source `10.10.0.190` | timeout، HTTP code `000` |
| `nat-gw-1` از `eth1` عمومی / source `188.121.120.198` | موفق، HTTP code `200` |

این paired test ثابت می‌کند که DHCP و اتصال لایه 2/3 شبکه خصوصی تا gateway برقرار است، اما خروجی اینترنت private path کار نمی‌کند.

## Netplan و routeها

### Bastion

- `public0`: `37.152.191.92/22`
- `private0`: `10.10.0.5/24`
- default route: `37.152.188.1` روی `public0`
- egress، DNS دیتابیس و HTTPS در 2026-07-19 سالم بود.

listenerها و سرویس‌های مشاهده‌شده روی Bastion:

- SSH روی `0.0.0.0:22` و `[::]:22`
- Squid روی `10.10.0.5:3128`، active و enabled
- Xray روی `127.0.0.1:10808` و `127.0.0.1:10809`، active و enabled
- UFW: `inactive`؛ کنترل دسترسی اصلی به policyهای پنل وابسته است.

### Load Balancerها

هر LB یک NIC خصوصی استاتیک و یک NIC عمومی DHCP دارد:

- `lb-1`: `eth0=10.10.0.7/24` به‌همراه VIP `10.10.0.6/24`؛ `eth1=185.231.180.132/22` و تنها default route از `185.231.180.1`
- `lb-2`: `eth0=10.10.0.8/24`؛ `eth1=37.152.187.55/23` و تنها default route از `37.152.186.1`
- Cloud-Init network روی هر دو LB غیرفعال است و Netplan فعال هر ماشین در `/etc/netplan/01-lb-public.yaml` یکپارچه شده است. DNSهای استاتیک `1.1.1.1` و `8.8.8.8` روی تعریف `eth0` حفظ شده‌اند و DHCP عمومی نیز مجاز به ارائه resolver است.
- `haproxy`، `keepalived` و `lb-public-filter.service` روی هر دو ماشین active هستند. فایروال میزبان روی `eth1` فقط DHCP، ICMP ضروری، connectionهای established و TCP/80/443 را می‌پذیرد؛ SSH عمومی در تست خارجی timeout شد.

### Nodeهای K3s

- `node-1`: default via `10.10.0.190` روی `eth1`
- `node-2`: default via `10.10.0.190` روی `eth1`
- `node-3`: default via `10.10.0.190` روی `eth1`
- DNS هر سه node: `1.1.1.1` و `1.0.0.1`
- Netplan هر node فقط `eth1` را تعریف می‌کند؛ public NIC قدیمی حذف شده است.
- Cloud-Init network روی هر سه node غیرفعال، hotplug rule حذف و marker `upgraded-network` نصب شده است.
- Public IPv4 ندارند.

### NAT Gateway `nat-gw-1`

Cloud-Init با `DataSourceOpenStackLocal` کامل شده و اکنون network config و hostname آن با فایل‌های `99-disable-network-config.cfg` و `99-preserve-hostname.cfg` قفل شده است.

- `eth0`: استاتیک `10.10.0.190/24`، بدون gateway و DNS
- `eth1`: DHCP عمومی `188.121.120.198/22` و تنها default route از `188.121.120.1`
- پورت/اینترفیس سابق `eth2`: از control plane detach شده است
- `net.ipv4.ip_forward=1`، `rp_filter=2` و redirectهای IPv4 خاموش
- `nftables.service`: active و enabled؛ input/output برابر accept و forward policy برابر drop
- MASQUERADE فقط برای `10.10.0.21/32`، `.22/32` و `.23/32` هنگام خروج از `eth1`
- `conntrack`: نصب‌شده؛ شمارش نهایی `71/8192` بود

پس از reboot، hostname، تنها route عمومی، nftables و source IP عمومی `188.121.120.198` حفظ شدند. خود Gateway به GitHub پاسخ `200`، از Docker Registry پاسخ `401` و به IP دیتابیس روی TCP/5432 اتصال موفق داشت.

Gateway مسیر فعال production هر سه node است. پس از Cutover، public source برابر `.198`، GitHub=`200`، Registry=`401`، DNS hostname دیتابیس و TCP/5432 هم از host و هم از Podهای pinned تأیید شدند. counterهای نهایی Gateway پس از تست‌ها شامل `allow new=75`، `MASQUERADE=75` و `established=2182` بود؛ counter ترافیک جدید public-to-private و catch-drop صفر باقی ماند.

## Bastion و دسترسی PuTTY

پروفایل PuTTY با نام `nat-gw-1` وجود دارد و اتصال آن در 2026-07-20 موفق بود:

- Host: `127.0.0.1`
- Port: `22190`
- User: `ubuntu`
- مقصد واقعی: `10.10.0.190:22`

forwarding پورت `22190` اکنون مستقیماً از saved profile مربوط به Bastion به Gateway می‌رود:

```text
Bastion -> 10.10.0.190:22
```

ورودی `L22190=10.10.0.190:22` به saved profile `Bastion` افزوده شد و process وابسته به `node-1` حذف شد. بنابراین تغییر default route روی `node-1` دیگر مسیر مدیریتی Gateway را قطع نمی‌کند.

یک listener قدیمی روی local port `22181` هنوز متعلق به session جاری PuTTY/Bastion بود و با پورت غیرفعال `10.10.0.181` مرتبط است؛ پروفایل `nat-gw-1` از آن استفاده نمی‌کند.

## لایه Load Balancer و VIP

روی هر دو LB در 2026-07-19:

- `haproxy.service`: active و enabled
- `keepalived.service`: active و enabled
- `k3s-egress-nat.service`: active و enabled
- `systemd-networkd`: active و enabled
- `net.ipv4.ip_forward=1`
- `net.ipv4.ip_nonlocal_bind=1`
- `net.ipv4.conf.eth0.send_redirects=0`

Listenerهای HAProxy روی هر دو LB:

- `0.0.0.0:80`
- `0.0.0.0:443`
- `10.10.0.6:6443`

HAProxy در `mode tcp` و با `roundrobin` کار می‌کند. TLS روی LB terminate نمی‌شود. backendهای زنده و یکسان روی هر دو LB:

| Frontend | Backendها |
| --- | --- |
| HTTP `:80` | `10.10.0.21:30081`، `10.10.0.22:30081`، `10.10.0.23:30081` |
| HTTPS `:443` | `10.10.0.21:30444`، `10.10.0.22:30444`، `10.10.0.23:30444` |
| Kubernetes API `10.10.0.6:6443` | `10.10.0.21:6443`، `10.10.0.22:6443`، `10.10.0.23:6443` |

`nginx.service` روی هیچ‌کدام از LBها وجود ندارد؛ edge واقعی HAProxy + Keepalived است.

VIP `10.10.0.6/24` هنگام بررسی روی `lb-1` قرار داشت. از Bastion اتصال TCP به این مسیرها موفق بود:

- `10.10.0.7:80` و `10.10.0.7:443`
- `10.10.0.8:80` و `10.10.0.8:443`
- `10.10.0.6:6443`

Keepalived همچنان `lb-1` را با priority `200` و `lb-2` را با priority `150` مدیریت می‌کند. failover کنترل‌شده در 2026-07-16 موفق بود، اما در audit جدید عمداً تکرار نشد.

health script فعلی فقط `pidof haproxy` را کنترل می‌کند؛ سلامت bind روی VIP، API و backendها را به‌صورت end-to-end بررسی نمی‌کند. health checkهای HAProxy نیز TCP-only هستند.

## NAT قدیمی روی LBها و مسیر فعال egress

روی هر دو LB ruleهای زیر همچنان وجود دارند:

- `MASQUERADE` برای `10.10.0.21/32`
- `MASQUERADE` برای `10.10.0.22/32`
- `MASQUERADE` برای `10.10.0.23/32`
- مستثنی‌کردن مقصدهای داخل `10.10.0.0/24`

این ruleها و سرویس `k3s-egress-nat` موقتاً برای rollback فعال و enabled مانده‌اند، اما هر سه node اکنون default route خود را به `nat-gw-1` می‌فرستند و این LBها ترافیک production egress را حمل نمی‌کنند. آن‌ها فقط پس از حداقل ۲۴ ساعت پایداری Gateway پاک‌سازی می‌شوند.

| ماشین | route انتخاب‌شده | نتیجه خروجی نهایی |
| --- | --- | --- |
| `lb-1` | `1.1.1.1 via 10.10.0.1` | timeout |
| `lb-2` | `1.1.1.1 via 10.10.0.1` | timeout |
| `node-1` | `1.1.1.1 via 10.10.0.190` | موفق؛ source `.198` |
| `node-2` | `1.1.1.1 via 10.10.0.190` | موفق؛ source `.198` |
| `node-3` | `1.1.1.1 via 10.10.0.190` | موفق؛ source `.198` |
| `bastion` | public gateway روی `public0` | `200` |

timeout خروجی خود LBها از private gateway همچنان توضیح می‌دهد چرا طراحی NAT قدیمی شکست خورد؛ مسیر جدید این وابستگی را با public NIC اختصاصی Gateway حذف کرده است.

## Kubernetes

هر سه node در بررسی نهایی 2026-07-20 `Ready` هستند:

| Node | Role | Internal IP | Version |
| --- | --- | --- | --- |
| `node-1` | control-plane, etcd | `10.10.0.21` | `v1.35.5+k3s1` |
| `node-2` | control-plane, etcd | `10.10.0.22` | `v1.35.5+k3s1` |
| `node-3` | control-plane, etcd | `10.10.0.23` | `v1.35.5+k3s1` |

Service ورودی همچنان `ingress-nginx/ingress-nginx-controller` از نوع NodePort است:

- HTTP: `30081/TCP`
- HTTPS: `30444/TCP`

Ingress hostهای مشاهده‌شده:

- `godarai.ir`
- `*.godarai.ir`
- `admin.godarai.ir`

در بررسی نهایی هیچ Pod با وضعیت غیر از `Running` یا `Succeeded` و هیچ container ناآماده‌ای وجود نداشت. همه Deploymentها، StatefulSetها و DaemonSetها به تعداد replica مورد انتظار رسیده‌اند. CoreDNS دارای دو Pod آماده روی دو node متفاوت است و HPA آن تعداد replica را بین ۲ و ۳ نگه می‌دارد.

### Job مهاجرت دیتابیس

Job شکست‌خورده 2026-07-17 به علت نبود DNS داخلی نتوانسته بود به دیتابیس متصل شود. خطای ثبت‌شده هنگام وقوع:

خطای ثبت‌شده هنگام وقوع:

```text
goose run: failed to connect to user=base-user database=default:
hostname resolving error: lookup c781656f0ab24ceb9040b7f8bdbe5f6c.db.arvandbaas.ir
on 10.43.0.10:53: server misbehaving
```

اکنون nodeها و Podهای ClusterFirst همین hostname را به `185.231.181.201` resolve می‌کنند و TCP دیتابیس روی `5432` نیز موفق است. هنگام reconcile کامل Argo CD در بازیابی 2026-07-20، hook مربوط به `migrate-main-db` به‌صورت خودکار اجرا و در ۲۵ ثانیه با `1/1 Complete` تمام شد؛ Argo CD سپس Job را طبق `HookSucceeded` حذف کرد. hookهای `migrate-clickhouse`، `support-admin-migrate` و `model-discovery-migrate` نیز موفق بودند؛ مورد آخر import تعداد ۲۲۹۸ مدل را کامل کرد.

## DNS و دسترسی عمومی

NSهای authoritative:

- `c.ns.arvancdn.ir`
- `p.ns.arvancdn.ir`

رکوردهای A در پنل با TTL دو دقیقه تنظیم شده‌اند:

- `@`: `37.152.187.55` و `185.231.180.132`
- `*`: `37.152.187.55` و `185.231.180.132`
- `www`: CNAME به `godarai.ir.`

در پذیرش نهایی، `https://godarai.ir/` پاسخ `200` داد. HTTPS مستقیم با SNI برابر `godarai.ir` روی هر دو IP نیز پاسخ `200` داد و HTTP به HTTPS با `308` redirect شد. authoritative DNS و resolverهای `1.1.1.1` و `8.8.8.8` از مبدأ بررسی فعلاً فقط `37.152.187.55` را برگرداندند، با اینکه هر دو مقدار در پنل ثبت شده‌اند؛ بنابراین توزیع یا failover پاسخ DNS در کنترل سرویس DNS آروان است و از این مشاهده نمی‌توان round-robin قطعی بین دو origin را نتیجه گرفت.

## ریسک‌ها و مسائل باز

1. **رفع‌شده با workaround — public edge روی Floating IPها timeout می‌شد.** مسیر production اکنون از public NICهای مستقیم استفاده می‌کند؛ Floating IPهای خراب detached و برای rollback حفظ شده‌اند.
2. **بالا — metadata شبکه OpenStack stale است.** MACهای public NIC حذف‌شده هنوز به nodeها اعلام می‌شوند و بدون marker `upgraded-network` می‌توانند Cloud-Init را در boot دچار خطا کنند.
3. **بالا — reboot اعتبارسنجی `node-2` با I/O error هنگام shutdown متوقف شد.** boot پس از power-cycle پاک بود و Longhorn اکنون سالم است، اما سلامت disk و block deviceهای transient باید پیش از reboot بعدی بررسی شود.
4. **متوسط — HA عمومی DNS-level است، نه یک public VIP مشترک.** هر دو origin سالم‌اند، اما authoritative DNS از مبدأ تست فقط IP `lb-2` را انتخاب کرد؛ failover خودکار باید جداگانه با قابلیت health-check/load-balancing ارائه‌دهنده اثبات شود.
5. **متوسط — health check Keepalived سطحی است.** فقط وجود process HAProxy را می‌سنجد، نه سلامت VIP/API/backend.
6. **متوسط — OS firewall روی بیشتر ماشین‌ها غیرفعال گزارش شده است.** روی هر دو LB اکنون nftables عمومی فعال است و فقط 80/443 را باز می‌گذارد؛ سایر ماشین‌ها همچنان به policy پنل وابسته‌اند.
7. **متوسط — ظرفیت CPU درخواستی روی `node-1` و `node-2` محدود است.** مصرف واقعی پایدار است، اما requests زمان‌بندی‌شده به‌ترتیب حدود ۹۰٪ و ۸۷٪ CPU بود؛ workloadهای جدید باید با ظرفیت‌سنجی اضافه شوند.
8. **متوسط — swap روی هیچ ماشین فعال نیست.** LBها و `nat-gw-1` فقط 1 GiB RAM دارند.

## گام‌های پیشنهادی بعدی

1. Gateway و سه node حداقل ۲۴ ساعت monitor شوند؛ conntrack بالاتر از ۷۰٪، packet drop، CoreDNS timeout و NodeNotReady هشدار ایجاد کند.
2. پس از این پنجره و در صورت پایداری، ruleهای NAT قدیمی، `k3s-egress-nat.service` و `ip_forward` روی هر دو LB حذف/غیرفعال شوند.
3. از ارائه‌دهنده خواسته شود metadata/Neutron portهای حذف‌شده nodeها را پاک کند؛ تا آن زمان marker `upgraded-network` نباید حذف شود.
4. پیش از reboot بعدی nodeها، خطاهای shutdown `node-2` و سلامت volumeهای Longhorn بررسی و snapshot/backup مناسب گرفته شود.
5. health-check و failover واقعی رکوردهای چندمقداری DNS آروان جداگانه بررسی شود؛ صرف ثبت دو A record تضمین public VIP یا failover سریع نیست.
6. Floating IPهای detached فقط پس از یک پنجره پایداری و تصمیم صریح درباره rollback release شوند.
7. alert برای egress Gateway، DNS خارجی از کلاستر، NAT/conntrack counters، سلامت هر دو IP مستقیم، CoreDNS replica count و disk usage nodeها اضافه شود.

## شواهد خوانده‌شده

شواهد 2026-07-19 شامل این موارد بود:

- `hostname`، زمان UTC، نسخه Ubuntu/kernel، CPU/RAM/disk/swap روی هر هفت VPS
- `ip -br link`، `ip -br address`، `ip route` و `ip route get`
- Netplan، `networkctl`، Cloud-Init datasource و DHCP leaseهای baseline مربوط به `nat-gw-1`
- وضعیت `haproxy`، `keepalived`، `k3s-egress-nat` و `systemd-networkd`
- listenerها با `ss -lnt`
- `sysctl`های forwarding/nonlocal-bind/send-redirects
- خروجی read-only از `iptables-save` و `nft list ruleset`
- forced-interface HTTPS از `eth0` و `eth1` روی `nat-gw-1`
- تست DNS دیتابیس و HTTPS از Bastion، LBها و nodeها
- وضعیت nodeها، serviceها، ingressها، podهای non-running و Job مایگریشن در K3s
- تنظیمات saved PuTTY profile `nat-gw-1` و process/listener تونل فعال
- authoritative DNS و HTTP/HTTPS به دامنه و هر دو Floating IP

در 2026-07-20 پیکربندی `nat-gw-1` و Netplan/Cloud-Init هر سه node تغییر کرد و Cutover production کامل شد. سپس public ingress هر دو LB به NIC عمومی مستقیم منتقل، DNS به‌روزرسانی و فایروال محدودکننده میزبان فعال شد. در بازیابی رخداد بعدی، snapshot تازه etcd گرفته شد، nodeها مرحله‌ای بازیابی و دیسک‌ها پاک‌سازی شدند؛ Podهای تشخیصی موقت پس از تست حذف شدند. reconcile خودکار Argo CD hookهای migration را با موفقیت اجرا کرد. NAT روی LBها برای rollback دست‌نخورده باقی ماند و هیچ Static Route سراسری در پنل ساخته نشد.
