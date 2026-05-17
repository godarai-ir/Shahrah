"use client";

import { useState } from 'react';
import { apiPost } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AdminImportPage() {
  const [token, setToken] = useState('');
  const [json, setJson] = useState('{\n  "providers": [],\n  "models": []\n}');

  const submit = async () => {
    const parsed = JSON.parse(json);
    const res = await apiPost('/admin/import/catalog-json', parsed, token);
    alert(JSON.stringify(res));
  };

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Import Catalog JSON</h1>
      <Input placeholder="Admin JWT" value={token} onChange={(e) => setToken(e.target.value)} />
      <textarea className="h-80 w-full rounded-xl border border-zinc-700 bg-zinc-950 p-3 font-mono text-xs" value={json} onChange={(e) => setJson(e.target.value)} />
      <Button onClick={submit}>Import</Button>
    </div>
  );
}
