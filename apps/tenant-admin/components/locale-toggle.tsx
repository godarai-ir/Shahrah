"use client";

import { Languages } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { useTransition } from "react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { SUPPORTED_LOCALES, resolveLocale, type AppLocale } from "@/lib/locale";
import { setBrowserLocalePreference } from "@/lib/locale-preference";

function useLocaleSwitcher() {
  const router = useRouter();
  const activeLocale = resolveLocale(useLocale());
  const [isPending, startTransition] = useTransition();

  function selectLocale(nextLocale: AppLocale) {
    if (nextLocale === activeLocale) return;

    setBrowserLocalePreference(nextLocale);
    startTransition(() => {
      router.refresh();
    });
  }

  return { activeLocale, isPending, selectLocale };
}

/** Standalone control for surfaces without a profile menu, such as login. */
export function LocaleToggle() {
  const { activeLocale, isPending, selectLocale } = useLocaleSwitcher();
  const t = useTranslations("language");

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          type="button"
          variant="outline"
          size="icon"
          aria-label={t("switch")}
          disabled={isPending}
          className="shrink-0"
        >
          <Languages aria-hidden="true" />
          <span className="sr-only">{t(`names.${activeLocale}`)}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" sideOffset={8} className="w-44 p-1 text-start">
        <DropdownMenuRadioGroup
          value={activeLocale}
          onValueChange={(value) => selectLocale(resolveLocale(value))}
        >
          {SUPPORTED_LOCALES.map((locale) => (
            <DropdownMenuRadioItem key={locale} value={locale} className="min-h-9 rounded-md">
              <span lang={locale}>{t(`names.${locale}`)}</span>
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/** Row for the profile dropdown; toggles between the two supported locales. */
export function LocaleToggleMenuItem() {
  const { activeLocale, isPending, selectLocale } = useLocaleSwitcher();
  const t = useTranslations("language");
  const nextLocale: AppLocale = activeLocale === "fa" ? "en" : "fa";

  return (
    <DropdownMenuItem
      className="min-h-10 justify-start gap-3 rounded-md px-3"
      disabled={isPending}
      onSelect={() => selectLocale(nextLocale)}
    >
      <Languages aria-hidden="true" />
      <span>{t("switch")}</span>
      <span lang={nextLocale} className="ms-auto text-xs text-muted-foreground">
        {t(`names.${nextLocale}`)}
      </span>
    </DropdownMenuItem>
  );
}
