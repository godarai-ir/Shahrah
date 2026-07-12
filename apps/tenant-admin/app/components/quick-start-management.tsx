"use client";

import Link from "next/link";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { Fragment, useEffect, useMemo, useState, type ReactNode } from "react";
import { useTranslations } from "next-intl";
import {
  ArrowUpRight,
  Building2,
  Check,
  CheckCircle2,
  CircleDollarSign,
  Copy,
  Gauge,
  KeyRound,
  Layers3,
  Lightbulb,
  Network,
  Play,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
  Users,
} from "lucide-react";

import { CodeBlock, CodePanel } from "@/components/app-ui";
import {
  deriveQuickStartJourney,
  type QuickStartAction,
  type QuickStartEntitlements,
  type QuickStartExploreCardID,
  type QuickStartResources,
  type QuickStartStepID,
  type QuickStartStepStatus,
  type QuickStartSubscription,
} from "@/app/components/quick-start-journey";
import { ProviderLogo } from "@/app/components/provider-account-wizard/provider-logo";
import type { ProviderCatalogProvider, ProvidersResponse } from "@/app/components/provider-account-wizard/types";
import {
  buildBashCurlGuideCommand,
  getTenantApiBaseURL,
} from "@/app/components/provider-account-wizard/utils";
import { AppCard, EmptyState, HelpLink, PageHeader, StatusBadge } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

type Banner = {
  kind: "success" | "error";
  message: string;
};

type QuickStartManagementProps = {
  initialEntitlements: QuickStartEntitlements;
  initialPermissions: string[];
  initialResources: QuickStartResources;
  initialSubscription: QuickStartSubscription;
  docsBaseHref: string;
};

type TokenType = "pat" | "vat";
type SnippetTab = "bash" | "python" | "node";
type TranslateValues = Record<string, string | number | Date>;
type TranslateFn = (key: string, values?: TranslateValues) => string;
type SampleRunResult = {
  content: string | null;
  raw: string;
  status: number;
  tokenType: TokenType;
};

const samplePrompt = "برای من سه مزیت اصلی استفاده از gateway را بنویس.";
const featuredProviderOrder = ["openai", "anthropic", "google", "azure", "mistral", "groq"] as const;
type FeaturedProvider = (typeof featuredProviderOrder)[number];

