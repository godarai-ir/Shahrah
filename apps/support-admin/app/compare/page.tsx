"use client";

import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/lib/api';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

type CompareRow = { canonicalId: string; provider?: { name: string }; status?: string; contextWindow?: number; maxOutputTokens?: number };

export default function ComparePage() {
  const [ids, setIds] = useState('openai/model-1,anthropic/model-1');
  const modelIds = useMemo(() => ids.split(',').map((s) => s.trim()).filter(Boolean), [ids]);

  const q = useQuery({ queryKey: ['compare', ids], queryFn: () => apiGet<CompareRow[]>(`/models/compare?ids=${encodeURIComponent(ids)}`) });

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-semibold">Compare Models</h1>
      <div className="flex gap-2">
        <Input value={ids} onChange={(e) => setIds(e.target.value)} placeholder="openai/model-1,anthropic/model-1" />
        <Button type="button" onClick={() => q.refetch()}>Compare</Button>
      </div>
      <p className="text-xs text-zinc-500">2 to 5 model canonical IDs, comma-separated.</p>

      <div className="overflow-x-auto rounded-xl border border-zinc-800">
        <table className="min-w-full text-sm">
          <thead className="bg-zinc-900">
            <tr>
              <th className="p-2 text-left">Field</th>
              {modelIds.map((id) => <th key={id} className="p-2 text-left">{id}</th>)}
            </tr>
          </thead>
          <tbody>
            {['provider','status','contextWindow','maxOutputTokens'].map((field) => (
              <tr key={field} className="border-t border-zinc-800">
                <td className="p-2 font-medium">{field}</td>
                {modelIds.map((id) => {
                  const row = q.data?.find((x) => x.canonicalId === id);
                  const val = field === 'provider' ? row?.provider?.name : row?.[field as keyof CompareRow];
                  return <td key={`${field}-${id}`} className="p-2">{String(val ?? '-')}</td>;
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
