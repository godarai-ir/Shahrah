"use client";

import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Check,
  Copy,
  KeyRound,
  Play,
  TerminalSquare,
} from "lucide-react";

import { CodeBlock } from "@/components/app-ui/CodeBlock";
import { CodePanel } from "@/components/app-ui/CodePanel";
import {
  deriveQuickStartJourney,
  type QuickStartAction,
  type QuickStartEntitlements,
  type QuickStartExploreCardID,
  type QuickStartJourneyStep,
  type QuickStartResources,
  type QuickStartStepID,
  type QuickStartSubscription,
} from "@/features/quick-start/journey";
import type { ProviderCatalogProvider } from "@/lib/contracts/provider-catalog";
import {
  createPersonalAccessToken,
  listQuickStartProviders,
  runQuickStartSample,
} from "@/features/quick-start/api";
import { buildBashCurlGuideCommand } from "@/lib/api-guides";
import { getAPIErrorMessage } from "@/lib/http/api-client";
import { getTenantApiBaseURL } from "@/lib/tenant-api-origin";
import { AppCard } from "@/components/app-ui/AppCard";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { PageHeader } from "@/components/app-ui/PageHeader";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import { useNumberFormatter } from "@/hooks/use-locale-format";
import {
  HeaderDescription,
  StepMarker,
  QuickStartStepActive,
  QuickStartStepDone,
  QuickStartStepPending,
  ExploreChip,
} from "./QuickStartParts";

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
type SampleRunResult = {
  content: string | null;
  raw: string;
  status: number;
  tokenType: TokenType;
};

const featuredProviderOrder = ["openai", "anthropic", "google", "azure", "mistral", "groq"] as const;
type FeaturedProvider = (typeof featuredProviderOrder)[number];

/** Chips are capped so the journey itself keeps the reader's attention. */
const exploreChipPriority: QuickStartExploreCardID[] = [
  "teams",
  "guardrails",
  "prompts",
  "billing_plan",
  "virtual_users",
  "provider_accounts",
  "metrics",
  "logs",
];
const exploreChipLimit = 4;

