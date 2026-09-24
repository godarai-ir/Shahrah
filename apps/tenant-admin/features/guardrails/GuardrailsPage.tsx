"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { Pencil, Plus, Trash2 } from "lucide-react";

import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppNumberField } from "@/components/app-ui/AppNumberField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { HelpLink } from "@/components/app-ui/HelpLink";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { SearchField } from "@/components/app-ui/SearchField";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { createGuardrail, deleteGuardrail, GuardrailAPIError, listGuardrails, updateGuardrail } from "./api";
import { GUARDRAIL_CATALOG, isMutateOnlyType, isValidateOnlyType, operationForType } from "./catalog";
import { buildPayload, emptyForm, formFromGuardrail, lines, parseJSONOrEmpty } from "./form";
import type { GuardrailFormInput, GuardrailType, TenantGuardrail } from "./types";
import { getLocaleDirection } from "@/lib/locale-direction";

const GUARDRAIL_COLUMN_CLASS = {
  name: "w-[52%]",
  type: "hidden w-36 @md:table-cell",
  execution: "hidden w-48 @lg:table-cell",
  status: "w-[28%]",
  updatedAt: "hidden w-40 @lg:table-cell",
  actions: "w-20 text-end",
} as const;

function labelForType(type: GuardrailType, t: ReturnType<typeof useTranslations<"guardrails">>) {
  return t(`types.${type}`);
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
  if (form.type === "regex_replace" && !form.pattern.trim()) return t("validation.pattern");
  if ((form.type === "sentence_count" || form.type === "word_count" || form.type === "character_count") && !form.min.trim() && !form.max.trim()) {
    return t("validation.range");
  }
  if (form.type === "json_schema") {
    try {
      parseJSONOrEmpty(form.schema);
    } catch {
      return t("validation.json");
    }
  }
  if (form.type === "json_keys" && lines(form.keys).length === 0) return t("validation.keys");
  if (form.type === "jwt_validation") {
    const hasVerifier = form.sharedSecret.trim() || form.inlineJwks.trim() || form.jwksUri.trim() || form.introspectionUrl.trim();
    if (!hasVerifier) return t("validation.jwtVerifier");
    if (form.inlineJwks.trim()) {
      try {
        parseJSONOrEmpty(form.inlineJwks);
      } catch {
        return t("validation.json");
      }
    }
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

export function GuardrailsPage({ docsHref }: { docsHref: string }) {
  const t = useTranslations("guardrails");
  const locale = useLocale();
  const direction = getLocaleDirection(locale);
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
  const [formNotice, setFormNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);
  const [deleteNotice, setDeleteNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

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
      setFormNotice({ kind: "error", message: validationError });
      return;
    }
    setFormNotice(null);
    setIsSaving(true);
    try {
      const payload = buildPayload(form, editing ? "edit" : "create");
      if (editing) {
        await updateGuardrail(editing.slug, payload);
      } else {
        await createGuardrail(payload);
      }
      setFormOpen(false);
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.saveSuccess") });
      await load();
    } catch (error) {
      setFormNotice({ kind: "error", message: localizedError(error, t("messages.saveError"), t) });
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
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteSuccess") });
      await load();
    } catch (error) {
      setDeleteNotice({ kind: "error", message: localizedError(error, t("messages.deleteError"), t) });
      throw error;
    } finally {
      setIsDeleting(false);
    }
  }

  return (
    <>
      <Card className="@container/guardrails">
        <CardHeader>
          <div className="flex min-w-0 flex-col gap-2 @md:flex-row @md:items-center">
            <SearchField
              containerClassName="min-w-0 flex-1"
              placeholder={t("searchPlaceholder")}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
            />
            <Button type="button" className="w-full @md:w-auto" onClick={openCreateDialog}>
              <Plus data-icon="inline-start" aria-hidden="true" />
              {t("actions.new")}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

          {tableItems.length === 0 && !isLoading ? (
            <EmptyState
              variant="inline"
              density="compact"
              title={hasSearchQuery ? t("emptySearch") : t("empty")}
              secondaryAction={!hasSearchQuery ? <HelpLink href={docsHref} /> : undefined}
            />
          ) : (
            <div className="overflow-hidden rounded-lg border">
              <Table className="table-fixed text-xs lg:text-sm">
                <TableHeader>
                  <TableRow>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.name}>{t("table.name")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.type}>{t("table.type")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.execution}>{t("table.execution")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.status}>{t("table.status")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.updatedAt}>{t("table.updatedAt")}</TableHead>
                    <TableHead className={GUARDRAIL_COLUMN_CLASS.actions}>{t("table.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tableItems.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.name}>
                        <div className="flex min-w-0 flex-col gap-1">
                          <span className="truncate font-medium" title={item.name}>{item.name}</span>
                          <div className="flex min-w-0 items-center gap-2">
                            <code className="truncate rounded bg-muted px-2 py-1 text-xs text-muted-foreground" dir="ltr">
                              {item.selector}
                            </code>
                            <CopyAction value={item.selector} copyLabel={t("actions.copySelector")} copiedLabel={t("actions.copied")} size="icon" />
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.type}>
                        <Badge variant="outline">{labelForType(item.type, t)}</Badge>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.execution}>
                        <div className="flex flex-wrap gap-1">
                          <Badge variant="secondary">{labelForOperation(item.operation, t)}</Badge>
                          <Badge variant="secondary">{labelForEnforcement(item.enforcement, t)}</Badge>
                        </div>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.status}>
                        <Badge variant={item.enabled ? "default" : "secondary"}>
                          {item.enabled ? t("status.enabled") : t("status.disabled")}
                        </Badge>
                      </TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.updatedAt}><DateTimeValue value={item.updated_at} locale={locale} format={{ year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }} /></TableCell>
                      <TableCell className={GUARDRAIL_COLUMN_CLASS.actions}>
                        <div className="flex justify-end gap-2">
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

      <FormSheet
        open={formOpen}
        onOpenChange={(open) => { setFormOpen(open); if (!open) setFormNotice(null); }}
        title={editing ? t("form.editTitle") : t("form.addTitle")}
        size="lg"
        pending={isSaving}
        closeLabel={t("actions.close")}
        bodyClassName="flex flex-col gap-4"
        footer={(
          <FormSheetActions
            secondary={<Button type="button" variant="outline" disabled={isSaving} onClick={() => { setFormOpen(false); setFormNotice(null); }}>
              {t("actions.cancel")}
            </Button>}
            primary={<Button type="button" disabled={isSaving} onClick={() => void submitForm()}>
              {isSaving ? t("actions.saving") : t("actions.save")}
            </Button>}
          />
        )}
      >
            <OperationNotice notice={formNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />

            <div className="grid gap-3 md:grid-cols-2">
              <DisplayNameField id="guardrail-name" label={t("form.name")} value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} />
              <IdentifierField
                id="guardrail-slug"
                label={t("form.identifier")}
                pattern="[A-Za-z0-9][A-Za-z0-9_-]{1,95}"
                minLength={2}
                maxLength={96}
                value={form.slug}
                disabled={Boolean(editing)}
                onChange={(event) => setForm({ ...form, slug: event.target.value })}
              />
            </div>

            <AppTextareaField id="guardrail-description" label={t("form.notes")} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />

            <div className="grid gap-3 md:grid-cols-3">
              <AppSelectField
                label={t("form.type")}
                dir={direction}
                value={form.type}
                onValueChange={(value) => {
                  const nextType = value as GuardrailType;
                  setForm({ ...form, type: nextType, operation: operationForType(nextType, form.operation) });
                }}
                options={GUARDRAIL_CATALOG.map(({ type }) => ({ value: type, label: t(`types.${type}`) }))}
              />
              <AppSelectField
                label={t("form.operation")}
                dir={direction}
                value={operationForType(form.type, form.operation)}
                disabled={isValidateOnlyType(form.type) || isMutateOnlyType(form.type)}
                onValueChange={(value) => setForm({ ...form, operation: value as GuardrailFormInput["operation"] })}
                options={[
                  { value: "validate", label: t("operations.validate") },
                  { value: "mutate", label: t("operations.mutate") },
                ]}
              />
              <AppSelectField
                label={t("form.enforcement")}
                dir={direction}
                value={form.enforcement}
                onValueChange={(value) => setForm({ ...form, enforcement: value as GuardrailFormInput["enforcement"] })}
                options={[
                  { value: "audit", label: t("enforcement.audit") },
                  { value: "enforce", label: t("enforcement.enforce") },
                  { value: "enforce_ignore_error", label: t("enforcement.enforce_ignore_error") },
                ]}
              />
            </div>

            {form.type === "keyword_blocklist" ? (
              <AppTextareaField
                  id="guardrail-terms"
                  label={t("form.terms")}
                  dir="ltr"
                  value={form.terms}
                  placeholder={"ignore previous instructions\ndrop table"}
                  onChange={(event) => setForm({ ...form, terms: event.target.value })}
                />
            ) : null}

            {form.type === "regex" ? (
              <AppTextareaField
                  id="guardrail-patterns"
                  label={t("form.patterns")}
                  dir="ltr"
                  value={form.patterns}
                  placeholder="(?i)password\\s*[:=]\\s*\\S+"
                  onChange={(event) => setForm({ ...form, patterns: event.target.value })}
                />
            ) : null}

            {form.type === "regex_replace" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <AppTextareaField
                    id="guardrail-pattern"
                    label={t("form.pattern")}
                    dir="ltr"
                    value={form.pattern}
                    placeholder="(?i)password\\s*[:=]\\s*\\S+"
                    onChange={(event) => setForm({ ...form, pattern: event.target.value })}
                  />
                <AppTextField id="guardrail-replacement" label={t("form.replacement")} dir="ltr" value={form.replacement} onChange={(event) => setForm({ ...form, replacement: event.target.value })} />
              </div>
            ) : null}

            {form.type === "sentence_count" || form.type === "word_count" || form.type === "character_count" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <AppNumberField id="guardrail-min" label={t("form.min")} value={form.min} onChange={(event) => setForm({ ...form, min: event.target.value })} />
                <AppNumberField id="guardrail-max" label={t("form.max")} value={form.max} onChange={(event) => setForm({ ...form, max: event.target.value })} />
              </div>
            ) : null}

            {form.type === "json_schema" ? (
              <AppCodeTextareaField id="guardrail-schema" label={t("form.schema")} value={form.schema} onChange={(event) => setForm({ ...form, schema: event.target.value })} />
            ) : null}

            {form.type === "json_keys" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <AppTextareaField id="guardrail-keys" label={t("form.keys")} dir="ltr" value={form.keys} placeholder={"answer\nconfidence"} onChange={(event) => setForm({ ...form, keys: event.target.value })} />
                <AppSelectField
                  label={t("form.keyMode")}
                  dir={direction}
                  value={form.keyMode}
                  onValueChange={(value) => setForm({ ...form, keyMode: value as GuardrailFormInput["keyMode"] })}
                  options={[
                    { value: "all", label: t("keyModes.all") },
                    { value: "any", label: t("keyModes.any") },
                    { value: "none", label: t("keyModes.none") },
                  ]}
                />
              </div>
            ) : null}

            {form.type === "contains_code" ? (
              <div className="grid gap-3 md:grid-cols-2">
                <AppTextareaField id="guardrail-languages" label={t("form.languages")} dir="ltr" value={form.languages} placeholder={"sql\npython\ntypescript"} onChange={(event) => setForm({ ...form, languages: event.target.value })} />
                <AppSelectField
                  label={t("form.codeMode")}
                  dir={direction}
                  value={form.codeMode}
                  onValueChange={(value) => setForm({ ...form, codeMode: value as GuardrailFormInput["codeMode"] })}
                  options={[
                    { value: "must_contain", label: t("codeModes.must_contain") },
                    { value: "must_not_contain", label: t("codeModes.must_not_contain") },
                  ]}
                />
              </div>
            ) : null}

            {form.type === "jwt_validation" ? (
              <div className="flex flex-col gap-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <AppTextField id="guardrail-token-regex" label={t("form.tokenRegex")} dir="ltr" value={form.tokenRegex} placeholder="Bearer capture regex" onChange={(event) => setForm({ ...form, tokenRegex: event.target.value })} />
                  <AppTextField id="guardrail-shared-secret" label={t("form.sharedSecret")} dir="ltr" value={form.sharedSecret} onChange={(event) => setForm({ ...form, sharedSecret: event.target.value })} />
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <AppTextField id="guardrail-issuer" label={t("form.issuer")} dir="ltr" value={form.issuer} onChange={(event) => setForm({ ...form, issuer: event.target.value })} />
                  <AppTextField id="guardrail-audience" label={t("form.audience")} dir="ltr" value={form.audience} onChange={(event) => setForm({ ...form, audience: event.target.value })} />
                </div>
                <div className="grid gap-3 md:grid-cols-2">
                  <AppTextField id="guardrail-jwks-uri" label={t("form.jwksUri")} dir="ltr" value={form.jwksUri} onChange={(event) => setForm({ ...form, jwksUri: event.target.value })} />
                  <AppTextField id="guardrail-introspection" label={t("form.introspectionUrl")} dir="ltr" value={form.introspectionUrl} onChange={(event) => setForm({ ...form, introspectionUrl: event.target.value })} />
                </div>
                <AppTextField id="guardrail-introspection-auth" label={t("form.introspectionAuthHeader")} dir="ltr" value={form.introspectionAuthHeader} placeholder="Bearer ..." onChange={(event) => setForm({ ...form, introspectionAuthHeader: event.target.value })} />
                <AppCodeTextareaField id="guardrail-inline-jwks" label={t("form.inlineJwks")} className="min-h-32" value={form.inlineJwks} onChange={(event) => setForm({ ...form, inlineJwks: event.target.value })} />
              </div>
            ) : null}

            {form.type === "inline_image_urls" ? (
              <div className="flex flex-col gap-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <AppNumberField id="guardrail-max-bytes" label={t("form.maxBytes")} value={form.maxBytes} onChange={(event) => setForm({ ...form, maxBytes: event.target.value })} />
                  <AppNumberField id="guardrail-timeout" label={t("form.timeoutMs")} value={form.timeoutMs} onChange={(event) => setForm({ ...form, timeoutMs: event.target.value })} />
                </div>
                <AppTextareaField id="guardrail-allowed-hosts" label={t("form.allowedHosts")} dir="ltr" value={form.allowedHosts} placeholder={"cdn.example.com\nimages.example.com"} onChange={(event) => setForm({ ...form, allowedHosts: event.target.value })} />
                <AppSwitchField id="guardrail-private-networks" label={t("form.allowPrivateNetworks")} checked={form.allowPrivateNetworks} onCheckedChange={(checked) => setForm({ ...form, allowPrivateNetworks: checked })} />
              </div>
            ) : null}

            {form.type !== "keyword_blocklist" && form.type !== "regex_replace" && form.type !== "sentence_count" && form.type !== "word_count" && form.type !== "character_count" && form.type !== "json_schema" && form.type !== "json_keys" && form.type !== "contains_code" && form.type !== "not_null" && form.type !== "jwt_validation" && form.type !== "inline_image_urls" ? (
              <AppTextField id="guardrail-redaction" label={t("form.redactionText")} dir="ltr" value={form.redactionText} onChange={(event) => setForm({ ...form, redactionText: event.target.value })} />
            ) : null}

            <AppSwitchField id="guardrail-enabled" label={t("form.enabled")} checked={form.enabled} onCheckedChange={(checked) => setForm({ ...form, enabled: checked })} />
      </FormSheet>

      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingDelete(null);
            setDeleteNotice(null);
          }
        }}
        title={t("deleteDialog.title")}
        description={t("deleteDialog.description", { name: pendingDelete?.name ?? "-" })}
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("actions.delete")}
        pendingLabel={t("deleteDialog.deleting")}
        onConfirm={confirmDelete}
        pending={isDeleting}
        closeOnConfirm={false}
        feedback={<OperationNotice notice={deleteNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />}
      />
    </>
  );
}
