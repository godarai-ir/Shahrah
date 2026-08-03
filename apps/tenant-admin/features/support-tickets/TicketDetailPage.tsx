"use client";

import { useCallback, useEffect, useRef, useState, type ComponentProps } from "react";
import Link from "next/link";
import { CalendarPlus, Clock3, Download, File, Flag, Headphones, LifeBuoy, LockKeyhole, MessagesSquare, Send, Star, Tag, UserRound } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppTextareaField, DateTimeValue, DefinitionList, DefinitionListItem, EmptyState, EntityIdentityHeader, PageContainer } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Attachment, AttachmentActions, AttachmentContent, AttachmentDescription, AttachmentMedia, AttachmentTitle } from "@/components/ui/attachment";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Message, MessageContent, MessageHeader } from "@/components/ui/message";
import { cn } from "@/lib/utils";
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
      {error ? <Alert variant="destructive"><AlertTitle>{t("states.errorTitle")}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
      <Card className="overflow-hidden">
        <CardHeader className="gap-3">
          <EntityIdentityHeader
            title={ticket.subject}
            identifier={ticket.public_id}
            status={<TicketStatusBadge status={ticket.status} label={t(`statuses.${ticket.status}`)} />}
            leading={<span className="flex size-11 items-center justify-center rounded-lg bg-primary/10 text-primary"><LifeBuoy aria-hidden="true" className="size-5" /></span>}
            copy={{ copyLabel: t("actions.copy"), copyingLabel: t("actions.copying"), copiedLabel: t("actions.copied") }}
          />
        </CardHeader>
        <CardContent className="border-t border-card-border pt-6">
          <DefinitionList columns={2}>
            <DefinitionListItem icon={Tag} label={t("fields.category")} value={<Badge variant="info">{t(`categories.${ticket.category}`)}</Badge>} />
            <DefinitionListItem icon={Flag} label={t("fields.priority")} value={<Badge variant={priorityBadgeVariant(ticket.priority)}>{t(`priorities.${ticket.priority}`)}</Badge>} />
            <DefinitionListItem icon={CalendarPlus} label={t("fields.createdAt")} value={<DateTimeValue value={ticket.created_at} locale={locale} fallback={t("values.notAvailable")} format={{ dateStyle: "medium", timeStyle: "short" }} />} />
            <DefinitionListItem icon={Clock3} label={t("fields.lastActivity")} value={<DateTimeValue value={ticket.last_message_at} locale={locale} fallback={t("values.notAvailable")} format={{ dateStyle: "medium", timeStyle: "short" }} />} />
            {ticket.resolved_at ? <DefinitionListItem icon={Clock3} label={t("fields.resolvedAt")} value={<DateTimeValue value={ticket.resolved_at} locale={locale} format={{ dateStyle: "medium", timeStyle: "short" }} />} /> : null}
            {ticket.closed_at ? <DefinitionListItem icon={LockKeyhole} label={t("fields.closedAt")} value={<DateTimeValue value={ticket.closed_at} locale={locale} format={{ dateStyle: "medium", timeStyle: "short" }} />} /> : null}
          </DefinitionList>
        </CardContent>
      </Card>

      <section className="flex flex-col gap-4" aria-labelledby="conversation-title">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary"><MessagesSquare aria-hidden="true" className="size-4" /></span>
            <h2 id="conversation-title" className="text-lg font-semibold">{t("conversation.title")}</h2>
          </div>
          <Badge variant="secondary" className="tabular-nums">{detail.messages.length}</Badge>
        </div>
        {detail.messages.map((message) => <MessageCard key={message.id} message={message} currentUserID={currentUserID} formatDate={formatDate} />)}
      </section>

      {canRate ? <RatingForm ticketID={ticket.public_id} initialScore={ticket.rating?.score} initialComment={ticket.rating?.comment} onSaved={(updated) => setDetail((current) => current ? { ...current, ticket: updated.ticket } : current)} /> : null}

      {canReply ? (
        <Card>
          <CardHeader><CardTitle className="flex items-center gap-2"><Send aria-hidden="true" className="size-5 text-primary" />{ticket.status === "resolved" ? t("reply.reopenTitle") : t("reply.title")}</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={reply} className="flex flex-col gap-4">
              <AppTextareaField label={t("fields.message")} value={body} onChange={(event) => setBody(event.target.value)} maxLength={10000} rows={6} disabled={submitting} />
              <Input type="file" multiple disabled={submitting} accept=".pdf,.png,.jpg,.jpeg,.webp,.txt,.csv,.log,.json" aria-label={t("attachments.add")} onChange={(event) => setFiles(Array.from(event.target.files ?? []).slice(0, 5))} />
              <Button type="submit" className="self-end" disabled={submitting}><Send aria-hidden="true" />{submitting ? t("actions.submitting") : ticket.status === "resolved" ? t("actions.sendAndReopen") : t("actions.send")}</Button>
            </form>
          </CardContent>
        </Card>
      ) : <p className="flex items-start gap-2 rounded-xl border border-neutral-soft-border bg-neutral-soft p-4 text-sm text-neutral-soft-foreground"><LockKeyhole aria-hidden="true" className="mt-0.5 size-4 shrink-0" />{t("reply.closed")}</p>}
    </PageContainer>
  );
}

function MessageCard({ message, currentUserID, formatDate }: { message: TicketMessage; currentUserID: string; formatDate: (value?: string) => string }) {
  const t = useTranslations("support");
  const support = message.author_type === "support_agent";
  const author = support ? t("conversation.support") : message.author_id === currentUserID ? t("conversation.you") : t("conversation.organization");
  return (
    <Message align={support ? "start" : "end"} role="article">
      <MessageContent className="max-w-3xl group-data-[align=end]/message:ms-auto">
        <MessageHeader className="flex-wrap justify-between gap-2 px-1">
          <strong className="flex items-center gap-1.5 text-foreground">{support ? <Headphones aria-hidden="true" className="size-4 text-info-soft-foreground" /> : <UserRound aria-hidden="true" className="size-4 text-primary" />}{author}</strong>
          <time dateTime={message.created_at}>{formatDate(message.created_at)}</time>
        </MessageHeader>
        <div className={cn("w-full rounded-xl border p-4 shadow-xs", support ? "border-info-soft-border bg-info-soft text-info-soft-foreground" : "border-card-border bg-card text-card-foreground")}>
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
      </MessageContent>
    </Message>
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

function priorityBadgeVariant(priority: TicketDetail["ticket"]["priority"]): ComponentProps<typeof Badge>["variant"] {
  switch (priority) {
    case "urgent": return "destructive";
    case "high": return "warning";
    case "low": return "secondary";
    default: return "outline";
  }
}
