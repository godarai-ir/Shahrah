import type { Metadata } from "next";
import localFont from "next/font/local";

import "./globals.css";
import { NotFoundPage } from "@/components/not-found-page";

const sans = localFont({
  src: [
    { path: "../public/fonts/vazirmatn/Vazirmatn-Regular.woff2", weight: "400", style: "normal" },
    { path: "../public/fonts/vazirmatn/Vazirmatn-Medium.woff2", weight: "500", style: "normal" },
    { path: "../public/fonts/vazirmatn/Vazirmatn-SemiBold.woff2", weight: "600", style: "normal" },
    { path: "../public/fonts/vazirmatn/Vazirmatn-Bold.woff2", weight: "700", style: "normal" },
  ],
  variable: "--font-vazirmatn",
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL || "https://godarai.ir"),
  title: "صفحه پیدا نشد | گدارAI",
  description: "مسیر درخواستی در فهرست مدل‌های گدارAI پیدا نشد.",
  applicationName: "گدارAI",
  manifest: "/catalog/brand/web/site.webmanifest",
  icons: {
    icon: [
      { url: "/catalog/brand/icons/favicon.svg", type: "image/svg+xml" },
      { url: "/catalog/brand/icons/favicon-light.ico", sizes: "any", media: "(prefers-color-scheme: light)" },
      { url: "/catalog/brand/icons/favicon-dark.ico", sizes: "any", media: "(prefers-color-scheme: dark)" },
      { url: "/catalog/brand/icons/favicon.ico", sizes: "any" },
    ],
    apple: [{ url: "/catalog/brand/icons/apple-touch-icon.png", sizes: "180x180" }],
    other: [{ rel: "mask-icon", url: "/catalog/brand/svg/safari-pinned-tab.svg" }],
  },
  openGraph: {
    title: "صفحه پیدا نشد | گدارAI",
    description: "مسیر درخواستی در فهرست مدل‌های گدارAI پیدا نشد.",
    siteName: "گدارAI",
    locale: "fa_IR",
    type: "website",
    images: [
      {
        url: "/catalog/brand/social/og-image-dark-1200x630.png",
        width: 1200,
        height: 630,
        type: "image/png",
        alt: "لوگوی گدارAI",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "صفحه پیدا نشد | گدارAI",
    description: "مسیر درخواستی در فهرست مدل‌های گدارAI پیدا نشد.",
    images: ["/catalog/brand/social/og-image-dark-1200x630.png"],
  },
  robots: { index: false, follow: false },
  other: {
    "msapplication-config": "/catalog/brand/web/browserconfig.xml",
  },
};

const copy = {
  eyebrow: "خطای ۴۰۴",
  title: "این مسیر به جایی نمی‌رسد",
  description:
    "صفحه‌ای که دنبالش بودید پیدا نشد؛ شاید نشانی آن تغییر کرده یا دیگر در دسترس نباشد.",
  hint: "نشانی را دوباره بررسی کنید یا از مسیرهای اصلی کاتالوگ به مقصدتان برسید.",
  homeAction: "بازگشت به فهرست مدل‌ها",
  providersAction: "مرور ارائه‌دهندگان",
  brand: "گدارAI",
  brandAria: "رفتن به فهرست مدل‌های گدارAI",
};

export default function GlobalNotFound() {
  return (
    <html
      lang="fa"
      dir="rtl"
      data-scroll-behavior="smooth"
      className={`${sans.variable} antialiased`}
    >
      <body className="site-canvas min-h-screen bg-background text-foreground">
        <NotFoundPage copy={copy} showBrand />
      </body>
    </html>
  );
}
