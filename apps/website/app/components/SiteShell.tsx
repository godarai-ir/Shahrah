'use client';

import type { ReactNode } from 'react';
import { Navbar } from './Navbar';
import { ThemeProvider } from './ThemeProvider';
import { Footer } from './Trust';

export function SiteShell({ children }: { children: ReactNode }) {
  return (
    <ThemeProvider defaultTheme="system" storageKey="shahrah-theme">
      <div
        className="min-h-screen bg-site-ground text-site-ink transition-colors duration-300 font-sans antialiased selection:bg-site-panel-deep"
        dir="rtl"
      >
        <Navbar />
        {children}
        <Footer />
      </div>
    </ThemeProvider>
  );
}
