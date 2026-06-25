"use client";

import { useEffect, useMemo, useState } from "react";
import { useLocale } from "next-intl";
import {
  AlertCircle,
  Braces,
  Check,
  Clipboard,
  DatabaseZap,
  FileCode2,
  FileText,
  History,
  Loader2,
  Play,
  Plus,
  Save,
  Search,
  Settings2,
  ShieldCheck,
  Sparkles,
} from "lucide-react";

import { requestJSON, unwrapData, normalizeError } from "@/app/components/provider-account-wizard/utils";
import type { ProviderCatalogModel, ProviderCatalogProvider } from "@/app/components/provider-account-wizard/types";
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
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

import {
  buildPromptUsageSnippet,
  extractPromptVariableNames,
  normalizePromptSlug,
  renderPromptText,
  type PromptEditorMessage,
} from "./prompt-management-utils";

type GatewayModel = {
  id: string;
  object: string;
  owned_by: string;
  provider?: string;
  aliases?: string[];
};

type GatewayModelListResponse = {
  data?: GatewayModel[];
};

type GuardrailItem = {
  id: string;
  slug: string;
  name: string;
  description?: string;
  selector?: string;
  type?: string;
};

type GuardrailListResponse = {
  data?: GuardrailItem[];
};

type PromptVariableMap = Record<string, string>;

type PromptModelSettings = {
  stream?: boolean;
  temperature?: number;
  top_p?: number;
  max_tokens?: number;
};

type PromptCacheConfig = {
  enabled: boolean;
  type?: "simple" | "smart";
  ttl?: number;
  namespace?: string;
  similarity_threshold?: number;
};

type PromptGuardrailConfig = {
  input?: string[];
  output?: string[];
  scope?: "all" | "last";
};

type StructuredOutputConfig = {
  enabled: boolean;
  format?: "json_schema" | "json_object";
  strict?: boolean;
  schema?: Record<string, unknown>;
};

type PromptManifest = {
  type: string;
  name: string;
  description?: string;
  messages: Array<{ role: string; content?: unknown }>;
  variables?: PromptVariableMap;
  model?: string;
  parameters?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  version_alias?: string;
  model_settings?: PromptModelSettings;
  cache?: PromptCacheConfig;
  guardrails?: PromptGuardrailConfig;
  structured_output?: StructuredOutputConfig;
  logging?: {
    enabled: boolean;
    store_logs: boolean;
  };
};

type PromptVersion = {
  id: string;
  prompt_id: string;
  version: number;
  fqn: string;
  version_alias?: string;
  name: string;
  description?: string;
  messages: Array<{ role: string; content?: unknown }>;
  variables?: PromptVariableMap;
  model?: string;
  parameters?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  manifest?: PromptManifest;
  commit_message?: string;
  created_by?: string;
  created_at?: string;
};

type Prompt = {
  id: string;
  name: string;
  slug: string;
  description?: string;
  latest_version: number;
  latest_version_detail?: PromptVersion;
  created_by?: string;
  created_at?: string;
  updated_at?: string;
};

