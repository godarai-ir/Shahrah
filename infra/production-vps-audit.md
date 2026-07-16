# گزارش بررسی استقرار VPSهای Production

تاریخ بررسی: 2026-07-16

دامنه بررسی اولیه: اتصال SSH فقط خواندنی به VPSها، بررسی DNS، و چند تست HTTP/HTTPS از ماشین محلی.

به‌روزرسانی عملیاتی 2026-07-16: پس از غیرفعال شدن Port Security شبکه خصوصی روی `lb-1` و `lb-2`، HAProxy روی هر دو Load Balancer به VIP خصوصی `10.10.0.6:6443` منتقل شد و failover کنترل‌شده Keepalived با موفقیت تست شد. سپس Public IPهای `node-1`، `node-2` و `node-3` از پنل provider حذف شدند و خروجی اینترنت این نودها از مسیر VIP/NAT روی Load Balancerها تأمین شد.

## خلاصه اجرایی

استقرار فعلی از یک کلاستر سه‌نودی K3s، دو Load Balancer عمومی، و یک Bastion تشکیل شده است. ترافیک عمومی روی `80` و `443` وارد `lb-1` و `lb-2` می‌شود، توسط HAProxy در حالت TCP عبور داده می‌شود، و به NodePortهای `ingress-nginx` روی هر سه نود K3s می‌رسد.

وضعیت زنده با فرض اولیه تفاوت دارد: روی `lb-1` و `lb-2`، `nginx` نصب/فعال نیست و لایه edge فعلی عملا HAProxy + Keepalived است.

پس از cutover تاریخ 2026-07-16، مسیر عمومی از fixed public IPهای قدیمی به Floating IPها منتقل شد: `37.32.31.230` برای `lb-1` و `37.32.31.139` برای `lb-2`. هر دو Floating IP روی `80` و `443` پاسخ سالم دادند.

HA خصوصی Kubernetes API اکنون از طریق VIP خصوصی `10.10.0.6:6443` فعال است. از Bastion اتصال TCP به VIP برقرار شد و Kubernetes API برای درخواست بدون credential پاسخ `401 Unauthorized` داد. در تست failover، VIP از `lb-1` به `lb-2` منتقل شد و پس از بازگشت `lb-1` دوباره روی نود با priority بالاتر نشست.

نودهای K3s اکنون private-only هستند. هیچ‌کدام Public IPv4 ندارند؛ SSH و Kubernetes API فقط از شبکه خصوصی و از طریق Bastion/Load Balancer قابل دسترسی است.

## فهرست ماشین‌ها

| نقش | Hostname | اینترفیس خصوصی | IP خصوصی | اینترفیس عمومی | IP عمومی |
| --- | --- | --- | --- | --- | --- |
| Bastion | `bastion` | `private0` / `fa:16:3e:e4:0c:1b` | `10.10.0.5/24` | `public0` / `fa:16:3e:19:75:91` | `37.152.191.92/22` |
| Load balancer | `lb-1` | `eth0` / `fa:16:3e:d8:b2:72` | `10.10.0.7/24` + VIP `10.10.0.6/24` | Floating IP | `37.32.31.230` |
| Load balancer | `lb-2` | `eth0` / `fa:16:3e:cc:41:ab` | `10.10.0.8/24` | Floating IP | `37.32.31.139` |
| K3s node | `node-1` | `eth1` / `fa:16:3e:3a:7d:ce` | `10.10.0.21/24` | حذف‌شده | ندارد |
| K3s node | `node-2` | `eth1` / `fa:16:3e:a7:91:15` | `10.10.0.22/24` | حذف‌شده | ندارد |
| K3s node | `node-3` | `eth1` / `fa:16:3e:1a:14:bd` | `10.10.0.23/24` | حذف‌شده | ندارد |

همه ماشین‌ها Ubuntu 24.04 با kernel `6.8.0-134-generic` هستند.

## وضعیت منابع

| ماشین | vCPU | RAM | دیسک root | نکته |
| --- | ---: | ---: | ---: | --- |
| `bastion` | 1 | 961 MiB | 23 GiB، مصرف 29% | Squid و listenerهای local Xray دیده شد |
| `lb-1` | 1 | 961 MiB | 23 GiB، مصرف 8% | HAProxy و Keepalived فعال |
| `lb-2` | 1 | 961 MiB | 23 GiB، مصرف 9% | HAProxy و Keepalived فعال |
| `node-1` | 4 | 11 GiB | 70 GiB، مصرف 77% | K3s فعال |
| `node-2` | 2 | 3.8 GiB | 70 GiB، مصرف 66% | K3s فعال |
| `node-3` | 2 | 3.8 GiB | 70 GiB، مصرف 78% | K3s فعال |

