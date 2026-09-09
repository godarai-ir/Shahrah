"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { copyTextToClipboard } from "@/components/app-ui/CopyableValue";
import { useLocale, useTranslations } from "next-intl";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Braces,
  Clipboard,
  Copy,
  Download,
  FileCode2,
  GitCompare,
  History,
  MessageSquareText,
  Play,
  Settings2,
} from "lucide-react";

import { getPrompt, listPromptUsers, listPromptVersions } from "@/features/prompts/api";
import { getAPIErrorMessage } from "@/lib/http/api-client";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { CodePanel } from "@/components/app-ui/CodePanel";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { SectionHeader } from "@/components/app-ui/SectionHeader";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { mobileAuthUIEnabled } from "@/lib/auth/mobile-auth-ui";
import { cn } from "@/lib/utils";

import {
  buildGatewaySnippet,
  buildPromptDiffRows,
  buildSdkSnippet,
  describePromptMessage,
  serializePromptSpec,
  type PromptCompareRow,
  type PromptDetailTab,
} from "./detail-utils";
import {
  type Notice,
  type Prompt,
  type PromptTranslationFn,
  type PromptVersion,
} from "@/lib/prompt-builder";
import { getLocaleDirection } from "@/lib/locale-direction";

type PromptDetailPageProps = {
  promptID: string;
};

