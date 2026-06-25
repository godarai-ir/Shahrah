"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Check, Copy, Pencil, Plus, ShieldCheck, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";

import { createGuardrail, deleteGuardrail, GuardrailAPIError, listGuardrails, updateGuardrail } from "./api";
import type { GuardrailFormInput, GuardrailPayload, GuardrailType, TenantGuardrail } from "./types";

const emptyForm: GuardrailFormInput = {
  slug: "",
  name: "",
  description: "",
  type: "keyword_blocklist",
  operation: "validate",
  enforcement: "enforce",
  enabled: true,
  terms: "",
  patterns: "",
  redactionText: "[***]",
};

function formatDateTime(value: string, locale: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  const dateLocale = locale === "fa" ? "fa-IR-u-ca-persian" : "en-US";
  return new Intl.DateTimeFormat(dateLocale, {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function lines(value: string) {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function labelForType(type: GuardrailType, t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (type === "keyword_blocklist") return t("types.keyword_blocklist");
  if (type === "regex") return t("types.regex");
  if (type === "secrets_detection") return t("types.secrets_detection");
  return t("types.basic_pii");
}

function labelForOperation(operation: GuardrailFormInput["operation"], t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (operation === "mutate") {
    return t.has("operations.mutate") ? t("operations.mutate") : "mutate";
  }
  return t.has("operations.validate") ? t("operations.validate") : "validate";
}

function labelForEnforcement(enforcement: GuardrailFormInput["enforcement"], t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (enforcement === "audit") {
    return t.has("enforcement.audit") ? t("enforcement.audit") : "audit";
  }
  if (enforcement === "enforce_ignore_error") {
    return t.has("enforcement.enforce_ignore_error") ? t("enforcement.enforce_ignore_error") : "enforce_ignore_error";
  }
  return t.has("enforcement.enforce") ? t("enforcement.enforce") : "enforce";
}

function formFromGuardrail(item: TenantGuardrail): GuardrailFormInput {
  return {
    slug: item.slug,
    name: item.name,
    description: item.description ?? "",
    type: item.type,
    operation: item.operation,
    enforcement: item.enforcement,
    enabled: item.enabled,
    terms: Array.isArray(item.config.terms) ? item.config.terms.join("\n") : "",
    patterns: Array.isArray(item.config.patterns) ? item.config.patterns.join("\n") : "",
    redactionText: typeof item.config.redaction_text === "string" ? item.config.redaction_text : "[***]",
  };
}

function buildPayload(form: GuardrailFormInput, mode: "create" | "edit"): GuardrailPayload {
  const type = form.type;
  const config: Record<string, unknown> = {};
  if (type === "keyword_blocklist") {
    config.terms = lines(form.terms);
  } else if (type === "regex") {
    config.patterns = lines(form.patterns);
    config.redaction_text = form.redactionText.trim() || "[***]";
  } else {
    config.redaction_text = form.redactionText.trim() || "[***]";
  }

  return {
    ...(mode === "create" ? { slug: form.slug.trim() } : {}),
    name: form.name.trim(),
    description: form.description.trim(),
    type,
    operation: type === "keyword_blocklist" ? "validate" : form.operation,
    enforcement: form.enforcement,
    enabled: form.enabled,
    config,
  };
}

function validateForm(form: GuardrailFormInput, t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (!/^[a-zA-Z0-9][a-zA-Z0-9_-]{1,95}$/.test(form.slug.trim())) {
    return t("validation.slug");
  }
  if (!form.name.trim()) return t("validation.name");
  if (form.type === "keyword_blocklist" && lines(form.terms).length === 0) {
    return t("validation.terms");
  }
  if (form.type === "regex" && lines(form.patterns).length === 0) {
    return t("validation.patterns");
  }
  return "";
}

function localizedError(error: unknown, fallback: string, t: ReturnType<typeof useTranslations<"guardrails">>) {
  if (!(error instanceof GuardrailAPIError)) {
    return error instanceof Error ? error.message : fallback;
  }
  if (error.status === 409) return t("messages.duplicate");
  if (error.status === 400) return t("messages.invalid");
  if (error.status === 403) return t("messages.forbidden");
  if (error.status === 404) return t("messages.notFound");
  if (error.status && error.status >= 500) return t("messages.serverError");
  return fallback;
}

export function GuardrailsPage() {
  const t = useTranslations("guardrails");
  const locale = useLocale();
  const direction = locale === "fa" ? "rtl" : "ltr";
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<TenantGuardrail[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [editing, setEditing] = useState<TenantGuardrail | null>(null);
  const [pendingDelete, setPendingDelete] = useState<TenantGuardrail | null>(null);
  const [form, setForm] = useState<GuardrailFormInput>(emptyForm);
  const [formOpen, setFormOpen] = useState(false);
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [copiedSelector, setCopiedSelector] = useState("");

  async function load() {
    setIsLoading(true);
    try {
      const response = await listGuardrails(query);
      setItems(response.data);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.loadError"), t) });
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const tableItems = useMemo(() => items, [items]);
  const hasSearchQuery = query.trim().length > 0;

  function openCreateDialog() {
    setEditing(null);
    setForm(emptyForm);
    setFormOpen(true);
  }

  function openEditDialog(item: TenantGuardrail) {
    setEditing(item);
    setForm(formFromGuardrail(item));
    setFormOpen(true);
  }

  async function submitForm() {
    if (isSaving) return;
    const validationError = validateForm(form, t);
    if (validationError) {
      setNotice({ kind: "error", message: validationError });
      return;
    }
    setIsSaving(true);
    try {
      const payload = buildPayload(form, editing ? "edit" : "create");
      if (editing) {
        await updateGuardrail(editing.slug, payload);
      } else {
        await createGuardrail(payload);
      }
      setFormOpen(false);
      setNotice({ kind: "success", message: t("messages.saveSuccess") });
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.saveError"), t) });
    } finally {
      setIsSaving(false);
    }
  }

  async function confirmDelete() {
    if (!pendingDelete || isDeleting) return;
    setIsDeleting(true);
    try {
      await deleteGuardrail(pendingDelete.slug);
      setPendingDelete(null);
      setNotice({ kind: "success", message: t("messages.deleteSuccess") });
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: localizedError(error, t("messages.deleteError"), t) });
    } finally {
      setIsDeleting(false);
    }
  }

  async function copySelector(selector: string) {
    await navigator.clipboard.writeText(selector);
    setCopiedSelector(selector);
    window.setTimeout(() => setCopiedSelector(""), 1600);
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="min-w-0">
              <CardTitle className="flex items-center gap-2">
                <ShieldCheck data-icon="inline-start" aria-hidden="true" />
                {t("title")}
              </CardTitle>
              <CardDescription>{t("description")}</CardDescription>
            </div>
            <Button type="button" onClick={openCreateDialog}>
              <Plus data-icon="inline-start" aria-hidden="true" />
              {t("actions.new")}
            </Button>
          </div>
          <Input placeholder={t("searchPlaceholder")} value={query} onChange={(event) => setQuery(event.target.value)} />
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {notice ? (
            <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
              <AlertTitle>{notice.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
              <AlertDescription>{notice.message}</AlertDescription>
            </Alert>
          ) : null}

          {tableItems.length === 0 && !isLoading ? (
            <div className="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">
              {hasSearchQuery ? t("emptySearch") : t("empty")}
            </div>
          ) : (
            <div className="overflow-x-auto rounded-lg border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("table.name")}</TableHead>
                    <TableHead>{t("table.type")}</TableHead>
                    <TableHead>{t("table.execution")}</TableHead>
                    <TableHead>{t("table.status")}</TableHead>
                    <TableHead>{t("table.updatedAt")}</TableHead>
                    <TableHead>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tableItems.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>
                        <div className="flex min-w-64 flex-col gap-1">
                          <span className="font-medium">{item.name}</span>
                          <div className="flex max-w-72 items-center gap-2">
                            <code className="truncate rounded bg-muted px-2 py-1 text-xs text-muted-foreground" dir="ltr">
                              {item.selector}
                            </code>
                            <Button
                              type="button"
                              size="icon"
                              variant="ghost"
                              aria-label={t("actions.copySelector")}
                              onClick={() => void copySelector(item.selector)}
                            >
                              {copiedSelector === item.selector ? (
                                <Check data-icon="inline-start" aria-hidden="true" />
                              ) : (
                                <Copy data-icon="inline-start" aria-hidden="true" />
                              )}
                            </Button>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{labelForType(item.type, t)}</Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          <Badge variant="secondary">{labelForOperation(item.operation, t)}</Badge>
                          <Badge variant="secondary">{labelForEnforcement(item.enforcement, t)}</Badge>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={item.enabled ? "default" : "secondary"}>
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatDateTime(item.updated_at, locale)}</TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button type="button" size="icon" variant="outline" aria-label={t("actions.edit")} onClick={() => openEditDialog(item)}>
                            <Pencil data-icon="inline-start" aria-hidden="true" />
                          </Button>
                          <Button type="button" size="icon" variant="outline" aria-label={t("actions.delete")} onClick={() => setPendingDelete(item)}>
                            <Trash2 data-icon="inline-start" aria-hidden="true" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={formOpen} onOpenChange={setFormOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("form.editTitle") : t("form.addTitle")}</DialogTitle>
            <DialogDescription>{t("form.description")}</DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="grid gap-3 md:grid-cols-2">
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-name">{t("form.name")}</Label>
                <Input id="guardrail-name" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-slug">{t("form.identifier")}</Label>
                <Input
                  id="guardrail-slug"
                  dir="ltr"
                  value={form.slug}
                  disabled={Boolean(editing)}
                  onChange={(event) => setForm({ ...form, slug: event.target.value })}
                />
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <Label htmlFor="guardrail-description">{t("form.notes")}</Label>
              <Textarea id="guardrail-description" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
            </div>

            <div className="grid gap-3 md:grid-cols-3">
              <div className="flex flex-col gap-2">
                <Label>{t("form.type")}</Label>
                <Select dir={direction} value={form.type} onValueChange={(value) => setForm({ ...form, type: value as GuardrailType, operation: value === "keyword_blocklist" ? "validate" : form.operation })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent dir={direction}>
                    <SelectGroup>
                      <SelectItem value="keyword_blocklist">{t("types.keyword_blocklist")}</SelectItem>
                      <SelectItem value="regex">{t("types.regex")}</SelectItem>
                      <SelectItem value="secrets_detection">{t("types.secrets_detection")}</SelectItem>
                      <SelectItem value="basic_pii">{t("types.basic_pii")}</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-2">
                <Label>{t("form.operation")}</Label>
                <Select dir={direction} value={form.type === "keyword_blocklist" ? "validate" : form.operation} disabled={form.type === "keyword_blocklist"} onValueChange={(value) => setForm({ ...form, operation: value as GuardrailFormInput["operation"] })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent dir={direction}>
                    <SelectGroup>
                      <SelectItem value="validate">{t("operations.validate")}</SelectItem>
                      <SelectItem value="mutate">{t("operations.mutate")}</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
              <div className="flex flex-col gap-2">
                <Label>{t("form.enforcement")}</Label>
                <Select dir={direction} value={form.enforcement} onValueChange={(value) => setForm({ ...form, enforcement: value as GuardrailFormInput["enforcement"] })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent dir={direction}>
                    <SelectGroup>
                      <SelectItem value="enforce">{t("enforcement.enforce")}</SelectItem>
                      <SelectItem value="audit">{t("enforcement.audit")}</SelectItem>
                      <SelectItem value="enforce_ignore_error">{t("enforcement.enforce_ignore_error")}</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {form.type === "keyword_blocklist" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-terms">{t("form.terms")}</Label>
                <Textarea
                  id="guardrail-terms"
                  dir="ltr"
                  value={form.terms}
                  placeholder="ignore previous instructions&#10;drop table"
                  onChange={(event) => setForm({ ...form, terms: event.target.value })}
                />
              </div>
            ) : null}

            {form.type === "regex" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-patterns">{t("form.patterns")}</Label>
                <Textarea
                  id="guardrail-patterns"
                  dir="ltr"
                  value={form.patterns}
                  placeholder="(?i)password\\s*[:=]\\s*\\S+"
                  onChange={(event) => setForm({ ...form, patterns: event.target.value })}
                />
              </div>
            ) : null}

            {form.type !== "keyword_blocklist" ? (
              <div className="flex flex-col gap-2">
                <Label htmlFor="guardrail-redaction">{t("form.redactionText")}</Label>
                <Input id="guardrail-redaction" dir="ltr" value={form.redactionText} onChange={(event) => setForm({ ...form, redactionText: event.target.value })} />
              </div>
            ) : null}

            <div className="flex items-center justify-between rounded-lg border px-3 py-2">
              <Label htmlFor="guardrail-enabled">{t("form.enabled")}</Label>
              <Switch id="guardrail-enabled" checked={form.enabled} onCheckedChange={(checked) => setForm({ ...form, enabled: checked })} />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" disabled={isSaving} onClick={() => setFormOpen(false)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" disabled={isSaving} onClick={() => void submitForm()}>
              {isSaving ? t("actions.saving") : t("actions.save")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={pendingDelete !== null} onOpenChange={(open) => !open && setPendingDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("deleteDialog.title")}</DialogTitle>
            <DialogDescription>{t("deleteDialog.description", { name: pendingDelete?.name ?? "-" })}</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" disabled={isDeleting} onClick={() => setPendingDelete(null)}>
              {t("actions.cancel")}
            </Button>
            <Button type="button" variant="destructive" disabled={isDeleting} onClick={() => void confirmDelete()}>
              {isDeleting ? t("deleteDialog.deleting") : t("actions.delete")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
