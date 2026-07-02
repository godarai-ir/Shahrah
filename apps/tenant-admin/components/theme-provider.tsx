"use client";

import React, { createContext, useContext, useEffect, useState } from "react";

type Theme = "dark" | "light";

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
  attribute?: string;
  enableSystem?: boolean;
  disableTransitionOnChange?: boolean;
};

type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
  theme: "light",
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  defaultTheme = "light",
  storageKey = "shahrah-theme",
  ...props
}: ThemeProviderProps) {
  const [theme, setTheme] = useState<Theme>(defaultTheme);

  useEffect(() => {
    const storedTheme = localStorage.getItem(storageKey);

    if (storedTheme === "dark" || storedTheme === "light") {
      const frame = window.requestAnimationFrame(() => {
        setTheme(storedTheme);
      });

      return () => window.cancelAnimationFrame(frame);
    }
  }, [storageKey]);

  useEffect(() => {
    const root = window.document.documentElement;

    if (props.disableTransitionOnChange) {
      root.classList.add("transition-none");
    }

    root.classList.remove("light", "dark");
    root.classList.add(theme);

    if (props.disableTransitionOnChange) {
      // Force reflow
      void root.offsetHeight;
      root.classList.remove("transition-none");
    }
  }, [theme, props.disableTransitionOnChange]);

  const value = {
    theme,
    setTheme: (theme: Theme) => {
      localStorage.setItem(storageKey, theme);
      setTheme(theme);
    },
  };

  return (
    <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }

  return context;
};
