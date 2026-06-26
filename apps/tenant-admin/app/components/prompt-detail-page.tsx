"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import {
  Clipboard,
  Copy,
  Download,
  FileCode2,
  GitCompare,
  History,
  PencilLine,
  Play,
} from "lucide-react";

import { normalizeError, requestJSON, unwrapData } from "@/app/components/provider-account-wizard/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { cn } from "@/lib/utils";

import { CodeBlock } from "./code-block";
import {
  buildGatewaySnippet,
  buildPromptDiffRows,
  buildSdkSnippet,
  describePromptMessage,
  serializePromptSpec,
  type PromptCompareRow,
  type PromptDetailTab,
} from "./prompt-detail-utils";
import {
  formatDate,
  type Notice,
  type Prompt,
  type PromptTranslationFn,
  type PromptVersion,
  type PromptVersionListResponse,
  type TenantAdminUserLookupResponse,
} from "./prompt-management-shared";

type PromptDetailPageProps = {
  promptID: string;
};

export default function PromptDetailPage(props: PromptDetailPageProps) {
  const t = useTranslations("prompts");
  const locale = useLocale();
  const router = useRouter();
  const pageDir = locale === "fa" ? "rtl" : "ltr";

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
    const payload = await requestJSON<{ data?: Prompt } | Prompt>(`/api/prompts/${encodeURIComponent(props.promptID)}`);
    const nextPrompt = unwrapData<Prompt>(payload);
    if (!nextPrompt) {
      throw new Error(t("messages.loadPromptsError"));
    }
    setPrompt(nextPrompt);
    setSelectedVersionID((current) => current || nextPrompt.latest_version_detail?.id || "");
    return nextPrompt;
  }, [props.promptID, t]);

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
      const payload = await requestJSON<TenantAdminUserLookupResponse>("/api/gateway/lookups/users");
      const users = payload.data ?? [];
      setUserLabelByID(
        Object.fromEntries(
          uniqueIDs.map((id) => {
            const match = users.find((user) => user.id === id);
            return [id, match?.display_name?.trim() || match?.email?.trim() || match?.mobile?.trim() || id];
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
      const payload = await requestJSON<PromptVersionListResponse>(`/api/prompts/${encodeURIComponent(props.promptID)}/versions?limit=100`);
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
      setNotice({ kind: "error", message: normalizeError(error, t("messages.loadPromptsError")) });
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
      <section className="flex flex-col gap-3 rounded-xl border bg-card p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <div className="text-xs text-muted-foreground">{t("detail.breadcrumb", { name: prompt?.name || t("header.title") })}</div>
            <h2 className="text-2xl font-semibold">{prompt?.name || t("header.title")}</h2>
            <p className="text-sm text-muted-foreground">{t("detail.description")}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={() => void copySelectedFqn()} disabled={!selectedVersion?.fqn}>
              <Clipboard />
              {t("actions.copyFqn")}
            </Button>
            <Button type="button" variant="outline" onClick={() => router.push(`/prompts/${props.promptID}/edit?intent=try`)}>
              <Play />
              {t("actions.tryPrompt")}
            </Button>
            <Button type="button" variant="outline" onClick={() => setHistoryOpen(true)}>
              <History />
              {t("actions.versionHistory")}
            </Button>
            <Button type="button" variant="outline" onClick={() => router.push(`/prompts/${props.promptID}/edit?intent=create-version`)}>
              <FileCode2 />
              {t("actions.createVersion")}
            </Button>
            <Button type="button" variant="outline" onClick={() => router.push(`/prompts/${props.promptID}/edit?intent=edit-latest`)}>
              <PencilLine />
              {t("actions.editLatestVersion")}
            </Button>
          </div>
        </div>

        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? t("notice.errorTitle") : t("notice.successTitle")}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <Card className="shadow-none">
          <CardContent className="grid gap-4 p-4 md:grid-cols-2 xl:grid-cols-5">
            <SummaryItem label={t("builder.fields.name")} value={prompt?.name || "-"} />
            <SummaryItem
              label={t("list.columns.latestVersion")}
              value={typeof prompt?.latest_version === "number" ? t("list.versionValue", { version: prompt.latest_version }) : "-"}
            />
            <SummaryItem label={t("list.columns.updatedAt")} value={formatDate(prompt?.updated_at, locale)} />
            <SummaryItem
              label={t("list.columns.createdBy")}
              value={userLabelByID[prompt?.created_by ?? ""] || prompt?.created_by || "-"}
            />
            <SummaryItem label={t("detail.summary.promptId")} value={prompt?.id || "-"} valueClassName="font-mono text-xs" />
            <SummaryItem label={t("list.columns.model")} value={selectedDescriptor?.model || "-"} />
            <SummaryItem label={t("versions.selectedFqn")} value={selectedVersion?.fqn || "-"} valueClassName="font-mono text-xs" />
            <SummaryItem label={t("detail.summary.createdAt")} value={formatDate(selectedVersion?.created_at, locale)} />
            <SummaryItem
              label={t("detail.summary.versionAuthor")}
              value={userLabelByID[selectedVersion?.created_by ?? ""] || selectedVersion?.created_by || "-"}
            />
            <SummaryItem
              label={t("builder.fields.description")}
              value={selectedDescriptor?.description || "-"}
              className="xl:col-span-2"
            />
          </CardContent>
        </Card>
      </section>

      {isLoading ? (
        <div className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">{t("detail.loading")}</div>
      ) : selectedVersion ? (
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as PromptDetailTab)} className="gap-4">
          <TabsList variant="line" className="h-auto w-full justify-start gap-2 rounded-none border-b p-0">
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

          <TabsContent value="prompt" forceMount>
            <PromptContent descriptor={selectedDescriptor} t={t} />
          </TabsContent>

          <TabsContent value="spec" forceMount>
            <CodePanel
              actions={
                <>
                  <Button type="button" variant="outline" size="sm" onClick={() => void copyText(selectedSpec, t("messages.specCopied"))}>
                    <Copy />
                    {t("actions.copy")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => downloadText(`${selectedVersion.name || "prompt"}-v${selectedVersion.version}.json`, selectedSpec)}
                  >
                    <Download />
                    {t("actions.download")}
                  </Button>
                </>
              }
              code={selectedSpec}
              description={t("detail.specDescription")}
              language="json"
              title={t("detail.tabs.spec")}
            />
          </TabsContent>

          <TabsContent value="sdk" forceMount>
            <CodePanel
              actions={
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void copyText(buildSdkSnippet(selectedVersion), t("messages.snippetCopied"))}
                >
                  <Copy />
                  {t("actions.copy")}
                </Button>
              }
              code={buildSdkSnippet(selectedVersion)}
              description={t("detail.sdkDescription")}
              language="python"
              title={t("detail.tabs.sdk")}
            />
          </TabsContent>

          <TabsContent value="gateway" forceMount>
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
        onCreateVersion={() => router.push(`/prompts/${props.promptID}/edit?intent=create-version`)}
        onSelectVersion={(version) => {
          setSelectedVersionID(version.id);
          setHistoryOpen(false);
        }}
        search={versionSearch}
        setSearch={setVersionSearch}
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
      setNotice({ kind: "error", message: t("messages.fqnUnavailable") });
      return;
    }
    await copyText(selectedVersion.fqn, t("messages.fqnCopied"));
  }

  async function copyText(text: string, successMessage: string) {
    await navigator.clipboard.writeText(text);
    setNotice({ kind: "success", message: successMessage });
  }
}

