'use client';

import Link from 'next/link';
import { use, useCallback, useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  idempotencyKey,
  supportTicketRequest,
  ticketDate,
  type SupportMessage,
  type SupportTicket,
  type SupportTicketDetail,
  type SupportTicketEvent,
  type SupportTicketStatus,
} from '@/lib/support-tickets';

export default function TicketDetailPage({ params }: { params: Promise<{ ticketId: string }> }) {
  const { ticketId } = use(params);
  const [detail, setDetail] = useState<SupportTicketDetail | null>(null);
  const [events, setEvents] = useState<SupportTicketEvent[]>([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [body, setBody] = useState('');
  const [files, setFiles] = useState<File[]>([]);
  const [nextStatus, setNextStatus] = useState<SupportTicketStatus>('in_progress');
  const [submitting, setSubmitting] = useState(false);
  const replyKey = useRef(idempotencyKey());

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [nextDetail, eventPayload] = await Promise.all([
        supportTicketRequest<SupportTicketDetail>(`/${encodeURIComponent(ticketId)}`),
        supportTicketRequest<{ data: SupportTicketEvent[] }>(`/${encodeURIComponent(ticketId)}/events?limit=100`),
      ]);
      setDetail(nextDetail);
      setEvents(eventPayload.data);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load ticket');
    } finally {
      setLoading(false);
    }
  }, [ticketId]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, [load]);
  /* eslint-enable react-hooks/set-state-in-effect */

  async function sendReply(event: React.FormEvent) {
    event.preventDefault();
    if (!body.trim() && files.length === 0) return setError('A message or attachment is required.');
    if ((nextStatus === 'resolved' || nextStatus === 'closed') && !window.confirm(`Send this reply and mark the ticket ${nextStatus}?`)) return;
    setSubmitting(true);
    setError('');
    const form = new FormData();
    form.set('payload', JSON.stringify({ body, next_status: nextStatus, idempotency_key: replyKey.current }));
    files.forEach((file) => form.append('files', file));
    try {
      const next = await supportTicketRequest<SupportTicketDetail>(`/${encodeURIComponent(ticketId)}/messages`, {
        method: 'POST', body: form, headers: { 'idempotency-key': replyKey.current },
      });
      setDetail(next);
      setBody('');
      setFiles([]);
      replyKey.current = idempotencyKey();
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Reply failed');
    } finally {
      setSubmitting(false);
    }
  }

  async function triage(input: { status?: SupportTicketStatus; priority?: string; assignee_id?: string }) {
    if (!detail) return;
    if ((input.status === 'resolved' || input.status === 'closed') && !window.confirm(`Change status to ${input.status}?`)) return;
    setSubmitting(true);
    setError('');
    try {
      const ticket = await supportTicketRequest<SupportTicket>(`/${encodeURIComponent(ticketId)}/triage`, {
        method: 'PATCH',
        headers: { 'content-type': 'application/json', 'idempotency-key': idempotencyKey() },
        body: JSON.stringify({ ...input, version: detail.ticket.version }),
      });
      setDetail({ ...detail, ticket });
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Triage failed. Refresh and retry.');
    } finally {
      setSubmitting(false);
    }
  }

  async function reopen() {
    if (!detail) return;
    const reason = window.prompt('Reason for reopening this closed ticket:')?.trim();
    if (!reason) return;
    setSubmitting(true);
    try {
      await supportTicketRequest<SupportTicket>(`/${encodeURIComponent(ticketId)}/reopen`, {
        method: 'POST',
        headers: { 'content-type': 'application/json', 'idempotency-key': idempotencyKey() },
        body: JSON.stringify({ reason, version: detail.ticket.version }),
      });
      await load();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Reopen failed');
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) return <p role="status">Loading ticket…</p>;
  if (!detail) return <div className="space-y-4"><p role="alert">{error || 'Ticket not found.'}</p><Link href="/admin/tickets">Back to queue</Link></div>;
  const ticket = detail.ticket;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div><Link className="text-sm text-support-accent hover:underline" href="/admin/tickets">← Ticket queue</Link>
          <h1 className="mt-2 text-2xl font-semibold">{ticket.subject}</h1>
          <p className="font-mono text-sm text-support-muted">{ticket.public_id}</p>
        </div>
        {ticket.status === 'closed' ? <Button className="border border-support-border bg-support-surface text-white hover:bg-support-surface-muted" disabled={submitting} onClick={() => void reopen()}>Reopen with reason</Button> : null}
      </div>

      {error ? <p role="alert" className="rounded-md border border-support-danger bg-support-surface p-3 text-sm text-support-danger">{error}</p> : null}

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex min-w-0 flex-col gap-4">
          <Card>
            <dl className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <Meta label="Tenant" value={ticket.tenant_slug || ticket.tenant_id} />
              <Meta label="Requester" value={ticket.requester_display_name || ticket.requester_email || ticket.requester_user_id} />
              <Meta label="Category" value={ticket.category} />
              <Meta label="Created" value={ticketDate(ticket.created_at)} />
              <Meta label="First response" value={ticketDate(ticket.first_support_response_at)} />
              <Meta label="Last activity" value={ticketDate(ticket.last_message_at)} />
            </dl>
          </Card>

          <section className="space-y-3" aria-labelledby="conversation-heading">
            <h2 id="conversation-heading" className="text-lg font-semibold">Public conversation</h2>
            {detail.messages.map((message) => <Message key={message.id} message={message} />)}
          </section>

          {ticket.status !== 'closed' ? (
            <Card>
              <form className="space-y-4" onSubmit={sendReply}>
                <label className="block space-y-2"><span className="text-sm font-medium">Reply</span>
                  <textarea className="min-h-36 w-full rounded-md border border-support-border bg-support-surface p-3 text-sm" maxLength={10000} value={body} onChange={(event) => setBody(event.target.value)} disabled={submitting} />
                </label>
                <Input type="file" multiple accept=".pdf,.png,.jpg,.jpeg,.webp,.txt,.csv,.log,.json" aria-label="Reply attachments" onChange={(event) => setFiles(Array.from(event.target.files ?? []).slice(0, 5))} />
                <label className="block space-y-2"><span className="text-sm font-medium">Status after reply</span>
                  <select className="w-full rounded-md border border-support-border bg-support-surface px-3 py-2 text-sm" value={nextStatus} onChange={(event) => setNextStatus(event.target.value as SupportTicketStatus)}>
                    <option value="in_progress">In progress</option><option value="waiting_for_customer">Waiting for customer</option><option value="resolved">Resolved</option>
                  </select>
                </label>
                <Button type="submit" disabled={submitting}>{submitting ? 'Sending…' : 'Send reply'}</Button>
              </form>
            </Card>
          ) : null}
        </div>

        <aside className="flex flex-col gap-4">
          <TriageCard ticket={ticket} disabled={submitting} onTriage={triage} />
          <Card>
            <h2 className="mb-3 font-semibold">Satisfaction</h2>
            {ticket.rating ? <div className="space-y-2"><p className="text-2xl font-semibold">{ticket.rating.score} / 5</p><p className="whitespace-pre-wrap text-sm text-support-muted">{ticket.rating.comment || 'No comment'}</p><p className="text-xs text-support-muted">Updated {ticketDate(ticket.rating.updated_at)}</p></div> : <p className="text-sm text-support-muted">No rating submitted.</p>}
          </Card>
          <Card>
            <h2 className="mb-3 font-semibold">Event history</h2>
            <ol className="space-y-3">
              {events.map((event) => <li key={event.id} className="border-s border-support-border ps-3 text-sm"><p className="font-medium">{event.type.replaceAll('_', ' ')}</p><p className="text-xs text-support-muted">{event.actor_role || event.actor_type} · {ticketDate(event.created_at)}</p>{event.request_id ? <p className="truncate font-mono text-xs text-support-muted">{event.request_id}</p> : null}</li>)}
            </ol>
          </Card>
        </aside>
      </div>
    </div>
  );
}

