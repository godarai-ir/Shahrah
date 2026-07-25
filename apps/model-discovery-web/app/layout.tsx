import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import localFont from "next/font/local";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages, getTranslations } from "next-intl/server";
import "./globals.css";
import { QueryProvider } from "@/lib/query-provider";
import { DisplaySettingsProvider } from "@/components/display-settings-provider";
import { DisplaySettingsDialog } from "@/components/display-settings-dialog";
import { CompareProvider } from "@/components/compare/compare-provider";
import { CompareTray } from "@/components/compare/compare-tray";
import { SiteFooter } from "@/components/site-footer";
import { SiteNavigation } from "@/components/site-navigation";
import { ThemeToggle } from "@/components/theme-toggle";

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

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("Metadata");

  return {
    title: t("title"),
    description: t("description"),
  };
}

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const locale = await getLocale();
  const messages = await getMessages();
  const t = await getTranslations("Navigation");

  return (
    <html
      lang={locale}
      dir="rtl"
      data-scroll-behavior="smooth"
      suppressHydrationWarning
      className={`${sans.variable} antialiased`}
    >
      <body className="site-canvas flex min-h-screen flex-col bg-background text-foreground">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <QueryProvider>
            <DisplaySettingsProvider>
              <CompareProvider>
                <header className="sticky top-0 z-50 border-b border-catalog-hero-border bg-catalog-hero/95 text-catalog-hero-foreground shadow-xs backdrop-blur-xl">
                  <nav
                    className="mx-auto flex h-16 w-full max-w-7xl items-center gap-3 px-4 sm:px-6 lg:px-8"
                    aria-label={t("mainLabel")}
                  >
                    <Link
                      href="https://godarai.ir/"
                      aria-label={t("brandAria")}
                      className="group flex shrink-0 items-center gap-2.5 rounded-xl text-catalog-hero-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    >
                      <span className="flex size-10 items-center justify-center rounded-xl border border-catalog-hero-border bg-background/10 shadow-xs transition-all group-hover:bg-background/15 group-hover:shadow-sm">
                        <Image src="/catalog/mini_logo.svg" alt="" width={30} height={30} priority className="size-7.5" />
                      </span>
                      <span className="hidden text-base font-bold tracking-tight sm:inline">
                        {t("brand")}
                      </span>
                    </Link>
                    <SiteNavigation />
                    <div className="flex shrink-0 items-center gap-1 rounded-xl border border-catalog-hero-border bg-background/10 p-1 shadow-xs">
                      <ThemeToggle />
                      <DisplaySettingsDialog />
                    </div>
                  </nav>
                </header>
                <main className="relative w-full flex-1 pb-28">
                  {children}
                </main>
                <SiteFooter />
                <CompareTray />
              </CompareProvider>
            </DisplaySettingsProvider>
          </QueryProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
