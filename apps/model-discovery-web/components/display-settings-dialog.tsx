"use client";

import { Settings } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useDisplaySettings } from "@/components/display-settings-provider";
import { CurrencyUnit, DateCalendar, PriceScale } from "@/lib/display-settings";

export function DisplaySettingsDialog() {
  const t = useTranslations("DisplaySettings");
  const { settings, updateSettings, resetSettings } = useDisplaySettings();

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          type="button"
          className="h-9 w-9 bg-muted p-0 text-foreground hover:bg-accent"
          aria-label={t("openAria")}
          title={t("title")}
        >
          <Settings className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="currency-unit">{t("currency")}</Label>
            <Select
              dir="rtl"
              value={settings.currencyUnit}
              onValueChange={(value) => updateSettings({ currencyUnit: value as CurrencyUnit })}
            >
              <SelectTrigger id="currency-unit">
                <SelectValue placeholder={t("selectCurrency")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="USD">{t("dollar")}</SelectItem>
                <SelectItem value="IRR">{t("rial")}</SelectItem>
                <SelectItem value="TOMAN">{t("toman")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="date-calendar">{t("calendar")}</Label>
            <Select
              dir="rtl"
              value={settings.dateCalendar}
              onValueChange={(value) => updateSettings({ dateCalendar: value as DateCalendar })}
            >
              <SelectTrigger id="date-calendar">
                <SelectValue placeholder={t("selectCalendar")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GREGORIAN">{t("gregorian")}</SelectItem>
                <SelectItem value="SHAMSI">{t("shamsi")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="price-scale">{t("priceUnit")}</Label>
            <Select
              dir="rtl"
              value={settings.priceScale}
              onValueChange={(value) => updateSettings({ priceScale: value as PriceScale })}
            >
              <SelectTrigger id="price-scale">
                <SelectValue placeholder={t("selectPriceUnit")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="PER_MILLION_TOKENS">{t("perMillionTokens")}</SelectItem>
                <SelectItem value="PER_TOKEN">{t("perToken")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-end">
            <Button type="button" className="bg-muted text-foreground hover:bg-accent" onClick={resetSettings}>
              {t("resetDefaults")}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