function TriageCard({ ticket, disabled, onTriage }: { ticket: SupportTicket; disabled: boolean; onTriage: (input: { status?: SupportTicketStatus; priority?: string; assignee_id?: string }) => Promise<void> }) {
  const [status, setStatus] = useState(ticket.status);
  const [priority, setPriority] = useState(ticket.priority);
  const [assigneeAction, setAssigneeAction] = useState('keep');
  return (
    <Card>
      <h2 className="mb-4 font-semibold">Triage</h2>
      <div className="space-y-4">
        <label className="block space-y-2"><span className="text-sm">Status</span><select className="w-full rounded-md border border-support-border bg-support-surface px-3 py-2 text-sm" value={status} onChange={(event) => setStatus(event.target.value as SupportTicketStatus)}>
          <option value="open">Open</option><option value="in_progress">In progress</option><option value="waiting_for_customer">Waiting for customer</option><option value="resolved">Resolved</option><option value="closed">Closed</option>
        </select></label>
        <label className="block space-y-2"><span className="text-sm">Priority</span><select className="w-full rounded-md border border-support-border bg-support-surface px-3 py-2 text-sm" value={priority} onChange={(event) => setPriority(event.target.value)}>
          <option value="low">Low</option><option value="normal">Normal</option><option value="high">High</option><option value="urgent">Urgent</option>
        </select></label>
        <label className="block space-y-2"><span className="text-sm">Assignee</span><select className="w-full rounded-md border border-support-border bg-support-surface px-3 py-2 text-sm" value={assigneeAction} onChange={(event) => setAssigneeAction(event.target.value)}>
          <option value="keep">Keep current ({ticket.assignee_id || 'unassigned'})</option><option value="self">Assign to me</option><option value="unassign">Unassign</option>
        </select></label>
        <Button disabled={disabled} onClick={() => void onTriage({
          status,
          priority,
          ...(assigneeAction === 'self' ? { assignee_id: 'self' } : assigneeAction === 'unassign' ? { assignee_id: '' } : {}),
        })}>Save triage</Button>
        <p className="text-xs text-support-muted">Version {ticket.version}. A stale update is rejected by the server.</p>
      </div>
    </Card>
  );
}

