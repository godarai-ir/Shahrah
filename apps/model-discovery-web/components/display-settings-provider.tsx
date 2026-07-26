"use client";

import { createContext, useContext, useEffect, useLayoutEffect, useMemo, useState } from 'react';
import { DisplaySettings, DISPLAY_SETTINGS_DEFAULTS, DISPLAY_SETTINGS_STORAGE_KEY } from '@/lib/display-settings';

type DisplaySettingsContextValue = {
  settings: DisplaySettings;
  setSettings: (next: DisplaySettings) => void;
  updateSettings: (patch: Partial<DisplaySettings>) => void;
  resetSettings: () => void;
};

const DisplaySettingsContext = createContext<DisplaySettingsContextValue | null>(null);

export function DisplaySettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettingsState] = useState<DisplaySettings>(DISPLAY_SETTINGS_DEFAULTS);

  useEffect(() => {
    try {
      const raw = window.localStorage.getItem(DISPLAY_SETTINGS_STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as Partial<DisplaySettings>;
      setSettingsState({ ...DISPLAY_SETTINGS_DEFAULTS, ...parsed });
    } catch {
      setSettingsState(DISPLAY_SETTINGS_DEFAULTS);
    }
  }, []);

  const setSettings = (next: DisplaySettings) => {
    setSettingsState(next);
    window.localStorage.setItem(DISPLAY_SETTINGS_STORAGE_KEY, JSON.stringify(next));
  };

  const updateSettings = (patch: Partial<DisplaySettings>) => {
    setSettingsState((prev) => {
      const next = { ...prev, ...patch };
      window.localStorage.setItem(DISPLAY_SETTINGS_STORAGE_KEY, JSON.stringify(next));
      return next;
    });
  };

  const resetSettings = () => {
    setSettingsState(DISPLAY_SETTINGS_DEFAULTS);
    window.localStorage.setItem(
      DISPLAY_SETTINGS_STORAGE_KEY,
      JSON.stringify(DISPLAY_SETTINGS_DEFAULTS),
    );
  };

  const value = useMemo(
    () => ({ settings, setSettings, updateSettings, resetSettings }),
    [settings],
  );

  useLayoutEffect(() => {
    const root = document.documentElement;
    const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const useDark = settings.themeMode === 'DARK'
      || (settings.themeMode === 'SYSTEM' && systemPrefersDark);

    root.classList.toggle('dark', useDark);
    root.style.colorScheme = useDark ? 'dark' : 'light';
  }, [settings.themeMode]);

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