روی ماشین‌ها swap فعال نیست.

## شبکه و Netplan

شبکه خصوصی `10.10.0.0/24` است و gateway آن `10.10.0.1` است.

روی Bastion، اینترفیس عمومی و خصوصی با netplan استاتیک تعریف شده است. روی دو LB پس از cutover، public fixed interface از guest OS حذف شد و route پیش‌فرض روی `eth0` از طریق gateway خصوصی `10.10.0.1` تنظیم شد.

روی نودهای K3s مدل شبکه اکنون private-only است:

- `eth1` اینترفیس خصوصی static داخل `10.10.0.0/24` است.
- `eth0` عمومی از guest OS و پنل provider حذف شده است.
- default route روی هر سه node به VIP خصوصی `10.10.0.6` اشاره می‌کند.
- policy routing مربوط به Public IPها حذف شده است.

نتیجه عملی: خروجی اینترنت nodeها از مسیر `node -> 10.10.0.6 -> lb-1/lb-2 NAT -> internet` انجام می‌شود. این مسیر با پاسخ `200` از `1.1.1.1`، پاسخ `401` از Docker Registry و پاسخ `200` از GitHub روی نودها تست شد.

## Bastion

`bastion` نقطه ورود SSH عمومی با IP `37.152.191.92` است. listenerهای مهم:

- `0.0.0.0:22` و `[::]:22` برای SSH.
- `10.10.0.5:3128` برای Squid.
- `127.0.0.1:10808` و `127.0.0.1:10809` برای Xray local.

UFW غیرفعال است. سرویس `ufw` در systemd وجود دارد، اما `ufw status verbose` مقدار `Status: inactive` برگرداند. این با توضیح شما درباره استفاده از Firewall پنل سازگار است.

## لایه Load Balancer

### وضعیت سرویس‌ها

روی هر دو LB:

- `haproxy.service`: فعال و enabled.
- `keepalived.service`: فعال و enabled.
- `k3s-egress-nat.service`: فعال و enabled.
- `nginx`: نصب/شناخته‌شده نیست.

بنابراین runbookها و مستندات عملیاتی باید وضعیت زنده HAProxy + Keepalived را منبع اصلی بدانند.

### مسیر عمومی HTTP/HTTPS

HAProxy روی هر دو LB این پورت‌ها را گوش می‌دهد:

- `0.0.0.0:80`
- `0.0.0.0:443`

Frontendها TCP passthrough هستند:

- `public_http` به backend `ingress_http`
- `public_https` به backend `ingress_https`

Backendها:

- HTTP: `10.10.0.21:30081`، `10.10.0.22:30081`، `10.10.0.23:30081`
- HTTPS: `10.10.0.21:30444`، `10.10.0.22:30444`، `10.10.0.23:30444`

HAProxy در mode `tcp` و با `roundrobin` کار می‌کند. TLS روی LB terminate نمی‌شود؛ ترافیک HTTPS به `ingress-nginx` داخل Kubernetes پاس داده می‌شود.

### مسیر خصوصی Kubernetes API

طراحی مورد انتظار:

- VIP خصوصی `10.10.0.6/24` روی `eth0`

وضعیت Keepalived:

- `lb-1`: حالت `MASTER`، priority `200`، peer برابر `10.10.0.8`
- `lb-2`: حالت `BACKUP`، priority `150`، peer برابر `10.10.0.7`
- health script فقط `pidof haproxy` را چک می‌کند.

وضعیت فعلی پس از اصلاح 2026-07-16:

- HAProxy روی هر دو LB به `10.10.0.6:6443` bind شده است.
- `net.ipv4.ip_nonlocal_bind=1` روی هر دو LB فعال است تا نود BACKUP هم بتواند پیش از مالک شدن VIP روی آن bind کند.
- از Bastion، `10.10.0.6:6443` باز است.
- درخواست بدون credential به `https://10.10.0.6:6443/version` پاسخ `401 Unauthorized` از Kubernetes API برمی‌گرداند.
- گواهی Kubernetes API شامل `10.10.0.6` و `k3s-api.internal` در SAN است.
- در تست کنترل‌شده، با توقف موقت Keepalived روی `lb-1`، VIP روی `lb-2` نشست و اتصال API برقرار ماند.
- پس از بازگرداندن `lb-1`، VIP به دلیل priority بالاتر به `lb-1` برگشت.

نتیجه: مسیر رسمی API خصوصی اکنون VIP `10.10.0.6:6443` است.

