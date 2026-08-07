# GodarAI Tenant Admin Design System — Components

> مستندات کامل کامپوننت‌های UI اپلیکیشن tenant-admin
> مبتنی بر Design Tokens و دارایی‌های رسمی برند GodarAI

---

## فهرست

1. [اصول طراحی](#اصول-طراحی)
2. [کامپوننت‌های پایه](#کامپوننت‌های-پایه)
3. [کامپوننت‌های فرم](#کامپوننت‌های-فرم)
4. [کامپوننت‌های نمایش داده](#کامپوننت‌های-نمایش-داده)
5. [کامپوننت‌های ناوبری](#کامپوننت‌های-ناوبری)
6. [کامپوننت‌های بازخورد](#کامپوننت‌های-بازخورد)
7. [کامپوننت‌های چیدمان](#کامپوننت‌های-چیدمان)
8. [میکرواینتراکشن‌ها](#میکرواینتراکشن‌ها)
9. [راهنمای دسترسی‌پذیری](#راهنمای-دسترسی‌پذیری)

---

## اصول طراحی

### پالت رنگی

| نقش | Light Mode | Dark Mode | کاربرد |
|-----|-----------|-----------|--------|
| Primary | Indigo `oklch(0.488 0.243 264.376)` | Deep Indigo `oklch(0.424 0.199 265.638)` | دکمه‌های اصلی، لینک‌ها، آیکون‌های فعال |
| Focus | Primary | Foreground پرکنتراست | focus ring و مرز کنترل فعال |
| Secondary | Neutral `oklch(0.967 0.001 286.375)` | Charcoal `oklch(0.274 0.006 286.033)` | دکمه‌های ثانویه و بک‌گراند المان‌ها |
| Success | `oklch(0.47 0.12 160)` | `oklch(0.72 0.14 155)` | پیام موفقیت، وضعیت فعال |
| Warning | `oklch(0.49 0.11 70)` | `oklch(0.76 0.14 80)` | هشدار، نیاز به توجه |
| Destructive | `oklch(0.577 0.245 27.325)` | `oklch(0.704 0.191 22.216)` | حذف، خطر، خطا |
| Info | Indigo | Light Blue | اطلاعات، راهنما |

### تایپوگرافی

فونت پایه: **Estedad** (فارسی و لاتین) — فونت متغیر با تمام وزن‌ها از Thin (100) تا Black (900). فونت Vazirmatn در دورهٔ ارزیابی به‌عنوان fallback نگه داشته شده است.

| استایل | اندازه | وزن | خط‌فاصله | کاربرد |
|--------|--------|-----|---------|--------|
| Heading 1 | 24px / 1.5rem | 600 | 1.25 | عنوان صفحات داشبورد |
| Heading 2 | 20px / 1.25rem | 600 | 1.25 | عنوان بخش‌ها |
| Heading 3 | 18px / 1.125rem | 600 | 1.25 | عنوان کارت‌ها |
| Heading 4 | 16px / 1rem | 500 | 1.5 | عنوان زیربخش‌ها |
| Body | 14px / 0.875rem | 400 | 1.5 | متن اصلی |
| Caption | 12px / 0.75rem | 400 | 1.5 | توضیحات، برچسب‌ها |
| Button | 14px / 0.875rem | 500 | 1.5 | متن دکمه‌ها |

### فاصله‌گذاری

سیستم 8pt با مقیاس ریز برای فواصل کوچک:

| Token | مقدار | کاربرد |
|-------|-------|--------|
| `--space-1` | 4px | فاصله داخلی آیکون‌ها |
| `--space-2` | 8px | فاصله بین المان‌های کوچک |
| `--space-3` | 12px | فاصله بین لیبل و اینپوت |
| `--space-4` | 16px | پدینگ کارت‌ها |
| `--space-6` | 24px | فاصله بین بخش‌ها |
| `--space-8` | 32px | فاصله بین ستون‌ها |

### شعاع گوشه

| Token | مقدار | کاربرد |
|-------|-------|--------|
| `--radius-sm` | 2px | بج‌ها، تگ‌ها |
| `--radius-md` | 4px | اینپوت‌ها، دکمه‌های کوچک |
| `--radius-lg` | 10px | کارت‌ها، دکمه‌های اصلی |
| `--radius-xl` | 14px | مودال‌ها، پنل‌ها |
| `--radius-2xl` | 16px | تصاویر پروفایل |
| `--radius-full` | 9999px | آواتارها، پیل‌ها |

---

## کامپوننت‌های پایه

### GodarAI Logo [`GodarAILogo`](apps/tenant-admin/components/app-ui/GodarAILogo.tsx)

```tsx
import { GodarAILogo } from "@/components/app-ui";

<GodarAILogo alt="GodarAI" className="h-10 w-auto max-w-full object-contain" />
<GodarAILogo variant="dark" decorative className="h-10 w-auto max-w-full object-contain" />
```

`auto` حالت پیش‌فرض است و از theme اپ پیروی می‌کند. `light` و `dark` برای
پس‌زمینه‌های ثابت و `mono-navy`/`mono-white` برای کاربرد تک‌رنگ هستند.
برای جلوگیری از اعوجاج همیشه ارتفاع را کنترل کنید و عرض را `auto` نگه دارید.

### Button [`Button`](apps/tenant-admin/components/ui/button.tsx)

```tsx
import { Button } from "@/components/ui/button";

// Variantهای اصلی
<Button variant="default">       // Primary — رنگ semantic برند
<Button variant="secondary">     // Secondary — بک‌گراند muted
<Button variant="outline">       // Outline — با border
<Button variant="ghost">         // Ghost — بدون بک‌گراند
<Button variant="destructive">   // Destructive — قرمز
<Button variant="link">          // Link — استایل لینک

// اندازه‌ها
<Button size="sm">      // کوچک — h-8
<Button size="default"> // پیش‌فرض — h-10
<Button size="lg">      // بزرگ — h-12
<Button size="icon">    // آیکون — مربع

// حالت‌ها
<Button disabled>       // غیرفعال
<Button loading>        // در حال بارگذاری (spinner)
```

**ویژگی‌های دسترسی‌پذیری:**
- حداقل ارتفاع 40px برای touch target (WCAG 2.5.5)
- نسبت کنتراست 4.5:1 برای متن روی دکمه‌ها
- حالت focus-visible با ring واضح

---

### Card [`Card`](apps/tenant-admin/components/ui/card.tsx)

```tsx
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";

<Card className="card-hover">
  <CardHeader>
    <CardTitle>عنوان کارت</CardTitle>
    <CardDescription>توضیحات کوتاه</CardDescription>
  </CardHeader>
  <CardContent>
    محتوای اصلی
  </CardContent>
  <CardFooter>
    اکشن‌ها
  </CardFooter>
</Card>
```

**انواع کارت:**
- **Stat Card**: نمایش آمار (مقدار، تغییر درصدی، آیکون)
- **Action Card**: کارت با اکشن اصلی (دکمه در footer)
- **Info Card**: کارت اطلاعاتی (بدون اکشن)
- **Interactive Card**: کارت قابل کلیک (hover effect)

---

### Badge [`Badge`](apps/tenant-admin/components/ui/badge.tsx)

```tsx
import { Badge } from "@/components/ui/badge";

<Badge>عادی</Badge>
<Badge variant="secondary">ثانویه</Badge>
<Badge variant="destructive">حذفی</Badge>
<Badge variant="outline">Outline</Badge>
<Badge variant="success">موفقیت</Badge>
<Badge variant="warning">هشدار</Badge>
```

**کاربردها:**
- وضعیت تراکنش: `success` (موفق)، `warning` (در انتظار)، `destructive` (ناموفق)
- برچسب نقش: `default` (ادمین)، `secondary` (کاربر)
- تعداد اعلان‌ها: badge قرمز با عدد

---

### Avatar [`Avatar`](apps/tenant-admin/components/ui/avatar.tsx)

```tsx
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

<Avatar>
  <AvatarImage src="/avatar.jpg" alt="نام کاربر" />
  <AvatarFallback>ع‌ک</AvatarFallback>
</Avatar>
```

**اندازه‌ها:**
- `h-8 w-8` — لیست کاربران
- `h-10 w-10` — هدر، پروفایل
- `h-12 w-12` — کارت پروفایل

---

### Separator [`Separator`](apps/tenant-admin/components/ui/separator.tsx)

```tsx
import { Separator } from "@/components/ui/separator";

<Separator />          // افقی
<Separator orientation="vertical" /> // عمودی
```

---

### Table [`Table`](apps/tenant-admin/components/ui/table.tsx)

```tsx
import { Table, TableHeader, TableBody, TableFooter, TableRow, TableHead, TableCell, TableCaption } from "@/components/ui/table";

<Table>
  <TableCaption>لیست کاربران</TableCaption>
  <TableHeader>
    <TableRow>
      <TableHead>نام</TableHead>
      <TableHead>ایمیل</TableHead>
      <TableHead>نقش</TableHead>
      <TableHead className="text-left">عملیات</TableHead>
    </TableRow>
  </TableHeader>
  <TableBody>
    <TableRow>
      <TableCell>رضا دلپاک</TableCell>
      <TableCell>r.delpak@example.com</TableCell>
      <TableCell><Badge>ادمین</Badge></TableCell>
      <TableCell className="text-left">
        <Button variant="ghost" size="sm">ویرایش</Button>
      </TableCell>
    </TableRow>
  </TableBody>
</Table>
```

---

## کامپوننت‌های فرم

### Input [`Input`](apps/tenant-admin/components/ui/input.tsx)

```tsx
import { Input } from "@/components/ui/input";

<Input placeholder="ایمیل خود را وارد کنید" />
<Input type="email" />
<Input type="password" />
<Input disabled />
<Input readOnly />
<Input className="border-destructive" /> // حالت خطا
```

### Label [`Label`](apps/tenant-admin/components/ui/label.tsx)

```tsx
import { Label } from "@/components/ui/label";

<Label htmlFor="email">ایمیل</Label>
```

### Form Pattern

```tsx
<div className="space-y-3">
  <Label htmlFor="email">ایمیل</Label>
  <Input id="email" type="email" placeholder="example@mail.com" />
  <p className="text-xs text-muted-foreground">ایمیل باید معتبر باشد</p>
  <p className="text-xs text-destructive">این فیلد الزامی است</p>
</div>
```

---

## کامپوننت‌های نمایش داده

### Alert [`Alert`](apps/tenant-admin/components/ui/alert.tsx)

```tsx
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";

<Alert variant="default">
  <AlertTitle>موفقیت</AlertTitle>
  <AlertDescription>عملیات با موفقیت انجام شد.</AlertDescription>
</Alert>

<Alert variant="destructive">
  <AlertTitle>خطا</AlertTitle>
  <AlertDescription>مشکلی در اتصال به سرور پیش آمده.</AlertDescription>
</Alert>
```

**انواع Alert:**
- `default` — اطلاعات عمومی (آبی)
- `destructive` — خطا (قرمز)
- `success` — موفقیت (سبز) — custom variant

---

### Sheet [`Sheet`](apps/tenant-admin/components/ui/sheet.tsx)

```tsx
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetTrigger } from "@/components/ui/sheet";

<Sheet>
  <SheetTrigger asChild>
    <Button variant="outline">باز کردن</Button>
  </SheetTrigger>
  <SheetContent side="right">
    <SheetHeader>
      <SheetTitle>عنوان پنل</SheetTitle>
      <SheetDescription>توضیحات پنل</SheetDescription>
    </SheetHeader>
    محتوای پنل
  </SheetContent>
</Sheet>
```

### Dropdown Menu

```tsx
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuLabel } from "@/components/ui/dropdown-menu";

<DropdownMenu>
  <DropdownMenuTrigger asChild>
    <Button variant="ghost">منو</Button>
  </DropdownMenuTrigger>
  <DropdownMenuContent align="end">
    <DropdownMenuLabel>تنظیمات</DropdownMenuLabel>
    <DropdownMenuItem>پروفایل</DropdownMenuItem>
    <DropdownMenuItem>تنظیمات</DropdownMenuItem>
    <DropdownMenuSeparator />
    <DropdownMenuItem className="text-destructive">خروج</DropdownMenuItem>
  </DropdownMenuContent>
</DropdownMenu>
```

---

## کامپوننت‌های ناوبری

### Admin Shell [`AdminShell`](apps/tenant-admin/app/components/admin-shell.tsx)

ساختار اصلی داشبورد با سایدبار و هدر:

```tsx
<AdminShell tenantSlug="my-tenant" currentUser={user}>
  <TenantAdminDashboard {...data} />
</AdminShell>
```

**ساختار ناوبری:**
- داشبورد (`/`)
- شارژ کیف پول (`/billing/topup`)
- حساب‌های ارائه‌دهنده (`/provider-accounts`)
- مدیریت دسترسی (`/users`)
- تنظیمات (`/settings`)

---

## کامپوننت‌های بازخورد

### Loading States

```tsx
// Skeleton loader
<div className="space-y-4">
  <div className="h-4 w-3/4 rounded bg-muted shimmer" />
  <div className="h-4 w-1/2 rounded bg-muted shimmer" />
</div>

// Spinner
<div className="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full" />
```

### Toast Notifications

```tsx
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";

showFeedbackToast({
  kind: "success",
  title: "موفق",
  description: "عملیات انجام شد",
});

showFeedbackToast({
  kind: "error",
  title: "خطا",
  description: "مشکلی پیش آمد",
});
```

Toast فقط برای نتیجه‌های موقتی و غیرمسدودکننده است. خطاهای قابل‌اصلاح فرم باید داخل همان فرم با `OperationNotice`، شرایط پایدار با `Alert` و تأیید عملیات مخرب با `ConfirmDialog` مبتنی بر `AlertDialog` نمایش داده شوند. کد feature نباید `sonner` را مستقیماً import کند.

---

## کامپوننت‌های چیدمان

### Dashboard Grid

```tsx
<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
  {/* 4 stat card in desktop, 2 in tablet, 1 in mobile */}
  <StatCard />
  <StatCard />
  <StatCard />
  <StatCard />
</div>
```

### Page Layout

```tsx
<div className="space-y-6">
  {/* Page Header */}
  <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <h1 className="text-2xl font-semibold">عنوان صفحه</h1>
    <Button>اکشن اصلی</Button>
  </div>
  
  {/* Content */}
  <div className="grid gap-6 md:grid-cols-2">
    {/* Cards, Tables, etc */}
  </div>
</div>
```

---

## میکرواینتراکشن‌ها

### Hover Effects

| المان | افکت | Duration |
|-------|------|----------|
| Card | `translateY(-2px)` + `shadow-md` | 200ms |
| Button | رنگ بک‌گراند + `shadow-primary` | 200ms |
| TableRow | `bg-muted/50` | 150ms |
| Nav Item | `bg-accent` + `border-r` | 150ms |

### Active States

| المان | افکت |
|-------|------|
| Button | `scale(0.98)` |
| Tab | `bg-accent` |
| Checkbox | چک‌مارک با animation |

### Loading Animations

- **Shimmer** — برای skeleton cards
- **Spinner** — برای دکمه‌ها و صفحات
- **Fade-in** — برای محتوای جدید

---

## راهنمای دسترسی‌پذیری

### WCAG 2.2 Compliance

| معیار | الزام | پیاده‌سازی |
|-------|-------|-----------|
| 1.4.3 Contrast | نسبت 4.5:1 | تمام رنگ‌ها OKLCH با کنتراست محاسبه‌شده |
| 2.4.7 Focus Visible | indicator واضح | `:focus-visible` با `outline` 2px |
| 2.5.5 Target Size | حداقل 44×44px | دکمه‌ها و لینک‌ها با `min-h-10` |
| 1.3.1 Info & Relationships | سمانتیک HTML | `<header>`, `<nav>`, `<main>`, `<section>` |
| 2.1.1 Keyboard | keyboard accessible | تمام تعامل‌ها با keyboard |
| 3.3.2 Labels | label برای form | `<Label htmlFor="...">` |
| 1.4.13 Hover Content | محتوا قابل پنهان‌سازی | tooltip‌ها با `aria-describedby` |

### RTL Support

- تمام کامپوننت‌ها `dir="rtl"` دارند
- فاصله‌گذاری از `space-x` بجای `space-x-reverse` استفاده می‌کند
- آیکون‌های جهت‌دار (فلش) باید mirror شوند

### Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## فایل‌های مرتبط

| فایل | مسیر | توضیح |
|------|------|-------|
| Design Tokens | `design-system/tokens.css` | متغیرهای CSS |
| Global Styles | `app/globals.css` | استایل‌های سراسری |
| Button | `components/ui/button.tsx` | کامپوننت دکمه |
| Card | `components/ui/card.tsx` | کامپوننت کارت |
| Input | `components/ui/input.tsx` | کامپوننت اینپوت |
| Alert | `components/ui/alert.tsx` | کامپوننت اعلان |
