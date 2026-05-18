"use client";

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

export default function AdminIndexingPage() {
  const [modelId, setModelId] = useState('');
  return (
    <div className="space-y-3">
      <h1 className="text-xl font-semibold">Indexing</h1>
      <div className="flex gap-2">
        <Button onClick={async () => { await fetch('/api/admin/indexing/rebuild', { method: 'POST' }); alert('Reindex queued'); }}>Rebuild Index</Button>
      </div>
      <div className="flex gap-2">
        <Input placeholder="Model ID" value={modelId} onChange={(e) => setModelId(e.target.value)} />
        <Button onClick={async () => { await fetch('/api/admin/indexing/model', { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ modelId }) }); alert('Index job queued'); }}>Index One</Button>
      </div>
    </div>
  );
}
