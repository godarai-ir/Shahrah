'use client';

import React, { createContext, useCallback, useContext, useEffect, useSyncExternalStore } from 'react';

type Theme = 'dark' | 'light';

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
};

type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
  theme: 'dark',
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);
const themeChangeEvent = 'shahrah-theme-change';

function readStoredTheme(storageKey: string, defaultTheme: Theme): Theme {
  if (typeof window === 'undefined') {
    return defaultTheme;
  }

  const storedTheme = window.localStorage.getItem(storageKey);

  return storedTheme === 'dark' || storedTheme === 'light' ? storedTheme : defaultTheme;
}

export function ThemeProvider({
  children,
  defaultTheme = 'dark',
  storageKey = 'vite-ui-theme',
  ...props
}: ThemeProviderProps) {
  const subscribe = useCallback((callback: () => void) => {
    window.addEventListener('storage', callback);
    window.addEventListener(themeChangeEvent, callback);

    return () => {
      window.removeEventListener('storage', callback);
      window.removeEventListener(themeChangeEvent, callback);
    };
  }, []);

  const theme = useSyncExternalStore(
    subscribe,
    () => readStoredTheme(storageKey, defaultTheme),
    () => defaultTheme,
  );

  useEffect(() => {
    const root = window.document.documentElement;

    root.classList.remove('light', 'dark');
    root.classList.add(theme);
  }, [theme]);

  const value = {
    theme,
    setTheme: (theme: Theme) => {
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(storageKey, theme);
        window.dispatchEvent(new Event(themeChangeEvent));
      }
    },
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
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
