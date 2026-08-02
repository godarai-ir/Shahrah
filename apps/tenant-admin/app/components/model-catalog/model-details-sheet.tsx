"use client";

import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { useMemo } from "react";
import { ExternalLink, FlaskConical, ScrollText, Settings2, TriangleAlert } from "lucide-react";

import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildPowerShellCurlGuideCommand,
  getCurlGuideEndpoint,
  getTenantApiBaseURL,
  resolveCurlGuideMode,
} from "@/app/components/provider-account-wizard/utils";
import { buildPlaygroundHref } from "@/app/components/playground-mode";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CodePanel } from "@/components/app-ui/CodePanel";
import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import type { CatalogModel } from "./types";
import { formatCatalogModelDisplayName } from "./formatters";

type ModelDetailsSheetProps = {
  model: CatalogModel | null;
  accountID: string;
  onAccountChange: (accountID: string) => void;
  onOpenChange: (open: boolean) => void;
};

export function ModelDetailsSheet({
  model,
  accountID,
  onAccountChange,
  onOpenChange,
}: ModelDetailsSheetProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();
  const selectedAvailability = useMemo(() => {
    if (!model) return null;
    return model.availability.find((item) => item.account_id === accountID) ?? model.availability[0] ?? null;
  }, [accountID, model]);

  const code = useMemo(() => {
    if (!model || !selectedAvailability) return null;
    const mode = resolveCurlGuideMode({
      provider: model.provider,
      model: model.model,
      display_name: model.display_name,
      mode: model.mode,
      types: model.types,
      capabilities: model.capabilities,
      features: model.features,
      input_modalities: model.input_modalities,
      output_modalities: model.output_modalities,
    });
    const endpoint = getCurlGuideEndpoint(mode);
    const baseUrl = getTenantApiBaseURL();
    const payload = buildCurlGuidePayload(mode, selectedAvailability.fqn);
    return {
      mode,
      bash: buildBashCurlGuideCommand({ baseUrl, endpoint, payload }),
      powershell: buildPowerShellCurlGuideCommand({
        baseUrl,
        endpoint,
        mode,
        modelFqn: selectedAvailability.fqn,
      }),
    };
  }, [model, selectedAvailability]);

  const playgroundHref = selectedAvailability && code
    ? buildPlaygroundHref(selectedAvailability.fqn, code.mode)
    : null;
  const logsHref = selectedAvailability && model
    ? `/logs?${new URLSearchParams({ model: selectedAvailability.fqn, provider: model.provider }).toString()}`
    : "/logs";
  const manageHref = selectedAvailability
    ? `/provider-accounts?${new URLSearchParams({ account: selectedAvailability.account_id, section: "models" }).toString()}`
    : "/provider-accounts";

  return (
    <Sheet open={Boolean(model)} onOpenChange={onOpenChange}>
      <SheetContent
        side={locale === "fa" ? "left" : "right"}
        closeLabel={t("actions.close")}
        className="flex min-w-0 w-full flex-col overflow-hidden p-0 sm:max-w-2xl"
      >
        <SheetHeader className="shrink-0 gap-2 px-6 pt-6 pe-12 text-start">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <SheetTitle
              className="min-w-0 truncate"
              dir="ltr"
              title={model ? formatCatalogModelDisplayName(model.model, model.display_name) : undefined}
            >
              {model ? formatCatalogModelDisplayName(model.model, model.display_name) : t("drawer.title")}
            </SheetTitle>
            {model?.is_deprecated ? <StatusBadge status="warning">{t("status.deprecated")}</StatusBadge> : null}
          </div>
          <SheetDescription>{t("drawer.description")}</SheetDescription>
        </SheetHeader>

        {model && selectedAvailability ? (
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-6 overflow-y-auto px-6 pb-6">
            <div className="flex flex-col gap-2">
              <AppSelectField
                id="model-catalog-account"
                label={t("drawer.account")}
                value={selectedAvailability.account_id}
                onValueChange={onAccountChange}
                optionTextDir="ltr"
                options={model.availability.map((availability) => ({
                  value: availability.account_id,
                  label: availability.account_name,
                }))}
              />
            </div>

            <div className="rounded-lg border border-border bg-card p-4">
              <p className="mb-2 text-xs font-medium text-muted-foreground">{t("drawer.fqn")}</p>
              <CopyableValue
                value={selectedAvailability.fqn}
                labels={{
                  valueAriaLabel: t("drawer.fqn"),
                  copy: t("actions.copyFqn"),
                  copying: t("actions.copying"),
                  copied: t("actions.copied"),
                }}
              />
            </div>

            {model.is_deprecated ? (
              <OperationNotice
                notice={{
                  kind: "warning",
                  message: model.replacement_model
                    ? t("drawer.deprecatedWithReplacement", { replacement: model.replacement_model })
                    : t("drawer.deprecated"),
                }}
                labels={{
                  success: t("notices.success"),
                  error: t("notices.error"),
                  warning: t("notices.warning"),
                  info: t("notices.info"),
                }}
                action={<TriangleAlert aria-hidden="true" />}
              />
            ) : null}

            <DefinitionList columns={2}>
              <DefinitionListItem label={t("columns.provider")} value={model.provider} valueDir="ltr" mono surface />
              <DefinitionListItem label={t("columns.context")} value={formatContext(model.context_window, t("values.na"))} valueDir="ltr" surface />
              <DefinitionListItem label={t("drawer.inputModalities")} value={formatModalityList(model.input_modalities, t)} surface />
              <DefinitionListItem label={t("drawer.outputModalities")} value={formatModalityList(model.output_modalities, t)} surface />
              <DefinitionListItem label={t("drawer.capabilities")} value={formatCapabilityList(selectedAvailability.effective_capabilities, t)} valueDir="auto" surface />
              <DefinitionListItem label={t("drawer.credentialSource")} value={t(`credential.${credentialKey(selectedAvailability.credential_source)}`)} surface />
              <DefinitionListItem label={t("drawer.publicInputPrice")} value={formatPrice(model.public_pricing?.input_usd_per_1m, t("values.na"))} valueDir="ltr" surface />
              <DefinitionListItem label={t("drawer.publicOutputPrice")} value={formatPrice(model.public_pricing?.output_usd_per_1m, t("values.na"))} valueDir="ltr" surface />
              <DefinitionListItem label={t("drawer.effectiveInputPrice")} value={formatPrice(selectedAvailability.effective_pricing?.input_usd_per_1m, t("values.na"))} valueDir="ltr" surface />
              <DefinitionListItem label={t("drawer.effectiveOutputPrice")} value={formatPrice(selectedAvailability.effective_pricing?.output_usd_per_1m, t("values.na"))} valueDir="ltr" surface />
            </DefinitionList>

            <PricingRatesTable
              publicRates={model.public_pricing?.rates ?? []}
              effectiveRates={selectedAvailability.effective_pricing?.rates ?? []}
              t={t}
            />

            <CapabilityReadinessTable
              rows={selectedAvailability.capability_readiness ?? model.readiness ?? []}
              t={t}
            />

            {code ? (
              <div className="flex flex-col gap-4">
                <CodePanel
                  title={t("drawer.bashExample")}
                  displayLanguage="Bash"
                  language="bash"
                  code={code.bash}
                  copy={{
                    copyLabel: t("actions.copyCode"),
                    copyingLabel: t("actions.copying"),
                    copiedLabel: t("actions.copied"),
                  }}
                />
                <CodePanel
                  title={t("drawer.powershellExample")}
                  displayLanguage="PowerShell"
                  language="powershell"
                  code={code.powershell}
                  copy={{
                    copyLabel: t("actions.copyCode"),
                    copyingLabel: t("actions.copying"),
                    copiedLabel: t("actions.copied"),
                  }}
                />
              </div>
            ) : null}
          </div>
        ) : null}
        {model && selectedAvailability ? (
          <SheetFooter className="shrink-0 border-t border-border bg-background p-4 sm:flex-row sm:flex-wrap">
            {playgroundHref ? (
              <Button asChild>
                <Link href={playgroundHref}>
                  <FlaskConical data-icon="inline-start" aria-hidden="true" />
                  {t("actions.openPlayground")}
                </Link>
              </Button>
            ) : null}
            <Button asChild variant="outline">
              <Link href={logsHref}>
                <ScrollText data-icon="inline-start" aria-hidden="true" />
                {t("actions.viewLogs")}
              </Link>
            </Button>
            <Button asChild variant="outline">
              <Link href={manageHref}>
                <Settings2 data-icon="inline-start" aria-hidden="true" />
                {t("actions.manageModel")}
              </Link>
            </Button>
            <Button asChild variant="ghost">
              <a href="/docs/guides/tenant-team-account-model" target="_blank" rel="noreferrer">
                <ExternalLink data-icon="inline-end" aria-hidden="true" />
                {t("actions.docs")}
              </a>
            </Button>
          </SheetFooter>
        ) : null}
      </SheetContent>
    </Sheet>
  );
}

