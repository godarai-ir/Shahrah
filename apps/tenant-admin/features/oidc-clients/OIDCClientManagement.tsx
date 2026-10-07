"use client";

import { useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import type { ColumnDef } from "@tanstack/react-table";
import { AppWindow, Plus } from "lucide-react";

import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { DataTableToolbar } from "@/components/app-ui/DataTableToolbar";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { RowActionsMenu } from "@/components/app-ui/RowActionsMenu";
import { SearchField } from "@/components/app-ui/SearchField";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { TypedConfirmDialog } from "@/components/app-ui/TypedConfirmDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/ui/data-table";
import { APIClientError } from "@/lib/http/api-client";

import {
  createOIDCClient,
  deleteOIDCClient,
  listOIDCClients,
  rotateOIDCClientSecret,
  updateOIDCClient,
  type OIDCClient,
} from "./api";
import { IntegrationDetails } from "./IntegrationDetails";
import { OIDCClientForm } from "./OIDCClientForm";
import type { OIDCClientInput } from "./input";

const KNOWN_ERROR_CODES = new Set(["oidc_client_invalid", "oidc_client_not_found", "oidc_client_conflict"]);

type SecretReveal = { clientName: string; clientID: string; secret: string; rotated: boolean };
type FormState = { open: boolean; client: OIDCClient | null; key: number };

export type OIDCClientManagementProps = {
  initialClients: OIDCClient[];
  /** Set when the initial load failed; "forbidden" means the caller is not a tenant admin. */
  initialError: "forbidden" | "unavailable" | null;
  issuerURL: string;
};

export default function OIDCClientManagement({ initialClients, initialError, issuerURL }: OIDCClientManagementProps) {
  const t = useTranslations("oidcClients");
  const locale = useLocale();

  const [clients, setClients] = useState(initialClients);
  const [query, setQuery] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [banner, setBanner] = useState<OperationNoticeValue | null>(
    initialError === "unavailable" ? { kind: "error", message: t("messages.loadError") } : null,
  );
  const [form, setForm] = useState<FormState>({ open: false, client: null, key: 0 });
  const [formNotice, setFormNotice] = useState<OperationNoticeValue | null>(null);
  const [saving, setSaving] = useState(false);
  const [rotateTarget, setRotateTarget] = useState<OIDCClient | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<OIDCClient | null>(null);
  const [dialogNotice, setDialogNotice] = useState<OperationNoticeValue | null>(null);
  const [dialogPending, setDialogPending] = useState(false);
  const [secret, setSecret] = useState<SecretReveal | null>(null);
  const [pendingToggleID, setPendingToggleID] = useState<string | null>(null);

  const noticeLabels = { success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") };

  function errorMessage(error: unknown, fallbackKey: string): string {
    if (error instanceof APIClientError && error.code && KNOWN_ERROR_CODES.has(error.code)) {
      return t(`apiErrors.${error.code}`);
    }
    if (error instanceof APIClientError && error.status === 403) return t("apiErrors.forbidden");
    return t(fallbackKey);
  }

  const visibleClients = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase();
    if (!needle) return clients;
    return clients.filter((client) =>
      [client.name, client.client_id, ...client.redirect_uris].some((value) => value.toLocaleLowerCase().includes(needle)),
    );
  }, [clients, query]);

  async function reload() {
    setIsLoading(true);
    try {
      setClients(await listOIDCClients());
      setBanner(null);
    } catch (error) {
      setBanner({ kind: "error", message: errorMessage(error, "messages.loadError") });
    } finally {
      setIsLoading(false);
    }
  }

  function openForm(client: OIDCClient | null) {
    setFormNotice(null);
    setForm((current) => ({ open: true, client, key: current.key + 1 }));
  }

  async function submitForm(input: OIDCClientInput) {
    setSaving(true);
    setFormNotice(null);
    try {
      if (form.client) {
        await updateOIDCClient(form.client.id, input);
        showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.updateSuccess") });
      } else {
        const created = await createOIDCClient(input);
        setSecret({ clientName: created.name, clientID: created.client_id, secret: created.client_secret, rotated: false });
      }
      setForm((current) => ({ ...current, open: false }));
      await reload();
    } catch (error) {
      setFormNotice({ kind: "error", message: errorMessage(error, form.client ? "messages.updateError" : "messages.createError") });
    } finally {
      setSaving(false);
    }
  }

  async function toggleActive(client: OIDCClient) {
    setPendingToggleID(client.id);
    try {
      await updateOIDCClient(client.id, { active: !client.active });
      showFeedbackToast({
        kind: "success",
        title: t("alerts.success"),
        description: client.active ? t("messages.deactivateSuccess") : t("messages.activateSuccess"),
      });
      await reload();
    } catch (error) {
      showFeedbackToast({ kind: "error", title: t("alerts.error"), description: errorMessage(error, "messages.updateError") });
    } finally {
      setPendingToggleID(null);
    }
  }

  function closeDialogs() {
    setRotateTarget(null);
    setDeleteTarget(null);
    setDialogNotice(null);
  }

  async function confirmRotate() {
    if (!rotateTarget) return;
    setDialogPending(true);
    try {
      const rotated = await rotateOIDCClientSecret(rotateTarget.id);
      closeDialogs();
      setSecret({ clientName: rotated.name, clientID: rotated.client_id, secret: rotated.client_secret, rotated: true });
      await reload();
    } catch (error) {
      setDialogNotice({ kind: "error", message: errorMessage(error, "messages.rotateError") });
    } finally {
      setDialogPending(false);
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) return;
    setDialogPending(true);
    try {
      await deleteOIDCClient(deleteTarget.id);
      closeDialogs();
      showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.deleteSuccess") });
      await reload();
    } catch (error) {
      setDialogNotice({ kind: "error", message: errorMessage(error, "messages.deleteError") });
    } finally {
      setDialogPending(false);
    }
  }

  const copyFailed = () =>
    showFeedbackToast({ kind: "error", title: t("alerts.error"), description: t("messages.copyError") });

  const columns: ColumnDef<OIDCClient>[] = [
    {
      id: "name",
      header: t("table.name"),
      meta: { label: t("table.name"), headerClassName: "w-[22%]" },
      cell: ({ row }) => (
        <span className="block truncate font-medium" dir="auto" title={row.original.name}>
          {row.original.name}
        </span>
      ),
    },
    {
      id: "clientID",
      header: t("table.clientID"),
      meta: { label: t("table.clientID") },
      cell: ({ row }) => (
        <span className="flex items-center gap-1">
          <code dir="ltr" className="truncate font-mono text-xs">{row.original.client_id}</code>
          <CopyAction
            value={row.original.client_id}
            copyLabel={t("actions.copyClientID")}
            copiedLabel={t("actions.copied")}
            size="icon-xs"
            onCopyError={copyFailed}
          />
        </span>
      ),
    },
    {
      id: "redirects",
      header: t("table.redirectURIs"),
      meta: { label: t("table.redirectURIs"), optional: true },
      cell: ({ row }) => {
        const [first, ...rest] = row.original.redirect_uris;
        return (
          <span className="flex min-w-0 items-center gap-2" title={row.original.redirect_uris.join("\n")}>
            <span dir="ltr" className="truncate font-mono text-xs text-muted-foreground">{first ?? "-"}</span>
            {rest.length > 0 ? (
              <span className="shrink-0 text-xs text-muted-foreground">{t("table.moreURIs", { count: rest.length })}</span>
            ) : null}
          </span>
        );
      },
    },
    {
      id: "secretRotatedAt",
      header: t("table.secretRotatedAt"),
      meta: { label: t("table.secretRotatedAt"), optional: true },
      cell: ({ row }) => (
        <DateTimeValue value={row.original.secret_rotated_at} locale={locale} fallback="-" className="text-muted-foreground" />
      ),
    },
    {
      id: "status",
      header: t("table.status"),
      meta: { label: t("table.status") },
      cell: ({ row }) => (
        <StatusBadge status={row.original.active ? "active" : "inactive"}>
          {row.original.active ? t("status.active") : t("status.inactive")}
        </StatusBadge>
      ),
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{t("table.actions")}</span>,
      meta: { headerClassName: "w-12", cellClassName: "w-12 text-end" },
      cell: ({ row }) => {
        const client = row.original;
        return (
          <RowActionsMenu
            label={t("actions.menu", { name: client.name })}
            actions={[
              { key: "edit", label: t("actions.edit"), onSelect: () => openForm(client) },
              {
                key: "toggle",
                label: client.active ? t("actions.deactivate") : t("actions.activate"),
                onSelect: () => void toggleActive(client),
                pending: pendingToggleID === client.id,
                pendingLabel: t("actions.saving"),
              },
              { key: "rotate", label: t("actions.rotateSecret"), onSelect: () => setRotateTarget(client) },
              { key: "delete", label: t("actions.delete"), onSelect: () => setDeleteTarget(client), destructive: true, separatorBefore: true },
            ]}
          />
        );
      },
    },
  ];

  if (initialError === "forbidden") {
    return (
      <EmptyState
        icon={AppWindow}
        variant="surface"
        title={t("forbidden.title")}
        description={t("forbidden.description")}
      />
    );
  }

  return (
    <div className="flex min-w-0 flex-col gap-6">
      <IntegrationDetails issuerURL={issuerURL} onCopyError={copyFailed} />

      <div className="flex min-w-0 flex-col gap-4">
        <OperationNotice notice={banner} labels={noticeLabels} />

        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
          <Button type="button" className="w-full sm:w-auto" onClick={() => openForm(null)}>
            <Plus data-icon="inline-start" />
            {t("actions.create")}
          </Button>
        </div>

        <DataTable
          tableId="oidc-clients"
          columns={columns}
          data={visibleClients}
          state={isLoading && clients.length === 0 ? "loading" : "ready"}
          skeletonRows={3}
          emptyLabel={t("table.empty")}
          noResultsSlot={
            <EmptyState
              icon={AppWindow}
              variant="table"
              title={clients.length === 0 ? t("table.empty") : t("table.noResults")}
              description={clients.length === 0 ? t("table.emptyDescription") : undefined}
            />
          }
          getRowId={(client) => client.id}
          toolbar={
            <DataTableToolbar
              search={
                <SearchField
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder={t("table.searchPlaceholder")}
                />
              }
            />
          }
        />
      </div>

      <OIDCClientForm
        key={form.key}
        open={form.open}
        client={form.client}
        pending={saving}
        notice={formNotice}
        onOpenChange={(open) => {
          if (saving) return;
          setForm((current) => ({ ...current, open }));
          if (!open) setFormNotice(null);
        }}
        onSubmit={(input) => void submitForm(input)}
      />

      <ConfirmDialog
        open={Boolean(rotateTarget)}
        onOpenChange={(open) => !open && !dialogPending && closeDialogs()}
        title={t("rotateDialog.title")}
        description={t("rotateDialog.description", { name: rotateTarget?.name ?? "" })}
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("rotateDialog.confirm")}
        pendingLabel={t("rotateDialog.pending")}
        pending={dialogPending}
        closeOnConfirm={false}
        onConfirm={confirmRotate}
        feedback={<OperationNotice notice={dialogNotice} labels={noticeLabels} />}
      />

      <TypedConfirmDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(open) => !open && !dialogPending && closeDialogs()}
        title={t("deleteDialog.title")}
        description={t("deleteDialog.description", { name: deleteTarget?.name ?? "" })}
        expectedValue={deleteTarget?.name ?? ""}
        inputLabel={t("deleteDialog.inputLabel", { name: deleteTarget?.name ?? "" })}
        inputDir="auto"
        cancelLabel={t("actions.cancel")}
        confirmLabel={t("deleteDialog.confirm")}
        pendingLabel={t("deleteDialog.pending")}
        onConfirm={confirmDelete}
        feedback={<OperationNotice notice={dialogNotice} labels={noticeLabels} />}
      />

      <OneTimeSecretDialog
        open={Boolean(secret)}
        onOpenChange={(open) => !open && setSecret(null)}
        value={secret?.secret ?? null}
        title={secret?.rotated ? t("secret.rotatedTitle") : t("secret.createdTitle")}
        description={t("secret.description", { name: secret?.clientName ?? "", clientID: secret?.clientID ?? "" })}
        warning={t("secret.warning")}
        closeLabel={t("actions.close")}
        copyLabels={{
          valueAriaLabel: t("secret.valueLabel"),
          copy: t("actions.copy"),
          copying: t("actions.copying"),
          copied: t("actions.copied"),
        }}
        onClear={() => setSecret(null)}
        onCopyError={copyFailed}
      />
    </div>
  );
}
