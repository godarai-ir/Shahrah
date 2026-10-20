'use client';

import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import { controlPlaneRequest, formatDate } from '@/lib/support-tenants';

type SeatTypeCount = { type: string; purchased: number; assigned: number; available: number; renewing: number };

type SeatSummary = {
  enforced: boolean;
  subscription_status?: string;
  period_end?: string | null;
  free_limit: number | null;
  free_used: number;
  unassigned: number;
  pending_requests: number;
  paid: SeatTypeCount[];
};

type SeatMember = {
  user_id: string;
  email?: string;
  name?: string;
  seat: string;
  pending_seat?: string;
  pending_effective_at?: string | null;
  personal_remaining_credits: string;
  personal_granted_credits: string;
  pool_used_credits: string;
};

/** One tenant's member seats, read-only: paid seats held, who holds which seat, and personal credits. */
export function TenantSeatsPanel({ tenantId }: { tenantId: string }) {
  const [summary, setSummary] = useState<SeatSummary>();
  const [members, setMembers] = useState<SeatMember[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    controlPlaneRequest<{ summary: SeatSummary; members: { items: SeatMember[] } }>(`tenants/${tenantId}/seats`)
      .then((payload) => { setSummary(payload.summary); setMembers(payload.members?.items ?? []); })
      .catch((cause: unknown) => setError(cause instanceof Error ? cause.message : 'Failed to load seats'));
  }, [tenantId]);

  if (error) return <p role="alert" className="text-sm text-red-300">{error}</p>;
  if (!summary) return <p className="text-sm text-zinc-400">Loading seats…</p>;

  const enforcement = summary.enforced
    ? 'Enforced'
    : `Not enforced (${summary.subscription_status || 'no subscription'})`;

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <Card><p className="text-sm text-zinc-400">Enforcement</p><p className="mt-2 font-semibold">{enforcement}</p></Card>
        <Card><p className="text-sm text-zinc-400">Free seats</p><p className="mt-2 text-xl font-semibold">{summary.free_used} / {summary.free_limit ?? '∞'}</p><p className="mt-1 text-xs text-zinc-500">{summary.unassigned} members without a seat</p></Card>
        {summary.paid.map((item) => (
          <Card key={item.type}><p className="text-sm capitalize text-zinc-400">{item.type} seats</p><p className="mt-2 text-xl font-semibold">{item.assigned} / {item.purchased}</p><p className="mt-1 text-xs text-zinc-500">{item.available} available · {item.renewing} renewing · paid to {formatDate(summary.period_end ?? undefined)}</p></Card>
        ))}
      </div>
      <Card>
        <h2 className="text-lg font-semibold">Members</h2>
        <p className="mt-1 text-sm text-zinc-400">{summary.pending_requests} open upgrade requests. Tenants change seats from their own admin panel.</p>
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="text-zinc-400"><tr><th className="py-2">Member</th><th>Seat</th><th>Personal credits</th><th>Pool used (month)</th></tr></thead>
            <tbody className="divide-y divide-zinc-800">
              {members.map((member) => (
                <tr key={member.user_id}>
                  <td className="py-3">{member.name || member.email || member.user_id}</td>
                  <td className="capitalize">{member.seat}{member.pending_seat ? <span className="text-zinc-500"> → {member.pending_seat} ({formatDate(member.pending_effective_at ?? undefined)})</span> : null}</td>
                  <td>{member.personal_remaining_credits} / {member.personal_granted_credits}</td>
                  <td>{member.pool_used_credits}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {members.length === 0 ? <p className="py-6 text-sm text-zinc-400">No members.</p> : null}
        </div>
      </Card>
    </div>
  );
}
