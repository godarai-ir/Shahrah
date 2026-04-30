# Todo List: تکمیل ماژول مدیریت دسترسی و کاربران

> مبنا: `plans/tenant-organization-team-design.md`، `plans/11_organization_team_virtualkey_design.md`، `plans/shahrah_access_control.md`
>
> وضعیت فعلی کد: اسکیمای دیتابیس هنوز حول `Organization` / `Team` / `VerificationToken` می‌چرخد، در `control-plane` فقط بخش محدودی از مدیریت کاربران tenant-admin و ساخت key پیاده شده، و در `tenant-admin` فعلاً مدیریت کاربران tenant-admin و دعوت‌نامه‌ها وجود دارد.

---

## تصمیم‌های پایه قبل از اجرا

- مرجع نهایی دامنه را `tenant-organization-team-design.md` در نظر بگیریم.
- برای جلوگیری از ری‌ورک، این نگاشت را رسمی کنیم:
  - `Organization` فعلی = لایه انتقالی برای `OrgUnit`
  - `VerificationToken` فعلی = لایه انتقالی برای `PAT` و `VAT`
- UI فاز بعدی فقط بعد از تثبیت قراردادهای API و permission matrix شروع شود.

---

## تصویر شکاف فعلی

### آنچه الان داریم

- مدل‌های پایه دیتابیس برای `organizations`, `teams`, `users`, `organization_memberships`, `team_memberships`, `verification_tokens`
- API محدود برای:
  - ساخت و لیست key
  - دعوت کاربر tenant-admin
  - فعال/غیرفعال کردن tenant-admin user
  - جستجوی access principal برای provider account access
- UI محدود در `tenant-admin` برای:
  - لیست کاربران tenant-admin
  - دعوت کاربر
  - فعال/غیرفعال‌سازی

### آنچه هنوز کامل نیست

- نبود model نهایی برای `OrgUnit`
- نبود model مستقل برای `VirtualUser`
- نبود تفکیک شفاف بین `PAT` و `VAT`
- نبود RBAC نهایی با scopeهای `tenant_admin` / `org_admin` / `team_admin` / `team_member` / `tenant_user`
- نبود endpointهای کامل مدیریت org unit / team / membership / virtual user / key lifecycle
- نبود enforcement واقعی permissionها روی endpointهای مدیریت و مصرف
- نبود UI عملیاتی برای teams / roles / memberships / virtual accounts / access policies

---

## فاز 1 — تثبیت دامنه و قراردادها

- [ ] یک ADR کوتاه بنویسیم و naming نهایی را قطعی کنیم: `OrgUnit`, `Team`, `User`, `VirtualUser`, `PAT`, `VAT`
- [ ] compatibility map بین schema فعلی و domain نهایی تعریف کنیم
- [ ] permission matrix نهایی برای همه roleها و همه endpointها تدوین شود
- [ ] ownership rule برای هر entity مشخص شود:
  - چه کسی create/update/delete می‌کند
  - چه کسی read usage / budgets / members را می‌بیند
  - چه کسی key صادر می‌کند
- [ ] قرارداد API نهایی برای لیست، جزئیات، create/update/delete و membership operations نوشته شود

**خروجی این فاز:** API contract ثابت و قابل استفاده برای backend و `tenant-admin`

---

## فاز 2 — مدل داده و migration

- [ ] schema نهایی برای `org_units` اضافه شود یا migration رسمی از `organizations` به `org_units` طراحی شود
- [ ] جدول `virtual_users` اضافه شود
- [ ] تفکیک منطقی/فیزیکی `PAT` و `VAT` مشخص شود:
  - یا دو جدول جدا
  - یا یک جدول با `key_type` و قیدهای معتبر
- [ ] membershipها بازبینی شوند:
  - `org_unit_memberships`
  - `team_memberships`
  - budget / permission overrides در سطح membership
