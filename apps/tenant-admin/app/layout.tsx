import type { Metadata } from "next";
import { JetBrains_Mono, Vazirmatn } from "next/font/google";
import { getLocale, getMessages } from "next-intl/server";

import AppProviders from "@/app/components/app-providers";
import { QueryProvider } from "@/app/components/query-provider";
import { Toaster } from "@/components/ui/sonner";
import { getLocaleDirection } from "@/lib/locale-direction";

import "./globals.css";

const vazirmatn = Vazirmatn({
  variable: "--font-vazirmatn",
  subsets: ["arabic", "latin"],
  display: "swap",
});

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "گُدارAI | پنل مدیریت فضای کار",
  description: "پنل فارسی مدیریت فضای کار برای کلیدها، مدل‌ها، کیف پول و تنظیمات حساب.",
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "any" },
      { url: "/favicon-32x32.png", type: "image/png", sizes: "32x32" },
      { url: "/favicon-16x16.png", type: "image/png", sizes: "16x16" },
    ],
    apple: [{ url: "/apple-touch-icon.png", sizes: "180x180" }],
  },
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();
  const messages = await getMessages();

  return (
    <html
      lang={locale}
      dir={getLocaleDirection(locale)}
      data-scroll-behavior="smooth"
      suppressHydrationWarning
      className={`${vazirmatn.variable} ${jetBrainsMono.variable} h-full antialiased`}
    >
      <body className="min-h-full bg-background text-foreground">
        <AppProviders locale={locale} messages={messages}>
          <QueryProvider>
            {children}
            <Toaster position="top-center" richColors />
          </QueryProvider>
        </AppProviders>
      </body>
    </html>
  );
}
