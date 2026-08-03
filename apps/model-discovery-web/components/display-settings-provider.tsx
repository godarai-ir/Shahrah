"use client";

import type { ResolvedTheme, ThemePreference } from '@shahrah/theme-sync';
import { useQuery } from '@tanstack/react-query';
import {
  applyThemePreference,
  getBrowserThemePreference,
  getSystemPrefersDark,
  migrateBrowserThemePreference,
  resolveThemePreference,
  setBrowserThemePreference,
  subscribeToThemePreference,
} from '@shahrah/theme-sync';
import {
  createContext,
  useContext,
  useEffect,
  useLayoutEffect,
  useState,
  useSyncExternalStore,
} from 'react';
import {
  DisplaySettings,
  DISPLAY_SETTINGS_DEFAULTS,
  DISPLAY_SETTINGS_STORAGE_KEY,
  ThemeMode,
} from '@/lib/display-settings';
import {
  ExchangeRateAvailability,
  ExchangeRateQuote,
  fetchCurrentExchangeRate,
  resolveExchangeRateAvailability,
} from '@/lib/exchange-rate';

const themeReadOptions = {
  defaultTheme: 'system' as const,
  legacyJsonStorage: [
    {key: DISPLAY_SETTINGS_STORAGE_KEY, property: 'themeMode'},
  ],
};

type DisplaySettingsContextValue = {
  settings: DisplaySettings;
  resolvedTheme: ResolvedTheme;
  setSettings: (next: DisplaySettings) => void;
  updateSettings: (patch: Partial<DisplaySettings>) => void;
  resetSettings: () => void;
  exchangeRate: ExchangeRateQuote | null;
  exchangeRateAvailability: ExchangeRateAvailability;
};

const DisplaySettingsContext = createContext<DisplaySettingsContextValue | null>(null);

function preferenceToThemeMode(preference: ThemePreference): ThemeMode {
  return preference.toUpperCase() as ThemeMode;
}

function themeModeToPreference(themeMode: ThemeMode): ThemePreference {
  return themeMode.toLowerCase() as ThemePreference;
}

function readLegacyDisplaySettings(): DisplaySettings {
  try {
    const raw = window.localStorage.getItem(DISPLAY_SETTINGS_STORAGE_KEY);
    if (!raw) return DISPLAY_SETTINGS_DEFAULTS;

    const parsed = JSON.parse(raw) as Partial<DisplaySettings>;
    return {...DISPLAY_SETTINGS_DEFAULTS, ...parsed};
  } catch {
    return DISPLAY_SETTINGS_DEFAULTS;
  }
}

function persistDisplaySettings(settings: DisplaySettings) {
  window.localStorage.setItem(DISPLAY_SETTINGS_STORAGE_KEY, JSON.stringify(settings));
}

export function DisplaySettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettingsState] = useState<DisplaySettings>(DISPLAY_SETTINGS_DEFAULTS);
  const exchangeRateQuery = useQuery({
    queryKey: ['current-exchange-rate'],
    queryFn: ({ signal }) => fetchCurrentExchangeRate(signal),
    staleTime: 60_000,
    refetchInterval: 5 * 60_000,
    retry: 1,
  });
  const exchangeRate = exchangeRateQuery.data ?? null;
  const exchangeRateAvailability = resolveExchangeRateAvailability(
    exchangeRate,
    exchangeRateQuery.isPending,
  );
  const systemPrefersDark = useSyncExternalStore(
    subscribeToThemePreference,
    getSystemPrefersDark,
    () => false,
  );
  const resolvedTheme: ResolvedTheme = resolveThemePreference(
    themeModeToPreference(settings.themeMode),
    systemPrefersDark,
  );

  useEffect(() => {
    const syncFromSharedTheme = () => {
      const nextPreference = getBrowserThemePreference(themeReadOptions);
      const nextThemeMode = preferenceToThemeMode(nextPreference);
      applyThemePreference(nextPreference);

      setSettingsState((current) => {
        if (current.themeMode === nextThemeMode) return current;
        const next = {...current, themeMode: nextThemeMode};
        persistDisplaySettings(next);
        return next;
      });
    };

    const unsubscribe = subscribeToThemePreference(syncFromSharedTheme);
    const frame = window.requestAnimationFrame(() => {
      const legacySettings = readLegacyDisplaySettings();
      const preference = migrateBrowserThemePreference(themeReadOptions);
      const migratedSettings = {
        ...legacySettings,
        themeMode: preferenceToThemeMode(preference),
      };

      persistDisplaySettings(migratedSettings);
      setSettingsState(migratedSettings);
      applyThemePreference(preference);
    });

    return () => {
      window.cancelAnimationFrame(frame);
      unsubscribe();
    };
  }, []);

  const setSettings = (next: DisplaySettings) => {
    setSettingsState(next);
    persistDisplaySettings(next);
    setBrowserThemePreference(themeModeToPreference(next.themeMode));
  };

  const updateSettings = (patch: Partial<DisplaySettings>) => {
    const next = {...settings, ...patch};
    setSettingsState(next);
    persistDisplaySettings(next);
    if (patch.themeMode) {
      setBrowserThemePreference(themeModeToPreference(patch.themeMode));
    }
  };

  const resetSettings = () => {
    setSettingsState(DISPLAY_SETTINGS_DEFAULTS);
    persistDisplaySettings(DISPLAY_SETTINGS_DEFAULTS);
    setBrowserThemePreference(themeModeToPreference(DISPLAY_SETTINGS_DEFAULTS.themeMode));
  };

  useLayoutEffect(() => {
    const preference = themeModeToPreference(settings.themeMode);
    applyThemePreference(preference);
  }, [resolvedTheme, settings.themeMode]);

  const value = {
    settings,
    resolvedTheme,
    setSettings,
    updateSettings,
    resetSettings,
    exchangeRate,
    exchangeRateAvailability,
  };

  return (
    <DisplaySettingsContext.Provider value={value}>
      {children}
    </DisplaySettingsContext.Provider>
  );
}

export function useDisplaySettings() {
  const ctx = useContext(DisplaySettingsContext);
  if (!ctx) {
    throw new Error('useDisplaySettings must be used within DisplaySettingsProvider');
  }
  return ctx;
}
