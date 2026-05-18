"use client";

import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useDisplaySettings } from '@/components/display-settings-provider';
import {
  CurrencyUnit,
  DateCalendar,
  PriceScale,
  DISPLAY_SETTINGS_DEFAULTS,
} from '@/lib/display-settings';

export default function DisplaySettingsPage() {
  const { settings, updateSettings, resetSettings } = useDisplaySettings();

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Display Settings</h1>
      <p className="text-sm text-zinc-400">
        These preferences are personal and stored only in your browser.
        Changes apply instantly.
      </p>

      <Card className="space-y-4">
        <label className="block space-y-1 text-sm">
          <span className="text-zinc-300">Currency Unit</span>
          <select
            name="currencyUnit"
            value={settings.currencyUnit}
            onChange={(e) =>
              updateSettings({ currencyUnit: e.target.value as CurrencyUnit })
            }
            className="w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2"
          >
            <option value="USD">USD (Dollar)</option>
            <option value="IRR">IRR (Rial)</option>
            <option value="TOMAN">Toman</option>
          </select>
        </label>

        <label className="block space-y-1 text-sm">
          <span className="text-zinc-300">Date Calendar</span>
          <select
            name="dateCalendar"
            value={settings.dateCalendar}
            onChange={(e) =>
              updateSettings({ dateCalendar: e.target.value as DateCalendar })
            }
            className="w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2"
          >
            <option value="GREGORIAN">Gregorian</option>
            <option value="SHAMSI">Shamsi (Persian)</option>
          </select>
        </label>

        <label className="block space-y-1 text-sm">
          <span className="text-zinc-300">Price Scale</span>
          <select
            name="priceScale"
            value={settings.priceScale}
            onChange={(e) =>
              updateSettings({ priceScale: e.target.value as PriceScale })
            }
            className="w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2"
          >
            <option value="PER_MILLION_TOKENS">Per million tokens</option>
            <option value="PER_TOKEN">Per token</option>
          </select>
        </label>

        <div className="rounded-md border border-zinc-800 bg-zinc-950 p-3 text-xs text-zinc-400">
          Defaults: {DISPLAY_SETTINGS_DEFAULTS.dateCalendar} | {DISPLAY_SETTINGS_DEFAULTS.currencyUnit} | {DISPLAY_SETTINGS_DEFAULTS.priceScale}
        </div>

        <div>
          <Button type="button" className="bg-zinc-700 text-white hover:bg-zinc-600" onClick={resetSettings}>
            Reset to defaults
          </Button>
        </div>
      </Card>
    </div>
  );
}
