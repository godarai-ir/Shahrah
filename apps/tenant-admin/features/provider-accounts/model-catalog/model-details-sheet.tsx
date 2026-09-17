"use client";

import Link from "next/link";
import { useLocale, useTranslations } from "next-intl";
import { useMemo } from "react";
import { BookOpen, FlaskConical, ScrollText, TriangleAlert } from "lucide-react";

import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildPowerShellCurlGuideCommand,
  getCurlGuideEndpoint,
  getTenantApiBaseURL,
  resolveCurlGuideMode,
} from "@/features/provider-accounts/utils";
import { buildModelDiscoveryDetailsHref } from "@/features/provider-accounts/model-editor-utils";
import { buildPlaygroundHref } from "@/lib/playground-mode";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { CopyAction } from "@/components/app-ui/CopyAction";
import { ModalityFlow } from "@/components/app-ui/ModalityFlow";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { TabbedCodePanel } from "@/components/app-ui/TabbedCodePanel";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { cn } from "@/lib/utils";

import type { CapabilityReadiness, CatalogModel } from "./types";
import { formatCatalogModelDisplayName } from "./formatters";
import { resolveModelAvailability } from "./readiness-summary";
import { useCurlGuideSamples } from "@/features/provider-accounts/use-curl-guide-samples";

type ModelDetailsSheetProps = {
  model: CatalogModel | null;
  accountID: string;
  onOpenChange: (open: boolean) => void;
};

export function ModelDetailsSheet({
  model,
  accountID,
  onOpenChange,
}: ModelDetailsSheetProps) {
  const t = useTranslations("modelCatalog");
  const locale = useLocale();
  const availability = model ? resolveModelAvailability(model, accountID) : null;
  const modelName = model ? formatCatalogModelDisplayName(model.model, model.display_name) : t("drawer.title");
  const links = useMemo(() => {
    if (!model || !availability) return null;
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
    return {
      playground: buildPlaygroundHref(availability.fqn, mode),
      logs: `/logs?${new URLSearchParams({ model: availability.fqn, provider: model.provider }).toString()}`,
      docs: buildModelDiscoveryDetailsHref(model),
    };
  }, [availability, model]);

  return (
    <Sheet open={Boolean(model)} onOpenChange={onOpenChange}>
      <SheetContent
        // inline-start: the right edge in RTL, the left edge in LTR.
        side={locale === "fa" ? "right" : "left"}
        closeLabel={t("actions.close")}
        className="flex w-full min-w-0 flex-col overflow-hidden p-0 sm:max-w-[460px]"
      >
        <SheetHeader className="shrink-0 gap-2 border-b border-border px-5 pb-4 pt-5 pe-12 text-start">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <SheetTitle className="min-w-0 truncate" dir="ltr" title={modelName}>{modelName}</SheetTitle>
            {model?.is_deprecated ? <StatusBadge status="error">{t("readiness.deprecated")}</StatusBadge> : null}
          </div>
          <SheetDescription className="sr-only">{t("drawer.description")}</SheetDescription>
          {availability ? (
            <div className="flex min-w-0 items-center gap-1 rounded-md border border-border bg-muted/30 ps-2">
              <span className="sr-only">{t("drawer.fqn")}</span>
              <code className="min-w-0 flex-1 truncate text-xs" dir="ltr" title={availability.fqn}>{availability.fqn}</code>
              <CopyAction
                value={availability.fqn}
                copyLabel={t("actions.copyFqn")}
                copyingLabel={t("actions.copying")}
                copiedLabel={t("actions.copied")}
              />
            </div>
          ) : null}
          {links ? (
            <div className="flex flex-wrap items-center gap-2 pt-1">
              {links.playground ? (
                <Button asChild size="sm">
                  <Link href={links.playground}>
                    <FlaskConical data-icon="inline-start" aria-hidden="true" />
                    {t("actions.openPlayground")}
                  </Link>
                </Button>
              ) : null}
              <Button asChild size="sm" variant="outline">
                <Link href={links.logs}>
                  <ScrollText data-icon="inline-start" aria-hidden="true" />
                  {t("actions.viewLogs")}
                </Link>
              </Button>
              <Button asChild size="sm" variant="outline">
                <a href={links.docs} target="_blank" rel="noreferrer">
                  <BookOpen data-icon="inline-start" aria-hidden="true" />
                  {t("actions.docs")}
                </a>
              </Button>
            </div>
          ) : null}
        </SheetHeader>

        {model ? (
          <ModelProfile
            model={model}
            accountID={accountID}
            showAccountSelector={false}
            className="min-h-0 flex-1 overflow-y-auto px-5 py-5"
          />
        ) : null}
      </SheetContent>
    </Sheet>
  );
}