### مسیر خروجی نودهای K3s

پس از حذف Public IP از nodeها، خروجی اینترنت آن‌ها از NAT روی Load Balancerها انجام می‌شود.

روی هر دو LB:

- `net.ipv4.ip_forward=1`
- `net.ipv4.conf.eth0.send_redirects=0`
- سرویس `k3s-egress-nat.service` فعال است.
- ruleهای `MASQUERADE` فقط برای sourceهای زیر تعریف شده‌اند:
  - `10.10.0.21/32`
  - `10.10.0.22/32`
  - `10.10.0.23/32`
- مقصدهای داخل `10.10.0.0/24` از NAT مستثنی هستند.

فایل‌های live روی LBها:

- `/usr/local/sbin/k3s-egress-nat.sh`
- `/etc/sysctl.d/70-k3s-egress-nat.conf`
- `/etc/systemd/system/k3s-egress-nat.service`

## Kubernetes

K3s روی هر سه node فعال است:

| Node | Status | Role | Internal IP | Version |
| --- | --- | --- | --- | --- |
| `node-1` | Ready | control-plane, etcd | `10.10.0.21` | `v1.35.5+k3s1` |
| `node-2` | Ready | control-plane, etcd | `10.10.0.22` | `v1.35.5+k3s1` |
| `node-3` | Ready | control-plane, etcd | `10.10.0.23` | `v1.35.5+k3s1` |

Service اصلی ingress:

- `ingress-nginx/ingress-nginx-controller`
- نوع: `NodePort`
- پورت‌ها: `80:30081/TCP` و `443:30444/TCP`

Ingress hostهای مهم:

- `godarai.ir`
- `*.godarai.ir`
- `admin.godarai.ir`

Podهای غیرسالم در آخرین چک Kubernetes:

- `observability/loki-chunks-cache-0`: وضعیت `Pending`
- `observability/prometheus-kube-prometheus-stack-prometheus-0`: وضعیت `ImagePullBackOff`

در یک خروجی گسترده‌تر، یک pod مربوط به `docs` هم موقتا `CrashLoopBackOff` دیده شد، اما در چک نهایی non-running نبود؛ آن مورد را فعلا transient در نظر بگیرید مگر دوباره تکرار شود.

## DNS و دسترسی عمومی

NSهای دامنه:

- `c.ns.arvancdn.ir`
- `p.ns.arvancdn.ir`

نمونه queryهای authoritative در زمان بررسی اولیه:

- `godarai.ir A` مقدار `95.38.235.75` را برگرداند.
- یک wildcard تصادفی زیر `godarai.ir` مقدار `95.38.233.36` را برگرداند.
- `www.godarai.ir CNAME` به `godarai.ir` اشاره می‌کند.

وضعیت پس از cutover Floating IP:

- نمونه query از NSهای Arvan برای `godarai.ir` مقدار `37.32.31.230` را برگرداند.
- `37.32.31.139` مستقیماً روی `80/443` سالم است، اما باید در DNS/Arvan نیز به‌عنوان رکورد فعال تأیید شود.

این رفتار با تصویر ذهنی «هر دو رکورد `@` و `*` همیشه هر دو IP را برمی‌گردانند» کاملا یکی نبود. ممکن است به تنظیمات Arvan، proxy/CDN، وزن‌دهی، یا وضعیت پنل مربوط باشد؛ قبل از اتکا به DNS برای HA باید در پنل Arvan دقیق بررسی شود.

تست‌های HTTP/HTTPS:

- `http://godarai.ir`: پاسخ `308` و redirect به HTTPS.
- `https://godarai.ir`: پاسخ `200 OK`.
- HTTP مستقیم به `37.32.31.230` و `37.32.31.139`: پاسخ `308`.
- HTTPS مستقیم به `37.32.31.230` و `37.32.31.139` با SNI برابر `godarai.ir`: پاسخ `200 OK`.

## نقاط ضعف و ریسک‌ها

1. ریسک قبلی HA خصوصی Kubernetes API از طریق VIP برطرف شد. VIP `10.10.0.6` از Bastion قابل اتصال است، HAProxy هر دو LB روی VIP bind شده‌اند، و failover کنترل‌شده موفق بود.

2. drift قبلی HAProxy بین `lb-1` و `lb-2` جمع شد. هر دو کانفیگ اکنون برای API خصوصی از `bind 10.10.0.6:6443` استفاده می‌کنند. باید مراقب بود تغییرات دستی بعدی دوباره این drift را ایجاد نکند.

