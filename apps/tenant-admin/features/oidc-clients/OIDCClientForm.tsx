"use client";

import { FormEvent, useState } from "react";
import { useTranslations } from "next-intl";

import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { AppTextField } from "@/components/app-ui/AppTextField";
import { FormSheet } from "@/components/app-ui/FormSheet";
import { FormSheetActions } from "@/components/app-ui/FormSheetActions";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { FieldGroup } from "@/components/ui/field";
import { Button } from "@/components/ui/button";

import type { OIDCClient } from "./api";
import { MAX_CLIENT_NAME_LENGTH, parseOIDCClientInput, splitURIList, type OIDCClientInput } from "./input";

const FORM_ID = "oidc-client-form";

type FieldErrors = Partial<Record<"name" | "redirect_uris" | "post_logout_redirect_uris", string>>;

export type OIDCClientFormProps = {
  open: boolean;
  /** The client being edited; omit to create a new one. */
  client: OIDCClient | null;
  pending: boolean;
  notice: OperationNoticeValue | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (input: OIDCClientInput) => void;
};

/** Keyed by the parent on open, so its local state always starts from the client being edited. */
export function OIDCClientForm({ open, client, pending, notice, onOpenChange, onSubmit }: OIDCClientFormProps) {
  const t = useTranslations("oidcClients");
  const [name, setName] = useState(client?.name ?? "");
  const [redirectURIs, setRedirectURIs] = useState((client?.redirect_uris ?? []).join("\n"));
  const [logoutURIs, setLogoutURIs] = useState((client?.post_logout_redirect_uris ?? []).join("\n"));
  const [active, setActive] = useState(client?.active ?? true);
  const [errors, setErrors] = useState<FieldErrors>({});

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = parseOIDCClientInput(
      {
        name,
        redirect_uris: splitURIList(redirectURIs),
        post_logout_redirect_uris: splitURIList(logoutURIs),
        ...(client ? { active } : {}),
      },
      { partial: Boolean(client) },
    );

    if (!result.ok) {
      const field = result.field === "active" || !result.field ? "name" : result.field;
      setErrors({ [field]: t(`errors.${result.code}`, { value: result.value ?? "" }) });
      return;
    }

    setErrors({});
    onSubmit(result.value);
  }

  const notices = { success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") };

  return (
    <FormSheet
      open={open}
      onOpenChange={onOpenChange}
      title={client ? t("form.editTitle") : t("form.createTitle")}
      description={client ? t("form.editDescription") : t("form.createDescription")}
      size="md"
      pending={pending}
      closeLabel={t("actions.close")}
      footer={
        <FormSheetActions
          secondary={
            <Button type="button" variant="outline" disabled={pending} onClick={() => onOpenChange(false)}>
              {t("actions.cancel")}
            </Button>
          }
          primary={
            <Button type="submit" form={FORM_ID} disabled={pending} aria-busy={pending}>
              {pending ? t("actions.saving") : client ? t("actions.save") : t("actions.create")}
            </Button>
          }
        />
      }
    >
      <OperationNotice notice={notice} labels={notices} />
      <form id={FORM_ID} noValidate onSubmit={submit}>
        <FieldGroup>
          <AppTextField
            label={t("fields.name")}
            description={t("fields.nameHelp")}
            value={name}
            maxLength={MAX_CLIENT_NAME_LENGTH}
            onChange={(event) => setName(event.target.value)}
            placeholder={t("fields.namePlaceholder")}
            error={errors.name}
            aria-invalid={Boolean(errors.name)}
            autoFocus
          />
          <AppCodeTextareaField
            label={t("fields.redirectURIs")}
            description={t("fields.redirectURIsHelp")}
            value={redirectURIs}
            onChange={(event) => setRedirectURIs(event.target.value)}
            placeholder="https://chat.example.com/oauth/openid/callback"
            rows={3}
            error={errors.redirect_uris}
            aria-invalid={Boolean(errors.redirect_uris)}
          />
          <AppCodeTextareaField
            label={t("fields.logoutURIs")}
            description={t("fields.logoutURIsHelp")}
            value={logoutURIs}
            onChange={(event) => setLogoutURIs(event.target.value)}
            placeholder="https://chat.example.com/login"
            rows={2}
            error={errors.post_logout_redirect_uris}
            aria-invalid={Boolean(errors.post_logout_redirect_uris)}
          />
          {client ? (
            <AppSwitchField
              label={t("fields.active")}
              description={t("fields.activeHelp")}
              checked={active}
              onCheckedChange={setActive}
            />
          ) : null}
        </FieldGroup>
      </form>
    </FormSheet>
  );
}
