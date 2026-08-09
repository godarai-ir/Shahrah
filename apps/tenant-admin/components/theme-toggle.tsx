"use client";

import { Moon, Sun } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { useTheme } from "@/components/theme-provider";

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const t = useTranslations("theme");
  const isDark = resolvedTheme === "dark";

  return (
    <Button
      type="button"
      variant="outline"
      size="icon"
      aria-label={t("toggle")}
      aria-pressed={isDark}
      onClick={() => setTheme(isDark ? "light" : "dark")}
      className="relative shrink-0"
    >
      <Sun
        data-icon="inline-start"
        className="rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0"
        aria-hidden="true"
      />
      <Moon
        data-icon="inline-start"
        className="absolute rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100"
        aria-hidden="true"
      />
      <span className="sr-only">{isDark ? t("dark") : t("light")}</span>
    </Button>
  );
}

export function ThemeToggleMenuItem() {
  const { resolvedTheme, setTheme } = useTheme();
  const t = useTranslations("theme");
  const isDark = resolvedTheme === "dark";

  return (
    <DropdownMenuItem
      className="min-h-10 justify-start gap-3 rounded-md px-3"
      onSelect={() => setTheme(isDark ? "light" : "dark")}
    >
      {isDark ? <Moon aria-hidden="true" /> : <Sun aria-hidden="true" />}
      <span>{t("toggle")}</span>
      <span className="ms-auto text-xs text-muted-foreground">
        {isDark ? t("dark") : t("light")}
      </span>
    </DropdownMenuItem>
  );
}