type CatalogTranslator = ReturnType<typeof useTranslations>;

function PricingRatesTable({
  publicRates,
  effectiveRates,
  t,
}: {
  publicRates: NonNullable<CatalogModel["public_pricing"]>["rates"];
  effectiveRates: NonNullable<CatalogModel["public_pricing"]>["rates"];
  t: CatalogTranslator;
}) {
  const effectiveByKey = new Map((effectiveRates ?? []).map((rate) => [rate.key, rate]));
  if (!publicRates?.length) return null;

  return (
    <section className="space-y-3" aria-labelledby="model-pricing-rates-title">
      <div>
        <h3 id="model-pricing-rates-title" className="text-sm font-semibold">{t("drawer.ratesTitle")}</h3>
        <p className="text-xs text-muted-foreground">{t("drawer.ratesDescription")}</p>
      </div>
      <div className="overflow-x-auto rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("rates.dimension")}</TableHead>
              <TableHead>{t("rates.unit")}</TableHead>
              <TableHead>{t("rates.public")}</TableHead>
              <TableHead>{t("rates.effective")}</TableHead>
              <TableHead>{t("rates.source")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {publicRates.map((rate) => {
              const effective = effectiveByKey.get(rate.key) ?? rate;
              const inherited = effective.unit_price_usd === rate.unit_price_usd;
              return (
                <TableRow key={rate.key}>
                  <TableCell dir="ltr" className="font-mono text-xs">{rate.dimension}</TableCell>
                  <TableCell dir="ltr">{rate.unit}</TableCell>
                  <TableCell dir="ltr">{rate.unit_price_usd}</TableCell>
                  <TableCell dir="ltr">{effective.unit_price_usd}</TableCell>
                  <TableCell>
                    <StatusBadge status={inherited ? "neutral" : "warning"}>
                      {inherited ? t("rates.inherited") : t("rates.overridden")}
                    </StatusBadge>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
    </section>
  );
}

function CapabilityReadinessTable({
  rows,
  t,
}: {
  rows: NonNullable<CatalogModel["readiness"]>;
  t: CatalogTranslator;
}) {
  if (!rows.length) return null;
  return (
    <section className="space-y-3" aria-labelledby="model-capability-readiness-title">
      <div>
        <h3 id="model-capability-readiness-title" className="text-sm font-semibold">{t("drawer.capabilityReadinessTitle")}</h3>
        <p className="text-xs text-muted-foreground">{t("drawer.capabilityReadinessDescription")}</p>
      </div>
      <div className="overflow-x-auto rounded-lg border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("capabilityReadiness.endpoint")}</TableHead>
              <TableHead>{t("capabilityReadiness.modality")}</TableHead>
              <TableHead>{t("capabilityReadiness.transport")}</TableHead>
              <TableHead>{t("capabilityReadiness.contract")}</TableHead>
              <TableHead>{t("capabilityReadiness.live")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={`${row.method ?? ""}:${row.endpoint}:${row.transport ?? ""}:${index}`}>
                <TableCell dir="ltr" className="font-mono text-xs">
                  {[row.method, row.endpoint].filter(Boolean).join(" ")}
                </TableCell>
                <TableCell>
                  {[row.input_modality, row.output_modality].filter(Boolean).join(" → ") || t("values.na")}
                </TableCell>
                <TableCell dir="ltr">{row.transport ?? "http"}</TableCell>
                <TableCell>
                  <StatusBadge status={row.state === "contract_ready" ? "success" : "warning"}>
                    {t(`capabilityReadiness.states.${row.state}`)}
                  </StatusBadge>
                </TableCell>
                <TableCell>
                  <StatusBadge status={row.live_validated ? "success" : "neutral"}>
                    {row.live_validated ? t("capabilityReadiness.yes") : t("capabilityReadiness.no")}
                  </StatusBadge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </section>
  );
}

function formatPrice(value: number | undefined, fallback: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return `$${new Intl.NumberFormat("en-US", { maximumFractionDigits: 6 }).format(value)} / 1M`;
}

function formatContext(value: number | undefined, fallback: string): string {
  return typeof value === "number" ? new Intl.NumberFormat("en-US").format(value) : fallback;
}

function formatModalityList(
  values: string[] | undefined,
  t: ReturnType<typeof useTranslations>,
): string {
  if (!values?.length) return t("values.na");
  return values.map((value) => {
    switch (value.trim().toLowerCase().replaceAll("-", "_")) {
      case "text":
        return t("modalities.text");
      case "image":
        return t("modalities.image");
      case "audio":
        return t("modalities.audio");
      case "video":
        return t("modalities.video");
      case "pdf":
        return t("modalities.pdf");
      case "text_to_speech":
        return t("modalities.textToSpeech");
      default:
        return value.replaceAll("_", " ");
    }
  }).join("، ");
}

function formatCapabilityList(
  values: string[] | undefined,
  t: ReturnType<typeof useTranslations>,
): string {
  if (!values?.length) return t("values.na");
  return values.map((value) => {
    switch (value.trim().toLowerCase().replaceAll("-", "_")) {
      case "chat":
        return t("capabilities.chat");
      case "tools":
      case "tool_calling":
        return t("capabilities.tools");
      case "vision":
        return t("capabilities.vision");
      case "embeddings":
      case "embedding":
        return t("capabilities.embeddings");
      case "text_to_speech":
        return t("capabilities.textToSpeech");
      case "moderation":
        return t("capabilities.moderation");
      case "image_generation":
        return t("capabilities.imageGeneration");
      case "audio_transcription":
      case "transcription":
        return t("capabilities.audioTranscription");
      case "audio_translation":
      case "translation":
        return t("capabilities.audioTranslation");
      default:
        return value;
    }
  }).join("، ");
}

function credentialKey(value: string): "managed" | "byok" {
  return value === "managed" ? "managed" : "byok";
}