- [ ] ستون‌ها و ایندکس‌های اجباری نهایی شوند:
  - `tenant_id`
  - `org_unit_id`
  - `team_id`
  - `user_id`
  - `virtual_user_id`
  - `token_hash`
- [ ] audit fields روی همه entityهای IAM بررسی و یکدست شوند
- [ ] migration strategy برای داده‌های موجود نوشته شود تا downtime و ambiguity نداشته باشیم

**نکته:** تا قبل از این فاز، ساخت UI کامل زودهنگام است.

---

## فاز 3 — RBAC و authorization enforcement

- [ ] enumها و typeهای RBAC در `packages/contracts` و لایه Go نهایی شوند
- [ ] middleware / helper مشترک برای تشخیص scope کاربر پیاده شود
- [ ] ruleهای دسترسی برای این سطوح enforce شود:
  - tenant-wide
  - org-unit scoped
  - team scoped
  - self-scoped
- [ ] team member permissions به‌صورت route-level و policy-level enforce شوند
- [ ] cross-tenant isolation برای همه queryها و mutationها تست شود
- [ ] permission checkها روی endpointهای management و endpointهای key usage هم‌راستا شوند
- [ ] audit log برای عملیات حساس فعال شود:
  - invite user
  - change role
  - add/remove membership
  - create/revoke/rotate key
  - activate/deactivate user or virtual user

---

## فاز 4 — Backend CRUD کامل مدیریت دسترسی و کاربران

### Org Unit / Organization

- [ ] `create org unit`
- [ ] `update org unit`
- [ ] `delete org unit`
- [ ] `get org unit details`
- [ ] `list org units`
- [ ] `add/remove/update org unit member`

### Team

- [ ] `create team`
- [ ] `update team`
- [ ] `delete team`
- [ ] `get team details`
- [ ] `list teams`
- [ ] `add/remove/update team member`
- [ ] `team member permissions list/update`
- [ ] پشتیبانی از team بدون org unit

### User

- [ ] `create user`
- [ ] `update user`
- [ ] `delete user`
- [ ] `get user details`
- [ ] `list users`
- [ ] `list user memberships`
- [ ] `activate/deactivate user`
- [ ] دعوت، پذیرش دعوت، expiry، resend، revoke

### Virtual User

- [ ] `create virtual user`
- [ ] `update virtual user`
- [ ] `delete virtual user`
- [ ] `get virtual user details`
- [ ] `list virtual users`
- [ ] attach/detach به team
- [ ] activate/deactivate virtual user

### Keys

- [ ] `generate PAT`
- [ ] `generate VAT`
- [ ] `list keys`
- [ ] `key details`
- [ ] `update key`
- [ ] `revoke/delete key`
- [ ] `regenerate/rotate key`
- [ ] masking و one-time reveal برای secret
- [ ] lifecycle rules برای keyهای user-bound و service-bound

---

## فاز 5 — Budget, model access, inheritance

- [ ] inheritance chain نهایی پیاده شود:
  - tenant
  - org unit
  - team
  - user / virtual user
  - key
- [ ] validation برای over-allocation پیاده شود
- [ ] model allowlist در همه سطوح enforce شود
- [ ] budget reset semantics و windowهای هم‌زمان مشخص و پیاده شود
- [ ] usage aggregation برای tenant / org unit / team / user / virtual user / key تکمیل شود
- [ ] soft-budget alert و hard-budget block از هم تفکیک شوند

---

## فاز 6 — Test و hardening

- [ ] unit test برای role helpers و permission evaluation
- [ ] integration test برای CRUD کامل entityهای IAM
- [ ] integration test برای isolation بین tenantها
- [ ] integration test برای permission denial cases
- [ ] integration test برای PAT/VAT lifecycle
- [ ] test برای invite acceptance, expiry, resend, revoke
- [ ] test برای budget inheritance و model restriction
- [ ] observability:
  - structured logs
  - audit events
  - admin metrics

