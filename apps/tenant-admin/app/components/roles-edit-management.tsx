"use client";

import { type FormEvent, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, ArrowRight, Save } from "lucide-react";

import { DisplayNameField, FormSection, getPermissionSelectionSummary, IdentifierField, PageContainer, PageHeader, PermissionEditor, PermissionSummary } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

import {
  filterSupportedPermissionKeys,
  filterSupportedPermissions,
  getPermissionEditorLabels,
  getPermissionSummaryLabels,
  translatePermissionDescription,
  translatePermissionGroup,
} from "./role-permissions";

type RoleDefinition = {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  permissions: string[];
  created_at: string;
  updated_at: string;
};

type PermissionItem = {
  key: string;
  group: string;
  description: string;
};

type RoleEditManagementProps = {
  initialRole: RoleDefinition;
  initialPermissionCatalog: PermissionItem[];
};

export default function RoleEditManagement({ initialRole, initialPermissionCatalog }: RoleEditManagementProps) {
  const t = useTranslations("accessManagement");
  const locale = useLocale();
  const router = useRouter();
  const isRtl = locale.startsWith("fa");
  const BackIcon = isRtl ? ArrowRight : ArrowLeft;

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [query, setQuery] = useState("");
  const [viewMode, setViewMode] = useState<"selected" | "all">("selected");
  const [error, setError] = useState<string | null>(null);

  const [form, setForm] = useState({
    display_name: initialRole.display_name,
    description: initialRole.description ?? "",
    permissions: filterSupportedPermissionKeys(initialRole.permissions ?? []),
  });

  const filteredPermissionCatalog = useMemo(
    () => filterSupportedPermissions(initialPermissionCatalog),
    [initialPermissionCatalog],
  );
  const selectionStats = useMemo(() => getPermissionSelectionSummary(form.permissions, filteredPermissionCatalog), [form.permissions, filteredPermissionCatalog]);
  const numberFormatter = useMemo(() => new Intl.NumberFormat(locale), [locale]);
  const editorLabels = useMemo(() => getPermissionEditorLabels(locale), [locale]);
  const summaryLabels = useMemo(() => getPermissionSummaryLabels(locale), [locale]);

  async function submitEdit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!form.display_name.trim()) {
      setError(t("rolesPage.form.displayName"));
      return;
    }
    if (form.permissions.length === 0) {
      setError(locale.startsWith("fa") ? "حداقل یک مجوز انتخاب کنید." : "Select at least one permission.");
      return;
    }

    setError(null);
    setIsSubmitting(true);

    try {
      await requestJSON(`/api/roles/${encodeURIComponent(initialRole.id)}`, {
        method: "PATCH",
        body: JSON.stringify(form),
      });
      router.push("/roles");
      router.refresh();
    } catch (submitError) {
      setError(normalizeError(submitError, t("messages.roleUpdateError")));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <PageContainer padding="none">
      <PageHeader
        eyebrow={t("title")}
        title={isRtl ? "ویرایش نقش" : "Edit role"}
        description={isRtl ? "مشخصات نقش و سطح دسترسی آن را به‌روزرسانی کنید." : "Update the role details and its access level."}
        titleAddon={
          <Badge variant={initialRole.is_system ? "secondary" : "outline"}>
            {initialRole.is_system ? t("rolesPage.systemRole") : t("rolesPage.customRole")}
          </Badge>
        }
        actions={
          <>
            <Button variant="outline" asChild>
              <Link href="/roles">
                <BackIcon data-icon="inline-start" aria-hidden="true" />
                {t("rolesPage.actions.cancel")}
              </Link>
            </Button>
            <Button type="submit" form="role-edit-form" disabled={isSubmitting}>
              <Save data-icon="inline-start" aria-hidden="true" />
              {isSubmitting ? `${t("rolesPage.actions.save")}...` : t("rolesPage.actions.save")}
            </Button>
          </>
        }
      />

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>{t("alerts.error")}</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid min-w-0 gap-6 xl:grid-cols-[minmax(0,1fr)_18rem]">
        <form id="role-edit-form" className="flex min-w-0 flex-col gap-6" onSubmit={submitEdit}>
          <FormSection
            title={isRtl ? "مشخصات نقش" : "Role details"}
            description={isRtl ? "نام نمایشی و توضیحات قابل مشاهده برای مدیران را ویرایش کنید." : "Edit the display name and administrator-facing description."}
          >
            <div className="grid gap-4 md:grid-cols-2">
              <IdentifierField id="role-slug" label={isRtl ? "شناسه فنی" : "Technical identifier"} value={initialRole.name} readOnly description={null} />
              <DisplayNameField
                id="role-display-name"
                label={t("rolesPage.form.displayName")}
                value={form.display_name}
                required
                error={error && !form.display_name.trim() ? t("rolesPage.form.displayName") : undefined}
                onChange={(event) => setForm((prev) => ({ ...prev, display_name: event.target.value }))}
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="role-description">{t("rolesPage.form.description")}</Label>
              <Textarea
                id="role-description"
                value={form.description}
                className="min-h-24 resize-y"
                onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
              />
            </div>
          </FormSection>

          <FormSection
            title={isRtl ? "مجوزهای نقش" : "Role permissions"}
            description={isRtl ? "مجوزها را جست‌وجو کنید و دسترسی‌های این نقش را در سطح گروه یا مورد تغییر دهید." : "Search permissions and adjust this role at group or item level."}
            actions={<Badge variant="secondary">{numberFormatter.format(selectionStats.count)}</Badge>}
          >
            <PermissionEditor
              permissions={filteredPermissionCatalog}
              value={form.permissions}
              onValueChange={(permissions) => setForm((prev) => ({ ...prev, permissions }))}
              query={query}
              onQueryChange={setQuery}
              mode={viewMode}
              onModeChange={setViewMode}
              labels={editorLabels}
              dir={isRtl ? "rtl" : "ltr"}
              formatGroupLabel={(group) => translatePermissionGroup(group, locale)}
              formatPermissionDescription={(permission) => translatePermissionDescription(permission, locale)}
              formatNumber={(value) => numberFormatter.format(value)}
            />
          </FormSection>
        </form>

        <aside className="xl:sticky xl:top-6 xl:h-fit">
          <PermissionSummary
            permissions={filteredPermissionCatalog}
            value={form.permissions}
            title={isRtl ? "خلاصه نقش" : "Role summary"}
            description={<span dir="ltr" className="font-mono">{initialRole.name}</span>}
            labels={summaryLabels}
            onClear={() => setForm((prev) => ({ ...prev, permissions: [] }))}
            formatNumber={(value) => numberFormatter.format(value)}
          />
        </aside>
      </div>
    </PageContainer>
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