export default function PromptDetailPage(props: PromptDetailPageProps) {
  const t = useTranslations("prompts");
  const locale = useLocale();
  const router = useRouter();
  const pageDir = getLocaleDirection(locale);

  const [activeTab, setActiveTab] = useState<PromptDetailTab>("prompt");
  const [compareBaseVersionID, setCompareBaseVersionID] = useState("");
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareTargetVersionID, setCompareTargetVersionID] = useState("");
  const [historyOpen, setHistoryOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [notice, setNotice] = useState<Notice>(null);
  const [prompt, setPrompt] = useState<Prompt | null>(null);
  const [selectedVersionID, setSelectedVersionID] = useState("");
  const [userLabelByID, setUserLabelByID] = useState<Record<string, string>>({});
  const [versionSearch, setVersionSearch] = useState("");
  const [versions, setVersions] = useState<PromptVersion[]>([]);

  const selectedVersion = useMemo(
    () => versions.find((item) => item.id === selectedVersionID) ?? prompt?.latest_version_detail ?? null,
    [prompt?.latest_version_detail, selectedVersionID, versions],
  );

  const selectedDescriptor = useMemo(
    () => (selectedVersion ? describePromptMessage(selectedVersion, t) : null),
    [selectedVersion, t],
  );

  const selectedSpec = useMemo(
    () => (selectedVersion ? serializePromptSpec(selectedVersion) : ""),
    [selectedVersion],
  );

  const isSelectedLatestVersion =
    Boolean(selectedVersion) && typeof prompt?.latest_version === "number" && selectedVersion?.version === prompt.latest_version;
  const selectedVersionPlaygroundHref = selectedVersion?.id
    ? `/playground?${new URLSearchParams({ promptID: props.promptID, versionID: selectedVersion.id }).toString()}`
    : "/playground";

  const filteredVersions = useMemo(() => {
    const query = versionSearch.trim().toLowerCase();
    if (!query) return versions;
    return versions.filter((version) => {
      const versionLabel = `${version.version}`.toLowerCase();
      const commit = (version.commit_message ?? "").toLowerCase();
      const creator = (userLabelByID[version.created_by ?? ""] ?? version.created_by ?? "").toLowerCase();
      return versionLabel.includes(query) || commit.includes(query) || creator.includes(query);
    });
  }, [userLabelByID, versionSearch, versions]);

  const compareBaseVersion = versions.find((item) => item.id === compareBaseVersionID) ?? selectedVersion ?? null;
  const compareTargetVersion =
    versions.find((item) => item.id === compareTargetVersionID)
    ?? versions.find((item) => item.id !== compareBaseVersion?.id)
    ?? null;

  const compareRows = useMemo<PromptCompareRow[]>(() => {
    if (!compareBaseVersion || !compareTargetVersion) return [];
    return buildPromptDiffRows(serializePromptSpec(compareBaseVersion), serializePromptSpec(compareTargetVersion));
  }, [compareBaseVersion, compareTargetVersion]);

  const loadPrompt = useCallback(async () => {
    const nextPrompt = await getPrompt(props.promptID);
    setPrompt(nextPrompt);
    setSelectedVersionID((current) => current || nextPrompt.latest_version_detail?.id || "");
    return nextPrompt;
  }, [props.promptID]);

  const loadUserLabels = useCallback(async (nextPrompt: Prompt | null, nextVersions: PromptVersion[]) => {
    const ids = [
      nextPrompt?.created_by ?? "",
      ...(nextVersions.map((item) => item.created_by ?? "")),
      nextPrompt?.latest_version_detail?.created_by ?? "",
    ]
      .map((value) => value.trim())
      .filter(Boolean);
    const uniqueIDs = [...new Set(ids)];
    if (uniqueIDs.length === 0) return;
    try {
      const payload = await listPromptUsers();
      const users = payload.data ?? [];
      setUserLabelByID(
        Object.fromEntries(
          uniqueIDs.map((id) => {
            const match = users.find((user) => user.id === id);
            return [id, match?.display_name?.trim() || match?.email?.trim() || (mobileAuthUIEnabled ? match?.mobile?.trim() : "") || id];
          }),
        ),
      );
    } catch {
      setUserLabelByID(Object.fromEntries(uniqueIDs.map((id) => [id, id])));
    }
  }, []);

  const loadPage = useCallback(async () => {
    setIsLoading(true);
    try {
      const loadedPrompt = await loadPrompt();
      const payload = await listPromptVersions(props.promptID);
      const nextVersions = payload.data ?? [];
      setVersions(nextVersions);
      const nextSelectedVersion = nextVersions.find((item) => item.id === loadedPrompt?.latest_version_detail?.id) ?? nextVersions[0] ?? null;
      if (nextSelectedVersion) {
        setSelectedVersionID(nextSelectedVersion.id);
        setCompareBaseVersionID(nextSelectedVersion.id);
        setCompareTargetVersionID(nextVersions.find((item) => item.id !== nextSelectedVersion.id)?.id || "");
      } else {
        setSelectedVersionID("");
        setCompareBaseVersionID("");
        setCompareTargetVersionID("");
      }
      await loadUserLabels(loadedPrompt, nextVersions);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("messages.loadPromptsError")) });
    } finally {
      setIsLoading(false);
    }
  }, [loadPrompt, loadUserLabels, props.promptID, t]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadPage();
    }, 0);

    return () => {
      window.clearTimeout(timer);
    };
  }, [loadPage]);

  return (
    <div dir={pageDir} className="flex flex-col gap-5">
      <section className="overflow-hidden rounded-xl border border-card-border bg-card shadow-xs">
        <div className="h-1 bg-primary" aria-hidden="true" />
        <div className="flex flex-col gap-5 p-4 sm:p-6">
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink asChild>
                  <Link href="/prompts">{t("header.title")}</Link>
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage dir="auto">{prompt?.name || t("header.title")}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex min-w-0 flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="flex min-w-0 flex-1 flex-col gap-3">
              <div className="flex min-w-0 flex-wrap items-center gap-2">
                <h2 className="min-w-0 break-words text-2xl font-semibold tracking-tight sm:text-3xl">
                  {prompt?.name || t("header.title")}
                </h2>
                {selectedVersion ? (
                  <>
                    <Badge variant="secondary">
                      {t("list.versionValue", { version: selectedVersion.version })}
                    </Badge>
                    <Badge variant={isSelectedLatestVersion ? "success" : "secondary"}>
                      {isSelectedLatestVersion ? t("detail.status.latest") : t("detail.status.previous")}
                    </Badge>
                    <Badge variant="outline">{t("detail.readOnlyBadge")}</Badge>
                  </>
                ) : null}
              </div>
              <p className="max-w-3xl text-sm leading-6 text-muted-foreground">
                {selectedDescriptor?.description || prompt?.description || t("detail.description")}
              </p>
            </div>

            <Button
              type="button"
              className="w-full lg:w-auto lg:shrink-0"
              onClick={() => router.push(selectedVersionPlaygroundHref)}
            >
              <Play data-icon="inline-start" />
              {t("actions.tryPrompt")}
            </Button>
          </div>

          <div className="grid grid-cols-1 gap-2 border-t border-border/60 pt-4 sm:grid-cols-2 xl:grid-cols-3">
            <Button type="button" size="sm" variant="outline" className="w-full" onClick={() => setHistoryOpen(true)}>
              <History data-icon="inline-start" />
              {t("actions.versionHistory")}
            </Button>
            <Button type="button" size="sm" variant="outline" className="w-full" onClick={() => void copySelectedFqn()} disabled={!selectedVersion?.fqn}>
              <Clipboard data-icon="inline-start" />
              {t("actions.copyFqn")}
            </Button>
            <Button type="button" size="sm" variant="outline" className="w-full" onClick={() => router.push(selectedVersionPlaygroundHref)}>
              <FileCode2 data-icon="inline-start" />
              {t("actions.createVersion")}
            </Button>
          </div>

          {notice ? (
            <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
              <AlertTitle>{notice.kind === "error" ? t("notice.errorTitle") : t("notice.successTitle")}</AlertTitle>
              <AlertDescription>{notice.message}</AlertDescription>
            </Alert>
          ) : null}
        </div>
      </section>

      {isLoading ? (
        <div className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">{t("detail.loading")}</div>
      ) : selectedVersion ? (
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as PromptDetailTab)} className="min-w-0 overflow-hidden rounded-2xl border bg-card p-4">
          <TabsList variant="line" className="grid h-auto w-full grid-cols-2 gap-0 rounded-none border-b p-0 sm:flex sm:justify-start sm:gap-2">
            <TabsTrigger value="prompt" className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
              {t("detail.tabs.prompt")}
            </TabsTrigger>
            <TabsTrigger value="spec" className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
              {t("detail.tabs.spec")}
            </TabsTrigger>
            <TabsTrigger value="sdk" className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
              {t("detail.tabs.sdk")}
            </TabsTrigger>
            <TabsTrigger value="gateway" className="rounded-none px-2 py-2 data-[state=active]:text-foreground">
              {t("detail.tabs.gateway")}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="prompt">
            <PromptContent
              descriptor={selectedDescriptor}
              locale={locale}
              prompt={prompt}
              selectedVersion={selectedVersion}
              t={t}
              userLabelByID={userLabelByID}
            />
          </TabsContent>

          <TabsContent value="spec">
            <CodePanel
              actions={
                <>
                  <Button type="button" variant="outline" size="sm" onClick={() => void copyText(selectedSpec, t("messages.specCopied"))}>
                    <Copy data-icon="inline-start" />
                    {t("actions.copy")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => downloadText(`${selectedVersion.name || "prompt"}-v${selectedVersion.version}.json`, selectedSpec)}
                  >
                    <Download data-icon="inline-start" />
                    {t("actions.download")}
                  </Button>
                </>
              }
              code={selectedSpec}
              description={t("detail.specDescription")}
              language="json"
              displayLanguage="JSON"
              title={t("detail.tabs.spec")}
            />
          </TabsContent>

          <TabsContent value="sdk">
            <CodePanel
              actions={
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void copyText(buildSdkSnippet(selectedVersion), t("messages.snippetCopied"))}
                >
                  <Copy data-icon="inline-start" />
                  {t("actions.copy")}
                </Button>
              }
              code={buildSdkSnippet(selectedVersion)}
              description={t("detail.sdkDescription")}
              language="python"
              displayLanguage="Python"
              title={t("detail.tabs.sdk")}
            />
          </TabsContent>

          <TabsContent value="gateway">
            <GatewayContent
              selectedVersion={selectedVersion}
              t={t}
              onCopy={(text) => void copyText(text, t("messages.snippetCopied"))}
            />
          </TabsContent>
        </Tabs>
      ) : (
        <div className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">{t("detail.empty")}</div>
      )}

      <PromptVersionHistorySheet
        currentVersionID={selectedVersionID}
        isOpen={historyOpen}
        onClose={() => setHistoryOpen(false)}
        onCompareOpen={() => {
          setCompareOpen(true);
          setHistoryOpen(false);
        }}
        onCreateVersion={() => router.push(selectedVersionPlaygroundHref)}
        onSelectVersion={(version) => {
          setSelectedVersionID(version.id);
          setHistoryOpen(false);
        }}
        search={versionSearch}
        setSearch={setVersionSearch}
        locale={locale}
        t={t}
        userLabelByID={userLabelByID}
        versions={filteredVersions}
      />

      <PromptCompareDialog
        baseVersion={compareBaseVersion}
        baseVersionID={compareBaseVersionID}
        isOpen={compareOpen}
        onClose={() => setCompareOpen(false)}
        onSelectBaseVersion={setCompareBaseVersionID}
        onSelectTargetVersion={setCompareTargetVersionID}
        rows={compareRows}
        t={t}
        targetVersion={compareTargetVersion}
        targetVersionID={compareTargetVersionID}
        versions={versions}
      />
    </div>
  );

  async function copySelectedFqn() {
    if (!selectedVersion?.fqn) {
      showFeedbackToast({ kind: "error", title: t("messages.fqnUnavailable") });
      return;
    }
    await copyText(selectedVersion.fqn, t("messages.fqnCopied"));
  }

  async function copyText(text: string, successMessage: string) {
    await copyTextToClipboard(text);
    showFeedbackToast({ kind: "success", title: successMessage });
  }
}