export default function QuickStartManagement(props: QuickStartManagementProps) {
  const t = useTranslations("quickStart");
  const common = useTranslations("common");
  const [resources, setResources] = useState(props.initialResources);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [copiedKey, setCopiedKey] = useState("");
  const [isPatDialogOpen, setIsPatDialogOpen] = useState(false);
  const [patName, setPatName] = useState("quick-start");
  const [patExpiresIn, setPatExpiresIn] = useState("7d");
  const [generatedPatSecret, setGeneratedPatSecret] = useState<string | null>(null);
  const [manualTokens, setManualTokens] = useState<Record<TokenType, string>>({ pat: "", vat: "" });
  const [selectedSnippetTab, setSelectedSnippetTab] = useState<SnippetTab>("bash");
  const [selectedTokenType, setSelectedTokenType] = useState<TokenType>("pat");
  const [isCreatingPat, setIsCreatingPat] = useState(false);
  const [isRunningSample, setIsRunningSample] = useState(false);
  const [sampleRunResult, setSampleRunResult] = useState<SampleRunResult | null>(null);
  const [isTestBenchOpen, setIsTestBenchOpen] = useState(false);
  const [labBanner, setLabBanner] = useState<Banner | null>(null);
  const [providerChoices, setProviderChoices] = useState<ProviderCatalogProvider[]>([]);

  const journey = useMemo(
    () =>
      deriveQuickStartJourney({
        subscription: props.initialSubscription,
        entitlements: props.initialEntitlements,
        permissions: props.initialPermissions,
        resources,
      }),
    [props.initialEntitlements, props.initialPermissions, props.initialSubscription, resources],
  );
  const featuredProviders = useMemo(() => {
    const ranked = new Map(featuredProviderOrder.map((slug, index) => [slug, index]));
    return providerChoices
      .map((item) => ({ item, featuredProvider: getFeaturedProvider(item.provider) }))
      .filter((entry): entry is { item: ProviderCatalogProvider; featuredProvider: FeaturedProvider } => Boolean(entry.featuredProvider))
      .sort((a, b) => (ranked.get(a.featuredProvider) ?? Number.MAX_SAFE_INTEGER) - (ranked.get(b.featuredProvider) ?? Number.MAX_SAFE_INTEGER))
      .map((entry) => entry.item);
  }, [providerChoices]);
  const completedStepCount = journey.steps.filter((step) => step.status === "done").length;

  const effectiveSelectedTokenType =
    journey.testBench.availableTokenTypes.includes(selectedTokenType)
      ? selectedTokenType
      : journey.testBench.availableTokenTypes[0] ?? journey.testBench.recommendedTokenType;

  const suggestedModel =
    journey.testBench.suggestedModelByTokenType[effectiveSelectedTokenType] ??
    journey.testBench.suggestedModelByTokenType.pat ??
    journey.testBench.suggestedModelByTokenType.vat ??
    "openai:default:gpt-4o-mini";

  const effectiveToken =
    effectiveSelectedTokenType === "pat"
      ? generatedPatSecret ?? manualTokens.pat.trim()
      : manualTokens.vat.trim();

  const tokenPlaceholder = effectiveSelectedTokenType === "pat" ? "YOUR_GodarAI_PAT" : "YOUR_GodarAI_VAT";
  const tokenForSnippet = effectiveToken || tokenPlaceholder;
  const baseUrl = getTenantApiBaseURL();
  const bashSnippet = buildBashCurlGuideCommand({
    baseUrl,
    endpoint: "/v1/chat/completions",
    payload: {
      model: suggestedModel,
      messages: [
        {
          role: "user",
          content: samplePrompt,
        },
      ],
    },
  }).replaceAll("<PAT_TOKEN>", tokenForSnippet);

  const pythonSnippet = [
    "from openai import OpenAI",
    "",
    "client = OpenAI(",
    `    api_key=${JSON.stringify(tokenForSnippet)},`,
    `    base_url=${JSON.stringify(`${baseUrl}/v1`)}`,
    ")",
    "",
    "response = client.chat.completions.create(",
    `    model=${JSON.stringify(suggestedModel)},`,
    "    messages=[",
    "        {",
    '            "role": "user",',
    `            "content": ${JSON.stringify(samplePrompt)}`,
    "        }",
    "    ]",
    ")",
    "",
    "print(response.choices[0].message.content)",
  ].join("\n");

  const nodeSnippet = [
    'import OpenAI from "openai";',
    "",
    "const client = new OpenAI({",
    `  apiKey: ${JSON.stringify(tokenForSnippet)},`,
    `  baseURL: ${JSON.stringify(`${baseUrl}/v1`)},`,
    "});",
    "",
    "const response = await client.chat.completions.create({",
    `  model: ${JSON.stringify(suggestedModel)},`,
    "  messages: [",
    "    {",
    '      role: "user",',
    `      content: ${JSON.stringify(samplePrompt)},`,
    "    },",
    "  ],",
    "});",
    "",
    "console.log(response.choices[0]?.message?.content);",
  ].join("\n");

  useEffect(() => {
    let isCancelled = false;
    void requestJSON<ProvidersResponse>("/api/provider-catalog/providers")
      .then((payload) => {
        if (isCancelled) return;
        setProviderChoices(payload.data ?? []);
      })
      .catch(() => {
        if (isCancelled) return;
        setProviderChoices([]);
      });
    return () => {
      isCancelled = true;
    };
  }, []);

  async function createPat() {
    if (isCreatingPat) return;
    if (!patName.trim()) {
      setBanner({ kind: "error", message: t("patDialog.validation") });
      return;
    }
    setIsCreatingPat(true);
    try {
      const payload = await requestJSON<{
        id: string;
        key: string;
      }>("/api/personal-access-tokens", {
        method: "POST",
        body: JSON.stringify({
          name: patName.trim(),
          expires_in: patExpiresIn.trim() || undefined,
        }),
      });
      setGeneratedPatSecret(payload.key);
      setManualTokens((current) => ({ ...current, pat: payload.key }));
      setResources((current) => ({
        ...current,
        personalAccessTokenCount: current.personalAccessTokenCount + 1,
      }));
      setIsPatDialogOpen(false);
      if (isTestBenchOpen) {
        setLabBanner({ kind: "success", message: t("messages.patCreated") });
      } else {
        setBanner({ kind: "success", message: t("messages.patCreated") });
      }
    } catch (error) {
      const message = normalizeError(error, t("messages.patCreateFailed"));
      if (isTestBenchOpen) {
        setLabBanner({ kind: "error", message });
      } else {
        setBanner({ kind: "error", message });
      }
    } finally {
      setIsCreatingPat(false);
    }
  }

  function handleAction(action: QuickStartAction) {
    if (action.type === "dialog") {
      setIsPatDialogOpen(true);
      return;
    }
    if (action.type === "anchor") {
      if (action.anchor === "test_bench") {
        setIsTestBenchOpen(true);
        return;
      }
      return;
    }
  }

  async function copyText(key: string, value: string) {
    try {
      if (navigator.clipboard?.writeText) {
        await copyTextToClipboard(value);
      } else {
        const textarea = document.createElement("textarea");
        textarea.value = value;
        textarea.setAttribute("readonly", "");
        textarea.style.position = "fixed";
        textarea.style.top = "-9999px";
        textarea.style.left = "-9999px";
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand("copy");
        document.body.removeChild(textarea);
      }
      setCopiedKey(key);
      window.setTimeout(() => setCopiedKey((current) => (current === key ? "" : current)), 1600);
    } catch {
      setBanner({ kind: "error", message: t("messages.copyFailed") });
    }
  }

  async function runSampleRequest() {
    if (isRunningSample) return;
    if (!journey.testBench.ready) {
      setLabBanner({ kind: "error", message: t("lab.notReadyTitle") });
      return;
    }
    if (!effectiveToken) {
      setLabBanner({ kind: "error", message: t(`lab.missingToken.${effectiveSelectedTokenType}`) });
      return;
    }

    setIsRunningSample(true);
    setLabBanner(null);
    try {
      const response = await fetch(`${baseUrl}/v1/chat/completions`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${effectiveToken}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          model: suggestedModel,
          messages: [
            {
              role: "user",
              content: samplePrompt,
            },
          ],
        }),
      });

      const text = await response.text();
      const payload = parsePayload(text);
      if (!response.ok) {
        throw new Error(extractMessage(payload) ?? `Request failed (${response.status})`);
      }

      setResources((current) => ({
        ...current,
        sampleRequestSucceeded: true,
      }));
      setSampleRunResult({
        content: extractSampleContent(payload),
        raw: formatSamplePayload(payload),
        status: response.status,
        tokenType: effectiveSelectedTokenType,
      });
      setLabBanner({ kind: "success", message: t("messages.sampleRequestSucceeded") });
    } catch (error) {
      setLabBanner({ kind: "error", message: normalizeError(error, t("messages.sampleRequestFailed")) });
    } finally {
      setIsRunningSample(false);
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={<HeaderDescription t={t} docsBaseHref={props.docsBaseHref} />}
      />

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="mx-auto w-full max-w-5xl">
        <AppCard
          title={t("journey.title")}
          description={t("journey.description") || undefined}
          actions={
            <StatusBadge status={completedStepCount === journey.steps.length ? "success" : "pending"}>
              {t("progress.value", { done: completedStepCount, total: journey.steps.length })}
            </StatusBadge>
          }
        >
          <ol className="flex flex-col gap-4">
            {journey.steps.map((step, index) => (
              <Fragment key={step.id}>
                <li className="relative ps-16">
                    {index < journey.steps.length - 1 ? (
                      <span className="absolute bottom-0 start-6 top-14 w-px bg-border" aria-hidden="true" />
                    ) : null}
                    <span
                      className={cn(
                        "absolute start-0 top-1 flex size-12 items-center justify-center rounded-lg border text-sm font-semibold",
                        step.status === "done"
                          ? "border-success/30 bg-success-muted text-success"
                          : step.status === "active"
                            ? "border-info/30 bg-info/10 text-info"
                            : step.status === "unavailable"
                              ? "border-destructive/20 bg-destructive/5 text-destructive"
                              : "border-border bg-muted/40 text-muted-foreground",
                      )}
                    >
                      {step.status === "done" ? <CheckCircle2 className="size-5" /> : formatCount(index + 1)}
                    </span>
                    <div className="bg-card p-4">
                      <div className="flex flex-col gap-3">
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0 flex-1">
                            <div className="flex flex-wrap items-center gap-2">
                              <h3 className="text-base font-semibold">{t(`steps.${step.id}.title`)}</h3>
                              <StatusBadge status={statusBadgeForStepStatus(step.status)}>
                                {t(`statuses.${step.status}`)}
                              </StatusBadge>
                            </div>
                          </div>
                          <StepActionButton
                            stepID={step.id}
                            status={step.status}
                            action={step.action}
                            t={t}
                            onAction={() => {
                              if (step.id === "inspect_activity") {
                                setResources((current) => ({
                                  ...current,
                                  telemetryVisited: true,
                                }));
                              }
                              handleAction(step.action);
                            }}
                          />
                        </div>
                        <div className="min-w-0">
                          <p className="mt-2 text-sm leading-7 text-muted-foreground">
                            {t(`steps.${step.id}.description`)}
                          </p>
                          {step.id === "connect_provider" && featuredProviders.length > 0 ? (
                            <div className="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                              {featuredProviders.map((providerItem) => (
                                <Link
                                  key={providerItem.provider}
                                  href={`/provider-accounts/new?provider=${encodeURIComponent(providerItem.provider)}`}
                                  className="group flex items-center gap-3 rounded-lg border border-border bg-background px-3 py-3 transition-colors hover:border-primary/30 hover:bg-accent"
                                >
                                  <span className="flex size-11 shrink-0 items-center justify-center rounded-md bg-muted">
                                    <ProviderLogo
                                      provider={providerItem.provider}
                                      displayName={providerItem.display_name}
                                      logoURL={providerItem.logo_url}
                                      size={26}
                                    />
                                  </span>
                                  <span className="min-w-0 flex-1 text-sm font-medium leading-6">
                                    {providerItem.display_name}
                                  </span>
                                  <ArrowUpRight className="size-4 shrink-0 text-muted-foreground transition group-hover:text-foreground rtl:-rotate-90" aria-hidden="true" />
                                </Link>
                              ))}
                            </div>
                          ) : null}
                          <div className="mt-3 flex w-full items-start gap-3 rounded-lg border border-warning/20 bg-warning-muted px-3 py-3 text-xs leading-6 text-foreground">
                            <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-warning/15 text-warning">
                              <Lightbulb className="size-4" aria-hidden="true" />
                            </span>
                            <p className="min-w-0 flex-1">{renderStepLearn(step.id, t, props.docsBaseHref)}</p>
                          </div>
                          {step.status === "unavailable" && step.blockedBy ? (
                            <p className="mt-3 text-xs leading-6 text-destructive">
                              {t(`blocked.${step.blockedBy}`)}
                            </p>
                          ) : null}
                        </div>
                      </div>
                    </div>
                  </li>
                {index < journey.steps.length - 1 ? <StepSeparator /> : null}
              </Fragment>
            ))}
          </ol>
        </AppCard>
      </section>

      <section>
        <AppCard
          title={
            <span className="flex items-center gap-2">
              <Sparkles className="text-info" aria-hidden="true" />
              <span>{t("explore.title")}</span>
            </span>
          }
          description={t("explore.description")}
          contentClassName="@container"
        >
          {journey.exploreCards.length > 0 ? (
            <div className="grid gap-3 @[28rem]:grid-cols-2 @[48rem]:grid-cols-3 @[64rem]:grid-cols-4 @[82rem]:grid-cols-5">
              {journey.exploreCards.map((cardID) => (
                <ExploreCard key={cardID} cardID={cardID} t={t} />
              ))}
            </div>
          ) : (
            <EmptyState icon={Sparkles} title={t("explore.title")} description={t("explore.description")} />
          )}
        </AppCard>
      </section>

      <Dialog open={isPatDialogOpen} onOpenChange={setIsPatDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("patDialog.title")}</DialogTitle>
            <DialogDescription>{t("patDialog.description")}</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="quick-start-pat-name">{t("patDialog.nameLabel")}</Label>
              <Input
                id="quick-start-pat-name"
                value={patName}
                onChange={(event) => setPatName(event.target.value)}
                placeholder="quick-start"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="quick-start-pat-expiry">{t("patDialog.expiryLabel")}</Label>
              <Input
                id="quick-start-pat-expiry"
                value={patExpiresIn}
                onChange={(event) => setPatExpiresIn(event.target.value)}
                placeholder="7d"
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setIsPatDialogOpen(false)} disabled={isCreatingPat}>
              {common("cancel")}
            </Button>
            <Button type="button" onClick={() => void createPat()} disabled={isCreatingPat}>
              {isCreatingPat ? <Spinner data-icon="inline-start" /> : <KeyRound data-icon="inline-start" />}
              {isCreatingPat ? t("patDialog.creating") : t("patDialog.submit")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Sheet open={isTestBenchOpen} onOpenChange={setIsTestBenchOpen}>
        <SheetContent side="left" className="flex h-full w-full max-w-3xl flex-col gap-0 p-0 sm:max-w-2xl">
          <SheetHeader className="border-b border-border px-6 py-5 text-start">
            <SheetTitle className="flex items-center gap-2">
              <TerminalSquare className="text-info" aria-hidden="true" />
              {t("lab.title")}
            </SheetTitle>
            <SheetDescription className="leading-7">{t("lab.description")}</SheetDescription>
          </SheetHeader>
          <div id="test_bench" className="custom-scrollbar flex-1 overflow-y-auto px-6 py-5" dir="rtl">
            <div className="flex flex-col gap-4">
              {labBanner ? (
                <Alert variant={labBanner.kind === "error" ? "destructive" : "default"}>
                  <AlertTitle>{labBanner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
                  <AlertDescription>{labBanner.message}</AlertDescription>
                </Alert>
              ) : null}

              {!journey.testBench.ready ? (
                <Alert>
                  <AlertTitle>{t("lab.notReadyTitle")}</AlertTitle>
                  <AlertDescription>
                    {resources.providerAccountCount === 0 ? t("lab.notReadyProvider") : t("lab.notReadyModels")}
                  </AlertDescription>
                </Alert>
              ) : null}

              <div className="rounded-lg border border-border bg-card p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0 text-start">
                    <div className="flex flex-wrap items-center gap-3">
                      <p className="text-sm font-semibold">{t("lab.keyTypeTitle")}</p>
                      <Tabs value={effectiveSelectedTokenType} onValueChange={(value) => setSelectedTokenType(value as TokenType)} className="gap-0">
                        <TabsList className="h-auto rounded-none bg-transparent p-0">
                          {journey.testBench.availableTokenTypes.map((tokenType) => (
                            <TabsTrigger key={tokenType} value={tokenType} className="h-7 flex-none rounded-md px-3 text-xs">
                              {t(`tokenLabels.${tokenType}`)}
                            </TabsTrigger>
                          ))}
                        </TabsList>
                      </Tabs>
                    </div>
                    <p className="mt-1 text-xs leading-6 text-muted-foreground">
                      {journey.recommendedTokenType === "vat" ? t("lab.recommendedVat") : t("lab.recommendedPat")}
                    </p>
                  </div>
                  {journey.testBench.availableTokenTypes.includes("pat") ? (
                    <Button type="button" variant="outline" size="sm" onClick={() => setIsPatDialogOpen(true)}>
                      <KeyRound data-icon="inline-start" />
                      {t("actions.createPat")}
                    </Button>
                  ) : null}
                </div>

                <Tabs value={effectiveSelectedTokenType} onValueChange={(value) => setSelectedTokenType(value as TokenType)} className="mt-4 gap-3">
                  {journey.testBench.availableTokenTypes.map((tokenType) => (
                    <TabsContent key={tokenType} value={tokenType} className="flex flex-col gap-3">
                      <div className="flex flex-col gap-2 text-right">
                        <Label htmlFor={`quick-start-token-${tokenType}`} className="text-right">
                          {t(`lab.tokenInputLabel.${tokenType}`)}
                        </Label>
                        <Input
                          id={`quick-start-token-${tokenType}`}
                          dir="rtl"
                          className="text-right"
                          value={tokenType === "pat" ? manualTokens.pat : manualTokens.vat}
                          onChange={(event) =>
                            setManualTokens((current) => ({
                              ...current,
                              [tokenType]: event.target.value,
                            }))
                          }
                          placeholder={t(`lab.tokenInputPlaceholder.${tokenType}`)}
                        />
                      </div>
                    </TabsContent>
                  ))}
                </Tabs>
              </div>

              {generatedPatSecret ? (
                <div className="rounded-lg border border-success/30 bg-success-muted p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-success">{t("lab.generatedTokenTitle")}</p>
                      <p className="mt-1 text-xs leading-6 text-foreground">{t("lab.generatedTokenDescription")}</p>
                      <code className="mt-3 block overflow-auto rounded-md border border-border bg-background px-3 py-2 text-xs" dir="ltr">
                        {generatedPatSecret}
                      </code>
                    </div>
                    <Button type="button" variant="outline" size="sm" onClick={() => void copyText("generated-token", generatedPatSecret)}>
                      <Copy data-icon="inline-start" />
                      {copiedKey === "generated-token" ? t("actions.copied") : t("actions.copyToken")}
                    </Button>
                  </div>
                </div>
              ) : null}

              <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                <div className="flex flex-col gap-2 text-start">
                  <Label htmlFor="quick-start-model">{t("lab.modelLabel")}</Label>
                  <Input id="quick-start-model" dir="ltr" readOnly value={suggestedModel} className="text-left font-mono" />
                </div>
                <div className="flex gap-2">
                  <Button type="button" onClick={() => void runSampleRequest()} disabled={!journey.testBench.ready || !effectiveToken || isRunningSample}>
                    {isRunningSample ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                    {isRunningSample ? t("lab.runningAction") : t("lab.runAction")}
                  </Button>
                </div>
              </div>

              <Tabs value={selectedSnippetTab} onValueChange={(value) => setSelectedSnippetTab(value as SnippetTab)}>
                <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <TabsList className="h-10 rounded-lg border border-border bg-muted p-1">
                      <TabsTrigger value="bash" className="h-8 px-3">{t("lab.tabs.bash")}</TabsTrigger>
                      <TabsTrigger value="python" className="h-8 px-3">{t("lab.tabs.python")}</TabsTrigger>
                      <TabsTrigger value="node" className="h-8 px-3">{t("lab.tabs.node")}</TabsTrigger>
                    </TabsList>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon-sm"
                      aria-label={copiedKey === `snippet-${selectedSnippetTab}` ? t("actions.copied") : t("actions.copySnippet")}
                      title={copiedKey === `snippet-${selectedSnippetTab}` ? t("actions.copied") : t("actions.copySnippet")}
                      onClick={() =>
                        void copyText(
                          `snippet-${selectedSnippetTab}`,
                          selectedSnippetTab === "bash"
                            ? bashSnippet
                            : selectedSnippetTab === "python"
                              ? pythonSnippet
                              : nodeSnippet,
                        )
                      }
                      disabled={!journey.testBench.ready}
                      className={cn(
                        "transition-all duration-200",
                        copiedKey === `snippet-${selectedSnippetTab}` && "border-success/30 bg-success/10 text-success hover:bg-success/15",
                      )}
                    >
                      {copiedKey === `snippet-${selectedSnippetTab}` ? (
                        <Check className="animate-in zoom-in-50 duration-200" aria-hidden="true" />
                      ) : (
                        <Copy aria-hidden="true" />
                      )}
                    </Button>
                  </div>
                  <TabsContent value="bash" className="mt-0">
                    <CodeBlock
                      code={bashSnippet}
                      language="bash"
                      highlightedTerms={[baseUrl, tokenPlaceholder]}
                      preClassName="text-xs leading-6 sm:text-xs sm:leading-6"
                    />
                  </TabsContent>
                  <TabsContent value="python" className="mt-0">
                    <CodeBlock
                      code={pythonSnippet}
                      language="python"
                      highlightedTerms={[baseUrl, tokenPlaceholder]}
                      preClassName="text-xs leading-6 sm:text-xs sm:leading-6"
                    />
                  </TabsContent>
                  <TabsContent value="node" className="mt-0">
                    <CodeBlock
                      code={nodeSnippet}
                      language="javascript"
                      highlightedTerms={[baseUrl, tokenPlaceholder]}
                      preClassName="text-xs leading-6 sm:text-xs sm:leading-6"
                    />
                  </TabsContent>
                </div>
              </Tabs>

              {!effectiveToken ? (
                <p className="text-xs leading-6 text-muted-foreground">{t(`lab.placeholderHint.${effectiveSelectedTokenType}`)}</p>
              ) : null}

              {sampleRunResult ? (
                <div className="flex flex-col gap-3 rounded-lg border border-success/30 bg-success-muted p-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                      <p className="text-sm font-semibold text-success">{t("lab.responseTitle")}</p>
                      <p className="mt-1 text-xs leading-6 text-foreground">
                        {t("lab.responseDescription", {
                          status: sampleRunResult.status,
                          tokenType: t(`tokenLabels.${sampleRunResult.tokenType}`),
                        })}
                      </p>
                    </div>
                  </div>

                  {sampleRunResult.content ? (
                    <div className="rounded-md border border-border bg-background p-3">
                      <p className="text-xs font-semibold text-muted-foreground">{t("lab.responsePreviewLabel")}</p>
                      <p className="mt-2 whitespace-pre-wrap text-sm leading-7">{sampleRunResult.content}</p>
                    </div>
                  ) : null}

                  <CodePanel
                    title={t("lab.rawResponseLabel")}
                    code={sampleRunResult.raw}
                    language="json"
                    displayLanguage="JSON"
                    copy={{ copyLabel: t("lab.copyResponse"), copiedLabel: t("actions.copied") }}
                  />
                </div>
              ) : null}
            </div>
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}

function HeaderDescription(props: {
  docsBaseHref: string;
  t: TranslateFn;
}) {
  return (
    <>
      {props.t("subtitle")}{" "}
      <HelpLink href={`${props.docsBaseHref}/quick-start`} />
    </>
  );
}

function StepSeparator() {
  return (
    <li className="flex justify-center ps-16" aria-hidden="true">
      <div className="flex w-2/3 max-w-sm items-center gap-3 text-muted-foreground">
        <span className="h-px flex-1 bg-border" />
        <span className="flex size-7 shrink-0 items-center justify-center rounded-full border border-border bg-background text-info">
          <Sparkles className="size-3.5" />
        </span>
        <span className="h-px flex-1 bg-border" />
      </div>
    </li>
  );
}

function StepActionButton(props: {
  action: QuickStartAction;
  onAction: () => void;
  status: QuickStartStepStatus;
  stepID: QuickStartStepID;
  t: TranslateFn;
}) {
  const isDisabled = props.status === "unavailable" || props.status === "locked";
  const label = props.t(`steps.${props.stepID}.${props.status === "done" ? "manageAction" : "action"}`);
  const icon = props.action.type === "anchor" ? <TerminalSquare data-icon="inline-end" /> : <KeyRound data-icon="inline-end" />;
  const buttonClassName = "shrink-0 max-w-full px-2.5 sm:px-3";
  const textClassName = "inline text-xs sm:text-sm";

  if (props.action.type === "link") {
    if (isDisabled) {
      return (
        <Button type="button" variant={props.status === "done" ? "outline" : "default"} className={buttonClassName} size="sm" disabled>
          <span className={textClassName}>{label}</span>
          <ArrowUpRight className="shrink-0 rtl:-rotate-90" data-icon="inline-end" />
        </Button>
      );
    }
    return (
      <Button
        type="button"
        asChild
        variant={props.status === "done" ? "outline" : "default"}
        className={buttonClassName}
        size="sm"
      >
        <Link href={props.action.href} onClick={props.onAction}>
          <span className={textClassName}>{label}</span>
          <ArrowUpRight className="shrink-0 rtl:-rotate-90" data-icon="inline-end" />
        </Link>
      </Button>
    );
  }

  return (
    <Button
      type="button"
      variant={props.status === "done" ? "outline" : "default"}
      className={buttonClassName}
      size="sm"
      onClick={props.onAction}
      disabled={isDisabled}
    >
      <span className={textClassName}>{label}</span>
      <span className="inline-flex shrink-0">{icon}</span>
    </Button>
  );
}

function ExploreCard(props: {
  cardID: QuickStartExploreCardID;
  t: TranslateFn;
}) {
  const hrefMap: Record<QuickStartExploreCardID, string> = {
    provider_accounts: "/provider-accounts",
    virtual_users: "/virtual-users",
    teams: "/teams",
    org_units: "/org-units",
    guardrails: "/guardrails",
    prompts: "/prompts",
    metrics: "/metrics",
    logs: "/logs",
    billing_plan: "/billing/plan",
  };

  const iconMap: Record<QuickStartExploreCardID, typeof Network> = {
    provider_accounts: Network,
    virtual_users: KeyRound,
    teams: Users,
    org_units: Building2,
    guardrails: ShieldCheck,
    prompts: Sparkles,
    metrics: Gauge,
    logs: Layers3,
    billing_plan: CircleDollarSign,
  };

  const Icon = iconMap[props.cardID];

  return (
    <Link
      href={hrefMap[props.cardID]}
      className="group flex h-full min-h-32 flex-col rounded-lg border border-card-border bg-card p-4 transition-colors hover:border-primary/30 hover:bg-accent @[28rem]:min-h-36"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
            <Icon aria-hidden="true" />
          </span>
          <p className="min-w-0 text-sm font-semibold leading-6">
            {props.t(`explore.cards.${props.cardID}.title`)}
          </p>
        </div>
        <ArrowUpRight className="mt-0.5 shrink-0 text-muted-foreground transition group-hover:text-foreground rtl:-rotate-90" aria-hidden="true" />
      </div>
      <p className="mt-3 text-xs leading-6 text-muted-foreground">
        {props.t(`explore.cards.${props.cardID}.description`)}
      </p>
    </Link>
  );
}

function renderStepLearn(stepID: QuickStartStepID, t: TranslateFn, docsBaseHref: string): ReactNode {
  const docsSlugByStep: Partial<Record<QuickStartStepID, string>> = {
    connect_provider: "/quick-start",
    create_pat: "/guides/api-keys",
    inspect_activity: "/observability-logs",
  };

  const docsSlug = docsSlugByStep[stepID];
  if (!docsSlug) {
    return t(`steps.${stepID}.learn`);
  }

  return (
    <>
      {t(`steps.${stepID}.learn`)}{" "}
      <HelpLink href={`${docsBaseHref}${docsSlug}`} />
    </>
  );
}

function normalizeFeaturedProvider(provider: string): string {
  const normalized = provider.trim().toLowerCase();
  if (normalized === "gemini" || normalized === "google-gemini") return "google";
  if (normalized === "azure-openai") return "azure";
  return normalized;
}

function getFeaturedProvider(provider: string): FeaturedProvider | null {
  const normalized = normalizeFeaturedProvider(provider);
  return (featuredProviderOrder as readonly string[]).includes(normalized) ? (normalized as FeaturedProvider) : null;
}

function statusBadgeForStepStatus(status: QuickStartStepStatus): "success" | "pending" | "error" | "neutral" {
  if (status === "done") return "success";
  if (status === "active") return "pending";
  if (status === "unavailable") return "error";
  return "neutral";
}

function formatCount(value: number) {
  return new Intl.NumberFormat("fa-IR").format(value);
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

function extractSampleContent(payload: unknown): string | null {
  if (!payload || typeof payload !== "object") return null;

  const candidate = payload as {
    choices?: Array<{ message?: { content?: unknown } }>;
    output_text?: unknown;
    data?: Array<{ url?: unknown }>;
  };

  if (typeof candidate.output_text === "string" && candidate.output_text.trim()) {
    return candidate.output_text.trim();
  }

  const messageContent = candidate.choices?.[0]?.message?.content;
  if (typeof messageContent === "string" && messageContent.trim()) {
    return messageContent.trim();
  }
  if (Array.isArray(messageContent)) {
    const text = messageContent
      .map((item) => {
        if (typeof item === "string") return item;
        if (!item || typeof item !== "object") return "";
        const part = item as { text?: unknown };
        return typeof part.text === "string" ? part.text : "";
      })
      .filter(Boolean)
      .join("\n")
      .trim();
    if (text) return text;
  }

  const firstURL = candidate.data?.find((item) => typeof item?.url === "string")?.url;
  return typeof firstURL === "string" ? firstURL : null;
}

function formatSamplePayload(payload: unknown): string {
  if (typeof payload === "string") return payload;
  try {
    return JSON.stringify(payload, null, 2);
  } catch {
    return String(payload);
  }
}
