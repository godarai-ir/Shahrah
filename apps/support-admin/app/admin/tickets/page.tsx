'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  supportTicketRequest,
  ticketDate,
  type SupportTicket,
  type SupportTicketCounts,
  type SupportTicketList,
} from '@/lib/support-tickets';

const queues = [
  ['needs_support_reply', 'Needs reply'],
  ['open', 'Open'],
  ['in_progress', 'In progress'],
  ['waiting_for_customer', 'Waiting for customer'],
  ['resolved', 'Resolved'],
  ['closed', 'Closed'],
  ['all', 'All'],
] as const;

const emptyCounts: SupportTicketCounts = {
  needs_support_reply: 0, open: 0, in_progress: 0, waiting_for_customer: 0, resolved: 0, closed: 0, all: 0,
};

export default function TicketsPage() {
  const [queue, setQueue] = useState<(typeof queues)[number][0]>('needs_support_reply');
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('');
  const [priority, setPriority] = useState('');
  const [tenantID, setTenantID] = useState('');
  const [assigneeID, setAssigneeID] = useState('');
  const [updatedFrom, setUpdatedFrom] = useState('');
  const [updatedTo, setUpdatedTo] = useState('');
  const [counts, setCounts] = useState(emptyCounts);
  const [payload, setPayload] = useState<SupportTicketList>({ data: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  async function load(cursor = '') {
    setLoading(true);
    setError('');
    const params = new URLSearchParams({ limit: '25' });
    if (queue === 'needs_support_reply') params.set('queue', queue);
    else if (queue !== 'all') params.set('status', queue);
    if (query.trim()) params.set('query', query.trim());
    if (category) params.set('category', category);
    if (priority) params.set('priority', priority);
    if (tenantID.trim()) params.set('tenant_id', tenantID.trim());
    if (assigneeID.trim()) params.set('assignee_id', assigneeID.trim());
    if (updatedFrom) params.set('updated_from', new Date(updatedFrom).toISOString());
    if (updatedTo) params.set('updated_to', new Date(updatedTo).toISOString());
    if (cursor) params.set('cursor', cursor);

    const countParams = new URLSearchParams(params);
    countParams.delete('queue');
    countParams.delete('status');
    countParams.delete('cursor');
    countParams.delete('limit');
    try {
      const [list, nextCounts] = await Promise.all([
        supportTicketRequest<SupportTicketList>(`?${params}`),
        supportTicketRequest<SupportTicketCounts>(`/counts?${countParams}`),
      ]);
      setPayload(list);
      setCounts(nextCounts);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load ticket queue');
    } finally {
      setLoading(false);
    }
  }

  /* eslint-disable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */
  useEffect(() => { void load(); }, [queue]);
  /* eslint-enable react-hooks/set-state-in-effect, react-hooks/exhaustive-deps */

  return (
    <div className="flex flex-col gap-6">
      <div>
        <p className="text-sm text-support-muted">Customer support operations</p>
        <h1 className="text-2xl font-semibold">Support tickets</h1>
      </div>

      <div className="flex flex-wrap gap-2" role="tablist" aria-label="Ticket queues">
        {queues.map(([value, label]) => (
          <Button key={value} type="button" role="tab" aria-selected={queue === value} className={queue === value ? '' : 'border border-support-border bg-support-surface text-white hover:bg-support-surface-muted'} onClick={() => setQueue(value)}>
            {label} <span className="ms-2 rounded-full bg-support-surface-muted px-2 py-0.5 text-xs">{counts[value]}</span>
          </Button>
        ))}
      </div>

      <Card>
        <form className="grid gap-3 lg:grid-cols-4" onSubmit={(event) => { event.preventDefault(); void load(); }}>
          <Input aria-label="Search tickets" placeholder="Ticket ID or subject" value={query} onChange={(event) => setQuery(event.target.value)} />
          <FilterSelect label="Category" value={category} onChange={setCategory} options={['technical', 'billing', 'account', 'feature_request', 'other']} />
          <FilterSelect label="Priority" value={priority} onChange={setPriority} options={['low', 'normal', 'high', 'urgent']} />
          <Input aria-label="Tenant ID" placeholder="Tenant ID" value={tenantID} onChange={(event) => setTenantID(event.target.value)} />
          <Input aria-label="Assignee ID" placeholder="Assignee ID" value={assigneeID} onChange={(event) => setAssigneeID(event.target.value)} />
          <Input type="datetime-local" aria-label="Last activity from" value={updatedFrom} onChange={(event) => setUpdatedFrom(event.target.value)} />
          <Input type="datetime-local" aria-label="Last activity to" value={updatedTo} onChange={(event) => setUpdatedTo(event.target.value)} />
          <Button type="submit" className="lg:col-start-4" disabled={loading}>{loading ? 'Loading…' : 'Apply'}</Button>
        </form>
      </Card>

      {error ? <p role="alert" className="rounded-md border border-support-danger bg-support-surface p-3 text-sm text-support-danger">{error}</p> : null}
      <div className="flex justify-end text-sm text-support-muted">
        Filtered rating average: {counts.average_rating == null ? '—' : `${counts.average_rating.toFixed(1)} / 5`}
      </div>
      <div className="overflow-x-auto rounded-xl border border-support-border">
        <table className="w-full text-left text-sm">
          <thead className="bg-support-surface-muted text-support-muted"><tr>
            <th className="p-3">ID</th><th className="p-3">Tenant</th><th className="p-3">Subject</th>
            <th className="p-3">Status</th><th className="p-3">Priority</th><th className="p-3">Assignee</th>
            <th className="p-3">Requester</th><th className="p-3">Waiting</th><th className="p-3">Last activity</th><th className="p-3">Rating</th>
          </tr></thead>
          <tbody className="divide-y divide-support-border">
            {payload.data.map((ticket) => <TicketRow key={ticket.id} ticket={ticket} />)}
          </tbody>
        </table>
        {!loading && payload.data.length === 0 ? <p className="p-10 text-center text-sm text-support-muted">No tickets match this queue and filter set.</p> : null}
      </div>
      {payload.next_cursor ? <Button className="self-end" onClick={() => void load(payload.next_cursor)}>Next page</Button> : null}
    </div>
  );
}

function TicketRow({ ticket }: { ticket: SupportTicket }) {
  const waitingSince = ticket.last_message_author_type === 'tenant_user' && ticket.status === 'open'
    ? elapsed(ticket.last_message_at)
    : '—';
  return (
    <tr className="bg-support-surface hover:bg-support-surface-muted">
      <td className="p-3"><Link className="font-mono text-support-accent hover:underline" href={`/admin/tickets/${ticket.public_id}`}>{ticket.public_id}</Link></td>
      <td className="p-3">{ticket.tenant_slug || ticket.tenant_id}</td>
      <td className="max-w-80 p-3"><span className="font-medium">{ticket.subject}</span><div className="text-xs text-support-muted">{ticket.category}</div></td>
      <td className="p-3">{ticket.status.replaceAll('_', ' ')}</td><td className="p-3">{ticket.priority}</td>
      <td className="p-3">{ticket.assignee_id || 'Unassigned'}</td>
      <td className="p-3">{ticket.requester_display_name || ticket.requester_email || ticket.requester_user_id}</td>
      <td className="whitespace-nowrap p-3">{waitingSince}</td><td className="whitespace-nowrap p-3">{ticketDate(ticket.last_message_at)}</td>
      <td className="p-3">{ticket.rating ? `${ticket.rating.score} / 5` : '—'}</td>
    </tr>
  );
}

function FilterSelect({ label, value, onChange, options }: { label: string; value: string; onChange: (value: string) => void; options: string[] }) {
  return (
    <select aria-label={label} className="rounded-md border border-support-border bg-support-surface px-3 py-2 text-sm" value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="">All {label.toLowerCase()}s</option>
      {options.map((option) => <option key={option} value={option}>{option.replaceAll('_', ' ')}</option>)}
    </select>
  );
}

function elapsed(value: string): string {
  const minutes = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 60000));
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  return hours < 24 ? `${hours}h` : `${Math.floor(hours / 24)}d`;
}
