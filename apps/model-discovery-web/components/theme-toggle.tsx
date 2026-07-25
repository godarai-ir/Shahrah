"use client";

import { Moon, Sun } from "lucide-react";
import { useTranslations } from "next-intl";
import { useDisplaySettings } from "@/components/display-settings-provider";
import { Button } from "@/components/ui/button";

export function ThemeToggle() {
  const t = useTranslations("DisplaySettings");
  const { settings, updateSettings } = useDisplaySettings();
  const isLight = settings.themeMode === "LIGHT";

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      className="text-catalog-hero-muted hover:bg-background/10 hover:text-catalog-hero-foreground"
      onClick={() => updateSettings({ themeMode: isLight ? "DARK" : "LIGHT" })}
      aria-label={isLight ? t("switchToDark") : t("switchToLight")}
      title={isLight ? t("switchToDark") : t("switchToLight")}
    >
      {isLight ? <Moon aria-hidden="true" /> : <Sun aria-hidden="true" />}
    </Button>
  );
}