export default function QuickStartManagement(props: QuickStartManagementProps) {
  const t = useTranslations("quickStart");
  const common = useTranslations("common");
  const numberFormatter = useNumberFormatter();
  const samplePrompt = t("samplePrompt");
  const [resources, setResources] = useState(props.initialResources);
  const [skippedStepIDs, setSkippedStepIDs] = useState<QuickStartStepID[]>([]);
  const [banner, setBanner] = useState<Banner | null>(null);
  const [patDialogNotice, setPatDialogNotice] = useState<Banner | null>(null);
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
  const [labBanner, setLabBanner] = useState<Banner | null>(null);
  const [providerChoices, setProviderChoices] = useState<ProviderCatalogProvider[]>([]);
  const testBenchRef = useRef<HTMLElement | null>(null);
  const tokenInputRef = useRef<HTMLInputElement | null>(null);

  const journey = useMemo(
    () =>
      deriveQuickStartJourney({
        subscription: props.initialSubscription,
        entitlements: props.initialEntitlements,
        permissions: props.initialPermissions,
        resources,
        skippedStepIDs,
      }),
    [props.initialEntitlements, props.initialPermissions, props.initialSubscription, resources, skippedStepIDs],
  );
  const isProviderStepOpen = journey.activeStepID === "connect_provider";
  const featuredProviders = useMemo(() => {
    const ranked = new Map(featuredProviderOrder.map((slug, index) => [slug, index]));
    return providerChoices
      .map((item) => ({ item, featuredProvider: getFeaturedProvider(item.provider) }))
      .filter((entry): entry is { item: ProviderCatalogProvider; featuredProvider: FeaturedProvider } => Boolean(entry.featuredProvider))
      .sort((a, b) => (ranked.get(a.featuredProvider) ?? Number.MAX_SAFE_INTEGER) - (ranked.get(b.featuredProvider) ?? Number.MAX_SAFE_INTEGER))
      .map((entry) => entry.item);
  }, [providerChoices]);
  const exploreChips = useMemo(
    () => exploreChipPriority.filter((cardID) => journey.exploreCards.includes(cardID)).slice(0, exploreChipLimit),
    [journey.exploreCards],
  );
  const progressPercent =
    journey.totalCount > 0 ? Math.min(100, Math.round((journey.doneCount / journey.totalCount) * 100)) : 100;

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

  const tokenPlaceholder = effectiveSelectedTokenType === "pat" ? "YOUR_GODARAI_PAT" : "YOUR_GODARAI_VAT";
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
    // The provider shortcuts only render inside the open first step, so nothing is
    // fetched once that step is done.
    if (!isProviderStepOpen) return;

    let isCancelled = false;
    void listQuickStartProviders()
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
  }, [isProviderStepOpen]);

  async function createPat() {
    if (isCreatingPat) return;
    if (!patName.trim()) {
      setPatDialogNotice({ kind: "error", message: t("patDialog.validation") });
      return;
    }
    setPatDialogNotice(null);
    setIsCreatingPat(true);
    try {
      const payload = await createPersonalAccessToken({
        name: patName.trim(),
        expiresIn: patExpiresIn.trim() || undefined,
      });
      setGeneratedPatSecret(payload.key);
      setManualTokens((current) => ({ ...current, pat: payload.key }));
      setResources((current) => ({
        ...current,
        personalAccessTokenCount: current.personalAccessTokenCount + 1,
        personalAccessTokenLabel: current.personalAccessTokenLabel ?? patName.trim(),
      }));
      setIsPatDialogOpen(false);
      showFeedbackToast({ kind: "success", title: t("messages.patCreated") });
    } catch (error) {
      const message = getAPIErrorMessage(error, t("messages.patCreateFailed"));
      setPatDialogNotice({ kind: "error", message });
    } finally {
      setIsCreatingPat(false);
    }
  }

  function focusTestBench() {
    const section = testBenchRef.current;
    if (!section) return;

    // scrollIntoView fights the sticky column on wide screens, so the page is moved
    // to the column's own offset instead.
    window.scrollTo({ top: Math.max(section.offsetTop - 24, 0), behavior: "smooth" });
    tokenInputRef.current?.focus({ preventScroll: true });
  }

  function handleAction(action: QuickStartAction) {
    if (action.type === "dialog") {
      setIsPatDialogOpen(true);
      return;
    }
    if (action.type === "anchor" && action.anchor === "test_bench") {
      focusTestBench();
    }
  }

  function skipStep(stepID: QuickStartStepID) {
    setSkippedStepIDs((current) => (current.includes(stepID) ? current : [...current, stepID]));
  }

  function resumeStep(stepID: QuickStartStepID) {
    setSkippedStepIDs((current) => current.filter((item) => item !== stepID));
  }

  function handleStepAction(step: QuickStartJourneyStep) {
    if (step.id === "inspect_activity") {
      setResources((current) => ({ ...current, telemetryVisited: true }));
    }
    handleAction(step.action);
  }

  async function copyText(key: string, value: string) {
    try {
      await copyTextToClipboard(value);
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
      const { payload, status } = await runQuickStartSample({
        baseUrl,
        token: effectiveToken,
        model: suggestedModel,
        prompt: samplePrompt,
      });

      setResources((current) => ({
        ...current,
        sampleRequestSucceeded: true,
        sampleRequestModel: suggestedModel,
      }));
      setSampleRunResult({
        content: extractSampleContent(payload),
        raw: formatSamplePayload(payload),
        status,
        tokenType: effectiveSelectedTokenType,
      });
      setLabBanner({ kind: "success", message: t("messages.sampleRequestSucceeded") });
    } catch (error) {
      setLabBanner({ kind: "error", message: getAPIErrorMessage(error, t("messages.sampleRequestFailed")) });
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

      {journey.estimatedMinutes > 0 ? (
        <p className="text-sm leading-6 text-muted-foreground">
          {t("progress.estimate", { minutes: journey.estimatedMinutes })}
        </p>
      ) : null}

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,420px)] lg:items-start">
        <div className="flex min-w-0 flex-col gap-6">
          <AppCard>
            <div className="flex flex-col gap-3">
              <div className="flex flex-wrap items-baseline justify-between gap-2">
                <p className="text-sm font-semibold">
                  {t("progress.value", { done: journey.doneCount, total: journey.totalCount })}
                </p>
                <p className="text-xs text-muted-foreground">
                  {t("progress.planPath", { plan: t(`planNames.${journey.plan}`) })}
                </p>
              </div>
              <Progress
                value={progressPercent}
                className="bg-success/15 [&_[data-slot=progress-indicator]]:bg-success"
              />
              <div>
                <StatusBadge status={journey.activeStepID ? "info" : "success"}>
                  {journey.activeStepID
                    ? t("progress.currentStep", { step: t(`steps.${journey.activeStepID}.title`) })
                    : t("progress.state.complete")}
                </StatusBadge>
              </div>
            </div>
          </AppCard>

          <AppCard title={t("journey.title")} description={t("journey.description") || undefined}>
            <ol className="flex flex-col gap-2">
              {journey.steps.map((step, index) => (
                <li key={step.id} className="relative ps-12 sm:ps-14">
                  {index < journey.steps.length - 1 ? (
                    <span className="absolute bottom-0 start-5 top-12 w-px bg-border sm:start-6" aria-hidden="true" />
                  ) : null}
                  <StepMarker status={step.status} label={numberFormatter.format(index + 1)} />
                  {step.status === "active" ? (
                    <QuickStartStepActive
                      step={step}
                      t={t}
                      docsBaseHref={props.docsBaseHref}
                      featuredProviders={step.id === "connect_provider" ? featuredProviders : []}
                      onAction={() => handleStepAction(step)}
                      onSkip={() => skipStep(step.id)}
                    />
                  ) : step.status === "done" ? (
                    <QuickStartStepDone step={step} t={t} onAction={() => handleStepAction(step)} />
                  ) : (
                    <QuickStartStepPending step={step} t={t} onResume={() => resumeStep(step.id)} />
                  )}
                </li>
              ))}
            </ol>
          </AppCard>

          {exploreChips.length > 0 ? (
            <section className="rounded-lg border border-dashed border-border p-4">
              <p className="text-xs font-medium text-muted-foreground">{t("explore.afterJourney")}</p>
              <div className="mt-3 flex flex-wrap gap-2">
                {exploreChips.map((cardID) => (
                  <ExploreChip key={cardID} cardID={cardID} t={t} />
                ))}
              </div>
            </section>
          ) : null}
        </div>

        <section
          id="test_bench"
          ref={testBenchRef}
          className="min-w-0 lg:sticky lg:top-24"
          aria-label={t("lab.title")}
        >
          <AppCard
            title={
              <span className="flex items-center gap-2">
                <TerminalSquare className="text-info" aria-hidden="true" />
                <span>{t("lab.title")}</span>
              </span>
            }
            description={t("lab.description")}
          >
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

              <div className="flex flex-col gap-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <Label htmlFor="quick-start-token" className="text-sm font-semibold">
                    {t("lab.keyTypeTitle")}
                  </Label>
                  {journey.testBench.availableTokenTypes.includes("pat") ? (
                    <Button type="button" variant="outline" size="sm" onClick={() => setIsPatDialogOpen(true)}>
                      <KeyRound data-icon="inline-start" />
                      {t("actions.createPat")}
                    </Button>
                  ) : null}
                </div>

                <Tabs
                  value={effectiveSelectedTokenType}
                  onValueChange={(value) => setSelectedTokenType(value as TokenType)}
                  className="gap-3"
                >
                  <TabsList className="h-9 w-full rounded-md border border-border bg-muted p-1">
                    {journey.testBench.availableTokenTypes.map((tokenType) => (
                      <TabsTrigger key={tokenType} value={tokenType} className="h-7 flex-1 rounded-md px-3 text-xs">
                        {t(`tokenLabels.${tokenType}`)}
                      </TabsTrigger>
                    ))}
                  </TabsList>
                </Tabs>

                <p className="text-xs leading-6 text-muted-foreground">
                  {journey.recommendedTokenType === "vat" ? t("lab.recommendedVat") : t("lab.recommendedPat")}
                </p>

                <div className="flex flex-col gap-2 text-start">
                  <Label htmlFor="quick-start-token" className="text-start">
                    {t(`lab.tokenInputLabel.${effectiveSelectedTokenType}`)}
                  </Label>
                  <Input
                    id="quick-start-token"
                    ref={tokenInputRef}
                    className="text-start"
                    value={manualTokens[effectiveSelectedTokenType]}
                    onChange={(event) =>
                      setManualTokens((current) => ({
                        ...current,
                        [effectiveSelectedTokenType]: event.target.value,
                      }))
                    }
                    placeholder={t(`lab.tokenInputPlaceholder.${effectiveSelectedTokenType}`)}
                  />
                  <p className="text-xs leading-6 text-muted-foreground">{t("lab.tokenLocalOnly")}</p>
                </div>
              </div>

              {generatedPatSecret ? (
                <div className="rounded-lg border border-success/30 bg-success-muted p-4">
                  <div className="flex flex-col gap-3">
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-success">{t("lab.generatedTokenTitle")}</p>
                      <p className="mt-1 text-xs leading-6 text-foreground">{t("lab.generatedTokenDescription")}</p>
                      <code className="mt-3 block overflow-auto rounded-md border border-border bg-background px-3 py-2 text-xs" dir="ltr">
                        {generatedPatSecret}
                      </code>
                    </div>
                    <div>
                      <Button type="button" variant="outline" size="sm" onClick={() => void copyText("generated-token", generatedPatSecret)}>
                        <Copy data-icon="inline-start" />
                        {copiedKey === "generated-token" ? t("actions.copied") : t("actions.copyToken")}
                      </Button>
                    </div>
                  </div>
                </div>
              ) : null}

              <div className="flex flex-col gap-3">
                <div className="flex flex-col gap-2 text-start">
                  <Label htmlFor="quick-start-model">{t("lab.modelLabel")}</Label>
                  <Input id="quick-start-model" dir="ltr" readOnly value={suggestedModel} className="text-start font-mono" />
                </div>
                <Button
                  type="button"
                  onClick={() => void runSampleRequest()}
                  disabled={!journey.testBench.ready || !effectiveToken || isRunningSample}
                >
                  {isRunningSample ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
                  {isRunningSample ? t("lab.runningAction") : t("lab.runAction")}
                </Button>
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
                  <div>
                    <p className="text-sm font-semibold text-success">{t("lab.responseTitle")}</p>
                    <p className="mt-1 text-xs leading-6 text-foreground">
                      {t("lab.responseDescription", {
                        status: sampleRunResult.status,
                        tokenType: t(`tokenLabels.${sampleRunResult.tokenType}`),
                      })}
                    </p>
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
          </AppCard>
        </section>
      </div>

      <Dialog open={isPatDialogOpen} onOpenChange={(open) => { setIsPatDialogOpen(open); if (!open) setPatDialogNotice(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("patDialog.title")}</DialogTitle>
            <DialogDescription>{t("patDialog.description")}</DialogDescription>
          </DialogHeader>
          <OperationNotice notice={patDialogNotice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />
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
    </div>
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
