"use client";

import type { ReactNode } from "react";
import {
  Banknote,
  CalendarDays,
  CheckCircle2,
  Gauge,
  Monitor,
  Moon,
  Palette,
  RotateCcw,
  Settings,
  SlidersHorizontal,
  Sun,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDisplaySettings } from "@/components/display-settings-provider";
import { CurrencyUnit, DateCalendar, PriceScale, ThemeMode } from "@/lib/display-settings";

export function DisplaySettingsDialog() {
  const t = useTranslations("DisplaySettings");
  const { settings, updateSettings, resetSettings } = useDisplaySettings();

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="text-catalog-hero-muted hover:bg-background/10 hover:text-catalog-hero-foreground"
          aria-label={t("openAria")}
          title={t("title")}
        >
          <Settings aria-hidden="true" />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-dvh max-w-xl overflow-y-auto p-0">
        <DialogHeader className="border-b border-border bg-muted/35 px-6 py-5 pe-14">
          <div className="flex items-start gap-3">
            <span className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-primary/15 bg-primary-soft text-primary-soft-foreground shadow-xs">
              <SlidersHorizontal className="size-5" aria-hidden="true" />
            </span>
            <div className="min-w-0">
              <DialogTitle>{t("title")}</DialogTitle>
              <DialogDescription className="mt-1">{t("description")}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="grid gap-3 px-6 py-5 sm:grid-cols-2">
          <SettingField
            icon={<Palette aria-hidden="true" />}
            htmlFor="theme-mode"
            label={t("theme")}
            description={t("themeDescription")}
          >
            <Select
              dir="rtl"
              value={settings.themeMode}
              onValueChange={(value) => updateSettings({ themeMode: value as ThemeMode })}
            >
              <SelectTrigger
                id="theme-mode"
                className="[&_[data-slot=select-value]]:flex [&_[data-slot=select-value]]:min-w-0 [&_[data-slot=select-value]]:items-center [&_[data-slot=select-value]]:gap-2 [&_[data-slot=select-value]]:whitespace-nowrap [&_[data-slot=select-value]_svg]:size-4 [&_[data-slot=select-value]_svg]:shrink-0"
              >
                <SelectValue placeholder={t("selectTheme")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="LIGHT">
                    <Sun aria-hidden="true" />
                    {t("light")}
                  </SelectItem>
                  <SelectItem value="DARK">
                    <Moon aria-hidden="true" />
                    {t("dark")}
                  </SelectItem>
                  <SelectItem value="SYSTEM">
                    <Monitor aria-hidden="true" />
                    {t("system")}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </SettingField>

          <SettingField
            icon={<Banknote aria-hidden="true" />}
            htmlFor="currency-unit"
            label={t("currency")}
            description={t("currencyDescription")}
          >
            <Select
              dir="rtl"
              value={settings.currencyUnit}
              onValueChange={(value) => updateSettings({ currencyUnit: value as CurrencyUnit })}
            >
              <SelectTrigger id="currency-unit">
                <SelectValue placeholder={t("selectCurrency")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="USD">{t("dollar")}</SelectItem>
                  <SelectItem value="IRR">{t("rial")}</SelectItem>
                  <SelectItem value="TOMAN">{t("toman")}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </SettingField>

          <SettingField
            icon={<CalendarDays aria-hidden="true" />}
            htmlFor="date-calendar"
            label={t("calendar")}
            description={t("calendarDescription")}
          >
            <Select
              dir="rtl"
              value={settings.dateCalendar}
              onValueChange={(value) => updateSettings({ dateCalendar: value as DateCalendar })}
            >
              <SelectTrigger id="date-calendar">
                <SelectValue placeholder={t("selectCalendar")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="GREGORIAN">{t("gregorian")}</SelectItem>
                  <SelectItem value="SHAMSI">{t("shamsi")}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </SettingField>

          <SettingField
            icon={<Gauge aria-hidden="true" />}
            htmlFor="price-scale"
            label={t("priceUnit")}
            description={t("priceUnitDescription")}
          >
            <Select
              dir="rtl"
              value={settings.priceScale}
              onValueChange={(value) => updateSettings({ priceScale: value as PriceScale })}
            >
              <SelectTrigger id="price-scale">
                <SelectValue placeholder={t("selectPriceUnit")} />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="PER_MILLION_TOKENS">{t("perMillionTokens")}</SelectItem>
                  <SelectItem value="PER_TOKEN">{t("perToken")}</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </SettingField>
        </div>

        <DialogFooter className="justify-between border-t border-border bg-muted/25 px-6 py-4">
          <span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
            <CheckCircle2 className="size-4 text-success-soft-foreground" aria-hidden="true" />
            {t("autoSaveHint")}
          </span>
          <Button type="button" variant="outline" size="lg" onClick={resetSettings}>
            <RotateCcw data-icon="inline-start" aria-hidden="true" />
            {t("resetDefaults")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function SettingField({
  icon,
  htmlFor,
  label,
  description,
  children,
}: {
  icon: ReactNode;
  htmlFor: string;
  label: string;
  description: string;
  children: ReactNode;
}) {
  return (
    <section className="rounded-xl border border-card-border bg-card p-4 shadow-xs">
      <div className="flex items-start gap-2.5">
        <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground [&_svg]:size-4">
          {icon}
        </span>
        <div className="min-w-0">
          <Label htmlFor={htmlFor} className="text-sm font-semibold">{label}</Label>
          <p className="mt-0.5 text-xs leading-5 text-muted-foreground">{description}</p>
        </div>
      </div>
      <div className="mt-3 [&_[data-slot=select-trigger]]:bg-background [&_[data-slot=select-trigger]]:shadow-xs">
        {children}
      </div>
    </section>
  );
}

