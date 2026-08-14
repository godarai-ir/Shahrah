'use client';

import { ArrowLeft, Download, Mail, Paperclip } from 'lucide-react';
import Link from 'next/link';
import { use, useCallback, useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import {
  resendReceivingRequest,
  type InboundEmailAttachment,
  type InboundEmailDetail,
} from '@/lib/resend-receiving';

export default function InboundEmailDetailPage({ params }: { params: Promise<{ emailId: string }> }) {
  const { emailId } = use(params);
  const [email, setEmail] = useState<InboundEmailDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');

    try {
      const nextEmail = await resendReceivingRequest<InboundEmailDetail>(`/${encodeURIComponent(emailId)}`);
      setEmail(nextEmail);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Failed to load received email');
    } finally {
      setLoading(false);
    }
  }, [emailId]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, [load]);
  /* eslint-enable react-hooks/set-state-in-effect */

  if (loading) return <p role="status" className="text-sm text-support-muted">Loading received email…</p>;

  if (!email) {
    return (
      <div className="flex flex-col gap-4">
        <Link className="inline-flex items-center gap-2 text-sm text-support-accent hover:underline" href="/admin/inbound-email">
          <ArrowLeft aria-hidden="true" className="size-4" />
          Inbound email
        </Link>
        <div role="alert" className="rounded-md border border-support-danger bg-support-surface p-4 text-sm text-support-danger">
          {error || 'Received email not found.'}
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <Link className="inline-flex items-center gap-2 text-sm text-support-accent hover:underline" href="/admin/inbound-email">
          <ArrowLeft aria-hidden="true" className="size-4" />
          Inbound email
        </Link>
        <div className="mt-3 flex items-start gap-3">
          <Mail aria-hidden="true" className="mt-1 size-6 shrink-0 text-support-accent" />
          <div className="min-w-0">
            <h1 className="break-words text-2xl font-semibold" dir="auto">{email.subject || '(No subject)'}</h1>
            <p className="mt-1 break-all font-mono text-xs text-support-muted">{email.id}</p>
          </div>
        </div>
      </div>

      <Card>
        <dl className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <EmailMeta label="From" value={email.from} autoDirection />
          <EmailMeta label="To" value={email.to.join(', ') || '—'} autoDirection />
          <EmailMeta label="Received" value={emailDate(email.created_at)} />
          <EmailMeta label="Reply to" value={email.reply_to?.join(', ') || '—'} autoDirection />
          <EmailMeta label="CC" value={email.cc?.join(', ') || '—'} autoDirection />
          <EmailMeta label="BCC" value={email.bcc?.join(', ') || '—'} autoDirection />
          <EmailMeta label="Received for" value={email.received_for.join(', ') || '—'} autoDirection />
          <EmailMeta label="Message ID" value={email.message_id} technical />
        </dl>
      </Card>

      <section className="flex flex-col gap-3" aria-labelledby="email-content-heading">
        <h2 id="email-content-heading" className="text-lg font-semibold">Message content</h2>
        {email.html ? (
          <div className="overflow-hidden rounded-xl border border-support-border bg-white">
            <iframe
              className="h-128 w-full bg-white"
              referrerPolicy="no-referrer"
              sandbox=""
              srcDoc={safeHtmlPreview(email.html)}
              title="Received email HTML preview"
            />
          </div>
        ) : email.text ? (
          <Card>
            <pre className="whitespace-pre-wrap break-words font-sans text-sm leading-7" dir="auto">{email.text}</pre>
          </Card>
        ) : (
          <Card className="text-sm text-support-muted">This email has no HTML or plain-text body.</Card>
        )}

        {email.html && email.text ? (
          <details className="rounded-xl border border-support-border bg-support-surface p-4">
            <summary className="cursor-pointer text-sm font-medium">Plain-text version</summary>
            <pre className="mt-4 whitespace-pre-wrap break-words font-sans text-sm leading-7" dir="auto">{email.text}</pre>
          </details>
        ) : null}
      </section>

      <section className="flex flex-col gap-3" aria-labelledby="attachments-heading">
        <div className="flex items-center gap-2">
          <Paperclip aria-hidden="true" className="size-5 text-support-muted" />
          <h2 id="attachments-heading" className="text-lg font-semibold">Attachments ({email.attachments.length})</h2>
        </div>
        {email.attachments.length ? (
          <div className="grid gap-3 md:grid-cols-2">
            {email.attachments.map((attachment) => <AttachmentCard key={attachment.id} attachment={attachment} />)}
          </div>
        ) : <Card className="text-sm text-support-muted">No attachments.</Card>}
      </section>

      <details className="rounded-xl border border-support-border bg-support-surface p-4">
        <summary className="cursor-pointer text-sm font-medium">Email headers</summary>
        <pre className="mt-4 overflow-x-auto whitespace-pre-wrap break-words font-mono text-xs leading-6 text-support-muted">
          {email.headers ? JSON.stringify(email.headers, null, 2) : 'No headers returned.'}
        </pre>
      </details>

      {email.raw ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-support-border bg-support-surface p-4">
          <div>
            <p className="text-sm font-medium">Raw message</p>
            <p className="mt-1 text-xs text-support-muted">Temporary Resend link, expires {emailDate(email.raw.expires_at)}.</p>
          </div>
          <a
            className="inline-flex items-center gap-2 rounded-md bg-support-accent px-3 py-2 text-sm font-medium text-black hover:opacity-90"
            href={email.raw.download_url}
            rel="noreferrer"
            target="_blank"
          >
            <Download aria-hidden="true" className="size-4" />
            Download raw email
          </a>
        </div>
      ) : null}
    </div>
  );
}

