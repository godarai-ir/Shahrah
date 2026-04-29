'use client';

import type { ReactNode } from 'react';
import { Navbar } from './Navbar';
import { ThemeProvider } from './ThemeProvider';
import { Footer } from './Trust';

export function SiteShell({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="shahrah-theme">
      <div
        className="min-h-screen bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white transition-colors duration-300 font-sans antialiased selection:bg-indigo-500/30"
        dir="rtl"
        style={{ fontFamily: "'Vazirmatn', sans-serif" }}
      >
        <Navbar />
        {children}
        <Footer />
      </div>
    </ThemeProvider>
  );
}
