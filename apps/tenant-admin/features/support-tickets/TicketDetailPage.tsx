"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Download, File, Star } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppTextareaField, CopyableValue, DefinitionList, DefinitionListItem, EmptyState, PageContainer } from "@/components/app-ui";
import { Attachment, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { TicketStatusBadge } from "./TicketStatusBadge";
import { createIdempotencyKey, readTicketResponse, type TicketDetail, type TicketMessage } from "./types";

export function TicketDetailPage({ ticketID, currentUserID }: { ticketID: string; currentUserID: string }) {
  const t = useTranslations("support");
  const locale = useLocale();
  const [detail, setDetail] = useState<TicketDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [body, setBody] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const replyKey = useRef(createIdempotencyKey());

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setDetail(await readTicketResponse<TicketDetail>(await fetch(`/api/support/tickets/${encodeURIComponent(ticketID)}`, { cache: "no-store" })));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t("errors.load"));
    } finally {
      setLoading(false);
    }
  }, [t, ticketID]);

  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => { void load(); }, [load]);
  /* eslint-enable react-hooks/set-state-in-effect */

  async function reply(event: React.FormEvent) {
    event.preventDefault();
    if (!body.trim() && files.length === 0) return setError(t("validation.messageRequired"));
    setSubmitting(true);
    setError("");
    const form = new FormData();
    form.set("payload", JSON.stringify({ body, idempotency_key: replyKey.current }));
    files.forEach((file) => form.append("files", file));
    try {
      const next = await readTicketResponse<TicketDetail>(await fetch(`/api/support/tickets/${encodeURIComponent(ticketID)}/messages`, {
        method: "POST", body: form, headers: { "Idempotency-Key": replyKey.current },
      }));
      setDetail(next);
      setBody("");
      setFiles([]);
      replyKey.current = createIdempotencyKey();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t("errors.reply"));
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) return <PageContainer><p role="status">{t("states.loading")}</p></PageContainer>;
  if (!detail) return <PageContainer><EmptyState title={t("states.notFound")} description={error} primaryAction={<Button asChild variant="outline"><Link href="/support/tickets">{t("actions.backToList")}</Link></Button>} /></PageContainer>;
  const ticket = detail.ticket;
  const formatDate = (value?: string) => value ? new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : t("values.notAvailable");
  const canReply = ticket.status !== "closed";
  const canRate = ticket.requester_user_id === currentUserID && (ticket.status === "resolved" || ticket.status === "closed");

  return (
    <PageContainer size="readable">
      {error ? <p role="alert" className="rounded-xl border border-destructive-soft-border bg-destructive-soft p-4 text-sm text-destructive-soft-foreground">{error}</p> : null}
      <Card>
        <CardHeader className="gap-3">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <CardTitle>{ticket.subject}</CardTitle>
            <TicketStatusBadge status={ticket.status} label={t(`statuses.${ticket.status}`)} />
          </div>
          <CopyableValue value={ticket.public_id} labels={{ valueAriaLabel: t("fields.id"), copy: t("actions.copy"), copying: t("actions.copying"), copied: t("actions.copied") }} />
        </CardHeader>
        <CardContent>
          <DefinitionList columns={2}>
            <DefinitionListItem label={t("fields.category")} value={t(`categories.${ticket.category}`)} />
            <DefinitionListItem label={t("fields.priority")} value={t(`priorities.${ticket.priority}`)} />
            <DefinitionListItem label={t("fields.createdAt")} value={formatDate(ticket.created_at)} />
            <DefinitionListItem label={t("fields.lastActivity")} value={formatDate(ticket.last_message_at)} />
            {ticket.resolved_at ? <DefinitionListItem label={t("fields.resolvedAt")} value={formatDate(ticket.resolved_at)} /> : null}
            {ticket.closed_at ? <DefinitionListItem label={t("fields.closedAt")} value={formatDate(ticket.closed_at)} /> : null}
          </DefinitionList>
        </CardContent>
      </Card>

      <section className="flex flex-col gap-4" aria-labelledby="conversation-title">
        <h2 id="conversation-title" className="text-lg font-semibold">{t("conversation.title")}</h2>
        {detail.messages.map((message) => <MessageCard key={message.id} message={message} currentUserID={currentUserID} formatDate={formatDate} />)}
      </section>

      {canRate ? <RatingForm ticketID={ticket.public_id} initialScore={ticket.rating?.score} initialComment={ticket.rating?.comment} onSaved={(updated) => setDetail((current) => current ? { ...current, ticket: updated.ticket } : current)} /> : null}

      {canReply ? (
        <Card>
          <CardHeader><CardTitle>{ticket.status === "resolved" ? t("reply.reopenTitle") : t("reply.title")}</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={reply} className="flex flex-col gap-4">
              <AppTextareaField label={t("fields.message")} value={body} onChange={(event) => setBody(event.target.value)} maxLength={10000} rows={6} disabled={submitting} />
              <Input type="file" multiple disabled={submitting} accept=".pdf,.png,.jpg,.jpeg,.webp,.txt,.csv,.log,.json" aria-label={t("attachments.add")} onChange={(event) => setFiles(Array.from(event.target.files ?? []).slice(0, 5))} />
              <Button type="submit" className="self-end" disabled={submitting}>{submitting ? t("actions.submitting") : ticket.status === "resolved" ? t("actions.sendAndReopen") : t("actions.send")}</Button>
            </form>
          </CardContent>
        </Card>
      ) : <p className="rounded-xl border border-border bg-muted p-4 text-sm text-muted-foreground">{t("reply.closed")}</p>}
    </PageContainer>
  );
}

