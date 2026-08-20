"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { BarChart3, Bot, Globe2, LockKeyhole, MessageCircle, Plus, RefreshCw, Rocket } from "lucide-react";
import Link from "next/link";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { CheckboxOptionList } from "@/components/app-ui/CheckboxOptionList";
import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  type ChatbotInput,
  createChatbot,
  getChatbotAnalytics,
  listAnswerModels,
  listChatbots,
  listKnowledgeBases,
  publishChatbot,
  unpublishChatbot,
} from "@/features/managed-rag/api";
import type { Chatbot, ChatbotAnalytics, KnowledgeBase } from "@/lib/contracts/managed-rag";
import { getAPIErrorMessage } from "@/lib/http/api-client";

export default function ChatbotsPage() {
  const t = useTranslations("managedRag");
  const [items, setItems] = useState<Chatbot[]>([]);
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBase[]>([]);
  const [answerModels, setAnswerModels] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [pendingID, setPendingID] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [analytics, setAnalytics] = useState<{ chatbot: Chatbot; data: ChatbotAnalytics } | null>(null);
  const [selectedKBs, setSelectedKBs] = useState<string[]>([]);
  const [accessMode, setAccessMode] = useState<"private" | "public">("private");
  const [retention, setRetention] = useState<"0" | "7" | "30">("30");
  const [defaultLocale, setDefaultLocale] = useState<"fa-IR" | "en-US">("fa-IR");
  const [answerLanguage, setAnswerLanguage] = useState<"query" | "fa" | "en">("query");
  const [answerModel, setAnswerModel] = useState("");
  const [notice, setNotice] = useState<{ kind: "success" | "error"; message: string } | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [chatbots, bases, models] = await Promise.all([listChatbots(), listKnowledgeBases(), listAnswerModels()]);
      setItems(chatbots.items);
      setKnowledgeBases(bases.items);
      setAnswerModels(models);
      setAnswerModel((current) => models.includes(current) ? current : models[0] ?? "");
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.load")) });
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    let active = true;
    void Promise.all([listChatbots(), listKnowledgeBases(), listAnswerModels()])
      .then(([chatbots, bases, models]) => {
        if (!active) return;
        setItems(chatbots.items);
        setKnowledgeBases(bases.items);
        setAnswerModels(models);
        setAnswerModel(models[0] ?? "");
        setNotice(null);
      })
      .catch((error: unknown) => {
        if (active) setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.load")) });
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => { active = false; };
  }, [t]);

  const knowledgeBaseOptions = useMemo(() => knowledgeBases.map((item) => ({
    value: item.id,
    label: item.name,
    description: t("chatbots.kbReady", { ready: item.ready_document_count }),
    disabled: item.ready_document_count === 0,
  })), [knowledgeBases, t]);

  async function handleCreate(formData: FormData) {
    if (selectedKBs.length === 0 || !answerModel) {
      setNotice({ kind: "error", message: selectedKBs.length === 0 ? t("chatbots.selectKB") : t("chatbots.selectModel") });
      return;
    }
    setPending(true);
    const input: ChatbotInput = {
      name: String(formData.get("name") ?? "").trim(),
      description: String(formData.get("description") ?? "").trim(),
      system_instruction: String(formData.get("instruction") ?? "").trim(),
      answer_model: answerModel,
      tone: "professional",
      default_locale: defaultLocale,
      answer_language: answerLanguage,
      access_mode: accessMode,
      retention_days: Number(retention) as 0 | 7 | 30,
      knowledge_base_ids: selectedKBs,
      allowed_origins: String(formData.get("origins") ?? "").split("\n").map((value) => value.trim()).filter(Boolean),
      branding: {},
    };
    try {
      const created = await createChatbot(input);
      setItems((current) => [created, ...current]);
      setCreateOpen(false);
      setSelectedKBs([]);
      setAccessMode("private");
      setRetention("30");
      setDefaultLocale("fa-IR");
      setAnswerLanguage("query");
      setAnswerModel(answerModels[0] ?? "");
      setNotice({ kind: "success", message: t("chatbots.created") });
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.create")) });
    } finally {
      setPending(false);
    }
  }

  async function togglePublication(chatbot: Chatbot) {
    setPendingID(chatbot.id);
    try {
      if (chatbot.published_version_id) await unpublishChatbot(chatbot.id);
      else await publishChatbot(chatbot.id);
      await load();
      setNotice({ kind: "success", message: chatbot.published_version_id ? t("chatbots.unpublished") : t("chatbots.published") });
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.publish")) });
    } finally {
      setPendingID(null);
    }
  }

  async function showAnalytics(chatbot: Chatbot) {
    setPendingID(chatbot.id);
    try {
      const data = await getChatbotAnalytics(chatbot.id);
      setAnalytics({ chatbot, data });
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.analytics")) });
    } finally {
      setPendingID(null);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <OperationNotice notice={notice} labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }} />
      <AppCard
        title={t("chatbots.title")}
        description={t("chatbots.description")}
        actions={<div className="flex gap-2"><Button type="button" variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw aria-hidden="true" />{t("actions.refresh")}</Button><Button type="button" size="sm" onClick={() => setCreateOpen(true)} disabled={knowledgeBases.length === 0}><Plus aria-hidden="true" />{t("chatbots.create")}</Button></div>}
      >
        {loading && items.length === 0 ? <div className="flex min-h-40 items-center justify-center"><Spinner /></div> : items.length === 0 ? (
          <EmptyState icon={Bot} title={t("chatbots.empty.title")} description={t("chatbots.empty.description")} primaryAction={<Button type="button" onClick={() => setCreateOpen(true)} disabled={knowledgeBases.length === 0}>{t("chatbots.create")}</Button>} variant="inline" />
        ) : (
          <Table>
            <TableHeader><TableRow><TableHead>{t("fields.name")}</TableHead><TableHead>{t("fields.access")}</TableHead><TableHead>{t("chatbots.knowledgeBases")}</TableHead><TableHead>{t("fields.status")}</TableHead><TableHead>{t("fields.publicID")}</TableHead><TableHead>{t("actions.title")}</TableHead></TableRow></TableHeader>
            <TableBody>{items.map((item) => (
              <TableRow key={item.id}>
                <TableCell><span className="font-medium" dir="auto">{item.name}</span><span className="block text-xs text-muted-foreground" dir="ltr">{item.answer_model}</span></TableCell>
                <TableCell><span className="inline-flex items-center gap-2">{item.access_mode === "private" ? <LockKeyhole aria-hidden="true" className="size-4 text-muted-foreground" /> : <Globe2 aria-hidden="true" className="size-4 text-muted-foreground" />}{t(`chatbots.access.${item.access_mode}`)}</span></TableCell>
                <TableCell>{item.knowledge_base_ids.length}</TableCell>
                <TableCell><StatusBadge status={item.published_version_id ? "success" : "neutral"}>{item.published_version_id ? t("chatbots.status.published") : t("chatbots.status.draft")}</StatusBadge></TableCell>
                <TableCell className="min-w-56"><CopyableValue value={item.public_id} labels={{ valueAriaLabel: t("fields.publicID"), copy: t("copy.copy"), copying: t("copy.copying"), copied: t("copy.copied") }} /></TableCell>
                <TableCell><div className="flex flex-wrap gap-2"><ChatAction chatbot={item} label={t("chat.action")} /><Button type="button" variant="outline" size="sm" onClick={() => void showAnalytics(item)} disabled={pendingID === item.id}>{pendingID === item.id ? <Spinner /> : <BarChart3 aria-hidden="true" />}{t("analytics.action")}</Button><Button type="button" variant={item.published_version_id ? "outline" : "default"} size="sm" onClick={() => void togglePublication(item)} disabled={pendingID === item.id}>{pendingID === item.id ? <Spinner /> : <Rocket aria-hidden="true" />}{item.published_version_id ? t("chatbots.unpublish") : t("chatbots.publish")}</Button></div></TableCell>
              </TableRow>
            ))}</TableBody>
          </Table>
        )}
      </AppCard>

      {analytics ? (
        <AppCard title={t("analytics.title", { name: analytics.chatbot.name })} description={t("analytics.description")} actions={<Button type="button" variant="ghost" size="sm" onClick={() => setAnalytics(null)}>{t("analytics.close")}</Button>}>
          <dl className="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
            <Metric label={t("analytics.queries")} value={String(analytics.data.query_count)} />
            <Metric label={t("analytics.refusals")} value={String(analytics.data.refusal_count)} />
            <Metric label={t("analytics.citationFailures")} value={String(analytics.data.citation_failure_count)} />
            <Metric label={t("analytics.latency")} value={t("analytics.milliseconds", { value: Math.round(analytics.data.average_latency_ms) })} />
            <Metric label={t("analytics.cost")} value={analytics.data.provider_cost_usd.toFixed(4)} dir="ltr" />
            <Metric label={t("analytics.platformCost")} value={new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(analytics.data.platform_cost_irr)} dir="ltr" />
          </dl>
        </AppCard>
      ) : null}

      <FormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title={t("chatbots.create")}
        description={t("chatbots.createDescription")}
        pending={pending}
        size="lg"
        footer={<><Button type="button" variant="outline" onClick={() => setCreateOpen(false)} disabled={pending}>{t("actions.cancel")}</Button><Button type="submit" form="create-chatbot-form" disabled={pending || selectedKBs.length === 0 || !answerModel}>{pending ? <Spinner /> : null}{t("actions.create")}</Button></>}
      >
        <form id="create-chatbot-form" className="grid gap-4 md:grid-cols-2" action={(data) => void handleCreate(data)}>
          <AppTextField name="name" label={t("fields.name")} required maxLength={160} dir="auto" />
          <AppSelectField
            label={t("fields.answerModel")}
            value={answerModel}
            onValueChange={setAnswerModel}
            options={answerModels.map((model) => ({ value: model, label: model }))}
            placeholder={t("chatbots.noModels")}
            description={t("chatbots.modelHint")}
            disabled={answerModels.length === 0}
            dir="ltr"
            optionTextDir="ltr"
          />
          <AppTextareaField name="description" label={t("fields.description")} className="min-h-24" maxLength={2000} dir="auto" />
          <AppTextareaField name="instruction" label={t("fields.instruction")} className="min-h-24" maxLength={8000} description={t("chatbots.instructionHint")} dir="auto" />
          <AppSelectField label={t("fields.access")} value={accessMode} onValueChange={(value) => setAccessMode(value as "private" | "public")} options={[{ value: "private", label: t("chatbots.access.private") }, { value: "public", label: t("chatbots.access.public") }]} />
          <AppSelectField label={t("fields.retention")} value={retention} onValueChange={(value) => setRetention(value as "0" | "7" | "30")} options={[{ value: "30", label: t("chatbots.retention.30") }, { value: "7", label: t("chatbots.retention.7") }, { value: "0", label: t("chatbots.retention.0") }]} />
          <AppSelectField label={t("fields.defaultLocale")} value={defaultLocale} onValueChange={(value) => setDefaultLocale(value as "fa-IR" | "en-US")} options={[{ value: "fa-IR", label: t("chatbots.locales.fa-IR") }, { value: "en-US", label: t("chatbots.locales.en-US") }]} />
          <AppSelectField label={t("fields.answerLanguage")} value={answerLanguage} onValueChange={(value) => setAnswerLanguage(value as "query" | "fa" | "en")} options={[{ value: "query", label: t("chatbots.answerLanguages.query") }, { value: "fa", label: t("chatbots.answerLanguages.fa") }, { value: "en", label: t("chatbots.answerLanguages.en") }]} />
          <CheckboxOptionList className="md:col-span-2" label={t("chatbots.knowledgeBases")} description={t("chatbots.kbHint")} options={knowledgeBaseOptions} value={selectedKBs} onValueChange={(value) => setSelectedKBs(value.slice(0, 5))} emptyText={t("knowledgeBases.empty.title")} />
          <AppTextareaField name="origins" fieldClassName="md:col-span-2" label={t("fields.origins")} description={t("chatbots.originsHint")} dir="ltr" className="min-h-20" maxLength={102400} />
        </form>
      </FormDialog>
    </div>
  );
}

function ChatAction({ chatbot, label }: { chatbot: Chatbot; label: string }) {
  if (!chatbot.published_version_id) {
    return <Button type="button" variant="outline" size="sm" disabled><MessageCircle aria-hidden="true" />{label}</Button>;
  }

  return (
    <Button asChild variant="outline" size="sm">
      <Link href={`/chatbots/${encodeURIComponent(chatbot.id)}/chat`}><MessageCircle data-icon="inline-start" aria-hidden="true" />{label}</Link>
    </Button>
  );
}

function Metric({ label, value, dir }: { label: string; value: string; dir?: "ltr" | "rtl" }) {
  return <div className="rounded-lg border border-border bg-muted/30 p-3"><dt className="text-xs text-muted-foreground">{label}</dt><dd className="mt-1 text-lg font-semibold" dir={dir}>{value}</dd></div>;
}
