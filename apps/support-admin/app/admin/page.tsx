"use client";

import Link from 'next/link';

export default function AdminHome() {
  const links = [
    ['/admin/providers', 'Providers'],
    ['/admin/tenants', 'Tenants'],
    ['/admin/models', 'Models'],
    ['/admin/model-series', 'Model Series'],
    ['/admin/pricing', 'Pricing'],
    ['/admin/capabilities', 'Capabilities'],
    ['/admin/import', 'Import JSON'],
    ['/admin/indexing', 'Indexing'],
    ['/admin/audit-log', 'Audit Log'],
    ['/admin/settings/display', 'Display Settings'],
		['/admin/settings/commercial', 'Commercial Signup Settings'],
  ];

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Admin</h1>
      <button
        type="button"
        className="rounded border border-zinc-700 px-3 py-1 text-sm text-zinc-300 hover:bg-zinc-900"
        onClick={async () => {
          await fetch('/api/auth/logout', { method: 'POST' });
          window.location.href = '/login';
        }}
      >
        Logout
      </button>
      <div className="grid gap-3 md:grid-cols-3">
        {links.map(([href, label]) => <Link key={href} href={href} className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">{label}</Link>)}
      </div>
    </div>
  );
}