export type ModelProfileProps = {
  model: CatalogModel;
  accountID: string;
  onAccountChange?: (accountID: string) => void;
  showAccountSelector?: boolean;
  className?: string;
};

export function ModelProfile({
  model,
  accountID,
  onAccountChange,
  showAccountSelector = true,
  className,
}: ModelProfileProps) {
  const t = useTranslations("modelCatalog");
  const curlGuideSamples = useCurlGuideSamples();
  const selectedAvailability = useMemo(() => resolveModelAvailability(model, accountID), [accountID, model]);
  const code = useMemo(() => {
    if (!selectedAvailability) return null;
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
    const payload = buildCurlGuidePayload(mode, selectedAvailability.fqn, curlGuideSamples);
    return {
      mode,
      bash: buildBashCurlGuideCommand({ baseUrl, endpoint, payload }),
      powershell: buildPowerShellCurlGuideCommand({ baseUrl, endpoint, mode, modelFqn: selectedAvailability.fqn, samples: curlGuideSamples }),
    };
  }, [curlGuideSamples, model, selectedAvailability]);

  if (!selectedAvailability) return null;

  return (
    <div className={cn("flex min-w-0 flex-col gap-6", className)}>
      {showAccountSelector && onAccountChange && model.availability.length > 1 ? (
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
      ) : null}

      {model.is_deprecated ? (
        <OperationNotice
          notice={{
            kind: "warning",
            message: model.replacement_model
              ? t("drawer.deprecatedWithReplacement", { replacement: model.replacement_model })
              : t("drawer.deprecated"),
          }}
          labels={{ success: t("notices.success"), error: t("notices.error"), warning: t("notices.warning"), info: t("notices.info") }}
          action={<TriangleAlert aria-hidden="true" />}
        />
      ) : null}

      <DefinitionList columns={2}>
        <DefinitionListItem
          label={t("columns.context")}
          value={formatContext(model.context_window, t("values.na"))}
          valueDir="ltr"
          valueClassName="text-lg font-semibold leading-6 tabular-nums"
          surface
        />
        <DefinitionListItem
          label={t("drawer.publicPrice")}
          value={(
            <PricePair
              inputLabel={t("drawer.inputPrice")}
              inputValue={formatPrice(model.public_pricing?.input_usd_per_1m, t("values.na"))}
              outputLabel={t("drawer.outputPrice")}
              outputValue={formatPrice(model.public_pricing?.output_usd_per_1m, t("values.na"))}
            />
          )}
          wrap
          surface
        />
        <DefinitionListItem
          label={t("columns.modalities")}
          value={(
            <ModalityFlow
              inputModalities={model.input_modalities}
              outputModalities={model.output_modalities}
              inputLabel={t("drawer.inputModalities")}
              outputLabel={t("drawer.outputModalities")}
              emptyLabel={t("values.na")}
              getModalityLabel={(modality) => formatModalityLabel(modality, t)}
              className="mt-1 flex-wrap"
            />
          )}
          className="sm:col-span-2"
          wrap
          surface
        />
        <DefinitionListItem label={t("drawer.capabilities")} value={formatCapabilityList(selectedAvailability.effective_capabilities, t)} valueDir="auto" surface />
        <DefinitionListItem label={t("drawer.credentialSource")} value={t(`credential.${credentialKey(selectedAvailability.credential_source)}`)} surface />
        {hasDifferentPrice(
          selectedAvailability.effective_pricing?.input_usd_per_1m,
          model.public_pricing?.input_usd_per_1m,
        ) || hasDifferentPrice(
          selectedAvailability.effective_pricing?.output_usd_per_1m,
          model.public_pricing?.output_usd_per_1m,
        ) ? (
          <DefinitionListItem
            label={t("drawer.effectivePrice")}
            value={(
              <PricePair
                inputLabel={t("drawer.inputPrice")}
                inputValue={formatPrice(selectedAvailability.effective_pricing?.input_usd_per_1m, t("values.na"))}
                outputLabel={t("drawer.outputPrice")}
                outputValue={formatPrice(selectedAvailability.effective_pricing?.output_usd_per_1m, t("values.na"))}
              />
            )}
            wrap
            surface
          />
        ) : null}
      </DefinitionList>

      <CapabilityReadinessList rows={selectedAvailability.capability_readiness ?? model.readiness ?? []} t={t} />
      <PricingRatesTable publicRates={model.public_pricing?.rates ?? []} effectiveRates={selectedAvailability.effective_pricing?.rates ?? []} t={t} />

      {code ? (
        <TabbedCodePanel
          title={t("drawer.codeExample")}
          titleHidden
          tabsLabel={t("drawer.codeExampleTabs")}
          tabsListVariant="line"
          tabs={[
            { value: "bash", label: t("drawer.bash"), language: "bash", code: code.bash, className: "rounded-none border-0" },
            { value: "powershell", label: t("drawer.powershell"), language: "powershell", code: code.powershell, className: "rounded-none border-0" },
          ]}
          copy={{ copyLabel: t("actions.copyCode"), copyingLabel: t("actions.copying"), copiedLabel: t("actions.copied"), showLabel: false }}
          headerClassName="p-2 md:p-2"
          contentClassName="p-0 md:p-0"
        />
      ) : null}
    </div>
  );
}

