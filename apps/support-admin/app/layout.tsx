import type { Metadata } from 'next';
import { IBM_Plex_Sans, IBM_Plex_Mono } from 'next/font/google';
import './globals.css';

const sans = IBM_Plex_Sans({ subsets: ['latin'], variable: '--font-sans', weight: ['400','500','600','700'] });
const mono = IBM_Plex_Mono({ subsets: ['latin'], variable: '--font-mono', weight: ['400','500'] });

export const metadata: Metadata = {
  title: 'Shahrah Support Admin',
  description: 'Internal operations and support administration',
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" className={`${sans.variable} ${mono.variable}`}>
      <body className="min-h-screen bg-zinc-950 text-zinc-100">
        {children}
      </body>
    </html>
  );
}
