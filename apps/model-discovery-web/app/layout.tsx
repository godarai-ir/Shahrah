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
      suppressHydrationWarning
      className={`${sans.variable} antialiased`}
    >
      <body className="site-canvas flex min-h-screen flex-col bg-background text-foreground">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <QueryProvider>
            <DisplaySettingsProvider>
              <CompareProvider>
                <header className="sticky top-0 z-20 border-b border-border bg-background/85 backdrop-blur-xl">
                  <nav
                    className="mx-auto flex w-full max-w-7xl items-center gap-2 px-4 py-2.5 sm:px-6 lg:px-8"
                    aria-label={t("mainLabel")}
                  >
                    <Link
                      href="/"
                      aria-label={t("brandAria")}
                      className="flex shrink-0 items-center gap-2 rounded-xl px-1 py-1 text-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    >
                      <span className="flex size-9 items-center justify-center rounded-xl border border-primary/20 bg-primary-soft">
                        <Image src="/catalog/mini_logo.svg" alt="" width={28} height={28} priority className="size-7" />
                      </span>
                      <span className="hidden font-bold tracking-tight sm:inline">{t("brand")}</span>
                    </Link>
                    <SiteNavigation />
                    <div className="flex shrink-0 items-center gap-1.5">
                      <ThemeToggle />
                      <DisplaySettingsDialog />
                    </div>
                  </nav>
                </header>
                <main className="relative mx-auto w-full max-w-7xl flex-1 px-4 py-5 pb-28 sm:px-6 sm:py-8 lg:px-8">
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
