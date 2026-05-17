"use client";

import Link from 'next/link';

export default function AdminHome() {
  const links = [
    ['/admin/providers', 'Providers'],
    ['/admin/models', 'Models'],
    ['/admin/pricing', 'Pricing'],
    ['/admin/capabilities', 'Capabilities'],
    ['/admin/import', 'Import JSON'],
    ['/admin/indexing', 'Indexing'],
    ['/admin/audit-log', 'Audit Log'],
  ];

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Admin</h1>
      <p className="text-zinc-400">Use JWT from <code>/api/v1/auth/login</code> with configured admin credentials.</p>
      <div className="grid gap-3 md:grid-cols-3">
        {links.map(([href, label]) => <Link key={href} href={href} className="rounded-xl border border-zinc-800 bg-zinc-900 p-4 hover:border-cyan-500">{label}</Link>)}
      </div>
    </div>
  );
}
