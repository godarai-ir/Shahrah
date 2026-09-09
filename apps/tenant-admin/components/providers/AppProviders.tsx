"use client";

import { NextIntlClientProvider, type AbstractIntlMessages } from "next-intl";
import { useEffect } from "react";

import { ThemeProvider } from "@/components/theme-provider";
import { DirectionProvider } from "@/components/ui/direction";
import { getLocaleDirection } from "@/lib/locale-direction";
import { applyDocumentLocale } from "@/lib/locale-preference";

type AppProvidersProps = {
  children: React.ReactNode;
  locale: string;
  messages: AbstractIntlMessages;
};

export default function AppProviders({
  children,
  locale,
  messages,
}: AppProvidersProps) {
  // Keep `<html lang>` and `<html dir>` in step with the server-resolved locale
  // after a language switch, which re-renders without a full document load.
  useEffect(() => {
    applyDocumentLocale(locale);
  }, [locale]);

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <DirectionProvider dir={getLocaleDirection(locale)}>
        <ThemeProvider defaultTheme="system" disableTransitionOnChange>
          {children}
        </ThemeProvider>
      </DirectionProvider>
    </NextIntlClientProvider>
  );
}
