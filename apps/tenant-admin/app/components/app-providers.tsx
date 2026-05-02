"use client";

import { NextIntlClientProvider, type AbstractIntlMessages } from "next-intl";

import { ThemeProvider } from "@/components/theme-provider";

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
      <ThemeProvider defaultTheme="light" disableTransitionOnChange>
        {children}
      </ThemeProvider>
    </NextIntlClientProvider>
  );
}