function PromptContent(props: {
  descriptor: ReturnType<typeof describePromptMessage> | null;
  locale: string;
  prompt: Prompt | null;
  selectedVersion: PromptVersion;
  t: PromptTranslationFn;
  userLabelByID: Record<string, string>;
}) {
  if (!props.descriptor) return null;

  const cachePills = [];
  if (props.descriptor.cache?.enabled) {
    cachePills.push(props.descriptor.cache.type ?? "simple");
    if (props.descriptor.cache.ttl) cachePills.push(`ttl ${props.descriptor.cache.ttl}`);
    if (props.descriptor.cache.namespace) cachePills.push(`ns ${props.descriptor.cache.namespace}`);
  }

  const structuredPills = [];
  if (props.descriptor.structuredOutput?.enabled) {
    structuredPills.push(props.descriptor.structuredOutput.format ?? "json_object");
    structuredPills.push(props.descriptor.structuredOutput.strict ? props.t("common.strict") : props.t("common.nonStrict"));
  }

  return (
    <div className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_380px]">
      <div className="flex min-w-0 flex-col gap-5">
        <Alert>
          <AlertTitle>{props.t("detail.readOnlyTitle", { version: props.selectedVersion.version })}</AlertTitle>
          <AlertDescription>{props.t("detail.readOnlyDescription")}</AlertDescription>
        </Alert>

        <Card className="shadow-none">
          <CardHeader>
            <div className="flex items-center gap-2">
              <MessageSquareText className="size-4 text-muted-foreground" />
              <CardTitle className="text-base">{props.t("builder.systemPrompt.title")}</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border bg-muted/20 p-5 text-sm leading-7" dir="auto">
              {props.descriptor.systemPrompt || "-"}
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{props.t("builder.conversation.title")}</CardTitle>
            <CardDescription>{props.t("builder.conversation.description")}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            {props.descriptor.conversationMessages.length === 0 ? (
              <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">{props.t("builder.conversation.empty")}</div>
            ) : (
              props.descriptor.conversationMessages.map((message, index) => (
                <div key={`${message.role}-${index}`} className="rounded-lg border bg-background p-4">
                  <div className="mb-2">
                    <Badge variant={message.role === "assistant" ? "secondary" : "outline"}>
                      {message.role === "assistant" ? props.t("roles.assistant") : props.t("roles.user")}
                    </Badge>
                  </div>
                  <div className="text-sm leading-7" dir="auto">
                    {typeof message.content === "string" ? message.content : JSON.stringify(message.content)}
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>

      <div className="flex min-w-0 flex-col gap-5 xl:sticky xl:top-6">
        <Card className="shadow-none">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Settings2 className="size-4 text-muted-foreground" />
              <CardTitle className="text-base">{props.t("detail.versionConfiguration")}</CardTitle>
            </div>
            <CardDescription>{props.descriptor.versionLabel}</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <section className="flex flex-col gap-2"><SectionHeader density="compact" title={props.t("config.model.title")} />
              <div className="rounded-md border bg-muted/20 px-3 py-2 font-mono text-xs" dir="ltr">{props.descriptor.model || "-"}</div>
            </section>
            <section className="flex flex-col gap-2"><SectionHeader density="compact" title={props.t("config.modelSettings.title")} />
              <DefinitionList columns={2}>
                <DefinitionListItem surface label={props.t("config.modelSettings.stream")} value={props.descriptor.modelSettings.stream ? props.t("common.enabled") : props.t("common.disabled")} />
                <DefinitionListItem surface label={props.t("config.modelSettings.temperature")} value={formatOptionalNumber(props.descriptor.modelSettings.temperature)} />
                <DefinitionListItem surface label={props.t("config.modelSettings.topP")} value={formatOptionalNumber(props.descriptor.modelSettings.top_p)} />
                <DefinitionListItem surface label={props.t("config.modelSettings.maxTokens")} value={formatOptionalNumber(props.descriptor.modelSettings.max_tokens)} />
              </DefinitionList>
            </section>
            <section className="flex flex-col gap-2"><SectionHeader density="compact" title={props.t("config.variables.title", { count: props.descriptor.variableNames.length })} />
              {props.descriptor.variableNames.length === 0 ? (
                <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.t("config.variables.empty")}</div>
              ) : (
                <div className="overflow-hidden rounded-md border">
                  <div className="grid grid-cols-[minmax(0,1fr)_96px] bg-muted/40 px-3 py-2 text-xs font-medium text-muted-foreground">
                    <span>{props.t("detail.variableName")}</span>
                    <span>{props.t("detail.variableType")}</span>
                  </div>
                  {props.descriptor.variableNames.map((name) => (
                    <div key={name} className="grid grid-cols-[minmax(0,1fr)_96px] border-t px-3 py-2 text-sm">
                      <span className="truncate font-mono text-xs" dir="ltr" title={name}>{name}</span>
                      <span className="text-xs text-muted-foreground">string</span>
                    </div>
                  ))}
                </div>
              )}
            </section>
            <section className="flex flex-col gap-2"><SectionHeader density="compact" title={props.t("detail.advancedSettings")} />
              <div className="flex flex-col gap-3">
                <PillGroup emptyLabel={props.t("config.cache.empty")} pills={cachePills} title={props.t("config.cache.title")} />
                <PillGroup
                  emptyLabel={props.t("config.inputGuardrails.empty")}
                  pills={props.descriptor.guardrails?.input ?? []}
                  title={props.t("config.inputGuardrails.title", { count: props.descriptor.guardrails?.input?.length ?? 0 })}
                />
                <PillGroup
                  emptyLabel={props.t("config.outputGuardrails.empty")}
                  pills={props.descriptor.guardrails?.output ?? []}
                  title={props.t("config.outputGuardrails.title", { count: props.descriptor.guardrails?.output?.length ?? 0 })}
                />
                <PillGroup emptyLabel={props.t("config.structuredOutput.empty")} pills={structuredPills} title={props.t("config.structuredOutput.title")} />
              </div>
            </section>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Braces className="size-4 text-muted-foreground" />
              <CardTitle className="text-base">{props.t("detail.technicalDetails")}</CardTitle>
            </div>
            <CardDescription>{props.t("detail.technicalDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <DefinitionList>
            <DefinitionListItem surface mono valueDir="ltr" label={props.t("versions.selectedFqn")} value={props.selectedVersion.fqn || "-"} />
            <div className="rounded-lg border border-border bg-muted/20 px-3 py-3">
              <div className="grid gap-3 sm:grid-cols-2">
                <div>
                  <dt className="text-xs text-muted-foreground">{props.t("list.columns.updatedAt")}</dt>
                  <dd className="mt-1 text-sm font-medium">
                    <DateTimeValue value={props.prompt?.updated_at} locale={props.locale} fallback="-" />
                  </dd>
                </div>
                <div className="border-t border-border/60 pt-3 sm:border-s sm:border-t-0 sm:ps-3 sm:pt-0">
                  <dt className="text-xs text-muted-foreground">{props.t("detail.summary.createdAt")}</dt>
                  <dd className="mt-1 text-sm font-medium">
                    <DateTimeValue value={props.selectedVersion.created_at} locale={props.locale} fallback="-" />
                  </dd>
                </div>
              </div>
            </div>
            <DefinitionListItem surface
              label={props.t("detail.summary.versionAuthor")}
              value={props.userLabelByID[props.selectedVersion.created_by ?? ""] || props.selectedVersion.created_by || "-"}
            />
            </DefinitionList>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{props.t("config.metadata.title")}</CardTitle>
          </CardHeader>
          <CardContent>
            {Object.keys(props.descriptor.metadata).length === 0 ? (
              <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.t("config.metadata.empty")}</div>
            ) : (
              <div className="flex flex-col gap-2">
                {Object.entries(props.descriptor.metadata).map(([key, value]) => (
                  <div key={key} className="rounded-md border bg-muted/20 px-3 py-2 text-sm">
                    <div className="font-medium">{key}</div>
                    <div dir="ltr" className="mt-1 font-mono text-xs text-muted-foreground">{JSON.stringify(value)}</div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function PillGroup(props: { emptyLabel: string; pills: string[]; title: string }) {
  const pills = props.pills.filter(Boolean);
  return (
    <div className="rounded-md border p-3">
      <div className="mb-2 text-xs font-medium text-muted-foreground">{props.title}</div>
      {pills.length === 0 ? (
        <div className="text-sm text-muted-foreground">{props.emptyLabel}</div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {pills.map((pill) => (
            <Badge key={pill} variant="outline">{pill}</Badge>
          ))}
        </div>
      )}
    </div>
  );
}

function formatOptionalNumber(value: unknown): string {
  return typeof value === "number" ? `${value}` : "-";
}

function GatewayContent(props: {
  selectedVersion: PromptVersion;
  t: PromptTranslationFn;
  onCopy: (text: string) => void;
}) {
  const streamingSnippet = buildGatewaySnippet(props.selectedVersion, true);
  const nonStreamingSnippet = buildGatewaySnippet(props.selectedVersion, false);

  return (
    <div className="flex flex-col gap-4">
      <div>
        <h3 className="text-base font-semibold">{props.t("detail.tabs.gateway")}</h3>
        <p className="text-sm text-muted-foreground">{props.t("detail.gatewayDescription")}</p>
      </div>
      <Tabs defaultValue="non-stream" className="gap-4">
        <TabsList>
          <TabsTrigger value="non-stream">{props.t("detail.gateway.nonStream")}</TabsTrigger>
          <TabsTrigger value="stream">{props.t("detail.gateway.stream")}</TabsTrigger>
        </TabsList>
        <TabsContent value="non-stream">
          <CodePanel
            actions={
              <Button type="button" variant="outline" size="sm" onClick={() => props.onCopy(nonStreamingSnippet)}>
                <Copy data-icon="inline-start" />
                {props.t("actions.copy")}
              </Button>
            }
            code={nonStreamingSnippet}
            description={props.t("detail.gateway.nonStreamDescription")}
            language="bash"
            displayLanguage="cURL"
            title={props.t("detail.gateway.nonStream")}
          />
        </TabsContent>
        <TabsContent value="stream">
          <CodePanel
            actions={
              <Button type="button" variant="outline" size="sm" onClick={() => props.onCopy(streamingSnippet)}>
                <Copy data-icon="inline-start" />
                {props.t("actions.copy")}
              </Button>
            }
            code={streamingSnippet}
            description={props.t("detail.gateway.streamDescription")}
            language="bash"
            displayLanguage="cURL"
            title={props.t("detail.gateway.stream")}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function PromptVersionHistorySheet(props: {
  currentVersionID: string;
  isOpen: boolean;
  onClose: () => void;
  onCompareOpen: () => void;
  onCreateVersion: () => void;
  onSelectVersion: (version: PromptVersion) => void;
  search: string;
  setSearch: (value: string) => void;
  locale: string;
  t: PromptTranslationFn;
  userLabelByID: Record<string, string>;
  versions: PromptVersion[];
}) {
  return (
    <Sheet open={props.isOpen} onOpenChange={(open) => { if (!open) props.onClose(); }}>
      <SheetContent
        side="left"
        closeLabel={props.t("actions.close")}
        className="flex h-full w-full flex-col overflow-hidden p-0 sm:max-w-lg"
      >
        <SheetHeader className="shrink-0 border-b border-border/60 px-4 pb-4 pt-5 sm:px-6">
          <div className="flex items-center gap-2 pe-8">
            <SheetTitle>{props.t("actions.versionHistory")}</SheetTitle>
            <Badge variant="secondary">{new Intl.NumberFormat(props.locale).format(props.versions.length)}</Badge>
          </div>
          <SheetDescription className="max-w-md leading-6">{props.t("detail.historyDescription")}</SheetDescription>
        </SheetHeader>

        <div className="flex min-h-0 flex-1 flex-col gap-4 p-4 sm:p-6">
          <div className="flex shrink-0 flex-col gap-3">
            <Input
              value={props.search}
              onChange={(event) => props.setSearch(event.target.value)}
              placeholder={props.t("detail.historySearchPlaceholder")}
              aria-label={props.t("detail.historySearchPlaceholder")}
            />
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <Button type="button" className="w-full" onClick={props.onCreateVersion}>
                <FileCode2 data-icon="inline-start" />
                {props.t("actions.createVersion")}
              </Button>
              <Button type="button" variant="outline" className="w-full" onClick={props.onCompareOpen}>
                <GitCompare data-icon="inline-start" />
                {props.t("actions.compareVersions")}
              </Button>
            </div>
          </div>

          <ScrollArea className="min-h-0 flex-1">
            <div className="flex flex-col gap-2 pe-3 pb-1">
              {props.versions.map((version) => (
                <button
                  key={version.id}
                  type="button"
                  className={cn(
                    "w-full rounded-lg border border-border bg-card p-3 text-start shadow-xs transition-colors hover:border-primary/40 hover:bg-muted/30 sm:p-4",
                    version.id === props.currentVersionID && "border-primary-subtle-border border-s-2 border-s-primary bg-primary-soft",
                  )}
                  onClick={() => props.onSelectVersion(version)}
                >
                  <div className="flex items-start justify-between gap-3">
                    <Badge variant={version.id === props.currentVersionID ? "default" : "secondary"}>
                      {props.t("versions.versionValue", { version: version.version })}
                    </Badge>
                    <DateTimeValue
                      value={version.created_at}
                      locale={props.locale}
                      fallback="-"
                      className="shrink-0 text-xs text-muted-foreground"
                    />
                  </div>
                  <div className="mt-3 truncate text-sm font-medium text-foreground">
                    {props.userLabelByID[version.created_by ?? ""] || version.created_by || "-"}
                  </div>
                  <div
                    className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground"
                    dir={version.commit_message ? "auto" : "ltr"}
                  >
                    {version.commit_message || version.fqn}
                  </div>
                </button>
              ))}
              {props.versions.length === 0 ? (
                <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">{props.t("versions.empty")}</div>
              ) : null}
            </div>
          </ScrollArea>
        </div>
      </SheetContent>
    </Sheet>
  );
}

function PromptCompareDialog(props: {
  baseVersion: PromptVersion | null;
  baseVersionID: string;
  isOpen: boolean;
  onClose: () => void;
  onSelectBaseVersion: (value: string) => void;
  onSelectTargetVersion: (value: string) => void;
  rows: PromptCompareRow[];
  t: PromptTranslationFn;
  targetVersion: PromptVersion | null;
  targetVersionID: string;
  versions: PromptVersion[];
}) {
  return (
    <FormDialog open={props.isOpen} onOpenChange={(open) => { if (!open) props.onClose(); }} title={props.t("detail.compareTitle")} description={props.t("detail.compareDescription")} size="xl" contentClassName="overflow-hidden" bodyClassName="overflow-auto">
          <div className="grid gap-3 md:grid-cols-2">
            <AppSelectField
              label={props.t("detail.compareBaseVersion")}
              value={props.baseVersionID}
              onValueChange={props.onSelectBaseVersion}
              options={props.versions.map((version) => ({
                value: version.id,
                label: props.t("versions.versionValue", { version: version.version }),
              }))}
            />
            <AppSelectField
              label={props.t("detail.compareTargetVersion")}
              value={props.targetVersionID}
              onValueChange={props.onSelectTargetVersion}
              options={props.versions.map((version) => ({
                value: version.id,
                label: props.t("versions.versionValue", { version: version.version }),
              }))}
            />
          </div>

          <div className="overflow-x-auto rounded-xl border">
            <div className="min-w-2xl">
              <div className="grid grid-cols-2 border-b bg-muted/30">
                <div className="border-e px-4 py-3 text-center font-medium">
                  {props.baseVersion ? props.t("versions.versionValue", { version: props.baseVersion.version }) : "-"}
                </div>
                <div className="px-4 py-3 text-center font-medium">
                  {props.targetVersion ? props.t("versions.versionValue", { version: props.targetVersion.version }) : "-"}
                </div>
              </div>
              <div className="grid grid-cols-2" dir="ltr">
                <div className="border-e">
                  {props.rows.map((row, index) => (
                    <DiffLine key={`left-${index}`} side="left" row={row} />
                  ))}
                </div>
                <div>
                  {props.rows.map((row, index) => (
                    <DiffLine key={`right-${index}`} side="right" row={row} />
                  ))}
                </div>
              </div>
            </div>
          </div>
    </FormDialog>
  );
}

function DiffLine(props: {
  row: PromptCompareRow;
  side: "left" | "right";
}) {
  const text = props.side === "left" ? props.row.leftText : props.row.rightText;
  const lineNumber = props.side === "left" ? props.row.leftLineNumber : props.row.rightLineNumber;
  const marker = props.row.change === "same"
    ? " "
    : props.row.change === "changed"
      ? "~"
      : props.row.change === "removed"
        ? props.side === "left" ? "-" : " "
        : props.side === "right" ? "+" : " ";

  const toneClassName = props.row.change === "same"
    ? ""
    : props.row.change === "changed"
      ? "bg-[var(--warning-soft)]"
      : props.row.change === "removed"
        ? props.side === "left" ? "bg-[var(--destructive-soft)]" : ""
        : props.side === "right" ? "bg-[var(--success-soft)]" : "";

  return (
    <div className={cn("grid min-h-8 grid-cols-[48px_24px_minmax(0,1fr)] font-mono text-xs", toneClassName)}>
      <div className="border-b px-2 py-1 text-end text-muted-foreground">{lineNumber ?? ""}</div>
      <div className="border-b px-1 py-1 text-center text-muted-foreground">{marker}</div>
      <div className="border-b px-3 py-1 whitespace-pre-wrap break-words">{text || " "}</div>
    </div>
  );
}

function downloadText(fileName: string, content: string) {
  const blob = new Blob([content], { type: "application/json;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  anchor.click();
  URL.revokeObjectURL(url);
}
