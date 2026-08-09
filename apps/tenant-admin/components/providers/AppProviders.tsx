"use client";

import { NextIntlClientProvider, type AbstractIntlMessages } from "next-intl";

import { ThemeProvider } from "@/components/theme-provider";
import { DirectionProvider } from "@/components/ui/direction";
import { getLocaleDirection } from "@/lib/locale-direction";

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
