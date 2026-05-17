"use client";

import { useState } from 'react';
import { apiPost } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

export default function AdminIndexingPage() {
  const [token, setToken] = useState('');
  const [modelId, setModelId] = useState('');
  return (
    <div className="space-y-3">
      <h1 className="text-xl font-semibold">Indexing</h1>
      <Input placeholder="Admin JWT" value={token} onChange={(e) => setToken(e.target.value)} />
      <div className="flex gap-2">
        <Button onClick={async () => { await apiPost('/admin/indexing/rebuild', {}, token); alert('Reindex queued'); }}>Rebuild Index</Button>
      </div>
      <div className="flex gap-2">
        <Input placeholder="Model ID" value={modelId} onChange={(e) => setModelId(e.target.value)} />
        <Button onClick={async () => { await apiPost('/admin/indexing/model', { modelId }, token); alert('Index job queued'); }}>Index One</Button>
      </div>
    </div>
  );
}
