"use client";

import { useEffect, useState } from "react";
import { useLocale, useTranslations } from "next-intl";
import { useRouter } from "next/navigation";

import { normalizeError, requestJSON, unwrapData } from "@/app/components/provider-account-wizard/utils";
import type { ProviderCatalogModel, ProviderCatalogProvider } from "@/app/components/provider-account-wizard/types";

import { PromptBuilder } from "./prompt-builder";
import {
  buildManifest,
  buildPromptSnippet,
  buildRunPayload,
  previewRunResponse,
  type GatewayModelListResponse,
  type GuardrailListResponse,
  type Notice,
  type Prompt,
  type PromptRunResponse,
  type PromptVersion,
  type PromptVersionListResponse,
} from "./prompt-management-shared";
import { usePromptBuilderState } from "./use-prompt-builder-state";

type PromptBuilderPageProps = {
  mode: "create" | "edit";
  promptID?: string;
};

export default function PromptBuilderPage(props: PromptBuilderPageProps) {
  const t = useTranslations("prompts");
  const locale = useLocale();
  const router = useRouter();
  const pageDir = locale === "fa" ? "rtl" : "ltr";
  const defaultSystemPrompt = t("builder.systemPrompt.default");

  const [notice, setNotice] = useState<Notice>(null);
  const [isPromptLoading, setIsPromptLoading] = useState(props.mode === "edit");
  const [isVersionsLoading, setIsVersionsLoading] = useState(props.mode === "edit");
  const [isSaving, setIsSaving] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [isModelsLoading, setIsModelsLoading] = useState(false);
  const [isGuardrailsLoading, setIsGuardrailsLoading] = useState(false);

  const [selectedPrompt, setSelectedPrompt] = useState<Prompt | null>(null);
  const [versions, setVersions] = useState<PromptVersion[]>([]);
  const [selectedVersionID, setSelectedVersionID] = useState("");
  const [runResponse, setRunResponse] = useState<PromptRunResponse | null>(null);
  const [runPreview, setRunPreview] = useState("");

  const [models, setModels] = useState<import("./prompt-management-shared").GatewayModel[]>([]);
  const [providerCatalog, setProviderCatalog] = useState<ProviderCatalogProvider[]>([]);
  const [providerModels, setProviderModels] = useState<ProviderCatalogModel[]>([]);
  const [guardrails, setGuardrails] = useState<import("./prompt-management-shared").GuardrailItem[]>([]);

  const {
    builder,
    canSave,
    detectedVariables,
    setBuilder,
    hydrateBuilder,
    addComposerMessage,
    updateMessage,
    removeMessage,
    addMetadataRow,
    updateMetadataRow,
    removeMetadataRow,
  } = usePromptBuilderState(defaultSystemPrompt, models);

  const selectedVersion = versions.find((item) => item.id === selectedVersionID) ?? selectedPrompt?.latest_version_detail ?? null;

  useEffect(() => {
    void Promise.all([loadModels(), loadGuardrails()]);
    if (props.mode === "edit" && props.promptID) {
      void loadPrompt(props.promptID);
      void loadVersions(props.promptID);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.mode, props.promptID]);

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
      setNotice({ kind: "error", message: normalizeError(error, t("messages.loadModelsError")) });
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
      setNotice({ kind: "error", message: normalizeError(error, t("messages.loadGuardrailsError")) });
    } finally {
      setIsGuardrailsLoading(false);
    }
  }

  async function loadPrompt(promptID: string) {
    setIsPromptLoading(true);
    try {
      const payload = await requestJSON<{ data?: Prompt } | Prompt>(`/api/prompts/${encodeURIComponent(promptID)}`);
      const prompt = unwrapData<Prompt>(payload);
      if (!prompt) {
        throw new Error(t("messages.loadPromptsError"));
      }
      setSelectedPrompt(prompt);
      if (prompt.latest_version_detail) {
        hydrateBuilder(prompt.latest_version_detail);
        setSelectedVersionID(prompt.latest_version_detail.id);
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.loadPromptsError")) });
    } finally {
      setIsPromptLoading(false);
    }
  }

  async function loadVersions(promptID: string, preferredVersionID = "") {
    setIsVersionsLoading(true);
    try {
      const payload = await requestJSON<PromptVersionListResponse>(`/api/prompts/${encodeURIComponent(promptID)}/versions?limit=100`);
      const nextVersions = payload.data ?? [];
      setVersions(nextVersions);
      const nextSelectedVersion = nextVersions.find((entry) => entry.id === preferredVersionID) ?? nextVersions[0] ?? null;
      if (nextSelectedVersion) {
        setSelectedVersionID(nextSelectedVersion.id);
        hydrateBuilder(nextSelectedVersion);
      }
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.loadVersionsError")) });
      setVersions([]);
    } finally {
      setIsVersionsLoading(false);
    }
  }

  async function savePrompt() {
    const manifestResult = buildManifest(builder, t);
    if (!manifestResult.ok) {
      setNotice({ kind: "error", message: manifestResult.message });
      return;
    }
    setIsSaving(true);
    setNotice(null);
    try {
      const path = props.mode === "edit" && props.promptID ? `/api/prompts/${encodeURIComponent(props.promptID)}/versions` : "/api/prompts";
      const saved = await requestJSON<PromptVersion>(path, {
        method: "POST",
        body: JSON.stringify({
          manifest: manifestResult.manifest,
          commit_message: builder.commitMessage.trim(),
        }),
      });
      if (props.mode === "create") {
        router.push(`/prompts/${saved.prompt_id}`);
        return;
      }
      setNotice({ kind: "success", message: t("messages.versionSaved") });
      await Promise.all([loadPrompt(saved.prompt_id), loadVersions(saved.prompt_id, saved.id)]);
    } catch (error) {
      setNotice({ kind: "error", message: normalizeError(error, t("messages.savePromptError")) });
    } finally {
      setIsSaving(false);
    }
  }

  async function runPrompt() {
    const manifestResult = buildManifest(builder, t);
    if (!manifestResult.ok) {
      setNotice({ kind: "error", message: manifestResult.message });
      return;
    }
    const runtimeRequest = buildRunPayload(builder, manifestResult.manifest, t);
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
      setNotice({ kind: "error", message: normalizeError(error, t("messages.runPromptError")) });
    } finally {
      setIsRunning(false);
    }
  }

  async function copySelectedFqn() {
    if (!selectedVersion?.fqn) return;
    await navigator.clipboard.writeText(selectedVersion.fqn);
    setNotice({ kind: "success", message: t("messages.fqnCopied") });
  }

  async function copySelectedSnippet() {
    if (!selectedVersion?.fqn) return;
    await navigator.clipboard.writeText(buildPromptSnippet(selectedVersion.fqn, selectedVersion.model));
    setNotice({ kind: "success", message: t("messages.snippetCopied") });
  }

  if (props.mode === "edit" && isPromptLoading) {
    return <div dir={pageDir} className="rounded-xl border bg-card p-6 text-sm text-muted-foreground">{t("list.description")}</div>;
  }

  return (
    <div dir={pageDir} className="flex flex-col gap-5">
      <section className="flex flex-col gap-3 rounded-xl border bg-card p-4">
        <div>
          <h2 className="text-2xl font-semibold">{props.mode === "edit" ? selectedPrompt?.name || t("header.title") : t("actions.createPrompt")}</h2>
          <p className="text-sm text-muted-foreground">{t("builder.description")}</p>
        </div>
      </section>

      <PromptBuilder
        builder={builder}
        canSave={canSave}
        detectedVariables={detectedVariables}
        isRunning={isRunning}
        isSaving={isSaving}
        locale={locale}
        mode={props.mode}
        notice={notice}
        onAddComposerMessage={addComposerMessage}
        onAddMetadataRow={addMetadataRow}
        onBuilderChange={(updater) => setBuilder(updater)}
        onCopySelectedFqn={copySelectedFqn}
        onCopySelectedSnippet={copySelectedSnippet}
        onRemoveMessage={removeMessage}
        onRemoveMetadataRow={removeMetadataRow}
        onRunPrompt={runPrompt}
        onSavePrompt={savePrompt}
        onSelectVersion={(version) => {
          setSelectedVersionID(version.id);
          hydrateBuilder(version);
        }}
        onUpdateMessage={updateMessage}
        onUpdateMetadataRow={updateMetadataRow}
        promptCatalog={{
          models,
          providerCatalog,
          providerModels,
          guardrails,
          isModelsLoading,
          isGuardrailsLoading,
        }}
        runPreview={runPreview}
        runResponse={runResponse}
        selectedPromptName={selectedPrompt?.name}
        selectedVersion={selectedVersion}
        selectedVersionID={selectedVersionID}
        t={t}
        versionsLoading={isVersionsLoading}
        versions={versions}
      />
    </div>
  );
}
