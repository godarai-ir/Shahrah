"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { FilePlus2, FolderOpen, Plus, Radio, Upload, WifiOff } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { AppTextField, AppTextareaField } from "@/components/app-ui/AppTextField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { RefreshButton } from "@/components/app-ui/RefreshButton";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  createKnowledgeBase,
  createTextDocument,
  listDocuments,
  listKnowledgeBases,
  uploadDocument,
} from "@/features/managed-rag/api";
import {
  DocumentIngestionStatus,
  processingDocumentStatuses,
} from "@/features/managed-rag/DocumentIngestionStatus";
import type { KnowledgeBase, KnowledgeBaseDocument } from "@/lib/contracts/managed-rag";
import { getAPIErrorMessage } from "@/lib/http/api-client";

export default function KnowledgeBasesPage() {
  const t = useTranslations("managedRag");
  const locale = useLocale();
  const [items, setItems] = useState<KnowledgeBase[]>([]);
  const [documents, setDocuments] = useState<KnowledgeBaseDocument[]>([]);
  const [selected, setSelected] = useState<KnowledgeBase | null>(null);
  const [loading, setLoading] = useState(true);
  const [documentsLoading, setDocumentsLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [documentOpen, setDocumentOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [notice, setNotice] = useState<{ kind: "error" | "success"; message: string } | null>(null);
  const [liveState, setLiveState] = useState<"live" | "refreshing" | "paused" | "error">("live");
  const [lastLiveUpdated, setLastLiveUpdated] = useState<Date | null>(null);
  const [changedDocumentIDs, setChangedDocumentIDs] = useState<Set<string>>(new Set());
  const documentsRef = useRef<KnowledgeBaseDocument[]>([]);
  const transitionTimersRef = useRef(new Map<string, ReturnType<typeof setTimeout>>());
  const documentRequestsRef = useRef(new Set<string>());
  const selectedIDRef = useRef<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const result = await listKnowledgeBases();
      setItems(result.items);
      setNotice(null);
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.load")) });
    } finally {
      setLoading(false);
    }
  }, [t]);

  const applyDocuments = useCallback((nextDocuments: KnowledgeBaseDocument[], animateChanges: boolean) => {
    const previousStatuses = new Map(documentsRef.current.map((document) => [document.id, document.status]));
    const transitioned = animateChanges
      ? nextDocuments.filter((document) => {
          const previous = previousStatuses.get(document.id);
          return previous !== undefined && previous !== document.status;
        }).map((document) => document.id)
      : [];

    documentsRef.current = nextDocuments;
    setDocuments(nextDocuments);

    if (transitioned.length === 0) return;

    setChangedDocumentIDs((current) => new Set([...current, ...transitioned]));
    for (const documentID of transitioned) {
      const existingTimer = transitionTimersRef.current.get(documentID);
      if (existingTimer) clearTimeout(existingTimer);

      const timer = setTimeout(() => {
        transitionTimersRef.current.delete(documentID);
        setChangedDocumentIDs((current) => {
          const next = new Set(current);
          next.delete(documentID);
          return next;
        });
      }, 1400);
      transitionTimersRef.current.set(documentID, timer);
    }
  }, []);

  const refreshSelectedDocuments = useCallback(async (
    knowledgeBaseID: string,
    options: { initial?: boolean; background?: boolean } = {},
  ) => {
    if (documentRequestsRef.current.has(knowledgeBaseID)) return;
    documentRequestsRef.current.add(knowledgeBaseID);

    if (options.initial) setDocumentsLoading(true);
    if (options.background) setLiveState("refreshing");

    try {
      const result = await listDocuments(knowledgeBaseID);
      if (selectedIDRef.current !== knowledgeBaseID) return;

      applyDocuments(result.items, !options.initial);
      const readyCount = result.items.filter((document) => document.status === "ready").length;
      setItems((current) => current.map((item) => item.id === knowledgeBaseID
        ? { ...item, document_count: result.total, ready_document_count: readyCount }
        : item));
      setSelected((current) => current?.id === knowledgeBaseID
        ? { ...current, document_count: result.total, ready_document_count: readyCount }
        : current);
      setLastLiveUpdated(new Date());
      setLiveState("live");
    } catch (error) {
      if (selectedIDRef.current !== knowledgeBaseID) return;

      if (options.background) {
        setLiveState("error");
      } else {
        setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.documents")) });
      }
    } finally {
      documentRequestsRef.current.delete(knowledgeBaseID);
      if (options.initial && selectedIDRef.current === knowledgeBaseID) setDocumentsLoading(false);
    }
  }, [applyDocuments, t]);

  const loadSelectedDocuments = useCallback(async (knowledgeBase: KnowledgeBase) => {
    selectedIDRef.current = knowledgeBase.id;
    setSelected(knowledgeBase);
    applyDocuments([], false);
    await refreshSelectedDocuments(knowledgeBase.id, { initial: true });
  }, [applyDocuments, refreshSelectedDocuments]);

  useEffect(() => {
    let active = true;
    void listKnowledgeBases()
      .then((result) => {
        if (!active) return;
        setItems(result.items);
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

  useEffect(() => () => {
    for (const timer of transitionTimersRef.current.values()) clearTimeout(timer);
    transitionTimersRef.current.clear();
  }, []);

  const hasProcessingDocuments = documents.some((document) => processingDocumentStatuses.has(document.status));
  const selectedID = selected?.id;

  useEffect(() => {
    if (!selectedID) return;

    let stopped = false;
    let pollTimer: ReturnType<typeof setTimeout> | undefined;

    const schedule = (delay: number) => {
      if (pollTimer) clearTimeout(pollTimer);
      pollTimer = setTimeout(() => void poll(), delay);
    };

    const poll = async () => {
      if (stopped) return;
      if (document.visibilityState === "hidden") {
        setLiveState("paused");
        schedule(10_000);
        return;
      }

      await refreshSelectedDocuments(selectedID, { background: true });
      if (!stopped) {
        const hasActiveWork = documentsRef.current.some((document) => processingDocumentStatuses.has(document.status));
        schedule(hasActiveWork ? 1_500 : 10_000);
      }
    };

    const handleVisibility = () => {
      if (document.visibilityState === "hidden") {
        setLiveState("paused");
        return;
      }

      schedule(0);
    };

    document.addEventListener("visibilitychange", handleVisibility);
    schedule(hasProcessingDocuments ? 1_500 : 10_000);

    return () => {
      stopped = true;
      if (pollTimer) clearTimeout(pollTimer);
      document.removeEventListener("visibilitychange", handleVisibility);
    };
  }, [hasProcessingDocuments, refreshSelectedDocuments, selectedID]);

  async function handleCreate(formData: FormData) {
    setPending(true);
    try {
      const created = await createKnowledgeBase({
        name: String(formData.get("name") ?? "").trim(),
        description: String(formData.get("description") ?? "").trim(),
      });
      setItems((current) => [created, ...current]);
      setCreateOpen(false);
      setNotice({ kind: "success", message: t("knowledgeBases.created") });
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.create")) });
    } finally {
      setPending(false);
    }
  }

  async function handleDocument(formData: FormData) {
    if (!selected) return;
    setPending(true);
    try {
      const file = formData.get("file");
      const created = file instanceof File && file.size > 0
        ? await uploadDocument(selected.id, file)
        : await createTextDocument(selected.id, {
            title: String(formData.get("title") ?? "").trim(),
            text: String(formData.get("text") ?? "").trim(),
          });
      applyDocuments([created, ...documentsRef.current.filter((document) => document.id !== created.id)], false);
      setDocumentOpen(false);
      setNotice({ kind: "success", message: t("documents.queued") });
      await load();
    } catch (error) {
      setNotice({ kind: "error", message: getAPIErrorMessage(error, t("errors.upload")) });
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <OperationNotice
        notice={notice}
        labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }}
      />
      <AppCard
        title={t("knowledgeBases.title")}
        description={t("knowledgeBases.description")}
        actions={(
          <div className="flex gap-2">
            <RefreshButton label={t("actions.refresh")} loadingLabel={t("actions.refreshing")} onClick={() => void load()} isLoading={loading} />
            <Button type="button" size="sm" onClick={() => setCreateOpen(true)}>
              <Plus aria-hidden="true" />{t("knowledgeBases.create")}
            </Button>
          </div>
        )}
      >
        {loading && items.length === 0 ? (
          <div className="flex min-h-40 items-center justify-center"><Spinner /></div>
        ) : items.length === 0 ? (
          <EmptyState
            icon={FolderOpen}
            title={t("knowledgeBases.empty.title")}
            description={t("knowledgeBases.empty.description")}
            primaryAction={<Button type="button" onClick={() => setCreateOpen(true)}>{t("knowledgeBases.create")}</Button>}
            variant="inline"
          />
        ) : (
          <Table>
            <TableHeader><TableRow>
              <TableHead>{t("fields.name")}</TableHead>
              <TableHead>{t("knowledgeBases.documents")}</TableHead>
              <TableHead>{t("knowledgeBases.ready")}</TableHead>
              <TableHead>{t("knowledgeBases.profile")}</TableHead>
              <TableHead>{t("actions.title")}</TableHead>
            </TableRow></TableHeader>
            <TableBody>{items.map((item) => (
              <TableRow key={item.id} data-state={selected?.id === item.id ? "selected" : undefined}>
                <TableCell><span className="font-medium text-foreground" dir="auto">{item.name}</span><span className="block text-xs text-muted-foreground" dir="auto">{item.description}</span></TableCell>
                <TableCell>{item.document_count}</TableCell>
                <TableCell>{item.ready_document_count}</TableCell>
                <TableCell><StatusBadge status="neutral" dir="ltr">{item.rag_profile_revision}</StatusBadge></TableCell>
                <TableCell><Button type="button" variant="outline" size="sm" onClick={() => void loadSelectedDocuments(item)}>{t("documents.view")}</Button></TableCell>
              </TableRow>
            ))}</TableBody>
          </Table>
        )}
      </AppCard>

      {selected ? (
        <AppCard
          title={t("documents.title", { name: selected.name })}
          description={t("documents.description")}
          actions={<Button type="button" size="sm" onClick={() => setDocumentOpen(true)}><FilePlus2 aria-hidden="true" />{t("documents.add")}</Button>}
        >
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-lg bg-muted/40 p-3" aria-live="polite">
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={liveState === "error" ? "warning" : liveState === "paused" ? "neutral" : "active"}>
                {liveState === "error" || liveState === "paused"
                  ? <WifiOff data-icon="inline-start" aria-hidden="true" />
                  : <Radio data-icon="inline-start" aria-hidden="true" className={liveState === "refreshing" ? "motion-safe:animate-pulse" : undefined} />}
                {t(`documents.live.${liveState}`)}
              </StatusBadge>
              {lastLiveUpdated ? (
                <span className="text-xs text-muted-foreground">
                  {t("documents.live.lastUpdated")}{" "}
                  <DateTimeValue value={lastLiveUpdated} locale={locale} format={{ timeStyle: "medium" }} />
                </span>
              ) : null}
            </div>
            <RefreshButton
              label={t("documents.live.refresh")}
              loadingLabel={t("documents.live.refreshingAction")}
              isLoading={liveState === "refreshing"}
              onClick={() => void refreshSelectedDocuments(selected.id, { background: true })}
            />
          </div>
          {documentsLoading ? <div className="flex min-h-28 items-center justify-center"><Spinner /></div> : documents.length === 0 ? (
            <EmptyState icon={Upload} title={t("documents.empty.title")} description={t("documents.empty.description")} variant="inline" />
          ) : (
            <Table>
              <TableHeader><TableRow><TableHead>{t("fields.title")}</TableHead><TableHead>{t("fields.type")}</TableHead><TableHead>{t("fields.language")}</TableHead><TableHead>{t("fields.status")}</TableHead><TableHead>{t("documents.chunks")}</TableHead></TableRow></TableHeader>
              <TableBody>{documents.map((document) => (
                <TableRow key={document.id}>
                  <TableCell className="font-medium" dir="auto">{document.title}</TableCell>
                  <TableCell>{t(`documents.source.${document.source_type}`)}</TableCell>
                  <TableCell><span dir="ltr">{document.detected_language || "—"}</span></TableCell>
                  <TableCell>
                    <DocumentIngestionStatus
                      status={document.status}
                      label={t(`documents.status.${document.status}`)}
                      changed={changedDocumentIDs.has(document.id)}
                    />
                  </TableCell>
                  <TableCell>{document.chunk_count}</TableCell>
                </TableRow>
              ))}</TableBody>
            </Table>
          )}
        </AppCard>
      ) : null}

      <FormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title={t("knowledgeBases.create")}
        description={t("knowledgeBases.createDescription")}
        pending={pending}
        footer={<><Button type="button" variant="outline" onClick={() => setCreateOpen(false)} disabled={pending}>{t("actions.cancel")}</Button><Button type="submit" form="create-kb-form" disabled={pending}>{pending ? <Spinner /> : null}{t("actions.create")}</Button></>}
      >
        <form id="create-kb-form" className="flex flex-col gap-4" action={(data) => void handleCreate(data)}>
          <AppTextField name="name" label={t("fields.name")} required maxLength={160} dir="auto" />
          <AppTextareaField name="description" label={t("fields.description")} maxLength={2000} dir="auto" />
        </form>
      </FormDialog>

      <FormDialog
        open={documentOpen}
        onOpenChange={setDocumentOpen}
        title={t("documents.add")}
        description={t("documents.addDescription")}
        pending={pending}
        size="md"
        footer={<><Button type="button" variant="outline" onClick={() => setDocumentOpen(false)} disabled={pending}>{t("actions.cancel")}</Button><Button type="submit" form="add-document-form" disabled={pending}>{pending ? <Spinner /> : null}{t("documents.submit")}</Button></>}
      >
        <form id="add-document-form" className="flex flex-col gap-4" action={(data) => void handleDocument(data)}>
          <AppTextField name="title" label={t("fields.title")} description={t("documents.textHint")} maxLength={255} dir="auto" />
          <AppTextareaField name="text" label={t("fields.text")} className="min-h-32" dir="auto" />
          <div className="flex flex-col gap-2">
            <label htmlFor="rag-document-file" className="text-sm font-medium text-foreground">{t("fields.file")}</label>
            <Input id="rag-document-file" name="file" type="file" accept=".pdf,.docx,.txt,.html,.htm" />
            <p className="text-xs text-muted-foreground">{t("documents.fileHint")}</p>
          </div>
        </form>
      </FormDialog>
    </div>
  );
}
