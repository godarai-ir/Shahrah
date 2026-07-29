import Link from 'next/link';
import { DisplaySettingsProvider } from '@/components/display-settings-provider';

const links = [
  ['/admin', 'Overview'],
  ['/admin/tenants', 'Tenants'],
  ['/admin/tickets', 'Tickets'],
  ['/admin/providers', 'Providers'],
  ['/admin/provider-connections', 'Provider Credentials'],
  ['/admin/models', 'Models'],
  ['/admin/pricing', 'Pricing'],
	['/admin/settings/commercial', 'Signup Settings'],
  ['/admin/audit-log', 'Audit Log'],
];

export default function AdminLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <DisplaySettingsProvider>
      <header className="sticky top-0 z-20 border-b border-zinc-800 bg-zinc-950/90 backdrop-blur">
        <nav className="mx-auto flex w-full max-w-7xl items-center gap-2 px-4 py-3">
          <span className="me-4 text-sm font-semibold text-white">Support Admin</span>
          {links.map(([href, label]) => (
            <Link key={href} href={href} className="rounded-md px-3 py-1.5 text-sm text-zinc-300 hover:bg-zinc-900 hover:text-white">
              {label}
            </Link>
          ))}
        </nav>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-6">{children}</main>
    </DisplaySettingsProvider>
  );
}
