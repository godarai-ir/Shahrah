'use client';

import type { ThemePreference, ResolvedTheme } from '@shahrah/theme-sync';
import {
  applyThemePreference,
  getBrowserThemePreference,
  getSystemPrefersDark,
  migrateBrowserThemePreference,
  resolveThemePreference,
  setBrowserThemePreference,
  subscribeToThemePreference,
} from '@shahrah/theme-sync';
import React, { createContext, useCallback, useContext, useEffect, useMemo, useSyncExternalStore } from 'react';

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: ThemePreference;
  storageKey?: string;
};

type ThemeProviderState = {
  theme: ThemePreference;
  resolvedTheme: ResolvedTheme;
  setTheme: (theme: ThemePreference) => void;
};

const initialState: ThemeProviderState = {
  theme: 'system',
  resolvedTheme: 'light',
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = 'shahrah-theme',
}: ThemeProviderProps) {
  const readOptions = useMemo(
    () => ({defaultTheme, legacyStorageKeys: [storageKey, 'vite-ui-theme', 'theme']}),
    [defaultTheme, storageKey],
  );
  const subscribe = useCallback(
    (callback: () => void) => subscribeToThemePreference(callback),
    [],
  );
  const getThemeSnapshot = useCallback(
    () => getBrowserThemePreference(readOptions),
    [readOptions],
  );
  const getResolvedSnapshot = useCallback(
    () => resolveThemePreference(getThemeSnapshot(), getSystemPrefersDark()),
    [getThemeSnapshot],
  );
  const theme = useSyncExternalStore(subscribe, getThemeSnapshot, () => defaultTheme);
  const resolvedTheme = useSyncExternalStore(
    subscribe,
    getResolvedSnapshot,
    () => resolveThemePreference(defaultTheme),
  );

  useEffect(() => {
    const migratedTheme = migrateBrowserThemePreference(readOptions);
    applyThemePreference(migratedTheme);
  }, [readOptions]);

  useEffect(() => {
    applyThemePreference(theme);
  }, [resolvedTheme, theme]);

  const value = useMemo<ThemeProviderState>(
    () => ({
      theme,
      resolvedTheme,
      setTheme: (nextTheme) => {
        const preference = setBrowserThemePreference(nextTheme);
        applyThemePreference(preference);
      },
    }),
    [resolvedTheme, theme],
  );

  return (
    <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined)
    throw new Error('useTheme must be used within a ThemeProvider');

  return context;
};
