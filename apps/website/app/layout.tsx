import type { Metadata } from "next";
import "./globals.css";
import { absoluteUrl } from "@/app/lib/seo";

export const metadata: Metadata = {
  metadataBase: new URL(absoluteUrl("/")),
  title: {
    default: "گدارAI | درگاه مدل‌های زبانی برای تیم‌های ایرانی",
    template: "%s | گدارAI",
  },
  description:
    "گدارAI درگاه مدل‌های زبانی برای اتصال، مسیریابی، کنترل هزینه و پایش مصرف مدل‌ها در تیم‌های ایرانی است.",
  applicationName: "گدارAI",
  manifest: "/site.webmanifest",
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "any" },
      { url: "/favicon-32x32.png", type: "image/png", sizes: "32x32" },
      { url: "/favicon-16x16.png", type: "image/png", sizes: "16x16" },
    ],
    apple: [{ url: "/apple-touch-icon.png", sizes: "180x180" }],
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
  },
  twitter: {
    card: "summary_large_image",
    title: "گدارAI | درگاه مدل‌های زبانی",
    description:
      "درگاه مدل‌های زبانی برای پرداخت ریالی، مشاهده‌پذیری و کنترل مصرف در ایران.",
  },
  robots: {
    index: true,
    follow: true,
  },
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
