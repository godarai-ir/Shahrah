"use client";

import { FormEvent, useState } from "react";
import { Activity } from "lucide-react";
import { useTranslations } from "next-intl";

import {
  AppTextField,
  DefinitionList,
  DefinitionListItem,
  FormDialog,
  OperationNotice,
  type OperationNoticeValue,
} from "@/components/app-ui";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import type { TenantProviderAccount } from "@/app/components/provider-account-wizard/types";
import { normalizeError, requestJSON, unwrapData } from "@/app/components/provider-account-wizard/utils";

export function ProviderConnectionTestDialog(props: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  account: TenantProviderAccount | null;
}) {
  const t = useTranslations("providerAccounts");
  const ux = useTranslations("providerAccounts.ux");
  const [modelHint, setModelHint] = useState("");
  const [notice, setNotice] = useState<OperationNoticeValue | null>(null);
  const [isTesting, setIsTesting] = useState(false);

  if (!props.account) return null;

  function resetDialogState() {
    setModelHint("");
    setNotice(null);
    setIsTesting(false);
  }

  function handleOpenChange(open: boolean) {
    if (!open) resetDialogState();
    props.onOpenChange(open);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.account) return;

    setIsTesting(true);
    setNotice(null);
    try {
      const trimmedModelHint = modelHint.trim();
      const result = await requestJSON<{ success: boolean; message?: string }>(
        `/api/tenant-provider-accounts/${encodeURIComponent(props.account.id)}/test-credentials`,
        {
          method: "POST",
          body: JSON.stringify({ model_hint: trimmedModelHint || undefined }),
        },
      );
      const payload = unwrapData<{ success: boolean; message?: string }>(result) ?? result;
      const message = payload.message ?? (payload.success ? t("messages.credentialValid") : t("messages.credentialInvalid"));
      setNotice({ kind: payload.success ? "success" : "error", message });
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.credentialTestFailed")) });
    } finally {
      setIsTesting(false);
    }
  }

  return (
    <FormDialog
      open={props.open}
      onOpenChange={handleOpenChange}
      title={ux("connectionTest.title")}
      description={ux("connectionTest.description", {
        name: props.account.name,
        provider: props.account.provider,
      })}
      size="md"
      pending={isTesting}
      footer={
        <Button type="submit" form="provider-connection-test-form" disabled={isTesting} aria-busy={isTesting}>
          {isTesting ? <Spinner data-icon="inline-start" /> : <Activity data-icon="inline-start" />}
          {isTesting ? ux("connectionTest.running") : ux("connectionTest.run")}
        </Button>
      }
    >
      <form id="provider-connection-test-form" className="flex flex-col gap-4" onSubmit={handleSubmit}>
        <DefinitionList columns={2}>
          <DefinitionListItem surface label={t("fields.provider")} value={props.account.provider} valueDir="ltr" />
          <DefinitionListItem surface label={t("fields.accountName")} value={props.account.name} valueDir="ltr" mono />
          <DefinitionListItem surface label={t("fields.authMode")} value={props.account.auth_mode} valueDir="ltr" />
          <DefinitionListItem
            surface
            label={ux("connectionTest.credentialHint")}
            value={props.account.credential_hint || ux("common.enabled")}
            valueDir="ltr"
            mono={Boolean(props.account.credential_hint)}
          />
        </DefinitionList>

        <AppTextField
          id="providerConnectionModelHint"
          label={ux("connectionTest.modelHintLabel")}
          placeholder={ux("connectionTest.modelHintPlaceholder")}
          value={modelHint}
          onChange={(event) => setModelHint(event.target.value)}
          disabled={isTesting}
          dir="ltr"
          className="font-mono"
        />

        <OperationNotice
          notice={notice}
          labels={{
            success: ux("connectionTest.successTitle"),
            error: ux("connectionTest.errorTitle"),
            warning: ux("connectionTest.warningTitle"),
            info: ux("connectionTest.infoTitle"),
          }}
        />
      </form>
    </FormDialog>
  );
}