function EmailMeta({
  label,
  value,
  autoDirection = false,
  technical = false,
}: {
  label: string;
  value: string;
  autoDirection?: boolean;
  technical?: boolean;
}) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-support-muted">{label}</dt>
      <dd
        className={`mt-1 break-words text-sm ${technical ? 'font-mono' : 'font-medium'}`}
        dir={autoDirection ? 'auto' : undefined}
      >
        {value}
      </dd>
    </div>
  );
}

function AttachmentCard({ attachment }: { attachment: InboundEmailAttachment }) {
  return (
    <Card className="flex items-start gap-3">
      <Paperclip aria-hidden="true" className="mt-0.5 size-5 shrink-0 text-support-muted" />
      <div className="min-w-0">
        <p className="truncate text-sm font-medium" dir="auto" title={attachment.filename ?? undefined}>
          {attachment.filename || 'Unnamed attachment'}
        </p>
        <p className="mt-1 text-xs text-support-muted">
          {attachment.content_type} · {formatBytes(attachment.size)}
        </p>
        <p className="mt-1 font-mono text-xs text-support-muted">{attachment.content_disposition || 'attachment'}</p>
      </div>
    </Card>
  );
}

function safeHtmlPreview(html: string): string {
  const contentSecurityPolicy = [
    "default-src 'none'",
    "img-src data:",
    "style-src 'unsafe-inline'",
    "font-src data:",
    "form-action 'none'",
    "base-uri 'none'",
  ].join('; ');

  return `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="${contentSecurityPolicy}"><meta name="referrer" content="no-referrer"><style>body{margin:16px;color:#111827;font-family:Arial,sans-serif;overflow-wrap:anywhere}img{max-width:100%;height:auto}</style></head><body>${html}</body></html>`;
}

function emailDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat('en-GB', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return 'Unknown size';
  if (bytes < 1024) return `${bytes} B`;

  const units = ['KB', 'MB', 'GB'];
  let value = bytes / 1024;
  let unit = units[0];

  for (let index = 1; index < units.length && value >= 1024; index += 1) {
    value /= 1024;
    unit = units[index];
  }

  return `${value.toFixed(value >= 10 ? 0 : 1)} ${unit}`;
}
