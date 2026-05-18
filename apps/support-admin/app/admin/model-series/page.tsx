"use client";

import { useEffect, useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

type Series = { slug: string; name: string; logoUrl?: string | null };

export default function AdminModelSeriesPage() {
  const [rows, setRows] = useState<Series[]>([]);
  const [slug, setSlug] = useState('');
  const [name, setName] = useState('');
  const [logoSlug, setLogoSlug] = useState('');
  const [file, setFile] = useState<File | null>(null);

  useEffect(() => {
    let cancelled = false;
    void fetch('/api/admin/model-series')
      .then((res) => (res.ok ? res.json() : []))
      .then((data) => {
        if (!cancelled) setRows(data);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const load = async () => {
    const res = await fetch('/api/admin/model-series');
    if (res.ok) setRows(await res.json());
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Model Series</h1>
      <div className="grid gap-2">
        {rows.map((s) => (
          <div key={s.slug} className="rounded border border-zinc-800 bg-zinc-900 p-3">
            <div className="font-medium">{s.name} ({s.slug})</div>
          </div>
        ))}
      </div>

      <form className="grid max-w-xl gap-3 rounded-xl border border-zinc-800 bg-zinc-900 p-4" onSubmit={async (e) => {
        e.preventDefault();
        await fetch('/api/admin/model-series', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ slug, name }) });
        await load();
      }}>
        <Input placeholder="slug" value={slug} onChange={(e) => setSlug(e.target.value)} />
        <Input placeholder="name" value={name} onChange={(e) => setName(e.target.value)} />
        <Button>Save Series</Button>
      </form>

      <form className="grid max-w-xl gap-3 rounded-xl border border-zinc-800 bg-zinc-900 p-4" onSubmit={async (e) => {
        e.preventDefault();
        if (!file) return;
        const fd = new FormData();
        fd.append('slug', logoSlug);
        fd.append('file', file);
        await fetch('/api/admin/model-series/logo', { method: 'POST', body: fd });
        await load();
      }}>
        <Input placeholder="series slug" value={logoSlug} onChange={(e) => setLogoSlug(e.target.value)} />
        <input type="file" accept="image/*" onChange={(e) => setFile(e.target.files?.[0] || null)} />
        <Button>Upload Series Logo</Button>
      </form>
    </div>
  );
}