type PromptListResponse = {
  data?: Prompt[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number | null;
};

type PromptVersionListResponse = {
  data?: PromptVersion[];
};

type PromptRunResponse = {
  choices?: Array<{
    message?: {
      role?: string;
      content?: unknown;
    };
  }>;
  usage?: {
    prompt_tokens?: number;
    completion_tokens?: number;
    total_tokens?: number;
  };
  model?: string;
  id?: string;
};

type MetadataRow = {
  id: string;
  key: string;
  value: string;
};

type Notice = { kind: "success" | "error"; message: string } | null;

type BuilderState = {
  name: string;
  description: string;
  selectedModel: string;
  selectedProvider: string;
  systemPrompt: string;
  messages: PromptEditorMessage[];
  composerRole: "user" | "assistant";
  composerText: string;
  commitMessage: string;
  variableValues: PromptVariableMap;
  stream: boolean;
  temperature: string;
  topP: string;
  maxTokens: string;
  cacheEnabled: boolean;
  cacheType: "simple" | "smart";
  cacheTTL: string;
  cacheNamespace: string;
  cacheSimilarityThreshold: string;
  inputGuardrails: string[];
  outputGuardrails: string[];
  guardrailScope: "all" | "last";
  structuredEnabled: boolean;
  structuredFormat: "json_schema" | "json_object";
  structuredStrict: boolean;
  structuredSchema: string;
  metadataRows: MetadataRow[];
};

const DEFAULT_SYSTEM_PROMPT = "You are a helpful assistant.";
const DEFAULT_SCHEMA = JSON.stringify(
  {
    type: "object",
    properties: {
      message: {
        type: "string",
      },
    },
    required: ["message"],
  },
  null,
  2,
);

const EMPTY_BUILDER: BuilderState = {
  name: "",
  description: "",
  selectedModel: "",
  selectedProvider: "",
  systemPrompt: DEFAULT_SYSTEM_PROMPT,
  messages: [],
  composerRole: "user",
  composerText: "",
  commitMessage: "",
  variableValues: {},
  stream: true,
  temperature: "0.2",
  topP: "",
  maxTokens: "",
  cacheEnabled: false,
  cacheType: "simple",
  cacheTTL: "600",
  cacheNamespace: "",
  cacheSimilarityThreshold: "0.9",
  inputGuardrails: [],
  outputGuardrails: [],
  guardrailScope: "last",
  structuredEnabled: false,
  structuredFormat: "json_schema",
  structuredStrict: true,
  structuredSchema: DEFAULT_SCHEMA,
  metadataRows: [],
};

function createMetadataRow(key = "", value = ""): MetadataRow {
  return {
    id: `meta_${Math.random().toString(36).slice(2, 10)}`,
    key,
    value,
  };
}

export default function PromptManagement() {
  const locale = useLocale();
  const pageDir = locale === "fa" ? "rtl" : "ltr";

  const [notice, setNotice] = useState<Notice>(null);
  const [isPromptsLoading, setIsPromptsLoading] = useState(false);
  const [isVersionsLoading, setIsVersionsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [isModelsLoading, setIsModelsLoading] = useState(false);
  const [isGuardrailsLoading, setIsGuardrailsLoading] = useState(false);

  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [promptTotal, setPromptTotal] = useState(0);
  const [query, setQuery] = useState("");
  const [pageSize, setPageSize] = useState(25);
  const [pageOffset, setPageOffset] = useState(0);
  const [selectedPromptID, setSelectedPromptID] = useState("");
  const [versions, setVersions] = useState<PromptVersion[]>([]);
  const [selectedVersionID, setSelectedVersionID] = useState("");

  const [builder, setBuilder] = useState<BuilderState>(EMPTY_BUILDER);
  const [runResponse, setRunResponse] = useState<PromptRunResponse | null>(null);
  const [runPreview, setRunPreview] = useState("");

  const [models, setModels] = useState<GatewayModel[]>([]);
  const [providerCatalog, setProviderCatalog] = useState<ProviderCatalogProvider[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);
  const [guardrails, setGuardrails] = useState<GuardrailItem[]>([]);

  const [modelPickerOpen, setModelPickerOpen] = useState(false);
  const [modelSearch, setModelSearch] = useState("");
  const [selectedProviderFilter, setSelectedProviderFilter] = useState("");
  const [cacheDialogOpen, setCacheDialogOpen] = useState(false);
  const [guardrailDialogMode, setGuardrailDialogMode] = useState<"input" | "output" | null>(null);
  const [structuredDialogOpen, setStructuredDialogOpen] = useState(false);

  const selectedPrompt = useMemo(() => prompts.find((item) => item.id === selectedPromptID) ?? null, [prompts, selectedPromptID]);
  const selectedVersion = useMemo(
    () => versions.find((item) => item.id === selectedVersionID) ?? selectedPrompt?.latest_version_detail ?? null,
    [selectedPrompt, selectedVersionID, versions],
  );

  const tableRows = useMemo(() => prompts, [prompts]);

  const detectedVariables = useMemo(
    () =>
      extractPromptVariableNames([
        builder.systemPrompt,
        ...builder.messages.map((message) => message.content),
        builder.composerText,
      ]),
    [builder.systemPrompt, builder.messages, builder.composerText],
  );

  const actualProviderGroups = useMemo(() => {
    const grouped = new Map<string, GatewayModel[]>();
    for (const model of models) {
      const key = (model.provider || "other").trim().toLowerCase();
      const list = grouped.get(key) ?? [];
      list.push(model);
      grouped.set(key, list);
    }
    return grouped;
  }, [models]);

  const providerOptions = useMemo(() => {
    const keys = [...actualProviderGroups.keys()].sort((left, right) => left.localeCompare(right));
    return keys.map((key) => {
      const providerInfo = providerCatalog.find((item) => item.provider.trim().toLowerCase() === key);
      return {
        id: key,
        label: providerInfo?.display_name || key,
        description: providerInfo?.description || "",
      };
    });
  }, [actualProviderGroups, providerCatalog]);

  const visibleProviderID = selectedProviderFilter || builder.selectedProvider || providerOptions[0]?.id || "";

  const visibleModels = useMemo(() => {
    const source = actualProviderGroups.get(visibleProviderID) ?? models;
    const q = modelSearch.trim().toLowerCase();
    if (!q) return source;
    return source.filter((item) => {
      const aliases = item.aliases?.join(" ") ?? "";
      return `${item.id} ${aliases}`.toLowerCase().includes(q);
    });
  }, [actualProviderGroups, builder.selectedProvider, modelSearch, models, visibleProviderID]);

  const visibleProviderCatalogModels = useMemo(() => {
    const source = providerModels.filter((item) => item.provider.trim().toLowerCase() === visibleProviderID);
    const byModel = new Map(source.map((item) => [item.model, item]));
    return visibleModels.map((item) => byModel.get(item.id) ?? byModel.get(item.aliases?.[0] ?? "") ?? null);
  }, [providerModels, visibleModels, visibleProviderID]);

  const selectedModelMetadata = useMemo(() => {
    const entry = models.find((item) => item.id === builder.selectedModel);
    const catalogEntry =
      providerModels.find((item) => item.model === builder.selectedModel)
      ?? providerModels.find((item) => (entry?.aliases ?? []).includes(item.model));
    return {
      model: entry,
      catalog: catalogEntry ?? null,
    };
  }, [builder.selectedModel, models, providerModels]);

  const inputGuardrailItems = useMemo(
    () => guardrails.filter((item) => builder.inputGuardrails.includes(item.selector || `guardrails/${item.slug}`)),
    [builder.inputGuardrails, guardrails],
  );
  const outputGuardrailItems = useMemo(
    () => guardrails.filter((item) => builder.outputGuardrails.includes(item.selector || `guardrails/${item.slug}`)),
    [builder.outputGuardrails, guardrails],
  );

  const canSave = builder.name.trim() !== "" && builder.selectedModel.trim() !== "" && (builder.systemPrompt.trim() !== "" || builder.messages.length > 0);

  useEffect(() => {
    setBuilder((current) => {
      const nextVariables = { ...current.variableValues };
      for (const variable of detectedVariables) {
        if (!(variable in nextVariables)) {
          nextVariables[variable] = "";
        }
      }
      for (const existingKey of Object.keys(nextVariables)) {
        if (!detectedVariables.includes(existingKey)) {
          delete nextVariables[existingKey];
        }
      }
      if (JSON.stringify(nextVariables) === JSON.stringify(current.variableValues)) {
        return current;
      }
      return { ...current, variableValues: nextVariables };
    });
  }, [detectedVariables]);

  useEffect(() => {
    void Promise.all([loadModels(), loadGuardrails()]);
    void loadPrompts(query, pageOffset, pageSize, "");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function loadPrompts(nextQuery: string, offset: number, limit: number, preferredPromptID: string) {
    setIsPromptsLoading(true);
    try {
      const payload = await requestJSON<PromptListResponse>(
        `/api/prompts?query=${encodeURIComponent(nextQuery)}&offset=${offset}&limit=${limit}`,
      );
      const rows = payload.data ?? [];
      setPrompts(rows);
      setPromptTotal(payload.total ?? rows.length);
      const preferred = rows.find((item) => item.id === preferredPromptID) ?? rows[0] ?? null;
      if (!preferred && !selectedPromptID) {
        resetBuilder();
      }
      if (preferred && (preferredPromptID || selectedPromptID)) {
        await selectPrompt(preferred.id, rows);
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Prompts could not be loaded.") });
    } finally {
      setIsPromptsLoading(false);
    }
  }

  async function loadModels() {
    setIsModelsLoading(true);
    try {
      const [modelPayload, providerPayload] = await Promise.all([
        requestJSON<GatewayModelListResponse>("/api/models"),
        requestJSON<{ data?: ProviderCatalogProvider[] }>("/api/provider-catalog/providers"),
      ]);
      const nextModels = modelPayload.data ?? [];
      const nextProviders = unwrapData<ProviderCatalogProvider[]>(providerPayload) ?? [];
      setModels(nextModels);
      setProviderCatalog(nextProviders);

      const providerIDs = Array.from(new Set(nextModels.map((item) => (item.provider || "other").trim().toLowerCase()).filter(Boolean)));
      const catalogResponses = await Promise.all(
        providerIDs
          .filter((provider) => provider !== "virtual" && provider !== "other")
          .map(async (provider) => {
            try {
              const response = await requestJSON<{ data?: ProviderCatalogModel[] }>(
                `/api/provider-catalog/providers/${encodeURIComponent(provider)}/models`,
              );
              return unwrapData<ProviderCatalogModel[]>(response) ?? [];
            } catch {
              return [];
            }
          }),
      );
      setProviderModels(catalogResponses.flat());
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Models could not be loaded.") });
    } finally {
      setIsModelsLoading(false);
    }
  }

  async function loadGuardrails() {
    setIsGuardrailsLoading(true);
    try {
      const payload = await requestJSON<GuardrailListResponse>("/api/guardrails?limit=200");
      setGuardrails(payload.data ?? []);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Guardrails could not be loaded.") });
    } finally {
      setIsGuardrailsLoading(false);
    }
  }

  async function selectPrompt(promptID: string, currentRows = prompts) {
    const item = currentRows.find((entry) => entry.id === promptID);
    if (!item) return;
    setSelectedPromptID(promptID);
    setRunResponse(null);
    setRunPreview("");
    hydrateBuilder(item.latest_version_detail ?? null);
    setIsVersionsLoading(true);
    try {
      const payload = await requestJSON<PromptVersionListResponse>(`/api/prompts/${encodeURIComponent(promptID)}/versions?limit=100`);
      const nextVersions = payload.data ?? [];
      setVersions(nextVersions);
      const nextSelectedVersionID = nextVersions[0]?.id ?? item.latest_version_detail?.id ?? "";
      setSelectedVersionID(nextSelectedVersionID);
      if (nextSelectedVersionID) {
        const nextSelectedVersion = nextVersions.find((entry) => entry.id === nextSelectedVersionID) ?? null;
        if (nextSelectedVersion) {
          hydrateBuilder(nextSelectedVersion);
        }
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Prompt versions could not be loaded.") });
      setVersions([]);
    } finally {
      setIsVersionsLoading(false);
    }
  }

  function hydrateBuilder(version: PromptVersion | null) {
    if (!version) {
      resetBuilder();
      return;
    }
    const manifest = version.manifest ?? buildFallbackManifest(version);
    const systemMessage = manifest.messages.find((message) => message.role === "system");
    const threadMessages = manifest.messages
      .filter((message) => message.role === "user" || message.role === "assistant")
      .map((message) => ({
        role: message.role as "user" | "assistant",
        content: typeof message.content === "string" ? message.content : "",
      }));
    const metadataRows = Object.entries(manifest.metadata ?? {}).map(([key, value]) => createMetadataRow(key, stringifyScalar(value)));
    setBuilder({
      name: manifest.name || version.name,
      description: manifest.description || version.description || "",
      selectedModel: manifest.model || version.model || "",
      selectedProvider: findProviderForModel(manifest.model || version.model || "", models),
      systemPrompt: typeof systemMessage?.content === "string" ? systemMessage.content : DEFAULT_SYSTEM_PROMPT,
      messages: threadMessages,
      composerRole: "user",
      composerText: "",
      commitMessage: "",
      variableValues: { ...(manifest.variables ?? version.variables ?? {}) },
      stream: manifest.model_settings?.stream ?? booleanFromUnknown(manifest.parameters?.stream, true),
      temperature: numberToInput(manifest.model_settings?.temperature ?? numberFromUnknown(manifest.parameters?.temperature)),
      topP: numberToInput(manifest.model_settings?.top_p ?? numberFromUnknown(manifest.parameters?.top_p)),
      maxTokens: numberToInput(manifest.model_settings?.max_tokens ?? numberFromUnknown(manifest.parameters?.max_tokens)),
      cacheEnabled: manifest.cache?.enabled ?? false,
      cacheType: manifest.cache?.type ?? "simple",
      cacheTTL: numberToInput(manifest.cache?.ttl ?? 600),
      cacheNamespace: manifest.cache?.namespace ?? "",
      cacheSimilarityThreshold: numberToInput(manifest.cache?.similarity_threshold ?? 0.9),
      inputGuardrails: manifest.guardrails?.input ?? [],
      outputGuardrails: manifest.guardrails?.output ?? [],
      guardrailScope: manifest.guardrails?.scope ?? "last",
      structuredEnabled: manifest.structured_output?.enabled ?? false,
      structuredFormat: manifest.structured_output?.format ?? "json_schema",
      structuredStrict: manifest.structured_output?.strict ?? true,
      structuredSchema: manifest.structured_output?.schema ? JSON.stringify(manifest.structured_output.schema, null, 2) : DEFAULT_SCHEMA,
      metadataRows,
    });
  }

  function resetBuilder() {
    setSelectedPromptID("");
    setSelectedVersionID("");
    setVersions([]);
    setBuilder(EMPTY_BUILDER);
    setRunResponse(null);
    setRunPreview("");
  }

  function addComposerMessage() {
    if (!builder.composerText.trim()) return;
    setBuilder((current) => ({
      ...current,
      messages: [
        ...current.messages,
        {
          role: current.composerRole,
          content: current.composerText.trim(),
        },
      ],
      composerText: "",
    }));
  }

  function updateMessage(index: number, patch: Partial<PromptEditorMessage>) {
    setBuilder((current) => ({
      ...current,
      messages: current.messages.map((message, messageIndex) => (messageIndex === index ? { ...message, ...patch } : message)),
    }));
  }

  function removeMessage(index: number) {
    setBuilder((current) => ({
      ...current,
      messages: current.messages.filter((_, messageIndex) => messageIndex !== index),
    }));
  }

  function addMetadataRow() {
    setBuilder((current) => ({
      ...current,
      metadataRows: [...current.metadataRows, createMetadataRow()],
    }));
  }

  function updateMetadataRow(rowID: string, patch: Partial<MetadataRow>) {
    setBuilder((current) => ({
      ...current,
      metadataRows: current.metadataRows.map((row) => (row.id === rowID ? { ...row, ...patch } : row)),
    }));
  }

  function removeMetadataRow(rowID: string) {
    setBuilder((current) => ({
      ...current,
      metadataRows: current.metadataRows.filter((row) => row.id !== rowID),
    }));
  }

  async function savePrompt() {
    const manifestResult = buildManifest(builder);
    if (!manifestResult.ok) {
      setNotice({ kind: "error", message: manifestResult.message });
      return;
    }
    setIsSaving(true);
    setNotice(null);
    try {
      const path = selectedPromptID ? `/api/prompts/${encodeURIComponent(selectedPromptID)}/versions` : "/api/prompts";
      const saved = await requestJSON<PromptVersion>(path, {
        method: "POST",
        body: JSON.stringify({
          manifest: manifestResult.manifest,
          commit_message: builder.commitMessage.trim(),
        }),
      });
      setNotice({ kind: "success", message: selectedPromptID ? "Prompt version saved." : "Prompt created." });
      setPageOffset(0);
      await loadPrompts(query, 0, pageSize, saved.prompt_id);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Prompt could not be saved.") });
    } finally {
      setIsSaving(false);
    }
  }

  async function runPrompt() {
    const manifestResult = buildManifest(builder);
    if (!manifestResult.ok) {
      setNotice({ kind: "error", message: manifestResult.message });
      return;
    }
    const runtimeRequest = buildRunPayload(builder, manifestResult.manifest);
    if (!runtimeRequest.ok) {
      setNotice({ kind: "error", message: runtimeRequest.message });
      return;
    }

    setIsRunning(true);
    setRunResponse(null);
    setRunPreview("");
    setNotice(null);
    try {
      const payload = await requestJSON<PromptRunResponse>("/api/prompts/run", {
        method: "POST",
        body: JSON.stringify(runtimeRequest.payload),
      });
      setRunResponse(payload);
      setRunPreview(previewRunResponse(payload));
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, "Prompt execution failed.") });
    } finally {
      setIsRunning(false);
    }
  }

  async function copySelectedFQN() {
    if (!selectedVersion?.fqn) return;
    await navigator.clipboard.writeText(selectedVersion.fqn);
    setNotice({ kind: "success", message: "Prompt FQN copied." });
  }

  async function copySelectedSnippet() {
    if (!selectedVersion?.fqn) return;
    await navigator.clipboard.writeText(buildPromptUsageSnippet(selectedVersion.fqn, selectedVersion.model));
    setNotice({ kind: "success", message: "Usage snippet copied." });
  }

  const pageCountEnd = Math.min(pageOffset + pageSize, promptTotal);
  const selectedGuardrailMode = guardrailDialogMode === "input" ? builder.inputGuardrails : builder.outputGuardrails;

  return (
    <div className="flex flex-col gap-5" dir={pageDir}>
      <section className="flex flex-col gap-3 rounded-xl border bg-card p-4">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 className="text-2xl font-semibold">Prompt Registry</h2>
            <p className="text-sm text-muted-foreground">
              Create reusable prompt templates, version them intentionally, and execute them through Shahrah Gateway.
            </p>
          </div>
          <Button
            type="button"
            onClick={() => {
              resetBuilder();
              setNotice(null);
            }}
          >
            <Plus data-icon="inline-start" />
            Create Prompt
          </Button>
        </div>

        {notice ? (
          <Alert variant={notice.kind === "error" ? "destructive" : "default"}>
            <AlertTitle>{notice.kind === "error" ? "Action failed" : "Done"}</AlertTitle>
            <AlertDescription>{notice.message}</AlertDescription>
          </Alert>
        ) : null}

        <Card className="shadow-none">
          <CardHeader className="border-b">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <CardTitle>All Prompts</CardTitle>
                <CardDescription>Browse saved prompts and load any version back into the playground.</CardDescription>
              </div>
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                <div className="relative min-w-[240px]">
                  <Search className="pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter") {
                        setPageOffset(0);
                        void loadPrompts(query, 0, pageSize, selectedPromptID);
                      }
                    }}
                    placeholder="Search prompts"
                    className="ps-9"
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => {
                    setPageOffset(0);
                    void loadPrompts(query, 0, pageSize, selectedPromptID);
                  }}
                >
                  Refresh
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent className="p-0">
            {isPromptsLoading ? (
              <div className="space-y-2 p-4">
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
              </div>
            ) : tableRows.length === 0 ? (
              <div className="p-10 text-center text-sm text-muted-foreground">No rows</div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Latest Version</TableHead>
                    <TableHead>Model</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Updated</TableHead>
                    <TableHead>Created By</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tableRows.map((item) => (
                    <TableRow
                      key={item.id}
                      className={cn("cursor-pointer", item.id === selectedPromptID && "bg-primary/5")}
                      onClick={() => void selectPrompt(item.id)}
                    >
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell>v{item.latest_version}</TableCell>
                      <TableCell className="font-mono text-xs">{item.latest_version_detail?.model || "-"}</TableCell>
                      <TableCell>{item.description || "-"}</TableCell>
                      <TableCell>{formatDate(item.updated_at, locale)}</TableCell>
                      <TableCell>{item.created_by || "-"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
            <div className="flex flex-col gap-3 border-t px-4 py-3 text-sm sm:flex-row sm:items-center sm:justify-between">
              <div className="text-muted-foreground">
                Showing {tableRows.length === 0 ? 0 : pageOffset + 1}-{pageCountEnd} of {promptTotal}
              </div>
              <div className="flex items-center gap-2">
                <label className="text-muted-foreground">Rows per page</label>
                <select
                  value={pageSize}
                  onChange={(event) => {
                    const next = Number(event.target.value);
                    setPageSize(next);
                    setPageOffset(0);
                    void loadPrompts(query, 0, next, selectedPromptID);
                  }}
                  className="h-9 rounded-md border bg-background px-2"
                >
                  <option value={10}>10</option>
                  <option value={25}>25</option>
                  <option value={50}>50</option>
                </select>
                <Button
                  type="button"
                  variant="outline"
                  disabled={pageOffset === 0}
                  onClick={() => {
                    const nextOffset = Math.max(pageOffset - pageSize, 0);
                    setPageOffset(nextOffset);
                    void loadPrompts(query, nextOffset, pageSize, selectedPromptID);
                  }}
                >
                  Previous
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  disabled={pageOffset + pageSize >= promptTotal}
                  onClick={() => {
                    const nextOffset = pageOffset + pageSize;
                    setPageOffset(nextOffset);
                    void loadPrompts(query, nextOffset, pageSize, selectedPromptID);
                  }}
                >
                  Next
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </section>

      <div className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
        <div className="space-y-5">
          <Card className="shadow-none">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Settings2 className="size-4" />
                Configuration
              </CardTitle>
              <CardDescription>Pick a model, configure runtime behavior, and prepare reusable template defaults.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-5">
              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Selected model</div>
                    <div className="text-xs text-muted-foreground">
                      {builder.selectedModel || "Choose a Shahrah model alias or virtual model"}
                    </div>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setModelPickerOpen(true)}>
                    {isModelsLoading ? <Loader2 className="animate-spin" /> : <Sparkles />}
                    Select model
                  </Button>
                </div>
                {selectedModelMetadata.catalog ? (
                  <div className="rounded-lg border bg-muted/20 p-3 text-sm">
                    <div className="font-medium">{selectedModelMetadata.catalog.display_name}</div>
                    <div className="mt-1 flex flex-wrap gap-2">
                      {selectedModelMetadata.catalog.context_window ? (
                        <Badge variant="outline">ctx {formatNumber(selectedModelMetadata.catalog.context_window, locale)}</Badge>
                      ) : null}
                      {(selectedModelMetadata.catalog.capabilities ?? []).slice(0, 4).map((capability) => (
                        <Badge key={capability} variant="secondary">
                          {capability}
                        </Badge>
                      ))}
                    </div>
                  </div>
                ) : null}
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="text-sm font-medium">Model settings</div>
                <div className="grid gap-3 sm:grid-cols-2">
                  <Field label="Stream">
                    <div className="flex h-10 items-center justify-between rounded-md border px-3">
                      <span className="text-sm text-muted-foreground">{builder.stream ? "Enabled" : "Disabled"}</span>
                      <Switch
                        checked={builder.stream}
                        onCheckedChange={(checked) => setBuilder((current) => ({ ...current, stream: checked }))}
                      />
                    </div>
                  </Field>
                  <Field label="Temperature">
                    <Input value={builder.temperature} onChange={(event) => setBuilder((current) => ({ ...current, temperature: event.target.value }))} />
                  </Field>
                  <Field label="Top P">
                    <Input value={builder.topP} onChange={(event) => setBuilder((current) => ({ ...current, topP: event.target.value }))} />
                  </Field>
                  <Field label="Max tokens">
                    <Input value={builder.maxTokens} onChange={(event) => setBuilder((current) => ({ ...current, maxTokens: event.target.value }))} />
                  </Field>
                </div>
                {builder.stream && builder.outputGuardrails.length > 0 ? (
                  <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>Streaming conflict</AlertTitle>
                    <AlertDescription>Output guardrails require streaming to be disabled in Shahrah Gateway.</AlertDescription>
                  </Alert>
                ) : null}
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Input variables ({detectedVariables.length})</div>
                    <div className="text-xs text-muted-foreground">
                      Detected from <code>{"{{variable_name}}"}</code> placeholders in messages.
                    </div>
                  </div>
                </div>
                {detectedVariables.length === 0 ? (
                  <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">No variables detected yet.</div>
                ) : (
                  <div className="space-y-2">
                    {detectedVariables.map((variable) => (
                      <Field key={variable} label={variable}>
                        <Input
                          value={builder.variableValues[variable] ?? ""}
                          onChange={(event) =>
                            setBuilder((current) => ({
                              ...current,
                              variableValues: { ...current.variableValues, [variable]: event.target.value },
                            }))
                          }
                          placeholder={`Value for ${variable}`}
                        />
                      </Field>
                    ))}
                  </div>
                )}
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Cache config</div>
                    <div className="text-xs text-muted-foreground">
                      Maps directly to Shahrah&apos;s `x-shahrah-cache-config` header.
                    </div>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setCacheDialogOpen(true)}>
                    <DatabaseZap />
                    Configure
                  </Button>
                </div>
                <ConfigSummaryPills
                  pills={
                    builder.cacheEnabled
                      ? [
                          `type ${builder.cacheType}`,
                          builder.cacheTTL ? `ttl ${builder.cacheTTL}s` : "",
                          builder.cacheNamespace ? `ns ${builder.cacheNamespace}` : "",
                          builder.cacheType === "smart" && builder.cacheSimilarityThreshold
                            ? `similarity ${builder.cacheSimilarityThreshold}`
                            : "",
                        ]
                      : ["disabled"]
                  }
                />
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Input guardrails ({builder.inputGuardrails.length})</div>
                    <div className="text-xs text-muted-foreground">Apply tenant guardrails before the request reaches the model.</div>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setGuardrailDialogMode("input")}>
                    <ShieldCheck />
                    Select
                  </Button>
                </div>
                <ConfigSummaryPills pills={inputGuardrailItems.map((item) => item.name)} emptyLabel="No input guardrails" />
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Output guardrails ({builder.outputGuardrails.length})</div>
                    <div className="text-xs text-muted-foreground">Apply tenant guardrails after the model response is received.</div>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setGuardrailDialogMode("output")}>
                    <ShieldCheck />
                    Select
                  </Button>
                </div>
                <ConfigSummaryPills pills={outputGuardrailItems.map((item) => item.name)} emptyLabel="No output guardrails" />
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="flex items-center justify-between gap-2">
                  <div>
                    <div className="text-sm font-medium">Structured output</div>
                    <div className="text-xs text-muted-foreground">Optional OpenAI-compatible `response_format` configuration.</div>
                  </div>
                  <Button type="button" variant="outline" onClick={() => setStructuredDialogOpen(true)}>
                    <Braces />
                    Configure
                  </Button>
                </div>
                <ConfigSummaryPills
                  pills={
                    builder.structuredEnabled
                      ? [builder.structuredFormat, builder.structuredStrict ? "strict" : "non-strict"]
                      : ["disabled"]
                  }
                />
              </section>

              <Separator />

              <section className="space-y-3">
                <div className="text-sm font-medium">Metadata</div>
                <div className="space-y-2">
                  {builder.metadataRows.length === 0 ? (
                    <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">No metadata rows added yet.</div>
                  ) : null}
                  {builder.metadataRows.map((row) => (
                    <div key={row.id} className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
                      <Input value={row.key} placeholder="key" onChange={(event) => updateMetadataRow(row.id, { key: event.target.value })} />
                      <Input value={row.value} placeholder="value" onChange={(event) => updateMetadataRow(row.id, { value: event.target.value })} />
                      <Button type="button" variant="outline" onClick={() => removeMetadataRow(row.id)}>
                        Remove
                      </Button>
                    </div>
                  ))}
                  <Button type="button" variant="outline" onClick={addMetadataRow}>
                    <Plus />
                    Add metadata
                  </Button>
                </div>
              </section>

              <Separator />

              <section className="space-y-2">
                <div className="text-sm font-medium">Logging</div>
                <div className="rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground">
                  Prompt-level log storage toggles are deferred in this v1 because Shahrah already enforces request/spend logging at gateway level.
                </div>
              </section>
            </CardContent>
          </Card>

          <Card className="shadow-none">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <History className="size-4" />
                Versions
              </CardTitle>
              <CardDescription>Immutable prompt versions remain the canonical saved runtime artifact.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {isVersionsLoading ? (
                <div className="space-y-2">
                  <Skeleton className="h-12 w-full" />
                  <Skeleton className="h-12 w-full" />
                </div>
              ) : versions.length === 0 ? (
                <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">Save a prompt first to start building version history.</div>
              ) : (
                <div className="space-y-2">
                  {versions.map((version) => (
                    <button
                      key={version.id}
                      type="button"
                      className={cn(
                        "w-full rounded-md border p-3 text-start transition hover:bg-accent",
                        version.id === selectedVersionID && "border-primary/40 bg-primary/5",
                      )}
                      onClick={() => {
                        setSelectedVersionID(version.id);
                        hydrateBuilder(version);
                      }}
                    >
                      <div className="flex items-center justify-between gap-2">
                        <span className="font-medium">Version {version.version}</span>
                        <Badge variant="secondary">{version.model || version.manifest?.model || "-"}</Badge>
                      </div>
                      <div className="mt-1 truncate text-xs text-muted-foreground" dir="ltr">
                        {version.commit_message || version.fqn}
                      </div>
                    </button>
                  ))}
                </div>
              )}
              {selectedVersion ? (
                <div className="rounded-lg border bg-muted/20 p-3">
                  <div className="text-xs text-muted-foreground">Selected FQN</div>
                  <div className="mt-1 font-mono text-xs" dir="ltr">
                    {selectedVersion.fqn}
                  </div>
                  <div className="mt-3 flex flex-wrap gap-2">
                    <Button type="button" variant="outline" size="sm" onClick={() => void copySelectedFQN()}>
                      <Clipboard />
                      Copy FQN
                    </Button>
                    <Button type="button" variant="outline" size="sm" onClick={() => void copySelectedSnippet()}>
                      <FileCode2 />
                      Copy SDK snippet
                    </Button>
                  </div>
                </div>
              ) : null}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-5">
          <Card className="shadow-none">
            <CardHeader className="border-b">
              <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                <div>
                  <div className="text-xs text-muted-foreground">
                    Prompt Registry {selectedPrompt ? `> ${selectedPrompt.name}` : "> New Prompt"}
                  </div>
                  <CardTitle>{selectedPrompt ? "Edit prompt and save a new version" : "New prompt playground"}</CardTitle>
                  <CardDescription>
                    Build the reusable template on the left, then run it through Shahrah Gateway before saving.
                  </CardDescription>
                </div>
                <Button type="button" onClick={() => void savePrompt()} disabled={!canSave || isSaving}>
                  {isSaving ? <Loader2 className="animate-spin" /> : <Save />}
                  {selectedPrompt ? "Save New Version" : "Save New Prompt"}
                </Button>
              </div>
            </CardHeader>
            <CardContent className="space-y-5 p-4">
              <div className="grid gap-3 md:grid-cols-2">
                <Field label="Prompt name">
                  <Input value={builder.name} onChange={(event) => setBuilder((current) => ({ ...current, name: event.target.value }))} placeholder="booking_greeting" />
                </Field>
                <Field label="Commit message">
                  <Input
                    value={builder.commitMessage}
                    onChange={(event) => setBuilder((current) => ({ ...current, commitMessage: event.target.value }))}
                    placeholder="What changed in this version?"
                  />
                </Field>
              </div>
              <Field label="Description">
                <Input value={builder.description} onChange={(event) => setBuilder((current) => ({ ...current, description: event.target.value }))} />
              </Field>

              <Card className="shadow-none">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm">SYSTEM PROMPT</CardTitle>
                  <CardDescription>Use reusable placeholders such as {`{{application}}`}.</CardDescription>
                </CardHeader>
                <CardContent>
                  <Textarea
                    value={builder.systemPrompt}
                    onChange={(event) => setBuilder((current) => ({ ...current, systemPrompt: event.target.value }))}
                    className="min-h-28"
                    dir="auto"
                  />
                </CardContent>
              </Card>

              <Card className="shadow-none">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm">Conversation template</CardTitle>
                  <CardDescription>Add reusable user or assistant turns that should be versioned with the prompt.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {builder.messages.length === 0 ? (
                    <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">No conversation messages yet.</div>
                  ) : (
                    builder.messages.map((message, index) => (
                      <div key={`message-${index}`} className="rounded-lg border p-3">
                        <div className="mb-2 flex items-center justify-between gap-3">
                          <select
                            value={message.role}
                            onChange={(event) => updateMessage(index, { role: event.target.value as "user" | "assistant" })}
                            className="h-9 rounded-md border bg-background px-2 text-sm"
                          >
                            <option value="user">User</option>
                            <option value="assistant">Assistant</option>
                          </select>
                          <Button type="button" variant="outline" size="sm" onClick={() => removeMessage(index)}>
                            Remove
                          </Button>
                        </div>
                        <Textarea
                          value={message.content}
                          onChange={(event) => updateMessage(index, { content: event.target.value })}
                          className="min-h-24"
                          dir="auto"
                        />
                      </div>
                    ))
                  )}

                  <div className="rounded-lg border bg-muted/20 p-3">
                    <div className="mb-2 flex items-center gap-2">
                      <select
                        value={builder.composerRole}
                        onChange={(event) => setBuilder((current) => ({ ...current, composerRole: event.target.value as "user" | "assistant" }))}
                        className="h-9 rounded-md border bg-background px-2 text-sm"
                      >
                        <option value="user">User</option>
                        <option value="assistant">Assistant</option>
                      </select>
                      <Button type="button" variant="outline" onClick={addComposerMessage} disabled={!builder.composerText.trim()}>
                        <Plus />
                        Add message
                      </Button>
                    </div>
                    <Textarea
                      value={builder.composerText}
                      onChange={(event) => setBuilder((current) => ({ ...current, composerText: event.target.value }))}
                      placeholder="Enter reusable message text..."
                      className="min-h-24"
                      dir="auto"
                    />
                  </div>
                </CardContent>
              </Card>

              <Card className="shadow-none">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm">Playground response</CardTitle>
                  <CardDescription>Run the current draft through Shahrah Gateway before saving.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="rounded-lg border bg-muted/20 p-3">
                    {runPreview ? (
                      <pre className="whitespace-pre-wrap break-words text-sm leading-6">{runPreview}</pre>
                    ) : (
                      <div className="text-sm text-muted-foreground">Response will appear here.</div>
                    )}
                  </div>
                  {runResponse?.usage ? (
                    <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                      <Badge variant="outline">prompt {runResponse.usage.prompt_tokens ?? 0}</Badge>
                      <Badge variant="outline">completion {runResponse.usage.completion_tokens ?? 0}</Badge>
                      <Badge variant="outline">total {runResponse.usage.total_tokens ?? 0}</Badge>
                    </div>
                  ) : null}
                  <div className="flex flex-wrap gap-2">
                    <Button type="button" onClick={() => void runPrompt()} disabled={isRunning || !builder.selectedModel}>
                      {isRunning ? <Loader2 className="animate-spin" /> : <Play />}
                      Run
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </CardContent>
          </Card>
        </div>
      </div>

      <Dialog open={modelPickerOpen} onOpenChange={setModelPickerOpen}>
        <DialogContent className="w-[min(1100px,95vw)] max-w-[1100px]">
          <DialogHeader>
            <DialogTitle>Select model</DialogTitle>
            <DialogDescription>Choose a real Shahrah model alias or virtual model that is already routable by the gateway.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <Input value={modelSearch} onChange={(event) => setModelSearch(event.target.value)} placeholder="Search models" />
            <div className="grid min-h-[380px] gap-3 md:grid-cols-[220px_minmax(0,1fr)]">
              <div className="space-y-1 rounded-lg border bg-muted/20 p-2">
                {providerOptions.map((provider) => (
                  <button
                    key={provider.id}
                    type="button"
                    className={cn(
                      "w-full rounded-md px-3 py-2 text-start text-sm transition hover:bg-background",
                      visibleProviderID === provider.id && "bg-primary/10 text-primary ring-1 ring-primary/30",
                    )}
                    onClick={() => setSelectedProviderFilter(provider.id)}
                  >
                    <div className="font-medium">{provider.label}</div>
                    {provider.description ? <div className="line-clamp-2 text-xs text-muted-foreground">{provider.description}</div> : null}
                  </button>
                ))}
              </div>
              <div className="space-y-2 rounded-lg border p-2">
                {visibleModels.length === 0 ? (
                  <div className="flex h-full items-center justify-center text-sm text-muted-foreground">No models available for this provider.</div>
                ) : (
                  visibleModels.map((model, index) => {
                    const catalog = visibleProviderCatalogModels[index];
                    const isSelected = builder.selectedModel === model.id;
                    return (
                      <button
                        key={model.id}
                        type="button"
                        className={cn(
                          "w-full rounded-md border p-3 text-start transition hover:bg-accent",
                          isSelected && "border-primary/40 bg-primary/5",
                        )}
                        onClick={() => {
                          setBuilder((current) => ({
                            ...current,
                            selectedModel: model.id,
                            selectedProvider: visibleProviderID,
                          }));
                          setModelPickerOpen(false);
                        }}
                      >
                        <div className="flex items-center justify-between gap-3">
                          <div className="min-w-0">
                            <div className="truncate font-medium">{catalog?.display_name || model.id}</div>
                            <div className="truncate text-xs text-muted-foreground" dir="ltr">
                              {model.id}
                            </div>
                          </div>
                          {isSelected ? <Check className="size-4 shrink-0 text-primary" /> : null}
                        </div>
                        {catalog ? (
                          <div className="mt-2 flex flex-wrap gap-2">
                            {catalog.context_window ? <Badge variant="outline">ctx {formatNumber(catalog.context_window, locale)}</Badge> : null}
                            {(catalog.capabilities ?? []).slice(0, 4).map((capability) => (
                              <Badge key={`${model.id}-${capability}`} variant="secondary">
                                {capability}
                              </Badge>
                            ))}
                          </div>
                        ) : (
                          <div className="mt-2 flex flex-wrap gap-2">
                            {(model.aliases ?? []).slice(0, 3).map((alias) => (
                              <Badge key={`${model.id}-${alias}`} variant="outline">
                                {alias}
                              </Badge>
                            ))}
                          </div>
                        )}
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={cacheDialogOpen} onOpenChange={setCacheDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Cache Config</DialogTitle>
            <DialogDescription>Configure Shahrah response cache using the same contract as `x-shahrah-cache-config`.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center justify-between rounded-md border px-3 py-2">
              <div>
                <div className="text-sm font-medium">Enable cache</div>
                <div className="text-xs text-muted-foreground">Use `simple` for exact-match and `smart` for semantic cache.</div>
              </div>
              <Switch
                checked={builder.cacheEnabled}
                onCheckedChange={(checked) => setBuilder((current) => ({ ...current, cacheEnabled: checked }))}
              />
            </div>
            <Field label="Cache type">
              <select
                value={builder.cacheType}
                onChange={(event) => setBuilder((current) => ({ ...current, cacheType: event.target.value as "simple" | "smart" }))}
                className="h-10 rounded-md border bg-background px-3 text-sm"
              >
                <option value="simple">simple</option>
                <option value="smart">smart</option>
              </select>
            </Field>
            <Field label="TTL seconds">
              <Input value={builder.cacheTTL} onChange={(event) => setBuilder((current) => ({ ...current, cacheTTL: event.target.value }))} />
            </Field>
            <Field label="Namespace">
              <Input value={builder.cacheNamespace} onChange={(event) => setBuilder((current) => ({ ...current, cacheNamespace: event.target.value }))} />
            </Field>
            {builder.cacheType === "smart" ? (
              <Field label="Similarity threshold">
                <Input value={builder.cacheSimilarityThreshold} onChange={(event) => setBuilder((current) => ({ ...current, cacheSimilarityThreshold: event.target.value }))} />
              </Field>
            ) : null}
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setCacheDialogOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={guardrailDialogMode !== null} onOpenChange={(open) => { if (!open) setGuardrailDialogMode(null); }}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{guardrailDialogMode === "input" ? "Input Guardrails" : "Output Guardrails"}</DialogTitle>
            <DialogDescription>
              {guardrailDialogMode === "input"
                ? "Attach tenant guardrails that validate or mutate the request before model execution."
                : "Attach tenant guardrails that validate or mutate the model response after execution."}
            </DialogDescription>
          </DialogHeader>
          {isGuardrailsLoading ? (
            <div className="space-y-2">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : (
            <div className="space-y-3">
              {guardrailDialogMode === "output" ? (
                <Alert variant={builder.stream ? "destructive" : "default"}>
                  <AlertCircle className="size-4" />
                  <AlertTitle>{builder.stream ? "Streaming must be disabled" : "Streaming is already disabled"}</AlertTitle>
                  <AlertDescription>
                    Shahrah currently rejects output guardrails when streaming is enabled.
                    {builder.stream ? " Disable stream in model settings before running or saving." : ""}
                  </AlertDescription>
                </Alert>
              ) : null}
              <Field label="Scope">
                <select
                  value={builder.guardrailScope}
                  onChange={(event) => setBuilder((current) => ({ ...current, guardrailScope: event.target.value as "all" | "last" }))}
                  className="h-10 rounded-md border bg-background px-3 text-sm"
                >
                  <option value="last">last</option>
                  <option value="all">all</option>
                </select>
              </Field>
              <div className="max-h-[360px] space-y-2 overflow-y-auto rounded-lg border p-2">
                {guardrails.map((guardrail) => {
                  const selector = guardrail.selector || `guardrails/${guardrail.slug}`;
                  const checked = selectedGuardrailMode.includes(selector);
                  return (
                    <label key={guardrail.id || guardrail.slug} className="flex items-start gap-3 rounded-md border p-3">
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={(event) => {
                          setBuilder((current) => {
                            const targetKey = guardrailDialogMode === "input" ? "inputGuardrails" : "outputGuardrails";
                            const currentValues = current[targetKey];
                            const nextValues = event.target.checked
                              ? [...currentValues, selector]
                              : currentValues.filter((value) => value !== selector);
                            return { ...current, [targetKey]: nextValues };
                          });
                        }}
                        className="mt-1"
                      />
                      <span className="min-w-0 flex-1">
                        <span className="block font-medium">{guardrail.name}</span>
                        <span className="block text-xs text-muted-foreground">{guardrail.description || guardrail.type || selector}</span>
                      </span>
                    </label>
                  );
                })}
              </div>
            </div>
          )}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setGuardrailDialogMode(null)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={structuredDialogOpen} onOpenChange={setStructuredDialogOpen}>
        <DialogContent className="w-[min(960px,95vw)] max-w-[960px]">
          <DialogHeader>
            <DialogTitle>Structured Output</DialogTitle>
            <DialogDescription>Configure OpenAI-compatible `response_format` for structured output.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex items-center justify-between rounded-md border px-3 py-2">
              <div>
                <div className="text-sm font-medium">Enable structured output</div>
                <div className="text-xs text-muted-foreground">Unsupported models should be handled explicitly by the UI and backend response.</div>
              </div>
              <Switch
                checked={builder.structuredEnabled}
                onCheckedChange={(checked) => setBuilder((current) => ({ ...current, structuredEnabled: checked }))}
              />
            </div>
            <div className="grid gap-3 sm:grid-cols-2">
              <Field label="Format">
                <select
                  value={builder.structuredFormat}
                  onChange={(event) => setBuilder((current) => ({ ...current, structuredFormat: event.target.value as "json_schema" | "json_object" }))}
                  className="h-10 rounded-md border bg-background px-3 text-sm"
                >
                  <option value="json_schema">json_schema</option>
                  <option value="json_object">json_object</option>
                </select>
              </Field>
              <Field label="Strict">
                <div className="flex h-10 items-center justify-between rounded-md border px-3">
                  <span className="text-sm text-muted-foreground">{builder.structuredStrict ? "Enabled" : "Disabled"}</span>
                  <Switch
                    checked={builder.structuredStrict}
                    onCheckedChange={(checked) => setBuilder((current) => ({ ...current, structuredStrict: checked }))}
                  />
                </div>
              </Field>
            </div>
            <Field label="JSON Schema">
              <Textarea
                value={builder.structuredSchema}
                onChange={(event) => setBuilder((current) => ({ ...current, structuredSchema: event.target.value }))}
                className="min-h-[300px] font-mono text-xs"
                dir="ltr"
                spellCheck={false}
              />
            </Field>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setStructuredDialogOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Field(props: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label>{props.label}</Label>
      {props.children}
    </div>
  );
}

function ConfigSummaryPills(props: { pills: string[]; emptyLabel?: string }) {
  const filtered = props.pills.filter(Boolean);
  if (filtered.length === 0) {
    return <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">{props.emptyLabel || "No values configured."}</div>;
  }
  return (
    <div className="flex flex-wrap gap-2">
      {filtered.map((pill) => (
        <Badge key={pill} variant="outline">
          {pill}
        </Badge>
      ))}
    </div>
  );
}

function buildFallbackManifest(version: PromptVersion): PromptManifest {
  return {
    type: "chat_prompt",
    name: version.name,
    description: version.description,
    messages: version.messages,
    variables: version.variables,
    model: version.model,
    parameters: version.parameters,
    metadata: version.metadata,
    version_alias: version.version_alias,
  };
}

function buildManifest(builder: BuilderState): { ok: true; manifest: PromptManifest } | { ok: false; message: string } {
  const name = normalizePromptSlug(builder.name);
  if (!/^[A-Za-z0-9_-]{1,256}$/.test(name)) {
    return { ok: false, message: "Prompt name must contain only letters, numbers, hyphens, and underscores." };
  }
  if (!builder.selectedModel.trim()) {
    return { ok: false, message: "Select a model before saving." };
  }
  const temperature = parseOptionalNumber(builder.temperature);
  const topP = parseOptionalNumber(builder.topP);
  const maxTokens = parseOptionalInteger(builder.maxTokens);
  if (temperature !== undefined && (temperature < 0 || temperature > 2)) {
    return { ok: false, message: "Temperature must be between 0 and 2." };
  }
  if (topP !== undefined && (topP < 0 || topP > 1)) {
    return { ok: false, message: "Top P must be between 0 and 1." };
  }
  if (maxTokens !== undefined && maxTokens < 1) {
    return { ok: false, message: "Max tokens must be at least 1." };
  }

  const metadata = buildMetadataObject(builder.metadataRows);
  if (!metadata.ok) return metadata;

  let structuredOutput: StructuredOutputConfig | undefined;
  if (builder.structuredEnabled) {
    if (builder.structuredFormat === "json_schema") {
      try {
        const schema = JSON.parse(builder.structuredSchema) as Record<string, unknown>;
        if (!schema || typeof schema !== "object" || Array.isArray(schema)) {
          return { ok: false, message: "Structured output schema must be a JSON object." };
        }
        structuredOutput = {
          enabled: true,
          format: "json_schema",
          strict: builder.structuredStrict,
          schema,
        };
      } catch {
        return { ok: false, message: "Structured output schema is not valid JSON." };
      }
    } else {
      structuredOutput = {
        enabled: true,
        format: "json_object",
        strict: builder.structuredStrict,
      };
    }
  }

  const cacheConfig = buildCacheConfig(builder);
  if (!cacheConfig.ok) return cacheConfig;

  if (builder.stream && builder.outputGuardrails.length > 0) {
    return { ok: false, message: "Output guardrails require streaming to be disabled." };
  }

  const promptMessages: Array<{ role: string; content?: unknown }> = [];
  if (builder.systemPrompt.trim()) {
    promptMessages.push({ role: "system", content: builder.systemPrompt.trim() });
  }
  for (const message of builder.messages) {
    if (message.content.trim()) {
      promptMessages.push({ role: message.role, content: message.content.trim() });
    }
  }
  if (promptMessages.length === 0) {
    return { ok: false, message: "Add at least one system, user, or assistant message." };
  }

  const modelSettings: PromptModelSettings = {
    stream: builder.stream,
    ...(temperature !== undefined ? { temperature } : {}),
    ...(topP !== undefined ? { top_p: topP } : {}),
    ...(maxTokens !== undefined ? { max_tokens: maxTokens } : {}),
  };
  const parameters: Record<string, unknown> = {
    stream: builder.stream,
    ...(temperature !== undefined ? { temperature } : {}),
    ...(topP !== undefined ? { top_p: topP } : {}),
    ...(maxTokens !== undefined ? { max_tokens: maxTokens } : {}),
  };

  return {
    ok: true,
    manifest: {
      type: "chat_prompt",
      name,
      description: builder.description.trim(),
      messages: promptMessages,
      variables: { ...builder.variableValues },
      model: builder.selectedModel.trim(),
      parameters,
      metadata: metadata.value,
      model_settings: modelSettings,
      cache: cacheConfig.value,
      guardrails: {
        input: builder.inputGuardrails,
        output: builder.outputGuardrails,
        scope: builder.guardrailScope,
      },
      structured_output: structuredOutput,
    },
  };
}

function buildCacheConfig(builder: BuilderState): { ok: true; value?: PromptCacheConfig } | { ok: false; message: string } {
  if (!builder.cacheEnabled) {
    return { ok: true, value: { enabled: false } };
  }
  const ttl = parseOptionalInteger(builder.cacheTTL);
  if (ttl === undefined || ttl < 1) {
    return { ok: false, message: "Cache TTL must be a positive integer." };
  }
  if (builder.cacheType === "smart") {
    const similarity = parseOptionalNumber(builder.cacheSimilarityThreshold);
    if (similarity === undefined || similarity < 0.8 || similarity > 1) {
      return { ok: false, message: "Smart cache similarity threshold must be between 0.8 and 1." };
    }
    return {
      ok: true,
      value: {
        enabled: true,
        type: "smart",
        ttl,
        namespace: builder.cacheNamespace.trim(),
        similarity_threshold: similarity,
      },
    };
  }
  return {
    ok: true,
    value: {
      enabled: true,
      type: "simple",
      ttl,
      namespace: builder.cacheNamespace.trim(),
    },
  };
}

function buildMetadataObject(rows: MetadataRow[]): { ok: true; value: Record<string, unknown> } | { ok: false; message: string } {
  const out: Record<string, unknown> = {};
  for (const row of rows) {
    const key = row.key.trim();
    const value = row.value.trim();
    if (!key && !value) continue;
    if (!key) {
      return { ok: false, message: "Metadata keys cannot be empty." };
    }
    if (key in out) {
      return { ok: false, message: `Duplicate metadata key: ${key}` };
    }
    out[key] = value;
  }
  return { ok: true, value: out };
}

function buildRunPayload(builder: BuilderState, manifest: PromptManifest) {
  const requiredVariables = extractPromptVariableNames([
    builder.systemPrompt,
    ...builder.messages.map((message) => message.content),
  ]);
  const missingVariables = requiredVariables.filter((variable) => !(builder.variableValues[variable] ?? "").trim());
  if (missingVariables.length > 0) {
    return { ok: false as const, message: `Provide runtime values for: ${missingVariables.join(", ")}` };
  }

  const renderedMessages: Array<{ role: string; content: string }> = [];
  const allMissing = new Set<string>();
  if (builder.systemPrompt.trim()) {
    const rendered = renderPromptText(builder.systemPrompt.trim(), builder.variableValues);
    rendered.missing.forEach((item) => allMissing.add(item));
    renderedMessages.push({ role: "system", content: rendered.rendered });
  }
  for (const message of builder.messages) {
    const rendered = renderPromptText(message.content.trim(), builder.variableValues);
    rendered.missing.forEach((item) => allMissing.add(item));
    renderedMessages.push({ role: message.role, content: rendered.rendered });
  }
  if (allMissing.size > 0) {
    return { ok: false as const, message: `Missing values for: ${[...allMissing].join(", ")}` };
  }

  const request: Record<string, unknown> = {
    model: manifest.model,
    messages: renderedMessages,
    stream: builder.stream,
    metadata: manifest.metadata ?? {},
  };
  const temperature = parseOptionalNumber(builder.temperature);
  const topP = parseOptionalNumber(builder.topP);
  const maxTokens = parseOptionalInteger(builder.maxTokens);
  if (temperature !== undefined) request.temperature = temperature;
  if (topP !== undefined) request.top_p = topP;
  if (maxTokens !== undefined) request.max_tokens = maxTokens;

  if (builder.structuredEnabled) {
    if (builder.structuredFormat === "json_schema") {
      const schema = JSON.parse(builder.structuredSchema) as Record<string, unknown>;
      request.response_format = {
        type: "json_schema",
        json_schema: {
          name: normalizePromptSlug(manifest.name || "prompt_output"),
          schema,
          strict: builder.structuredStrict,
        },
      };
    } else {
      request.response_format = { type: "json_object" };
    }
  }

  const cacheConfig = builder.cacheEnabled
    ? builder.cacheType === "smart"
      ? {
          type: "smart",
          ttl: parseOptionalInteger(builder.cacheTTL),
          namespace: builder.cacheNamespace.trim() || undefined,
          similarity_threshold: parseOptionalNumber(builder.cacheSimilarityThreshold),
        }
      : {
          type: "simple",
          ttl: parseOptionalInteger(builder.cacheTTL),
          namespace: builder.cacheNamespace.trim() || undefined,
        }
    : null;

  return {
    ok: true as const,
    payload: {
      request,
      cache_config: cacheConfig,
      guardrails: {
        input: builder.inputGuardrails,
        output: builder.outputGuardrails,
        scope: builder.guardrailScope,
      },
    },
  };
}

function previewRunResponse(response: PromptRunResponse): string {
  const content = response.choices?.[0]?.message?.content;
  if (typeof content === "string") {
    return content;
  }
  return JSON.stringify(response, null, 2);
}

function parseOptionalNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function parseOptionalInteger(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function numberToInput(value?: number): string {
  return typeof value === "number" && Number.isFinite(value) ? `${value}` : "";
}

function numberFromUnknown(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim()) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : undefined;
  }
  return undefined;
}

function booleanFromUnknown(value: unknown, fallback = false): boolean {
  if (typeof value === "boolean") return value;
  if (typeof value === "string") return value.trim().toLowerCase() === "true";
  return fallback;
}

function stringifyScalar(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return `${value}`;
  return JSON.stringify(value);
}

function findProviderForModel(modelID: string, models: GatewayModel[]): string {
  const found = models.find((item) => item.id === modelID);
  return found?.provider?.trim().toLowerCase() || "";
}

function formatNumber(value: number, locale: string) {
  return new Intl.NumberFormat(locale).format(value);
}

function formatDate(value: string | undefined, locale: string) {
  if (!value) return "-";
  return new Intl.DateTimeFormat(locale, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
