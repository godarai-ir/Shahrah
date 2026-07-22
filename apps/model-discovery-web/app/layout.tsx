import type { Metadata } from 'next';
import { IBM_Plex_Sans, IBM_Plex_Mono } from 'next/font/google';
import './globals.css';
import Link from 'next/link';
import { QueryProvider } from '@/lib/query-provider';
import { DisplaySettingsProvider } from '@/components/display-settings-provider';
import { DisplaySettingsDialog } from '@/components/display-settings-dialog';
import { CompareProvider } from '@/components/compare/compare-provider';
import { CompareTray } from '@/components/compare/compare-tray';
import { SiteFooter } from '@/components/site-footer';

const sans = IBM_Plex_Sans({ subsets: ['latin'], variable: '--font-sans', weight: ['400','500','600','700'] });
const mono = IBM_Plex_Mono({ subsets: ['latin'], variable: '--font-mono', weight: ['400','500'] });

export const metadata: Metadata = {
  title: 'GodarAI Model Catalog',
  description: 'LLM model discovery and catalog management',
};

const links = [
  ['/models', 'Models'],
  ['/providers', 'Providers'],
  ['/compare', 'Compare'],
];

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={`${sans.variable} ${mono.variable}`}>
      <body className="flex min-h-screen flex-col bg-zinc-950 text-zinc-100">
        <QueryProvider>
          <DisplaySettingsProvider>
            <CompareProvider>
              <header className="sticky top-0 z-20 border-b border-zinc-800 bg-zinc-950/90 backdrop-blur">
                <nav className="mx-auto flex w-full max-w-7xl items-center gap-2 px-4 py-3">
                  <div className="flex items-center gap-2">
                    {links.map(([href, label]) => (
                      <Link key={href} href={href} className="rounded-md px-3 py-1.5 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-white">
                        {label}
                      </Link>
                    ))}
                  </div>
                  <div className="ms-auto">
                    <DisplaySettingsDialog />
                  </div>
                </nav>
              </header>
              <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-6 pb-28">{children}</main>
              <SiteFooter />
              <CompareTray />
            </CompareProvider>
          </DisplaySettingsProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