function SummaryItem(props: {
  label: string;
  value: string;
  className?: string;
  valueClassName?: string;
}) {
  return (
    <div className={cn("space-y-1", props.className)}>
      <div className="text-xs text-muted-foreground">{props.label}</div>
      <div className={cn("text-sm font-medium", props.valueClassName)} dir={props.valueClassName?.includes("font-mono") ? "ltr" : "auto"}>
        {props.value}
      </div>
    </div>
  );
}

function PromptContent(props: {
  descriptor: ReturnType<typeof describePromptMessage> | null;
  t: PromptTranslationFn;
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
    <div className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
      <div className="space-y-5">
        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{props.t("config.title")}</CardTitle>
            <CardDescription>{props.descriptor.versionLabel}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Section title={props.t("config.model.title")}>
              <div className="font-medium">{props.descriptor.model || "-"}</div>
            </Section>
            <Section title={props.t("config.modelSettings.title")}>
              <div className="flex flex-wrap gap-2">
                <Badge variant="outline">{props.descriptor.modelSettings.stream ? props.t("common.enabled") : props.t("common.disabled")}</Badge>
                {typeof props.descriptor.modelSettings.temperature === "number" ? <Badge variant="outline">temp {props.descriptor.modelSettings.temperature}</Badge> : null}
                {typeof props.descriptor.modelSettings.top_p === "number" ? <Badge variant="outline">top_p {props.descriptor.modelSettings.top_p}</Badge> : null}
                {typeof props.descriptor.modelSettings.max_tokens === "number" ? <Badge variant="outline">max_tokens {props.descriptor.modelSettings.max_tokens}</Badge> : null}
              </div>
            </Section>
            <Section title={props.t("config.variables.title", { count: props.descriptor.variableNames.length })}>
              {props.descriptor.variableNames.length === 0 ? (
                <div className="text-sm text-muted-foreground">{props.t("config.variables.empty")}</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {props.descriptor.variableNames.map((name) => (
                    <Badge key={name} variant="outline">{name}</Badge>
                  ))}
                </div>
              )}
            </Section>
            <Section title={props.t("config.cache.title")}>
              {cachePills.length === 0 ? (
                <div className="text-sm text-muted-foreground">{props.t("config.cache.empty")}</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {cachePills.map((pill) => (
                    <Badge key={pill} variant="outline">{pill}</Badge>
                  ))}
                </div>
              )}
            </Section>
            <Section title={props.t("config.inputGuardrails.title", { count: props.descriptor.guardrails?.input?.length ?? 0 })}>
              <div className="text-sm text-muted-foreground">
                {(props.descriptor.guardrails?.input?.length ?? 0) > 0
                  ? props.descriptor.guardrails?.input?.join(", ")
                  : props.t("config.inputGuardrails.empty")}
              </div>
            </Section>
            <Section title={props.t("config.outputGuardrails.title", { count: props.descriptor.guardrails?.output?.length ?? 0 })}>
              <div className="text-sm text-muted-foreground">
                {(props.descriptor.guardrails?.output?.length ?? 0) > 0
                  ? props.descriptor.guardrails?.output?.join(", ")
                  : props.t("config.outputGuardrails.empty")}
              </div>
            </Section>
            <Section title={props.t("config.structuredOutput.title")}>
              {structuredPills.length === 0 ? (
                <div className="text-sm text-muted-foreground">{props.t("config.structuredOutput.empty")}</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {structuredPills.map((pill) => (
                    <Badge key={pill} variant="outline">{pill}</Badge>
                  ))}
                </div>
              )}
            </Section>
            <Section title={props.t("config.metadata.title")}>
              {Object.keys(props.descriptor.metadata).length === 0 ? (
                <div className="text-sm text-muted-foreground">{props.t("config.metadata.empty")}</div>
              ) : (
                <div className="space-y-2">
                  {Object.entries(props.descriptor.metadata).map(([key, value]) => (
                    <div key={key} className="rounded-md border bg-muted/20 px-3 py-2 text-sm">
                      <div className="font-medium">{key}</div>
                      <div dir="ltr" className="mt-1 font-mono text-xs text-muted-foreground">{JSON.stringify(value)}</div>
                    </div>
                  ))}
                </div>
              )}
            </Section>
          </CardContent>
        </Card>
      </div>

      <div className="space-y-5">
        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{props.t("builder.systemPrompt.title")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg border bg-muted/20 p-4 text-sm" dir="auto">
              {props.descriptor.systemPrompt || "-"}
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{props.t("builder.conversation.title")}</CardTitle>
            <CardDescription>{props.t("builder.conversation.description")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {props.descriptor.conversationMessages.length === 0 ? (
              <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">{props.t("builder.conversation.empty")}</div>
            ) : (
              props.descriptor.conversationMessages.map((message, index) => (
                <div key={`${message.role}-${index}`} className="rounded-lg border p-4">
                  <div className="mb-2">
                    <Badge variant={message.role === "assistant" ? "secondary" : "outline"}>
                      {message.role === "assistant" ? props.t("roles.assistant") : props.t("roles.user")}
                    </Badge>
                  </div>
                  <div className="text-sm" dir="auto">
                    {typeof message.content === "string" ? message.content : JSON.stringify(message.content)}
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function Section(props: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <div className="text-sm font-medium">{props.title}</div>
      {props.children}
    </section>
  );
}

function CodePanel(props: {
  actions?: React.ReactNode;
  code: string;
  description: string;
  language: "bash" | "json" | "python";
  title: string;
}) {
  return (
    <Card className="shadow-none">
      <CardHeader className="border-b">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <CardTitle className="text-base">{props.title}</CardTitle>
            <CardDescription>{props.description}</CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">{props.actions}</div>
        </div>
      </CardHeader>
      <CardContent className="p-4">
        <CodeBlock code={props.code} language={props.language} maxHeightClassName="max-h-[min(34rem,60dvh)]" />
      </CardContent>
    </Card>
  );
}

function GatewayContent(props: {
  selectedVersion: PromptVersion;
  t: PromptTranslationFn;
  onCopy: (text: string) => void;
}) {
  const streamingSnippet = buildGatewaySnippet(props.selectedVersion, true);
  const nonStreamingSnippet = buildGatewaySnippet(props.selectedVersion, false);

  return (
    <Card className="shadow-none">
      <CardHeader className="border-b">
        <CardTitle className="text-base">{props.t("detail.tabs.gateway")}</CardTitle>
        <CardDescription>{props.t("detail.gatewayDescription")}</CardDescription>
      </CardHeader>
      <CardContent className="p-4">
        <Tabs defaultValue="non-stream" className="gap-4">
          <TabsList>
            <TabsTrigger value="non-stream">{props.t("detail.gateway.nonStream")}</TabsTrigger>
            <TabsTrigger value="stream">{props.t("detail.gateway.stream")}</TabsTrigger>
          </TabsList>
          <TabsContent value="non-stream" forceMount>
            <CodePanel
              actions={
                <Button type="button" variant="outline" size="sm" onClick={() => props.onCopy(nonStreamingSnippet)}>
                  <Copy />
                  {props.t("actions.copy")}
                </Button>
              }
              code={nonStreamingSnippet}
              description={props.t("detail.gateway.nonStreamDescription")}
              language="bash"
              title={props.t("detail.gateway.nonStream")}
            />
          </TabsContent>
          <TabsContent value="stream" forceMount>
            <CodePanel
              actions={
                <Button type="button" variant="outline" size="sm" onClick={() => props.onCopy(streamingSnippet)}>
                  <Copy />
                  {props.t("actions.copy")}
                </Button>
              }
              code={streamingSnippet}
              description={props.t("detail.gateway.streamDescription")}
              language="bash"
              title={props.t("detail.gateway.stream")}
            />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
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
  t: PromptTranslationFn;
  userLabelByID: Record<string, string>;
  versions: PromptVersion[];
}) {
  return (
    <Sheet open={props.isOpen} onOpenChange={(open) => { if (!open) props.onClose(); }}>
      <SheetContent side="right" className="w-[94vw] max-w-xl overflow-y-auto">
        <SheetHeader>
          <div className="flex items-center justify-between gap-3">
            <div>
              <SheetTitle>{props.t("actions.versionHistory")}</SheetTitle>
              <SheetDescription>{props.t("detail.historyDescription")}</SheetDescription>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={props.onCompareOpen}>
              <GitCompare />
              {props.t("actions.compareVersions")}
            </Button>
          </div>
        </SheetHeader>
        <div className="mt-6 space-y-4">
          <div className="flex gap-2">
            <Input
              value={props.search}
              onChange={(event) => props.setSearch(event.target.value)}
              placeholder={props.t("detail.historySearchPlaceholder")}
            />
            <Button type="button" onClick={props.onCreateVersion}>
              {props.t("actions.createVersion")}
            </Button>
          </div>
          <div className="space-y-2">
            {props.versions.map((version) => (
              <button
                key={version.id}
                type="button"
                className={cn(
                  "w-full rounded-lg border p-4 text-start transition hover:bg-accent",
                  version.id === props.currentVersionID && "border-primary/40 bg-primary/5",
                )}
                onClick={() => props.onSelectVersion(version)}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="font-medium">{props.t("versions.versionValue", { version: version.version })}</div>
                  <div className="text-xs text-muted-foreground">{formatDate(version.created_at, "en")}</div>
                </div>
                <div className="mt-1 text-sm text-muted-foreground">
                  {props.userLabelByID[version.created_by ?? ""] || version.created_by || "-"}
                </div>
                <div className="mt-2 truncate text-xs text-muted-foreground" dir="ltr">
                  {version.commit_message || version.fqn}
                </div>
              </button>
            ))}
            {props.versions.length === 0 ? (
              <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">{props.t("versions.empty")}</div>
            ) : null}
          </div>
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
    <Dialog open={props.isOpen} onOpenChange={(open) => { if (!open) props.onClose(); }}>
      <DialogContent className="max-h-[90vh] max-w-[calc(100vw-2rem)] overflow-hidden p-0 sm:max-w-6xl">
        <DialogHeader className="border-b px-6 py-4">
          <DialogTitle>{props.t("detail.compareTitle")}</DialogTitle>
          <DialogDescription>{props.t("detail.compareDescription")}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 overflow-auto p-6">
          <div className="grid gap-3 md:grid-cols-2">
            <label className="space-y-2 text-sm">
              <span className="font-medium">{props.t("detail.compareBaseVersion")}</span>
              <select
                value={props.baseVersionID}
                onChange={(event) => props.onSelectBaseVersion(event.target.value)}
                className="h-10 w-full rounded-md border bg-background px-3"
              >
                {props.versions.map((version) => (
                  <option key={version.id} value={version.id}>
                    {props.t("versions.versionValue", { version: version.version })}
                  </option>
                ))}
              </select>
            </label>
            <label className="space-y-2 text-sm">
              <span className="font-medium">{props.t("detail.compareTargetVersion")}</span>
              <select
                value={props.targetVersionID}
                onChange={(event) => props.onSelectTargetVersion(event.target.value)}
                className="h-10 w-full rounded-md border bg-background px-3"
              >
                {props.versions.map((version) => (
                  <option key={version.id} value={version.id}>
                    {props.t("versions.versionValue", { version: version.version })}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="overflow-hidden rounded-xl border">
            <div className="grid grid-cols-2 border-b bg-muted/30">
              <div className="border-r px-4 py-3 text-center font-medium">
                {props.baseVersion ? props.t("versions.versionValue", { version: props.baseVersion.version }) : "-"}
              </div>
              <div className="px-4 py-3 text-center font-medium">
                {props.targetVersion ? props.t("versions.versionValue", { version: props.targetVersion.version }) : "-"}
              </div>
            </div>
            <div className="grid grid-cols-2" dir="ltr">
              <div className="border-r">
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
      </DialogContent>
    </Dialog>
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
    ? "bg-slate-950/0"
    : props.row.change === "changed"
      ? "bg-amber-100/60 dark:bg-amber-950/30"
      : props.row.change === "removed"
        ? props.side === "left" ? "bg-rose-100/70 dark:bg-rose-950/30" : "bg-slate-950/0"
        : props.side === "right" ? "bg-emerald-100/70 dark:bg-emerald-950/30" : "bg-slate-950/0";

  return (
    <div className={cn("grid min-h-8 grid-cols-[48px_24px_minmax(0,1fr)] font-mono text-xs", toneClassName)}>
      <div className="border-b px-2 py-1 text-right text-muted-foreground">{lineNumber ?? ""}</div>
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