---

## فاز 7 — UI در `tenant-admin` (مرحله بعدی شما)

### پیش‌نیاز UI

- [ ] endpointها و response shapeها نهایی شوند
- [ ] statusها و enumها stable شوند
- [ ] namespaceهای i18n برای IAM طراحی شوند

### صفحات و جریان‌ها

- [ ] صفحه `Users`
  - لیست
  - جستجو
  - فیلتر role/status
  - دعوت
  - فعال/غیرفعال
  - مشاهده membershipها
- [ ] صفحه `Teams`
  - لیست تیم‌ها
  - create/edit/delete
  - اعضا
  - permissions
  - بودجه و مدل‌های مجاز
- [ ] صفحه `Org Units`
  - لیست
  - create/edit/delete
  - اعضا
  - تیم‌های زیرمجموعه
- [ ] صفحه `Virtual Accounts`
  - لیست virtual user
  - create/edit/deactivate
  - اتصال به team
  - VAT management
- [ ] صفحه `Roles & Access`
  - matrix نقش‌ها
  - team member permissions
  - access summaries
- [ ] صفحه `Keys`
  - PAT/VAT list
  - create/revoke/rotate
  - masked display
  - scope, models, expiry, spend
- [ ] صفحه `Audit Logs`
  - role changes
  - membership changes
  - key events

### الزامات UI

- [ ] همه متن‌ها با `next-intl`
- [ ] عدم استفاده از hardcoded strings
- [ ] هم‌ترازی `messages/fa.json` و `messages/en.json`
- [ ] stateهای loading / empty / error / success برای همه flowها
- [ ] optimistic update فقط در جایی که rollback مطمئن داریم

---

## ترتیب اجرایی پیشنهادی

1. نهایی‌سازی domain و API contract
2. migration دیتابیس و modelهای نهایی
3. RBAC enforcement
4. backend CRUD برای org unit / team / user / virtual user / key
5. budgets و inheritance
6. test و hardening
7. UI در `apps/tenant-admin`

---

## Backlog دقیق برای شروع فوری

### Sprint 1

- [ ] نهایی‌کردن naming و permission matrix
- [ ] طراحی schema نهایی `org_units` و `virtual_users`
- [ ] تصمیم نهایی PAT/VAT storage model
- [ ] تعریف contract endpointهای IAM

### Sprint 2

- [ ] migration دیتابیس
- [ ] پیاده‌سازی repo/service برای org units و virtual users
- [ ] پیاده‌سازی RBAC helpers و middleware
- [ ] تکمیل endpointهای membership management

### Sprint 3

- [ ] پیاده‌سازی lifecycle کامل keyها
- [ ] پیاده‌سازی budget/model inheritance
- [ ] integration testهای اصلی

### Sprint 4

- [ ] شروع UI `tenant-admin` با صفحات `Users`, `Teams`, `Org Units`
- [ ] تکمیل `Virtual Accounts`, `Keys`, `Roles & Access`
- [ ] polish نهایی i18n و UX states

---

## پیشنهاد اجرایی

اگر بخواهیم ری‌ورک کم شود، قدم بعدی باید این باشد:

- اول: من یک **IAM backend implementation plan** بر اساس همین Todo به taskهای فایل‌به‌فایل تبدیل کنم
- بعد: وارد فاز UI `tenant-admin` شویم

این ترتیب از ساخت UI روی API ناپایدار جلوگیری می‌کند.

---

## Phase 1 Status

- [x] ADR-level naming decisions completed in `plans/13_iam_phase1_foundation.md`
- [x] Compatibility map completed in `plans/13_iam_phase1_foundation.md`
- [x] Permission matrix completed in `plans/13_iam_phase1_foundation.md`
- [x] Ownership rules completed in `plans/13_iam_phase1_foundation.md`
- [x] Stable IAM API contract completed in `plans/13_iam_phase1_foundation.md`
