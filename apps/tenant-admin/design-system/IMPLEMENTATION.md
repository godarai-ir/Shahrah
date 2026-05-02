# Shahrah Design System — Implementation Guide

> راهنمای پیاده‌سازی Design System جدید برای tenant-admin
> مبتنی بر یکپارچگی بصری با وب‌سایت Shahrah

---

## 1. نمای کلی

این Design System برای یکپارچه‌سازی ظاهر بصری tenant-admin با وب‌سایت Shahrah طراحی شده است. تمام کامپوننت‌ها بر اساس Design Tokens مشترک ساخته شده‌اند و از استانداردهای WCAG 2.2 پشتیبانی می‌کنند.

### تغییرات کلیدی

| بخش | قبل | بعد |
|-----|-----|-----|
| پالت رنگی | آبی-بنفش عمومی | Indigo/Purple با برند Shahrah |
| تایپوگرافی | Vazirmatn (بدون وزن) | Vazirmatn (9 وزن) |
| گوشه‌ها | `rounded-md` | `rounded-lg` و `rounded-xl` |
| سایه‌ها | `shadow-sm` ساده | سایه‌های سفارشی با glow |
| انیمیشن‌ها | بدون انیمیشن | Microinteractions هدفمند |
| دسترس‌پذیری | پایه | WCAG 2.2 کامل |

---

## 2. ساختار فایل‌ها

```
apps/tenant-admin/
├── app/
│   ├── globals.css              # Global styles + Design Tokens import
│   └── components/
│       ├── admin-shell.tsx      # Redesigned shell with sidebar + header
│       ├── stat-card.tsx        # New stat card component
│       ├── login-form.tsx       # Redesigned auth form
│       └── ...                  # Other components
├── components/
│   └── ui/
│       ├── button.tsx           # Enhanced button with loading state
│       ├── card.tsx             # Enhanced card with hover effect
│       ├── input.tsx            # Enhanced input with focus ring
│       ├── alert.tsx            # Enhanced alert with icons
│       ├── badge.tsx            # Enhanced badge with variants
│       └── ...                  # Other UI components
├── design-system/
│   ├── tokens.css               # Design Tokens (colors, spacing, etc.)
│   ├── COMPONENTS.md            # Component documentation
│   └── IMPLEMENTATION.md        # This file
└── public/
    └── brand/
        └── shahrah-logo.png     # Brand logo
```

---

## 3. مراحل مهاجرت

### مرحله 1: نصب وابستگی‌های جدید

```bash
cd apps/tenant-admin

# آیکون‌های جدید برای کامپوننت‌ها
pnpm add lucide-react

# اطمینان از نصب بودن shadcn/ui
pnpm add @radix-ui/react-slot class-variance-authority
```

### مرحله 2: جایگزینی فایل‌های UI

فایل‌های زیر در `components/ui/` بازنویسی شده‌اند:

| فایل | تغییرات اصلی | Breaking Changes |
|------|-------------|-----------------|
| `button.tsx` | اضافه شدن `loading` prop، تغییر اندازه‌ها | `h-9` → `h-10` برای default |
| `card.tsx` | اضافه شدن `card-hover` class، responsive padding | بدون تغییر |
| `input.tsx` | focus ring جدید، border رنگی | بدون تغییر |
| `alert.tsx` | اضافه شدن variantهای `success`, `warning`, `info` + icon prop | نیاز به بررسی استفاده‌های موجود |
| `badge.tsx` | rounded-full، variantهای جدید | بدون تغییر |

### مرحله 3: بروزرسانی globals.css

فایل `app/globals.css` باید با محتوای جدید جایگزین شود:

```css
/* محتوای فعلی را با فایل جدید جایگزین کنید */
@import "tailwindcss";
@import "../design-system/tokens.css";

/* بقیه محتوا از فایل جدید */
```

### مرحله 4: بروزرسانی AdminShell

`app/components/admin-shell.tsx` بازنویسی شده است. تغییرات:

1. Sidebar navigation با استایل جدید
2. Header responsive بهبود یافته
3. User profile card در پایین sidebar
4. Mobile menu با Sheet component

### مرحله 5: بروزرسانی Login Form

`app/components/login-form.tsx` بازنویسی شده:

1. Background gradient هم‌راستا با website
2. Logo display بهبود یافته
3. Captcha با UI بهتر
4. OTP flow با انیمیشن fade-in

---

## 4. استفاده از Design Tokens

### در CSS

```css
.my-element {
  background-color: var(--background);
  color: var(--foreground);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  transition: all var(--duration-normal) var(--ease-out);
}
```

### در Tailwind

```tsx
<div className="bg-background text-foreground rounded-lg p-4 transition-all duration-200 ease-out">
  {/* Content */}
</div>
```

### افزودن Token جدید

فایل `design-system/tokens.css` را ویرایش کنید:

```css
:root {
  --my-new-token: oklch(0.53 0.17 255);
}

@theme inline {
  --color-my-new: var(--my-new-token);
}
```

---

## 5. کامپوننت‌های جدید

### StatCard

```tsx
import { Wallet, Key, TrendingUp } from "lucide-react";
import { StatCard } from "@/app/components/stat-card";

<StatCard
  title="موجودی"
  value="۱,۲۵۰,۰۰۰ ریال"
  subtitle="آخرین بروزرسانی: امروز"
  icon={Wallet}
  trend={{ value: 12, positive: true }}
  color="indigo"
/>
```

