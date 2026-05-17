import type { Metadata } from 'next';
import { IBM_Plex_Sans, IBM_Plex_Mono } from 'next/font/google';
import './globals.css';
import Link from 'next/link';
import { QueryProvider } from '@/lib/query-provider';

const sans = IBM_Plex_Sans({ subsets: ['latin'], variable: '--font-sans', weight: ['400','500','600','700'] });
const mono = IBM_Plex_Mono({ subsets: ['latin'], variable: '--font-mono', weight: ['400','500'] });

export const metadata: Metadata = {
  title: 'Shahrah Model Catalog',
  description: 'LLM model discovery and catalog management',
};

const links = [
  ['/', 'Home'],
  ['/models', 'Models'],
  ['/providers', 'Providers'],
  ['/compare', 'Compare'],
  ['/admin', 'Admin'],
];

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={`${sans.variable} ${mono.variable}`}>
      <body className="min-h-screen bg-zinc-950 text-zinc-100">
        <QueryProvider>
          <header className="sticky top-0 z-20 border-b border-zinc-800 bg-zinc-950/90 backdrop-blur">
            <nav className="mx-auto flex max-w-7xl gap-2 px-4 py-3">
              {links.map(([href, label]) => (
                <Link key={href} href={href} className="rounded-md px-3 py-1.5 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-white">
                  {label}
                </Link>
              ))}
            </nav>
          </header>
          <main className="mx-auto max-w-7xl px-4 py-6">{children}</main>
        </QueryProvider>
      </body>
    </html>
  );
}
