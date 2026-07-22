import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { IBM_Plex_Mono, Vazirmatn } from "next/font/google";
import { NextIntlClientProvider } from "next-intl";
import { getLocale, getMessages, getTranslations } from "next-intl/server";
import "./globals.css";
import { QueryProvider } from "@/lib/query-provider";
import { DisplaySettingsProvider } from "@/components/display-settings-provider";
import { DisplaySettingsDialog } from "@/components/display-settings-dialog";
import { CompareProvider } from "@/components/compare/compare-provider";
import { CompareTray } from "@/components/compare/compare-tray";
import { SiteFooter } from "@/components/site-footer";

const sans = Vazirmatn({
  subsets: ["arabic", "latin"],
  variable: "--font-vazirmatn",
  display: "swap",
});

const mono = IBM_Plex_Mono({
  subsets: ["latin"],
  variable: "--font-ibm-plex-mono",
  weight: ["400", "500"],
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
  const links = [
    ["/models", t("models")],
    ["/providers", t("providers")],
    ["/compare", t("compare")],
  ] as const;

  return (
    <html lang={locale} dir="rtl" className={`${sans.variable} ${mono.variable} antialiased`}>
      <body className="flex min-h-screen flex-col bg-background text-foreground">
        <NextIntlClientProvider locale={locale} messages={messages}>
          <QueryProvider>
            <DisplaySettingsProvider>
              <CompareProvider>
                <header className="sticky top-0 z-20 border-b border-border bg-background/90 backdrop-blur">
                  <nav className="mx-auto flex w-full max-w-7xl items-center gap-2 px-4 py-2.5 sm:px-6 lg:px-8" aria-label={t("mainLabel")}>
                    <Link
                      href="/"
                      aria-label={t("brandAria")}
                      className="flex shrink-0 items-center gap-2 rounded-lg px-1.5 py-1 text-foreground transition-colors hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    >
                      <Image src="/catalog/mini_logo.svg" alt="" width={30} height={30} priority className="size-8" />
                      <span className="font-bold tracking-tight">{t("brand")}</span>
                    </Link>
                    <div className="scrollbar-none flex min-w-0 flex-1 items-center gap-1 overflow-x-auto">
                      {links.map(([href, label]) => (
                        <Link
                          key={href}
                          href={href}
                          className="shrink-0 rounded-md px-2 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:px-3 sm:text-sm"
                        >
                          {label}
                        </Link>
                      ))}
                    </div>
                    <div className="shrink-0">
                      <DisplaySettingsDialog />
                    </div>
                  </nav>
                </header>
                <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-5 pb-28 sm:px-6 sm:py-8 lg:px-8">{children}</main>
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
