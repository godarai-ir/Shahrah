"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Paperclip, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppSelectField, AppTextField, AppTextareaField, FormSection, PageContainer } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createIdempotencyKey, readTicketResponse, type TicketCategory, type TicketDetail } from "./types";

const MAX_FILES = 5;
const MAX_FILE_BYTES = 10 * 1024 * 1024;
const MAX_TOTAL_BYTES = 25 * 1024 * 1024;
const ALLOWED_EXTENSIONS = new Set(["pdf", "png", "jpg", "jpeg", "webp", "txt", "csv", "log", "json"]);

export function NewTicketPage() {
  const t = useTranslations("support");
  const router = useRouter();
  const [subject, setSubject] = useState("");
  const [category, setCategory] = useState<TicketCategory>("technical");
  const [body, setBody] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const idempotencyKey = useRef(createIdempotencyKey());

  function updateFiles(incoming: File[]) {
    const next = [...files, ...incoming];
    if (next.length > MAX_FILES) return setError(t("validation.fileCount"));
    if (next.some((file) => file.size > MAX_FILE_BYTES)) return setError(t("validation.fileSize"));
    if (next.reduce((sum, file) => sum + file.size, 0) > MAX_TOTAL_BYTES) return setError(t("validation.totalSize"));
    if (next.some((file) => !ALLOWED_EXTENSIONS.has(file.name.split(".").pop()?.toLowerCase() ?? ""))) return setError(t("validation.fileType"));
    setError("");
    setFiles(next);
  }

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    const trimmedSubject = subject.trim();
    if (trimmedSubject.length < 5 || trimmedSubject.length > 160) return setError(t("validation.subject"));
    if (!body.trim() && files.length === 0) return setError(t("validation.messageRequired"));
    if (body.length > 10000) return setError(t("validation.messageLength"));
    setSubmitting(true);
    setError("");
    const form = new FormData();
    form.set("payload", JSON.stringify({ subject: trimmedSubject, category, body, idempotency_key: idempotencyKey.current }));
    files.forEach((file) => form.append("files", file));
    try {
      const response = await fetch("/api/support/tickets", {
        method: "POST",
        body: form,
        headers: { "Idempotency-Key": idempotencyKey.current },
      });
      const detail = await readTicketResponse<TicketDetail>(response);
      router.push(`/support/tickets/${detail.ticket.public_id}`);
      router.refresh();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t("errors.create"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <PageContainer size="readable">
      <form onSubmit={submit} className="flex flex-col gap-6">
        <FormSection title={t("new.detailsTitle")} description={t("new.detailsDescription")}>
          <div className="grid gap-5">
            <AppTextField label={t("fields.subject")} value={subject} onChange={(event) => setSubject(event.target.value)} minLength={5} maxLength={160} required disabled={submitting} />
            <AppSelectField
              label={t("fields.category")}
              value={category}
              onValueChange={(value) => setCategory(value as TicketCategory)}
              disabled={submitting}
              options={(["technical", "billing", "account", "feature_request", "other"] as const).map((value) => ({ value, label: t(`categories.${value}`) }))}
            />
            <AppTextareaField label={t("fields.message")} description={t("new.messageHint")} value={body} onChange={(event) => setBody(event.target.value)} maxLength={10000} rows={8} disabled={submitting} />
          </div>
        </FormSection>
        <FormSection title={t("attachments.title")} description={t("attachments.policy")}>
          <div className="flex flex-col gap-3">
            <Input
              type="file"
              multiple
              disabled={submitting || files.length >= MAX_FILES}
              aria-label={t("attachments.add")}
              accept=".pdf,.png,.jpg,.jpeg,.webp,.txt,.csv,.log,.json"
              onChange={(event) => {
                updateFiles(Array.from(event.target.files ?? []));
                event.currentTarget.value = "";
              }}
            />
            {files.map((file, index) => (
              <div key={`${file.name}-${file.lastModified}`} className="flex items-center gap-3 rounded-lg border border-border bg-card p-3">
                <Paperclip className="size-4 text-muted-foreground" aria-hidden="true" />
                <span className="min-w-0 flex-1 truncate text-sm" dir="auto">{file.name}</span>
                <span className="text-xs text-muted-foreground" dir="ltr">{formatBytes(file.size)}</span>
                <Button type="button" variant="ghost" size="icon-sm" aria-label={t("attachments.remove", { name: file.name })} onClick={() => setFiles((current) => current.filter((_, itemIndex) => itemIndex !== index))}>
                  <X aria-hidden="true" />
                </Button>
              </div>
            ))}
          </div>
        </FormSection>
        {error ? <Alert variant="destructive"><AlertTitle>{t("states.errorTitle")}</AlertTitle><AlertDescription>{error}</AlertDescription></Alert> : null}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" disabled={submitting} onClick={() => router.back()}>{t("actions.cancel")}</Button>
          <Button type="submit" disabled={submitting}>{submitting ? t("actions.submitting") : t("actions.submit")}</Button>
        </div>
      </form>
    </PageContainer>
  );
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KiB`;
  return `${(value / 1024 / 1024).toFixed(1)} MiB`;
}
