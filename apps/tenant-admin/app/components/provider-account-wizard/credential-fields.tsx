"use client";

import { useTranslations } from "next-intl";
import { Pencil } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import type { ProviderAuthMode } from "./types";
export function CredentialFields(props: {
  authMode: ProviderAuthMode;
  apiKey: string;
  setAPIKey: (value: string) => void;
  apiKeyTouched?: boolean;
  setAPIKeyTouched?: (value: boolean) => void;
  apiKeyHint?: string;
  secretRef: string;
  setSecretRef: (value: string) => void;
  azureTenantID: string;
  setAzureTenantID: (value: string) => void;
  azureClientID: string;
  setAzureClientID: (value: string) => void;
  awsAccessKeyID: string;
  setAWSAccessKeyID: (value: string) => void;
  awsSecretAccessKey: string;
  setAWSSecretAccessKey: (value: string) => void;
  awsRegion: string;
  setAWSRegion: (value: string) => void;
  googleServiceAccountJSON: string;
  setGoogleServiceAccountJSON: (value: string) => void;
  t: ReturnType<typeof useTranslations>;
  surface?: boolean;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const hasStoredAPIKey = props.authMode === "api_key" && !props.apiKeyTouched && !!props.apiKeyHint;

  function handleEditStoredAPIKey() {
    if (!hasStoredAPIKey) return;
    props.setAPIKeyTouched?.(true);
    props.setAPIKey("");
  }

  function handleAPIKeyChange(value: string) {
    props.setAPIKeyTouched?.(true);
    props.setAPIKey(value);
  }

  return (
    <section className={cn("space-y-4", props.surface === false ? "pt-1" : "rounded-lg border border-border/70 bg-background p-4 sm:p-5")}>
      <div className="space-y-1">
        <h4 className="text-base font-semibold">{ux("setupSteps.accountCredentials")}</h4>
        <p className="text-sm leading-6 text-muted-foreground">{props.t("step2.connectionDescription")}</p>
      </div>
      <div className="space-y-4">
        {props.authMode === "api_key" ? (
          <div className="grid gap-3 md:grid-cols-2">
            <Field>
              <FieldLabel htmlFor="providerApiKey">{props.t("fields.apiKeyOptional")}</FieldLabel>
              {hasStoredAPIKey ? (
                <div className="flex h-10 items-center justify-between gap-2 rounded-md border border-input bg-muted/30 px-3">
                  <span className="min-w-0 truncate font-mono text-sm text-foreground" dir="ltr">
                    {props.apiKeyHint}
                  </span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 shrink-0"
                    onClick={handleEditStoredAPIKey}
                    aria-label={ux("credentials.editStoredApiKey")}
                  >
                    <Pencil className="h-4 w-4" />
                  </Button>
                </div>
              ) : (
                <Input
                  id="providerApiKey"
                  value={props.apiKey}
                  onChange={(event) => handleAPIKeyChange(event.target.value)}
                  placeholder={ux("placeholders.apiKey")}
                />
              )}
            </Field>
            <Field>
              <FieldLabel htmlFor="providerSecretRef">{props.t("fields.secretRefOptional")}</FieldLabel>
              <Input id="providerSecretRef" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefProvider")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "azure_entra_cert" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field>
              <FieldLabel htmlFor="azureTenantID">{props.t("fields.azureTenantID")}</FieldLabel>
              <Input id="azureTenantID" value={props.azureTenantID} onChange={(event) => props.setAzureTenantID(event.target.value)} placeholder={ux("placeholders.tenantId")} />
            </Field>
            <Field>
              <FieldLabel htmlFor="azureClientID">{props.t("fields.azureClientID")}</FieldLabel>
              <Input id="azureClientID" value={props.azureClientID} onChange={(event) => props.setAzureClientID(event.target.value)} placeholder={ux("placeholders.clientId")} />
            </Field>
            <Field>
              <FieldLabel htmlFor="providerSecretRefAzure">{props.t("fields.secretRefOptional")}</FieldLabel>
              <Input id="providerSecretRefAzure" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefAzure")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "aws_sigv4" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field>
              <FieldLabel htmlFor="awsAccessKeyID">{props.t("fields.awsAccessKeyID")}</FieldLabel>
              <Input id="awsAccessKeyID" value={props.awsAccessKeyID} onChange={(event) => props.setAWSAccessKeyID(event.target.value)} placeholder={ux("placeholders.awsAccessKey")} />
            </Field>
            <Field>
              <FieldLabel htmlFor="awsSecretAccessKey">{props.t("fields.awsSecretAccessKey")}</FieldLabel>
              <Input id="awsSecretAccessKey" value={props.awsSecretAccessKey} onChange={(event) => props.setAWSSecretAccessKey(event.target.value)} placeholder={ux("placeholders.secret")} />
            </Field>
            <Field>
              <FieldLabel htmlFor="awsRegion">{props.t("fields.awsRegion")}</FieldLabel>
              <Input id="awsRegion" value={props.awsRegion} onChange={(event) => props.setAWSRegion(event.target.value)} placeholder={ux("placeholders.awsRegion")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "google_service_account" ? (
          <div className="grid gap-3">
            <Field>
              <FieldLabel htmlFor="googleServiceAccountJSON">{props.t("fields.googleServiceAccountJSON")}</FieldLabel>
              <textarea
                id="googleServiceAccountJSON"
                value={props.googleServiceAccountJSON}
                onChange={(event) => props.setGoogleServiceAccountJSON(event.target.value)}
                placeholder={ux("placeholders.googleServiceJson")}
                className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </Field>
            <Field>
              <FieldLabel htmlFor="providerSecretRefGoogle">{props.t("fields.secretRefOptional")}</FieldLabel>
              <Input id="providerSecretRefGoogle" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefGoogle")} />
            </Field>
          </div>
        ) : null}
      </div>
    </section>
  );
}


