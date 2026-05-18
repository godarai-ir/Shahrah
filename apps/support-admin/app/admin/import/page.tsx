"use client";

import { useState } from 'react';
import { Button } from '@/components/ui/button';

export default function AdminImportPage() {
  const [json, setJson] = useState('{\n  "providers": [],\n  "models": []\n}');

  const submit = async () => {
    const parsed = JSON.parse(json);
    const res = await fetch('/api/admin/import', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(parsed),
    }).then((r) => r.json());
    alert(JSON.stringify(res));
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Import Catalog JSON</h1>
      <textarea className="h-80 w-full rounded-xl border border-zinc-700 bg-zinc-950 p-3 font-mono text-xs" value={json} onChange={(e) => setJson(e.target.value)} />
      <Button onClick={submit}>Import</Button>
    </div>
  );
}