### Alert با آیکون سفارشی

```tsx
import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";

<Alert variant="info" icon={<AlertCircle className="h-4 w-4" />}>
  <AlertDescription>اطلاعات نمایشی</AlertDescription>
</Alert>
```

### Button با حالت Loading

```tsx
import { Button } from "@/components/ui/button";

<Button loading={isLoading} onClick={handleSubmit}>
  ذخیره
</Button>
```

---

## 6. راهنمای دسترسی‌پذیری (WCAG 2.2)

### نسبت کنتراست

تمام ترکیب‌های رنگی باید حداقل نسبت 4.5:1 را داشته باشند:

| ترکیب | نسبت | وضعیت |
|-------|------|-------|
| foreground / background | 15:1 | ✅ |
| primary / primary-foreground | 7:1 | ✅ |
| muted-foreground / background | 4.6:1 | ✅ |
| destructive / destructive-foreground | 5:1 | ✅ |

### Touch Targets

حداقل اندازه المان‌های تعاملی: 44×44px

```tsx
// ✅ خوب - min-w-[44px]
<Button size="sm" className="min-w-[44px]" />

// ❌ بد - خیلی کوچک
<Button className="w-6" />
```

### Focus Management

تمام المان‌های تعاملی باید `:focus-visible` داشته باشند:

```css
:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}
```

### Screen Reader Support

```tsx
// Label برای screen reader
<Button aria-label="بستن" size="icon">
  <X className="h-4 w-4" />
</Button>

// متن مخفی برای screen reader
<span className="sr-only">این متن فقط برای screen reader است</span>
```

---

## 7. راهنمای RTL

### جهت صفحه

```tsx
// در layout.tsx
<html dir="rtl" lang="fa">
```

### فاصله‌گذاری

```tsx
// ❌ استفاده نکنید - چپ‌چین
<div className="space-x-2">

// ✅ استفاده کنید - خودکار RTL
<div className="gap-2 flex">

// ✅ برای آیکون‌های جهت‌دار
<ArrowLeft className="rotate-180" /> {/* Mirror for RTL */}
```

### ناوبری

```tsx
// Sheet از راست باز شود (مناسب RTL)
<SheetContent side="right">
```

---

## 8. راهنمای انیمیشن

### Microinteractions

| المان | انیمیشن | trigger |
|-------|---------|---------|
| Card | `translateY(-2px) + shadow-md` | hover |
| Button | `background-color + glow` | hover |
| Icon | `scale(1.1)` | hover |
| Nav Item | `bg-accent + border-r` | active |

### Page Transitions

```tsx
<div className="animate-fade-in">
  {/* Page content */}
</div>
```

### Loading States

```tsx
// Shimmer effect
<div className="shimmer rounded-lg h-4 w-3/4" />

// Spinner
<div className="animate-spin h-4 w-4 border-2 border-primary border-t-transparent rounded-full" />
```

### Reduced Motion

برای کاربرانی که انیمیشن را غیرفعال کرده‌اند:

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## 9. تست و اعتبارسنجی

### تست کنتراست

```bash
# نصب ابزار
npm install -g @axe-core/cli

# اجرا
axe ./dist --reporter html
```

### تست Responsive

| اندازه | عرض | دستگاه |
|--------|-----|--------|
| Mobile | < 768px | موبایل |
| Tablet | 768px - 1024px | تبلت |
| Desktop | > 1024px | دسکتاپ |

### تست دسترسی‌پذیری

- Lighthouse Accessibility score > 90
- axe-core violations = 0
- Keyboard navigation complete

---

## 10. رفع مشکلات رایج

### مشکل: رنگ‌ها اعمال نمی‌شوند

```bash
# کش Tailwind را پاک کنید
rm -rf .next
rm -rf node_modules/.cache

# دوباره بسازید
pnpm build
```

### مشکل: انیمیشن‌ها کار نمی‌کنند

```css
/* اطمینان از import درست */
@import "../design-system/tokens.css";
```

### مشکل: Focus ring نمایش داده نمی‌شود

```css
/* اطمینان از وجود :focus-visible */
:focus-visible {
  outline: 2px solid var(--ring);
  outline-offset: 2px;
}
```

---

## 11. چک‌لیست قبل از Deploy

- [ ] تمام تست‌های واحد عبور شده
- [ ] Lighthouse Accessibility > 90
- [ ] تست keyboard navigation در Chrome, Firefox, Safari
- [ ] تست screen reader با NVDA یا VoiceOver
- [ ] تست responsive در 3 اندازه صفحه
- [ ] تست dark/light mode
- [ ] تست RTL
- [ ] بررسی کنتراست رنگ‌ها
- [ ] بررسی reduced motion

---

## 12. منابع بیشتر

- [WCAG 2.2 Guidelines](https://www.w3.org/TR/WCAG22/)
- [OKLCH Color Space](https://oklch.com/)
- [Vazirmatn Font](https://github.com/rastikerdar/vazirmatn)
- [TailwindCSS Documentation](https://tailwindcss.com/)
- [Radix UI Primitives](https://www.radix-ui.com/)
