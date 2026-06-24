"use client";

import { Settings } from "lucide-react";
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
  const { settings, updateSettings, resetSettings } = useDisplaySettings();

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          type="button"
          className="h-9 w-9 bg-zinc-900 p-0 text-zinc-100 hover:bg-zinc-800"
          aria-label="Open display settings"
          title="Display settings"
        >
          <Settings className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Display Settings</DialogTitle>
          <DialogDescription>
            Personal preferences for this browser. Changes are applied instantly.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="currency-unit">Currency</Label>
            <Select
              value={settings.currencyUnit}
              onValueChange={(value) => updateSettings({ currencyUnit: value as CurrencyUnit })}
            >
              <SelectTrigger id="currency-unit">
                <SelectValue placeholder="Select currency" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="USD">Dollar (USD)</SelectItem>
                <SelectItem value="IRR">Rial (IRR)</SelectItem>
                <SelectItem value="TOMAN">Toman</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="date-calendar">Date Calendar</Label>
            <Select
              value={settings.dateCalendar}
              onValueChange={(value) => updateSettings({ dateCalendar: value as DateCalendar })}
            >
              <SelectTrigger id="date-calendar">
                <SelectValue placeholder="Select calendar" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GREGORIAN">Gregorian</SelectItem>
                <SelectItem value="SHAMSI">Shamsi</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="price-scale">Price Unit</Label>
            <Select
              value={settings.priceScale}
              onValueChange={(value) => updateSettings({ priceScale: value as PriceScale })}
            >
              <SelectTrigger id="price-scale">
                <SelectValue placeholder="Select price unit" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="PER_MILLION_TOKENS">Per Million Token</SelectItem>
                <SelectItem value="PER_TOKEN">Per Token</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center justify-end">
            <Button type="button" className="bg-zinc-800 text-zinc-100 hover:bg-zinc-700" onClick={resetSettings}>
              Reset Defaults
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

