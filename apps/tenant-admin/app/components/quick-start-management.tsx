"use client";

import Link from "next/link";
import { useMemo, useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  ArrowUpRight,
  BookOpen,
  Building2,
  CheckCircle2,
  CircleDollarSign,
  Copy,
  Gauge,
  KeyRound,
  Layers3,
  Loader2,
  Network,
  Play,
  RefreshCcw,
  ShieldCheck,
  Sparkles,
  TerminalSquare,
  Users,
} from "lucide-react";

import { CodeBlock } from "@/app/components/code-block";
import { StatCard } from "@/app/components/stat-card";
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
import {
  buildBashCurlGuideCommand,
  getTenantApiBaseURL,
} from "@/app/components/provider-account-wizard/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Textarea } from "@/components/ui/textarea";
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
};

type TokenType = "pat" | "vat";
type SnippetTab = "bash" | "python";
type TranslateValues = Record<string, string | number | Date | null | undefined>;
type TranslateFn = (key: string, values?: TranslateValues) => string;
type SampleRunResult = {
  content: string | null;
  raw: string;
  status: number;
  tokenType: TokenType;
};

const samplePrompt = "برای من سه مزیت اصلی استفاده از gateway را بنویس.";

export default function QuickStartManagement(props: QuickStartManagementProps) {
  const t = useTranslations("quickStart");
  const common = useTranslations("common");
  const router = useRouter();
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
  const [isRefreshing, startRefreshing] = useTransition();

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

  const tokenPlaceholder = effectiveSelectedTokenType === "pat" ? "YOUR_SHAHRAH_PAT" : "YOUR_SHAHRAH_VAT";
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

  const stepCountLabel = t("progress.value", {
    done: formatCount(journey.doneCount),
    total: formatCount(journey.totalCount),
  });

  function refreshStatus() {
    setBanner(null);
    setSampleRunResult(null);
    startRefreshing(() => {
      router.refresh();
    });
  }

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
      setBanner({ kind: "success", message: t("messages.patCreated") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.patCreateFailed")) });
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
      document.getElementById(action.anchor)?.scrollIntoView({ behavior: "smooth", block: "start" });
      return;
    }
  }

  async function copyText(key: string, value: string) {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedKey(key);
      window.setTimeout(() => setCopiedKey((current) => (current === key ? "" : current)), 1600);
    } catch {
      setBanner({ kind: "error", message: t("messages.copyFailed") });
    }
  }

  async function runSampleRequest() {
    if (isRunningSample) return;
    if (!journey.testBench.ready) {
      setBanner({ kind: "error", message: t("lab.notReadyTitle") });
      return;
    }
    if (!effectiveToken) {
      setBanner({ kind: "error", message: t(`lab.missingToken.${effectiveSelectedTokenType}`) });
      return;
    }

    setIsRunningSample(true);
    setBanner(null);
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
      setBanner({ kind: "success", message: t("messages.sampleRequestSucceeded") });
    } catch (error) {
      setBanner({ kind: "error", message: normalizeError(error, t("messages.sampleRequestFailed")) });
    } finally {
      setIsRunningSample(false);
    }
  }

  return (
    <div className="flex flex-col gap-6 px-1 py-1 sm:px-2 sm:py-2">
      <section className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(21rem,0.8fr)]">
        <Card className="overflow-hidden rounded-2xl border-info/30 bg-info/5 shadow-sm">
          <CardHeader className="gap-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="outline">{t(`planNames.${journey.plan}`)}</Badge>
              <Badge variant="secondary">{stepCountLabel}</Badge>
              <Badge variant={journey.doneCount === journey.totalCount ? "success" : "outline"}>
                {t(`progress.state.${journey.doneCount === journey.totalCount ? "complete" : "inProgress"}`)}
              </Badge>
            </div>
            <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0">
                <CardDescription className="text-sm font-semibold text-info">{t("eyebrow")}</CardDescription>
                <CardTitle className="mt-1 text-2xl tracking-tight sm:text-3xl">{t("title")}</CardTitle>
                <CardDescription className="mt-3 max-w-3xl leading-7 text-foreground/80">
                  {t("subtitle")}
                </CardDescription>
                <p className="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground">
                  {t(`planHints.${journey.plan}`)}
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-2">
                <Button type="button" variant="outline" onClick={refreshStatus} disabled={isRefreshing}>
                  {isRefreshing ? <Loader2 className="size-4 animate-spin" /> : <RefreshCcw className="size-4" />}
                  {t("actions.refreshStatus")}
                </Button>
              </div>
            </div>
          </CardHeader>
        </Card>

        <div className="grid gap-4 sm:grid-cols-3 xl:grid-cols-1">
          <StatCard
            title={t("stats.providers.title")}
            value={formatCount(resources.providerAccountCount)}
            subtitle={t("stats.providers.subtitle")}
            icon={Network}
            color={resources.providerAccountCount > 0 ? "green" : "amber"}
          />
          <StatCard
            title={t("stats.tokens.title")}
            value={formatCount(resources.personalAccessTokenCount + resources.virtualUserCount)}
            subtitle={t("stats.tokens.subtitle")}
            icon={KeyRound}
            color={resources.personalAccessTokenCount + resources.virtualUserCount > 0 ? "green" : "amber"}
          />
          <StatCard
            title={t("stats.controls.title")}
            value={formatCount(resources.rateLimitRuleCount + resources.budgetLimitRuleCount + resources.guardrailCount)}
            subtitle={t("stats.controls.subtitle")}
            icon={ShieldCheck}
            color={resources.rateLimitRuleCount + resources.budgetLimitRuleCount + resources.guardrailCount > 0 ? "green" : "indigo"}
          />
        </div>
      </section>

      {banner ? (
        <Alert variant={banner.kind === "error" ? "destructive" : "default"}>
          <AlertTitle>{banner.kind === "error" ? t("alerts.error") : t("alerts.success")}</AlertTitle>
          <AlertDescription>{banner.message}</AlertDescription>
        </Alert>
      ) : null}

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
        <Card className="rounded-2xl border-border/70 shadow-sm">
          <CardHeader>
            <CardTitle>{t("journey.title")}</CardTitle>
            <CardDescription>{t("journey.description")}</CardDescription>
          </CardHeader>
          <CardContent>
            <ol className="space-y-4">
              {journey.steps.map((step, index) => (
                <li key={step.id} className="relative ps-16">
                  {index < journey.steps.length - 1 ? (
                    <span className="absolute start-[1.65rem] top-12 h-[calc(100%-1.5rem)] w-px bg-border" aria-hidden="true" />
                  ) : null}
                  <span
                    className={cn(
                      "absolute start-0 top-1 flex size-12 items-center justify-center rounded-2xl border text-sm font-semibold",
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
                  <div className="rounded-2xl border border-border/70 bg-card/95 p-4 shadow-sm">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <h3 className="text-base font-semibold">{t(`steps.${step.id}.title`)}</h3>
                          <Badge variant={badgeVariantForStatus(step.status)}>
                            {t(`statuses.${step.status}`)}
                          </Badge>
                        </div>
                        <p className="mt-2 text-sm leading-7 text-muted-foreground">
                          {t(`steps.${step.id}.description`)}
                        </p>
                        <p className="mt-3 rounded-xl bg-muted/40 px-3 py-2 text-xs leading-6 text-foreground/80">
                          <strong>{t("journey.learnLabel")}</strong> {t(`steps.${step.id}.learn`)}
                        </p>
                        {step.status === "unavailable" && step.blockedBy ? (
                          <p className="mt-3 text-xs leading-6 text-destructive">
                            {t(`blocked.${step.blockedBy}`)}
                          </p>
                        ) : null}
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
                  </div>
                </li>
              ))}
            </ol>
          </CardContent>
        </Card>

        <div id="test_bench" className="space-y-6">
          <Card className="rounded-2xl border-border/70 shadow-sm">
            <CardHeader>
              <div className="flex items-start justify-between gap-3">
                <div>
                  <CardTitle className="flex items-center gap-2">
                    <TerminalSquare className="size-5 text-info" />
                    {t("lab.title")}
                  </CardTitle>
                  <CardDescription className="mt-2 leading-7">{t("lab.description")}</CardDescription>
                </div>
                <Badge variant="outline">{t(`tokenLabels.${journey.testBench.recommendedTokenType}`)}</Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {!journey.testBench.ready ? (
                <Alert>
                  <AlertTitle>{t("lab.notReadyTitle")}</AlertTitle>
                  <AlertDescription>
                    {resources.providerAccountCount === 0
                      ? t("lab.notReadyProvider")
                      : t("lab.notReadyModels")}
                  </AlertDescription>
                </Alert>
              ) : null}

              <div className="rounded-2xl border border-border/70 bg-muted/20 p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div>
                    <p className="text-sm font-semibold">{t("lab.keyTypeTitle")}</p>
                    <p className="mt-1 text-xs leading-6 text-muted-foreground">
                      {journey.recommendedTokenType === "vat" ? t("lab.recommendedVat") : t("lab.recommendedPat")}
                    </p>
                  </div>
                  {journey.testBench.availableTokenTypes.includes("pat") ? (
                    <Button type="button" variant="outline" size="sm" onClick={() => setIsPatDialogOpen(true)}>
                      <KeyRound className="size-4" />
                      {t("actions.createPat")}
                    </Button>
                  ) : null}
                </div>

                <Tabs value={effectiveSelectedTokenType} onValueChange={(value) => setSelectedTokenType(value as TokenType)} className="mt-4 gap-3">
                  <TabsList variant="line" className="justify-start">
                    {journey.testBench.availableTokenTypes.map((tokenType) => (
                      <TabsTrigger key={tokenType} value={tokenType}>
                        {t(`tokenLabels.${tokenType}`)}
                      </TabsTrigger>
                    ))}
                  </TabsList>

                  {journey.testBench.availableTokenTypes.map((tokenType) => (
                    <TabsContent key={tokenType} value={tokenType} className="space-y-3">
                      <div className="space-y-2">
                        <Label htmlFor={`quick-start-token-${tokenType}`}>{t(`lab.tokenInputLabel.${tokenType}`)}</Label>
                        <Textarea
                          id={`quick-start-token-${tokenType}`}
                          dir="ltr"
                          rows={3}
                          value={tokenType === "pat" ? manualTokens.pat : manualTokens.vat}
                          onChange={(event) =>
                            setManualTokens((current) => ({
                              ...current,
                              [tokenType]: event.target.value,
                            }))
                          }
                          placeholder={t(`lab.tokenInputPlaceholder.${tokenType}`)}
                        />
                        <p className="text-xs leading-6 text-muted-foreground">{t("lab.localOnlyHint")}</p>
                      </div>
                    </TabsContent>
                  ))}
                </Tabs>
              </div>

              {generatedPatSecret ? (
                <div className="rounded-2xl border border-success/30 bg-success-muted p-4">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-success">{t("lab.generatedTokenTitle")}</p>
                      <p className="mt-1 text-xs leading-6 text-foreground/80">{t("lab.generatedTokenDescription")}</p>
                      <code className="mt-3 block overflow-auto rounded-xl bg-background/80 px-3 py-2 text-xs" dir="ltr">
                        {generatedPatSecret}
                      </code>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => void copyText("generated-token", generatedPatSecret)}
                    >
                      <Copy className="size-4" />
                      {copiedKey === "generated-token" ? t("actions.copied") : t("actions.copyToken")}
                    </Button>
                  </div>
                </div>
              ) : null}

              <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
                <div className="space-y-2">
                  <Label htmlFor="quick-start-model">{t("lab.modelLabel")}</Label>
                  <Input id="quick-start-model" dir="ltr" readOnly value={suggestedModel} />
                </div>
                <div className="flex gap-2">
                  <Button type="button" onClick={() => void runSampleRequest()} disabled={!journey.testBench.ready || !effectiveToken || isRunningSample}>
                    {isRunningSample ? <Loader2 className="size-4 animate-spin" /> : <Play className="size-4" />}
                    {isRunningSample ? t("lab.runningAction") : t("lab.runAction")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      void copyText(
                        `snippet-${selectedSnippetTab}`,
                        selectedSnippetTab === "bash" ? bashSnippet : pythonSnippet,
                      )
                    }
                    disabled={!journey.testBench.ready}
                  >
                    <Copy className="size-4" />
                    {copiedKey === `snippet-${selectedSnippetTab}` ? t("actions.copied") : t("actions.copySnippet")}
                  </Button>
                </div>
              </div>

              <p className="text-xs leading-6 text-muted-foreground">{t("lab.browserRunHint")}</p>

              <Tabs value={selectedSnippetTab} onValueChange={(value) => setSelectedSnippetTab(value as SnippetTab)}>
                <TabsList variant="line" className="justify-start">
                  <TabsTrigger value="bash">{t("lab.tabs.bash")}</TabsTrigger>
                  <TabsTrigger value="python">{t("lab.tabs.python")}</TabsTrigger>
                </TabsList>
                <TabsContent value="bash">
                  <CodeBlock code={bashSnippet} language="bash" />
                </TabsContent>
                <TabsContent value="python">
                  <CodeBlock code={pythonSnippet} language="python" />
                </TabsContent>
              </Tabs>

              {!effectiveToken ? (
                <p className="text-xs leading-6 text-muted-foreground">
                  {t(`lab.placeholderHint.${effectiveSelectedTokenType}`)}
                </p>
              ) : null}

              {sampleRunResult ? (
                <div className="space-y-3 rounded-2xl border border-success/30 bg-success-muted p-4">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div>
                      <p className="text-sm font-semibold text-success">{t("lab.responseTitle")}</p>
                      <p className="mt-1 text-xs leading-6 text-foreground/80">
                        {t("lab.responseDescription", {
                          status: sampleRunResult.status,
                          tokenType: t(`tokenLabels.${sampleRunResult.tokenType}`),
                        })}
                      </p>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => void copyText("sample-response", sampleRunResult.raw)}
                    >
                      <Copy className="size-4" />
                      {copiedKey === "sample-response" ? t("actions.copied") : t("lab.copyResponse")}
                    </Button>
                  </div>

                  {sampleRunResult.content ? (
                    <div className="rounded-xl bg-background/80 p-3">
                      <p className="text-xs font-semibold text-muted-foreground">{t("lab.responsePreviewLabel")}</p>
                      <p className="mt-2 whitespace-pre-wrap text-sm leading-7">{sampleRunResult.content}</p>
                    </div>
                  ) : null}

                  <div>
                    <p className="mb-2 text-xs font-semibold text-muted-foreground">{t("lab.rawResponseLabel")}</p>
                    <CodeBlock code={sampleRunResult.raw} language="json" />
                  </div>
                </div>
              ) : null}
            </CardContent>
          </Card>

          <Card className="rounded-2xl border-border/70 shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Sparkles className="size-5 text-info" />
                {t("explore.title")}
              </CardTitle>
              <CardDescription>{t("explore.description")}</CardDescription>
            </CardHeader>
            <CardContent className="@container">
              <div className="grid gap-3 @[28rem]:grid-cols-2 @[48rem]:grid-cols-2 @[64rem]:grid-cols-3 @[82rem]:grid-cols-4">
              {journey.exploreCards.map((cardID) => (
                <ExploreCard key={cardID} cardID={cardID} t={t} />
              ))}
              </div>
            </CardContent>
          </Card>
        </div>
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
              {isCreatingPat ? <Loader2 className="size-4 animate-spin" /> : <KeyRound className="size-4" />}
              {isCreatingPat ? t("patDialog.creating") : t("patDialog.submit")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
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

  if (props.action.type === "link") {
    if (isDisabled) {
      return (
        <Button type="button" variant={props.status === "done" ? "outline" : "default"} className="shrink-0" disabled>
          {label}
          <ArrowUpRight className="size-4" />
        </Button>
      );
    }
    return (
      <Button
        type="button"
        asChild
        variant={props.status === "done" ? "outline" : "default"}
        className="shrink-0"
      >
        <Link href={props.action.href} onClick={props.onAction}>
          {label}
          <ArrowUpRight className="size-4" />
        </Link>
      </Button>
    );
  }

  return (
    <Button
      type="button"
      variant={props.status === "done" ? "outline" : "default"}
      className="shrink-0"
      onClick={props.onAction}
      disabled={isDisabled}
    >
      {label}
      {props.action.type === "anchor" ? <TerminalSquare className="size-4" /> : <KeyRound className="size-4" />}
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
    prompts: BookOpen,
    metrics: Gauge,
    logs: Layers3,
    billing_plan: CircleDollarSign,
  };

  const Icon = iconMap[props.cardID];

  return (
    <Link
      href={hrefMap[props.cardID]}
      className="group flex h-full min-h-32 flex-col rounded-2xl border border-border/70 bg-card/90 p-4 transition-colors hover:border-primary/30 hover:bg-muted/20 @[28rem]:min-h-36"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Icon className="size-5" />
          </span>
          <p className="min-w-0 text-sm font-semibold leading-6 @[48rem]:text-[0.95rem]">
            {props.t(`explore.cards.${props.cardID}.title`)}
          </p>
        </div>
        <ArrowUpRight className="mt-0.5 size-4 shrink-0 text-muted-foreground transition group-hover:text-foreground" />
      </div>
      <p className="mt-3 text-xs leading-6 text-muted-foreground @[48rem]:text-[0.8rem]">
        {props.t(`explore.cards.${props.cardID}.description`)}
      </p>
    </Link>
  );
}

function badgeVariantForStatus(status: QuickStartStepStatus): "outline" | "secondary" | "success" | "destructive" {
  if (status === "done") return "success";
  if (status === "active") return "secondary";
  if (status === "unavailable") return "destructive";
  return "outline";
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