function Message({ message }: { message: SupportMessage }) {
  const support = message.author_type === 'support_agent';
  return (
    <article className={`rounded-xl border p-4 ${support ? 'border-support-accent bg-support-surface-muted' : 'border-support-border bg-support-surface'}`}>
      <header className="mb-3 flex flex-wrap justify-between gap-2 text-sm"><strong>{support ? 'Support agent' : 'Tenant user'}</strong><time className="text-support-muted">{ticketDate(message.created_at)}</time></header>
      {message.body ? <p className="whitespace-pre-wrap break-words text-sm leading-7">{message.body}</p> : null}
      {message.attachments.length ? <ul className="mt-4 space-y-2">{message.attachments.map((attachment) => <li key={attachment.id} className="flex items-center justify-between gap-3 rounded-md border border-support-border p-3 text-sm"><span className="min-w-0 truncate" dir="auto">{attachment.original_filename}</span><span className="text-xs text-support-muted">{attachment.scan_status}</span>{attachment.scan_status === 'clean' ? <a className="text-support-accent hover:underline" href={`/api/admin/control-plane/support/ticket-attachments/${encodeURIComponent(attachment.id)}/download`}>Download</a> : null}</li>)}</ul> : null}
    </article>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return <div><dt className="text-xs text-support-muted">{label}</dt><dd className="mt-1 truncate text-sm font-medium">{value}</dd></div>;
}
