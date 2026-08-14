'use client';

import { ArrowLeft, ArrowRight, Mail, Paperclip, RefreshCw, Search } from 'lucide-react';
import Link from 'next/link';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import {
  resendReceivingRequest,
  type InboundEmailList,
  type InboundEmailSummary,
} from '@/lib/resend-receiving';

const pageSize = 25;

export default function InboundEmailPage() {
  const [payload, setPayload] = useState<InboundEmailList>({ object: 'list', has_more: false, data: [] });
  const [cursor, setCursor] = useState('');
  const [cursorHistory, setCursorHistory] = useState<string[]>([]);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(async (after: string) => {
    setLoading(true);
    setError('');

    const params = new URLSearchParams({ limit: String(pageSize) });
    if (after) params.set('after', after);

    try {
      const nextPayload = await resendReceivingRequest<InboundEmailList>(`?${params}`);
      setPayload(nextPayload);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load received emails');
    } finally {
      setLoading(false);
    }
  }, []);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(cursor); }, [cursor, load]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const visibleEmails = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    if (!normalizedQuery) return payload.data;

    return payload.data.filter((email) => [
      email.subject,
      email.from,
      email.to.join(' '),
      email.message_id,
    ].some((value) => value.toLowerCase().includes(normalizedQuery)));
  }, [payload.data, query]);

  function nextPage() {
    const nextCursor = payload.data.at(-1)?.id;
    if (!payload.has_more || !nextCursor) return;

    setCursorHistory((history) => [...history, cursor]);
    setCursor(nextCursor);
  }

  function previousPage() {
    setCursor(cursorHistory.at(-1) ?? '');
    setCursorHistory((history) => history.slice(0, -1));
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-sm text-support-muted">Resend receiving inbox</p>
          <h1 className="text-2xl font-semibold">Inbound email</h1>
          <p className="mt-1 text-sm text-support-muted">
            Messages received for the configured Resend account.
          </p>
        </div>
        <Button type="button" aria-busy={loading} disabled={loading} onClick={() => void load(cursor)}>
          <RefreshCw aria-hidden="true" className={`size-4 ${loading ? 'animate-spin' : ''}`} />
          {loading ? 'Refreshing…' : 'Refresh'}
        </Button>
      </div>

      <Card>
        <label className="relative block">
          <span className="sr-only">Search the current email page</span>
          <Search aria-hidden="true" className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-support-muted" />
          <Input
            className="ps-10"
            placeholder="Search sender, recipient, subject, or message ID"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        <p className="mt-2 text-xs text-support-muted">Search applies to the current page of results.</p>
      </Card>

      {error ? (
        <div role="alert" className="rounded-md border border-support-danger bg-support-surface p-4 text-sm text-support-danger">
          <p className="font-medium">Could not load inbound email.</p>
          <p className="mt-1">{error}</p>
        </div>
      ) : null}

      <div className="overflow-x-auto rounded-xl border border-support-border">
        <table className="w-full min-w-200 text-left text-sm">
          <thead className="bg-support-surface-muted text-support-muted">
            <tr>
              <th className="p-3">Received</th>
              <th className="p-3">From</th>
              <th className="p-3">Subject</th>
              <th className="p-3">To</th>
              <th className="p-3 text-center">Attachments</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-support-border">
            {visibleEmails.map((email) => <InboundEmailRow key={email.id} email={email} />)}
          </tbody>
        </table>
        {!loading && visibleEmails.length === 0 ? (
          <div className="flex flex-col items-center gap-2 p-10 text-center text-sm text-support-muted">
            <Mail aria-hidden="true" className="size-8" />
            <p>{query.trim() ? 'No emails match this search on the current page.' : 'No received emails were found.'}</p>
          </div>
        ) : null}
        {loading && payload.data.length === 0 ? (
          <p role="status" className="p-10 text-center text-sm text-support-muted">Loading received emails…</p>
        ) : null}
      </div>

      <div className="flex items-center justify-between gap-3">
        <Button
          type="button"
          className="border border-support-border bg-support-surface text-white hover:bg-support-surface-muted"
          disabled={loading || cursorHistory.length === 0}
          onClick={previousPage}
        >
          <ArrowLeft aria-hidden="true" className="size-4" />
          Previous
        </Button>
        <span className="text-xs text-support-muted">Up to {pageSize} emails per page</span>
        <Button
          type="button"
          disabled={loading || !payload.has_more || payload.data.length === 0}
          onClick={nextPage}
        >
          Next
          <ArrowRight aria-hidden="true" className="size-4" />
        </Button>
      </div>
    </div>
  );
}

function InboundEmailRow({ email }: { email: InboundEmailSummary }) {
  return (
    <tr className="bg-support-surface hover:bg-support-surface-muted">
      <td className="whitespace-nowrap p-3 text-support-muted">{emailDate(email.created_at)}</td>
      <td className="max-w-64 p-3">
        <span className="block truncate" dir="auto" title={email.from}>{email.from}</span>
      </td>
      <td className="max-w-96 p-3">
        <Link
          className="block truncate font-medium text-support-accent hover:underline"
          href={`/admin/inbound-email/${encodeURIComponent(email.id)}`}
          title={email.subject || '(No subject)'}
        >
          {email.subject || '(No subject)'}
        </Link>
        <span className="mt-1 block truncate font-mono text-xs text-support-muted" title={email.message_id}>
          {email.message_id}
        </span>
      </td>
      <td className="max-w-64 p-3">
        <span className="block truncate" dir="auto" title={email.to.join(', ')}>{email.to.join(', ') || '—'}</span>
      </td>
      <td className="p-3 text-center">
        {email.attachments.length ? (
          <span className="inline-flex items-center gap-1" title={`${email.attachments.length} attachment(s)`}>
            <Paperclip aria-hidden="true" className="size-4 text-support-muted" />
            {email.attachments.length}
          </span>
        ) : '—'}
      </td>
    </tr>
  );
}

function emailDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat('en-GB', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}
