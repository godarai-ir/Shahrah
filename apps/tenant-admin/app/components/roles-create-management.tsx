"use client";

import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { Plus } from "lucide-react";

import { DisplayNameField } from "@/components/app-ui/DisplayNameField";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { IdentifierField } from "@/components/app-ui/IdentifierField";
import { PermissionEditor } from "@/components/app-ui/PermissionEditor";
import { PermissionSummary } from "@/components/app-ui/PermissionSummary";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";

import {
  filterSupportedPermissions,
  getPermissionEditorLabels,
  getPermissionSummaryLabels,
  translatePermissionDescription,
  translatePermissionGroup,
} from "./role-permissions";

type PermissionItem = {
  key: string;
  group: string;
  description: string;
};

type Banner = {
  kind: "success" | "error";
  message: string;
};

type RoleCreateManagementProps = {
  initialPermissionCatalog: PermissionItem[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

type RoleFormState = {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
};

const initialForm: RoleFormState = { name: "", display_name: "", description: "", permissions: [] };

export default function RoleCreateManagement({ initialPermissionCatalog, open, onOpenChange }: RoleCreateManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();

  const [form, setForm] = useState<RoleFormState>(initialForm);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [query, setQuery] = useState("");
  const [viewMode, setViewMode] = useState<"selected" | "all">("all");
  const [error, setError] = useState<string | null>(null);
  const [banner, setBanner] = useState<Banner | null>(null);

  const filteredPermissionCatalog = useMemo(
    () => filterSupportedPermissions(initialPermissionCatalog),
    [initialPermissionCatalog],
  );
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const editorLabels = useMemo(() => getPermissionEditorLabels(locale), [locale]);
  const summaryLabels = useMemo(() => getPermissionSummaryLabels(locale), [locale]);

  function validate(next: RoleFormState): string | null {
    if (!next.name.trim()) return t("rolesPage.form.name");
    if (!next.display_name.trim()) return t("rolesPage.form.displayName");
    if (next.permissions.length === 0) return locale.startsWith("fa") ? "حداقل یک مجوز انتخاب کنید." : "Select at least one permission.";
    return null;
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBanner(null);

    const validation = validate(form);
    if (validation) {
      setError(validation);
      return;
    }
    setError(null);

    setIsSubmitting(true);
    try {
      await requestJSON("/api/roles", {
        method: "POST",
        body: JSON.stringify(form),
      });
      setForm(initialForm);
      setQuery("");
      setViewMode("all");
      onOpenChange(false);
      router.refresh();
    } catch (submitError) {
      setBanner({ kind: "error", message: normalizeError(submitError, t("messages.roleCreateError")) });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      title={t("rolesPage.createTitle")}
      description={locale.startsWith("fa") ? "اطلاعات نقش و مجوزهای آن را وارد کنید." : "Enter the role details and permissions."}
      size="lg"
      side="left"
      pending={isSubmitting}
      closeLabel={t("rolesPage.actions.cancel")}
      footer={
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            {t("rolesPage.actions.cancel")}
          </Button>
          <Button type="submit" form="create-role-form" disabled={isSubmitting}>
            <Plus data-icon="inline-start" aria-hidden="true" />
            {isSubmitting ? t("rolesPage.form.creating") : t("rolesPage.form.create")}
          </Button>
        </div>
      }
    >
      <form id="create-role-form" className="flex flex-col gap-5" onSubmit={onSubmit}>
        {banner ? (
          <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
            <AlertDescription>{banner.message}</AlertDescription>
          </Alert>
        ) : null}

              <section className="flex flex-col gap-3">
                <FieldGroup className="grid gap-3 md:grid-cols-2">
                  <IdentifierField id="role-name" label={t("rolesPage.form.name")} value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} placeholder="tenant_finance_admin" />
                  <DisplayNameField id="role-display-name" label={t("rolesPage.form.displayName")} value={form.display_name} onChange={(e) => setForm((p) => ({ ...p, display_name: e.target.value }))} placeholder={locale.startsWith("fa") ? "مدیر مالی فضای کاری" : "Tenant Finance Admin"} />
                  <Field className="md:col-span-2">
                    <FieldLabel htmlFor="role-description">{t("rolesPage.form.description")}</FieldLabel>
                    <Textarea id="role-description" value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} className="min-h-24" />
                  </Field>
                </FieldGroup>
              </section>

              <Separator />

              <section className="flex flex-col gap-3">
                <PermissionEditor
                  permissions={filteredPermissionCatalog}
                  value={form.permissions}
                  onValueChange={(permissions) => setForm((prev) => ({ ...prev, permissions }))}
                  query={query}
                  onQueryChange={setQuery}
                  mode={viewMode}
                  onModeChange={setViewMode}
                  labels={editorLabels}
                  dir={locale.startsWith("fa") ? "rtl" : "ltr"}
                  formatGroupLabel={(group) => translatePermissionGroup(group, locale)}
                  formatPermissionDescription={(permission) => translatePermissionDescription(permission, locale)}
                  formatNumber={(value) => numberFormatter.format(value)}
                />
              </section>

              <FieldError>{error}</FieldError>

        <PermissionSummary
          permissions={filteredPermissionCatalog}
          value={form.permissions}
          title={locale.startsWith("fa") ? "خلاصه انتخاب‌ها" : "Selection summary"}
          labels={summaryLabels}
          onClear={() => setForm((prev) => ({ ...prev, permissions: [] }))}
          formatNumber={(value) => numberFormatter.format(value)}
        />
      </form>
    </FormSheet>
  );
}

async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, {
    ...options,
    headers,
    cache: "no-store",
  });
  const text = await response.text();
  const payload = parsePayload(text);
  if (!response.ok) {
    throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
  }
  return payload as T;
}

function parsePayload(text: string): unknown {
  if (!text.trim()) return {};
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return { message: text };
  }
}

function extractMessage(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}
