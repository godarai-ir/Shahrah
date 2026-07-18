"use client";

import { ExternalLink } from "lucide-react";
import type { ChangeEvent } from "react";

import { AppCodeTextareaField, AppTextField } from "@/components/app-ui";
import type { ProviderConnectionSchema, ProviderConnectionSchemaField } from "./types";
import type { ProviderCredentialValues } from "./credential-schema";

function fieldDescription(field: ProviderConnectionSchemaField, writeOnlyLabel: string, documentationLabel: string) {
  const text = [field.help_text, field.secret ? writeOnlyLabel : null].filter(Boolean).join(" ");
  if (!field.documentation_url) return text || undefined;
  return (
    <span className="flex flex-wrap items-center gap-2">
      {text ? <span>{text}</span> : null}
      <a
        href={field.documentation_url}
        target="_blank"
        rel="noreferrer"
        className="inline-flex items-center gap-1 font-medium text-primary underline-offset-4 hover:underline"
      >
        {documentationLabel}
        <ExternalLink className="size-3.5" aria-hidden="true" />
      </a>
    </span>
  );
}

export function ProviderCredentialSchemaFields({
  schema,
  values,
  onValueChange,
  disabled,
  connectionFieldsTitle,
  credentialFieldsTitle,
  writeOnlyLabel,
  documentationLabel,
}: {
  schema: ProviderConnectionSchema;
  values: ProviderCredentialValues;
  onValueChange: (key: string, value: string) => void;
  disabled?: boolean;
  connectionFieldsTitle: string;
  credentialFieldsTitle: string;
  writeOnlyLabel: string;
  documentationLabel: string;
}) {
  const groups = [
    { scope: "connection" as const, title: connectionFieldsTitle },
    { scope: "credential" as const, title: credentialFieldsTitle },
  ];

  return (
    <div className="flex flex-col gap-5">
      {groups.map((group) => {
        const fields = schema.fields.filter((field) => field.scope === group.scope);
        if (fields.length === 0) return null;
        return (
          <section key={group.scope} className="flex flex-col gap-3" aria-labelledby={`provider-${group.scope}-fields`}>
            <h4 id={`provider-${group.scope}-fields`} className="text-sm font-semibold">{group.title}</h4>
            <div className="grid gap-4 md:grid-cols-2">
              {fields.map((field) => {
                const label = `${field.label}${field.required ? " *" : ""}`;
                const shared = {
                  id: `provider-credential-${field.key}`,
                  label,
                  description: fieldDescription(field, writeOnlyLabel, documentationLabel),
                  value: values[field.key] ?? "",
                  disabled,
                  required: field.required,
                  onChange: (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => onValueChange(field.key, event.target.value),
                };
                if (field.type === "textarea") {
                  return <AppCodeTextareaField key={field.key} {...shared} className="min-h-32" />;
                }
                return (
                  <AppTextField
                    key={field.key}
                    {...shared}
                    type={field.secret || field.type === "password" ? "password" : field.type === "url" ? "url" : "text"}
                    dir="ltr"
                    autoComplete={field.secret ? "new-password" : "off"}
                    className="text-start"
                  />
                );
              })}
            </div>
          </section>
        );
      })}
    </div>
  );
}
