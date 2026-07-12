"use client";

import { FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, Plus } from "lucide-react";

import { IdentifierField, PermissionEditor, PermissionSummary } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
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
};

type RoleFormState = {
  name: string;
  display_name: string;
  description: string;
  permissions: string[];
};

export default function RoleCreateManagement({ initialPermissionCatalog }: RoleCreateManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();

  const [form, setForm] = useState<RoleFormState>({ name: "", display_name: "", description: "", permissions: [] });
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
      setBanner({ kind: "success", message: t("messages.roleCreated") });
      router.push("/roles");
      router.refresh();
    } catch (submitError) {
      setBanner({ kind: "error", message: normalizeError(submitError, t("messages.roleCreateError")) });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <p className="text-sm text-muted-foreground">
        {locale.startsWith("fa")
          ? "ایجاد نقش در صفحه جدا انجام می‌شود تا جریان ثبت دقیق‌تر و بدون شلوغی باشد."
          : "Create a role in a dedicated page for a focused, low-noise workflow."}
      </p>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-12">
        <Card className="lg:col-span-8">
          <CardHeader>
            <CardTitle>{t("rolesPage.createTitle")}</CardTitle>
            <CardDescription>
              {locale.startsWith("fa") ? "مرحله ۱: اطلاعات نقش. مرحله ۲: انتخاب مجوز. مرحله ۳: ثبت." : "Step 1: Role details. Step 2: Permission selection. Step 3: Submit."}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form className="flex flex-col gap-5" onSubmit={onSubmit}>
              <section className="flex flex-col gap-3">
                <FieldGroup className="grid gap-3 md:grid-cols-2">
                  <IdentifierField id="role-name" label={t("rolesPage.form.name")} value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} placeholder="tenant_finance_admin" />
                  <Field>
                    <FieldLabel htmlFor="role-display-name">{t("rolesPage.form.displayName")}</FieldLabel>
                    <Input id="role-display-name" value={form.display_name} onChange={(e) => setForm((p) => ({ ...p, display_name: e.target.value }))} placeholder={locale.startsWith("fa") ? "مدیر مالی فضای کاری" : "Tenant Finance Admin"} />
                  </Field>
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

              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={isSubmitting} className="gap-2">
                  <Plus className="h-4 w-4" />
                  {isSubmitting ? t("rolesPage.form.creating") : t("rolesPage.form.create")}
                </Button>
                <Button asChild type="button" variant="ghost" className="gap-2">
                  <Link href="/roles">
                    <ArrowLeft className="h-4 w-4" />
                    {locale.startsWith("fa") ? "بازگشت به فهرست نقش‌ها" : "Back to roles list"}
                  </Link>
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <PermissionSummary
          permissions={filteredPermissionCatalog}
          value={form.permissions}
          title={locale.startsWith("fa") ? "خلاصه انتخاب‌ها" : "Selection summary"}
          labels={summaryLabels}
          onClear={() => setForm((prev) => ({ ...prev, permissions: [] }))}
          formatNumber={(value) => numberFormatter.format(value)}
          className="lg:col-span-4 lg:sticky lg:top-6 lg:h-fit"
        />
      </div>
    </div>
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
