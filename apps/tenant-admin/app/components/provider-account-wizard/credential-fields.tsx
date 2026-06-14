"use client";

import { useTranslations } from "next-intl";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { ProviderAuthMode } from "./types";
import { Field } from "./field";
export function CredentialFields(props: {
  authMode: ProviderAuthMode;
  apiKey: string;
  setAPIKey: (value: string) => void;
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
}) {
  const ux = useTranslations("providerAccounts.ux");
  return (
    <Card className="border-border/70">
      <CardHeader>
        <CardTitle className="text-base">{ux("setupSteps.accountCredentials")}</CardTitle>
        <CardDescription>{props.t("step2.connectionDescription")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {props.authMode === "api_key" ? (
          <div className="grid gap-3 md:grid-cols-2">
            <Field label={props.t("fields.apiKeyOptional")} id="providerApiKey">
              <Input id="providerApiKey" value={props.apiKey} onChange={(event) => props.setAPIKey(event.target.value)} placeholder={ux("placeholders.apiKey")} />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRef">
              <Input id="providerSecretRef" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefProvider")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "azure_entra_cert" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field label={props.t("fields.azureTenantID")} id="azureTenantID">
              <Input id="azureTenantID" value={props.azureTenantID} onChange={(event) => props.setAzureTenantID(event.target.value)} placeholder={ux("placeholders.tenantId")} />
            </Field>
            <Field label={props.t("fields.azureClientID")} id="azureClientID">
              <Input id="azureClientID" value={props.azureClientID} onChange={(event) => props.setAzureClientID(event.target.value)} placeholder={ux("placeholders.clientId")} />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRefAzure">
              <Input id="providerSecretRefAzure" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefAzure")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "aws_sigv4" ? (
          <div className="grid gap-3 md:grid-cols-3">
            <Field label={props.t("fields.awsAccessKeyID")} id="awsAccessKeyID">
              <Input id="awsAccessKeyID" value={props.awsAccessKeyID} onChange={(event) => props.setAWSAccessKeyID(event.target.value)} placeholder={ux("placeholders.awsAccessKey")} />
            </Field>
            <Field label={props.t("fields.awsSecretAccessKey")} id="awsSecretAccessKey">
              <Input id="awsSecretAccessKey" value={props.awsSecretAccessKey} onChange={(event) => props.setAWSSecretAccessKey(event.target.value)} placeholder={ux("placeholders.secret")} />
            </Field>
            <Field label={props.t("fields.awsRegion")} id="awsRegion">
              <Input id="awsRegion" value={props.awsRegion} onChange={(event) => props.setAWSRegion(event.target.value)} placeholder={ux("placeholders.awsRegion")} />
            </Field>
          </div>
        ) : null}

        {props.authMode === "google_service_account" ? (
          <div className="grid gap-3">
            <Field label={props.t("fields.googleServiceAccountJSON")} id="googleServiceAccountJSON">
              <textarea
                id="googleServiceAccountJSON"
                value={props.googleServiceAccountJSON}
                onChange={(event) => props.setGoogleServiceAccountJSON(event.target.value)}
                placeholder={ux("placeholders.googleServiceJson")}
                className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </Field>
            <Field label={props.t("fields.secretRefOptional")} id="providerSecretRefGoogle">
              <Input id="providerSecretRefGoogle" value={props.secretRef} onChange={(event) => props.setSecretRef(event.target.value)} placeholder={ux("placeholders.secretRefGoogle")} />
            </Field>
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}


