import type { Metadata, Viewport } from "next";
import { createThemeBootstrapScript } from "@shahrah/theme-sync";
import { headers } from "next/headers";
import { JetBrains_Mono, Vazirmatn } from "next/font/google";
import { getLocale, getMessages, getTranslations } from "next-intl/server";

import AppProviders from "@/app/components/app-providers";
import { QueryProvider } from "@/app/components/query-provider";
import { Toaster } from "@/components/ui/sonner";
import { getLocaleDirection } from "@/lib/locale-direction";
import { resolveRequestOriginFromHeaders } from "@/lib/tenant";

import "./globals.css";

const themeBootstrapScript = createThemeBootstrapScript({
  defaultTheme: "system",
  legacyStorageKeys: ["shahrah-theme", "vite-ui-theme"],
});

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

export async function generateMetadata(): Promise<Metadata> {
  const requestHeaders = await headers();
  const fallbackOrigin =
    process.env.NEXT_PUBLIC_TENANT_ADMIN_URL?.trim() ||
    process.env.SHAHRAH_TENANT_ADMIN_URL?.trim() ||
    "https://godarai.ir";
  const origin = resolveRequestOriginFromHeaders(requestHeaders, fallbackOrigin);
  const socialImage = new URL("/brand/social/og-image-dark-1200x630.png", origin).toString();
  const title = "گدارAI | پنل مدیریت فضای کار";
  const description = "پنل فارسی مدیریت فضای کار برای کلیدها، مدل‌ها، کیف پول و تنظیمات حساب.";

  return {
    metadataBase: new URL(origin),
    title,
    description,
    applicationName: "گدارAI",
    manifest: "/brand/web/site.webmanifest",
    icons: {
      icon: [
        { url: "/brand/icons/favicon.svg", type: "image/svg+xml" },
        { url: "/brand/icons/favicon-light.ico", sizes: "any", media: "(prefers-color-scheme: light)" },
        { url: "/brand/icons/favicon-dark.ico", sizes: "any", media: "(prefers-color-scheme: dark)" },
        { url: "/brand/icons/favicon.ico", sizes: "any" },
      ],
      apple: [{ url: "/brand/icons/apple-touch-icon.png", sizes: "180x180" }],
      other: [{ rel: "mask-icon", url: "/brand/svg/safari-pinned-tab.svg" }],
    },
    openGraph: {
      title,
      description,
      siteName: "گدارAI",
      locale: "fa_IR",
      type: "website",
      images: [{ url: socialImage, width: 1200, height: 630, type: "image/png", alt: "لوگوی گدارAI" }],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [socialImage],
    },
    robots: {
      index: false,
      follow: false,
      nocache: true,
    },
    other: {
      "msapplication-config": "/brand/web/browserconfig.xml",
    },
  };
}

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#FFFFFF" },
    { media: "(prefers-color-scheme: dark)", color: "#06182D" },
  ],
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();
  const messages = await getMessages();
  const common = await getTranslations("common");

  return (
    <html
      lang={locale}
      dir={getLocaleDirection(locale)}
      data-scroll-behavior="smooth"
      suppressHydrationWarning
      className={`${vazirmatn.variable} ${jetBrainsMono.variable} h-full antialiased`}
    >
      <head>
        <script dangerouslySetInnerHTML={{__html: themeBootstrapScript}} />
      </head>
      <body className="min-h-full bg-background text-foreground">
        <AppProviders locale={locale} messages={messages}>
          <QueryProvider>
            {children}
            <Toaster
              position="top-center"
              dir={getLocaleDirection(locale)}
              containerAriaLabel={common("feedback.notifications")}
              closeButtonAriaLabel={common("feedback.closeNotification")}
            />
          </QueryProvider>
        </AppProviders>
      </body>
    </html>
  );
}