function PricePair({
  inputLabel,
  inputValue,
  outputLabel,
  outputValue,
}: {
  inputLabel: string;
  inputValue: string;
  outputLabel: string;
  outputValue: string;
}) {
  return (
    <span className="mt-1 flex flex-wrap items-center gap-x-4 gap-y-1">
      <span className="flex items-baseline gap-1 whitespace-nowrap">
        <span className="text-xs font-normal text-muted-foreground">{inputLabel}</span>
        <bdi className="font-semibold" dir="auto">{inputValue}</bdi>
      </span>
      <span className="flex items-baseline gap-1 whitespace-nowrap">
        <span className="text-xs font-normal text-muted-foreground">{outputLabel}</span>
        <bdi className="font-semibold" dir="auto">{outputValue}</bdi>
      </span>
    </span>
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

function CapabilityReadinessList({
  rows,
  t,
}: {
  rows: CapabilityReadiness[];
  t: CatalogTranslator;
}) {
  if (!rows.length) return null;
  return (
    <section className="space-y-3" aria-labelledby="model-capability-readiness-title">
      <div>
        <h3 id="model-capability-readiness-title" className="text-sm font-semibold">{t("profile.capabilityReadiness")}</h3>
        <p className="text-xs text-muted-foreground">{t("drawer.capabilityReadinessDescription")}</p>
      </div>
      <ul className="divide-y divide-border rounded-lg border border-border">
        {rows.map((row, index) => {
          const modality = [row.input_modality, row.output_modality].filter(Boolean).join(" → ");
          return (
            <li key={`${row.method ?? ""}:${row.endpoint}:${row.transport ?? ""}:${index}`} className="flex flex-col gap-1.5 px-3 py-2.5">
              <div className="flex min-w-0 items-center justify-between gap-2">
                <code className="min-w-0 truncate text-xs" dir="ltr" title={row.endpoint}>
                  {[row.method, row.endpoint].filter(Boolean).join(" ")}
                </code>
                <StatusBadge status={capabilityStateStatus(row.state)} className="shrink-0">
                  {t(`capabilityReadiness.states.${row.state}`)}
                </StatusBadge>
              </div>
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
                {modality ? <span>{t("capabilityReadiness.modality")}: <bdi>{modality}</bdi></span> : null}
                <span>{t("capabilityReadiness.transport")}: <bdi dir="ltr">{row.transport ?? "http"}</bdi></span>
                <span>
                  {t("capabilityReadiness.live")}: {row.live_validated ? t("capabilityReadiness.yes") : t("capabilityReadiness.no")}
                </span>
              </div>
              {row.reason ? (
                // Server reasons are not translated; English text must not inherit the RTL page direction.
                <p className="text-xs leading-5 text-foreground/80" dir={isLatinText(row.reason) ? "ltr" : "auto"}>
                  {row.reason}
                </p>
              ) : null}
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function capabilityStateStatus(state: CapabilityReadiness["state"]) {
  switch (state) {
    case "contract_ready":
      return "success";
    case "pricing_incomplete":
      return "warning";
    case "unsupported_transport":
      return "info";
    default:
      return "error";
  }
}

function isLatinText(value: string): boolean {
  return !/[\u0590-\u08FF\uFB50-\uFEFF]/.test(value);
}

function formatPrice(value: number | undefined, fallback: string): string {
  if (typeof value !== "number" || !Number.isFinite(value)) return fallback;
  return `$${new Intl.NumberFormat("en-US", { maximumFractionDigits: 6 }).format(value)} / 1M`;
}

function formatContext(value: number | undefined, fallback: string): string {
  return typeof value === "number" ? new Intl.NumberFormat("en-US").format(value) : fallback;
}

function formatModalityLabel(
  value: string,
  t: ReturnType<typeof useTranslations>,
): string {
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
}

function hasDifferentPrice(effective: number | undefined, publicPrice: number | undefined): boolean {
  return typeof effective === "number"
    && Number.isFinite(effective)
    && effective !== publicPrice;
}

function formatCapabilityList(
  values: string[] | undefined,
  t: CatalogTranslator,
): string {
  if (!values?.length) return t("values.na");
  return values.map((value) => formatCapabilityLabel(value, t)).join(t("listSeparator"));
}

export function formatCapabilityLabel(value: string, t: CatalogTranslator): string {
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
}

function credentialKey(value: string): "managed" | "byok" {
  return value === "managed" ? "managed" : "byok";
}