3. HA عمومی بیشتر DNS-level به نظر می‌رسد، نه public VIP واقعی. اگر DNS یک LB خراب را برگرداند، کاربر تا پایان cache/TTL ممکن است خطا بگیرد، مگر اینکه Arvan health-check/proxy فعال و درست داشته باشد.

4. هر runbook یا ذهنیتی که Nginx را edge اصلی بداند، با وضعیت زنده سازگار نیست و باید HAProxy + Keepalived را منبع اصلی بداند.

5. OS firewall غیرفعال است. با توجه به استفاده از Firewall پنل، این الزاماً اشتباه نیست، اما defense-in-depth حذف شده و هر خطای پنل می‌تواند listenerهای سیستم‌عامل را مستقیم expose کند.

6. نودهای K3s دیگر Public IP ندارند. این ریسک exposure عمومی SSH و پورت‌های Kubernetes را کاهش داده است، اما سلامت `k3s-egress-nat.service` روی LBها اکنون برای image pull، GitOps، دسترسی به registry و وابستگی‌های بیرونی حیاتی است.

7. health check در Keepalived سطحی است. فقط وجود process HAProxy را چک می‌کند، نه bind بودن به VIP درست، نه reachability پورت `6443`، و نه سلامت backendها.

8. health checkهای HAProxy هم TCP-only هستند. برای passthrough منطقی است، ولی خرابی‌های لایه بالاتر ingress/controller/app را تشخیص نمی‌دهد.

9. LBها فقط 1 GiB RAM دارند و swap ندارند. مصرف فعلی سبک است، اما در شرایط log burst، package operation، یا رشد غیرمنتظره memory حاشیه امن کم است.

10. مصرف دیسک nodeها به 66 تا 78 درصد رسیده است. مخصوصا `node-1` و `node-3` نیاز به threshold هشدار و سیاست cleanup/retention دارند.

11. بخشی از observability خراب یا ناقص است. Prometheus در `ImagePullBackOff` و Loki chunks cache در `Pending` بود؛ این توان تشخیص سریع مشکل HA را ضعیف می‌کند.

12. بخشی از کانفیگ manual و بخشی versioned است. NAT خروجی nodeها فعلاً روی LBها به‌صورت systemd/script live پیاده‌سازی شده و باید در runbookها و automation آینده نگه داشته شود.

## بررسی‌های پیشنهادی بعدی

این‌ها پیشنهاد برای بررسی‌اند، نه تغییر فوری:

1. از ruleهای Firewall پنل برای هر شش VPS export یا screenshot بگیرید و با intent امنیتی تطبیق دهید.
2. endpoint رسمی Kubernetes API اکنون `10.10.0.6:6443` است. kubeconfigها و tunnelها را به این VIP هم‌راستا نگه دارید.
3. Port Security شبکه خصوصی را برای `lb-1` و `lb-2` خاموش نگه دارید؛ برای نودهای K3s خاموش کردن آن لازم نیست.
4. کانفیگ `/etc/haproxy/haproxy.cfg` روی دو LB را با repo مقایسه کنید و هر drift جدید را قبل از reload اصلاح کنید.
5. سلامت `k3s-egress-nat.service`، `ip_forward` و ruleهای `MASQUERADE` را به monitoring اضافه کنید.
6. رفتار Arvan DNS/CDN برای `@` و `*` را بررسی کنید: proxy، weight، health check، یا round-robin ساده.
7. رکورد دوم DNS برای `37.32.31.139` را در Arvan بررسی کنید؛ در نمونه query پس از cutover فقط `37.32.31.230` دیده شد.
8. قبل از اتکا به alerting، وضعیت podهای مشکل‌دار observability را بررسی کنید.
9. وضعیت runbookها را پس از هر تغییر عملیاتی با «HAProxy + Keepalived + NAT خروجی nodeها» هم‌راستا نگه دارید.

## شواهد خوانده‌شده

دسته فرمان‌های خوانده‌شده از طریق SSH:

- `hostname`، `date -Is`، نسخه OS/kernel، CPU/RAM/disk.
- محتوای `/etc/netplan/*`.
- `ip -br link`، `ip -br addr`، `ip route`، `ip rule`.
- `ufw status verbose` و وضعیت systemd سرویس‌ها.
- listenerها با `ss -tulpen`.
- `/etc/haproxy/haproxy.cfg`.
- `/etc/keepalived/keepalived.conf` و `check_haproxy.sh`.
- خروجی‌های `k3s kubectl get nodes/services/ingress/pods`.
- queryهای authoritative DNS برای `godarai.ir`، `www.godarai.ir` و wildcard.
- تست‌های HEAD برای HTTP/HTTPS با resolve مستقیم IP.