function MessageCard({ message, currentUserID, formatDate }: { message: TicketMessage; currentUserID: string; formatDate: (value?: string) => string }) {
  const t = useTranslations("support");
  const support = message.author_type === "support_agent";
  const author = support ? t("conversation.support") : message.author_id === currentUserID ? t("conversation.you") : t("conversation.organization");
  return (
    <article className={`flex ${support ? "justify-start" : "justify-end"}`}>
      <div className="w-full max-w-3xl rounded-xl border border-border bg-card p-4 shadow-xs">
        <header className="mb-3 flex flex-wrap items-center justify-between gap-2 text-sm">
          <strong>{author}</strong><time className="text-muted-foreground" dateTime={message.created_at}>{formatDate(message.created_at)}</time>
        </header>
        {message.body ? <p className="whitespace-pre-wrap break-words text-sm leading-7">{message.body}</p> : null}
        {message.attachments.length ? <div className="mt-4 grid gap-2">{message.attachments.map((attachment) => (
          <Attachment key={attachment.id} size="sm" state={attachment.scan_status === "clean" ? "done" : attachment.scan_status === "rejected" ? "error" : "processing"}>
            <AttachmentMedia><File aria-hidden="true" /></AttachmentMedia>
            <AttachmentContent>
              <AttachmentTitle dir="auto">{attachment.original_filename}</AttachmentTitle>
              <AttachmentDescription>{t(`attachments.scan.${attachment.scan_status}`)} · <span dir="ltr">{formatBytes(attachment.size_bytes)}</span></AttachmentDescription>
            </AttachmentContent>
            {attachment.scan_status === "clean" ? <AttachmentActions><Button asChild size="icon-sm" variant="outline"><a href={`/api/support/ticket-attachments/${encodeURIComponent(attachment.id)}/download`} aria-label={t("attachments.download", { name: attachment.original_filename })}><Download aria-hidden="true" /></a></Button></AttachmentActions> : null}
          </Attachment>
        ))}</div> : null}
      </div>
    </article>
  );
}

function RatingForm({ ticketID, initialScore, initialComment, onSaved }: { ticketID: string; initialScore?: number; initialComment?: string; onSaved: (detail: TicketDetail) => void }) {
  const t = useTranslations("support");
  const [score, setScore] = useState(initialScore ?? 0);
  const [comment, setComment] = useState(initialComment ?? "");
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (score < 1 || score > 5) return setMessage(t("validation.rating"));
    setSaving(true);
    setMessage("");
    try {
      const response = await fetch(`/api/support/tickets/${encodeURIComponent(ticketID)}/rating`, {
        method: "PUT",
        headers: { "content-type": "application/json", "Idempotency-Key": createIdempotencyKey() },
        body: JSON.stringify({ score, comment }),
      });
      const ticket = await readTicketResponse<TicketDetail["ticket"]>(response);
      onSaved({ ticket, messages: [] });
      setMessage(t("rating.saved"));
    } catch (cause) {
      setMessage(cause instanceof Error ? cause.message : t("errors.rating"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card>
      <CardHeader><CardTitle>{t("rating.title")}</CardTitle></CardHeader>
      <CardContent>
        <form onSubmit={save} className="flex flex-col gap-4">
          <fieldset>
            <legend className="mb-2 text-sm font-medium">{t("rating.scoreLabel")}</legend>
            <div className="flex gap-2" dir="ltr">
              {[1, 2, 3, 4, 5].map((value) => (
                <Button key={value} type="button" variant={score === value ? "default" : "outline"} size="icon" aria-label={t("rating.scoreValue", { score: value })} aria-pressed={score === value} onClick={() => setScore(value)}>
                  <span className="sr-only">{value}</span><Star className={score >= value ? "fill-current" : ""} aria-hidden="true" />
                </Button>
              ))}
            </div>
          </fieldset>
          <AppTextareaField label={t("rating.comment")} value={comment} onChange={(event) => setComment(event.target.value)} maxLength={500} rows={3} />
          {message ? <p role="status" className="text-sm text-muted-foreground">{message}</p> : null}
          <Button type="submit" className="self-end" disabled={saving}>{saving ? t("actions.submitting") : t("rating.submit")}</Button>
        </form>
      </CardContent>
    </Card>
  );
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`;
  return `${(value / 1024 / 1024).toFixed(1)} MiB`;
}
