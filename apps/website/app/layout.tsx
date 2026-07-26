import type { Metadata, Viewport } from "next";
import "./globals.css";
import { defaultSocialImage, siteUrl } from "@/app/lib/seo";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "گدارAI | درگاه مدل‌های زبانی برای تیم‌های ایرانی",
    template: "%s | گدارAI",
  },
  description:
    "گدارAI درگاه مدل‌های زبانی برای اتصال، مسیریابی، کنترل هزینه و پایش مصرف مدل‌ها در تیم‌های ایرانی است.",
  applicationName: "گدارAI",
  manifest: "/brand/web/site.webmanifest",
  icons: {
    icon: [
      { url: "/brand/icons/favicon.svg", type: "image/svg+xml" },
      { url: "/brand/icons/favicon-light.ico", sizes: "any", media: "(prefers-color-scheme: light)" },
      { url: "/brand/icons/favicon-dark.ico", sizes: "any", media: "(prefers-color-scheme: dark)" },
      { url: "/brand/icons/favicon.ico", sizes: "any" },
      { url: "/brand/icons/favicon-32x32.png", type: "image/png", sizes: "32x32" },
      { url: "/brand/icons/favicon-16x16.png", type: "image/png", sizes: "16x16" },
    ],
    apple: [{ url: "/brand/icons/apple-touch-icon.png", sizes: "180x180" }],
    other: [{ rel: "mask-icon", url: "/brand/svg/safari-pinned-tab.svg" }],
  },
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "گدارAI | درگاه مدل‌های زبانی برای تیم‌های ایرانی",
    description:
      "مصرف مدل‌های زبانی را با گدارAI از یک مسیر قابل‌پیگیری مدیریت، پایش و کنترل کنید.",
    url: "/",
    siteName: "گدارAI",
    locale: "fa_IR",
    type: "website",
    images: [
      {
        url: defaultSocialImage,
        width: 1200,
        height: 630,
        type: "image/png",
        alt: "لوگوی گدارAI",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "گدارAI | درگاه مدل‌های زبانی",
    description:
      "درگاه مدل‌های زبانی برای پرداخت ریالی، مشاهده‌پذیری و کنترل مصرف در ایران.",
    images: [defaultSocialImage],
  },
  robots: {
    index: true,
    follow: true,
  },
  other: {
    "msapplication-config": "/brand/web/browserconfig.xml",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#FFFFFF" },
    { media: "(prefers-color-scheme: dark)", color: "#06182D" },
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="fa" dir="rtl" suppressHydrationWarning className="h-full antialiased" style={{ fontFamily: 'Vazirmatn, sans-serif' }}>
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
