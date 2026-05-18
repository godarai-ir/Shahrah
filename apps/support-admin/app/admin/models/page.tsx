"use client";

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';

type ModelRow = { canonicalId: string; displayName: string; enabled: boolean; provider: { slug: string }; modelSeries?: { slug: string; name: string } | null };

export default function AdminModelsPage() {
  const [rows, setRows] = useState<ModelRow[]>([]);

  useEffect(() => {
    let cancelled = false;
    void fetch('/api/admin/models')
      .then((res) => (res.ok ? res.json() : []))
      .then((data) => {
        if (!cancelled) setRows(data);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const load = async () => {
    const res = await fetch('/api/admin/models');
    if (res.ok) setRows(await res.json());
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Admin Models</h1>
      <div className="flex gap-2">
        <Link href="/admin/models/new" className="inline-block rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black">Create / Upsert Model</Link>
        <Link href="/admin/model-series" className="inline-block rounded-md bg-zinc-700 px-3 py-2 text-sm font-medium text-white">Manage Model Series</Link>
      </div>
      <div className="space-y-2">
        {rows.map((m) => (
          <div key={m.canonicalId} className="flex items-center justify-between rounded border border-zinc-800 bg-zinc-900 p-3">
            <div>
              <div className="font-medium">{m.displayName}</div>
              <div className="text-xs text-zinc-500">{m.canonicalId} | series: {m.modelSeries?.name || '-'}</div>
            </div>
            <Button onClick={async () => {
              await fetch('/api/admin/models/enabled', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ canonicalId: m.canonicalId, enabled: !m.enabled }) });
              await load();
            }}>{m.enabled ? 'Disable' : 'Enable'}</Button>
          </div>
        ))}
      </div>
    </div>
  );
}
