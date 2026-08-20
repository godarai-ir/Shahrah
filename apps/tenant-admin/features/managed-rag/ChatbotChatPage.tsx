"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import {
  Bot,
  ChevronDown,
  CircleStop,
  ExternalLink,
  FileText,
  KeyRound,
  LockKeyhole,
  MessageCircle,
  RotateCcw,
  Send,
  UserRound,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { useEffect, useId, useMemo, useRef, useState } from "react";

import { CopyableValue } from "@/components/app-ui/CopyableValue";
import { resolveDisplayNameDirection } from "@/components/app-ui/DisplayNameField";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { MarkdownRenderer } from "@/components/app-ui/MarkdownRenderer";
import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { OneTimeSecretDialog } from "@/components/app-ui/OneTimeSecretDialog";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { InputGroup, InputGroupAddon, InputGroupTextarea } from "@/components/ui/input-group";
import { Message, MessageAvatar, MessageContent, MessageGroup, MessageHeader } from "@/components/ui/message";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Spinner } from "@/components/ui/spinner";
import { createChatbotSessionToken, getChatbot } from "@/features/managed-rag/api";
import type { ManagedRAGUIMessage } from "@/features/managed-rag/chat-types";
import type { Chatbot, Citation } from "@/lib/contracts/managed-rag";
import { getAPIErrorMessage } from "@/lib/http/api-client";
import { cn } from "@/lib/utils";

export default function ChatbotChatPage({ chatbotID }: { chatbotID: string }) {
  const t = useTranslations("managedRag");
  const [chatbot, setChatbot] = useState<Chatbot | null>(null);
  const [loading, setLoading] = useState(true);
  const [input, setInput] = useState("");
  const [stage, setStage] = useState<"retrieving" | "reranking" | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [sessionToken, setSessionToken] = useState<string | null>(null);
  const [tokenExpiresAt, setTokenExpiresAt] = useState<string | null>(null);
  const [issuingToken, setIssuingToken] = useState(false);
  const [tokenError, setTokenError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const reactID = useId();
  const chatID = `${chatbotID}-${reactID}`;
  const transport = useMemo(() => new DefaultChatTransport<ManagedRAGUIMessage>({
    api: `/api/managed-rag/chatbots/${encodeURIComponent(chatbotID)}/preview`,
  }), [chatbotID]);
  const {
    messages,
    sendMessage,
    status,
    error,
    stop,
    setMessages,
    clearError,
  } = useChat<ManagedRAGUIMessage>({
    id: chatID,
    transport,
    onData: (part) => {
      if (part.type === "data-status") setStage(part.data.stage);
    },
    onFinish: () => setStage(null),
    onError: () => setStage(null),
  });

  useEffect(() => {
    let active = true;
    void getChatbot(chatbotID)
      .then((item) => {
        if (active) setChatbot(item);
      })
      .catch((caught: unknown) => {
        if (active) setLoadError(getAPIErrorMessage(caught, t("errors.load")));
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => { active = false; };
  }, [chatbotID, t]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" });
  }, [messages, stage]);

  const busy = status === "submitted" || status === "streaming";
  const inputDirection = resolveDisplayNameDirection(input)
    ?? resolveDisplayNameDirection(t("chat.placeholder"))
    ?? "auto";

  async function submit() {
    const text = input.trim();
    if (!text || busy || !chatbot?.published_version_id) return;
    setInput("");
    clearError();
    await sendMessage({ text });
  }

  function resetConversation() {
    if (busy) stop();
    setMessages([]);
    clearError();
    setStage(null);
  }

  async function issueSessionToken() {
    setIssuingToken(true);
    setTokenError(null);

    try {
      const result = await createChatbotSessionToken(chatbotID);
      setSessionToken(result.token);
      setTokenExpiresAt(result.expires_at);
    } catch (caught: unknown) {
      setTokenError(getAPIErrorMessage(caught, t("chat.openWebUI.tokenError")));
    } finally {
      setIssuingToken(false);
    }
  }

  if (loading) {
    return <div className="flex min-h-96 items-center justify-center"><Spinner label={t("chat.loading")} /></div>;
  }

  if (loadError || !chatbot) {
    return <OperationNotice notice={{ kind: "error", message: loadError ?? t("errors.load") }} labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }} />;
  }

  const openAIBaseURL = gatewayBaseURL(chatbot.public_id);

  return (
    <>
      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_20rem]">
      <Card className="min-w-0 overflow-hidden">
        <CardHeader className="border-b border-border">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="flex items-center gap-3">
              <Avatar className="size-10"><AvatarFallback><Bot aria-hidden="true" className="size-5" /></AvatarFallback></Avatar>
              <CardTitle dir="auto">{chatbot.name}</CardTitle>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusBadge status={chatbot.access_mode === "private" ? "neutral" : "success"}>
                {chatbot.access_mode === "private" ? <LockKeyhole aria-hidden="true" /> : <ExternalLink aria-hidden="true" />}
                {t(`chatbots.access.${chatbot.access_mode}`)}
              </StatusBadge>
              <Button type="button" variant="outline" size="sm" onClick={resetConversation}>
                <RotateCcw data-icon="inline-start" aria-hidden="true" />{t("chat.reset")}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="p-0 md:p-0">
          {!chatbot.published_version_id ? (
            <Alert className="m-4">
              <AlertTitle>{t("chat.unpublished.title")}</AlertTitle>
              <AlertDescription>{t("chat.unpublished.description")}</AlertDescription>
            </Alert>
          ) : (
            <>
              <ScrollArea className="h-96 px-4 py-5 md:px-6" aria-label={t("chat.messagesLabel")}>
                {messages.length === 0 ? (
                  <EmptyState icon={MessageCircle} title={t("chat.empty.title")} description={t("chat.empty.description")} variant="inline" />
                ) : (
                  <MessageGroup>
                    {messages.map((message) => <ChatMessage key={message.id} message={message} />)}
                    {busy && stage ? <ChatStatus stage={stage} /> : null}
                    <div ref={bottomRef} />
                  </MessageGroup>
                )}
              </ScrollArea>
              <div className="border-t border-border p-4 md:p-5">
                {error ? <OperationNotice notice={{ kind: "error", message: t("errors.chat") }} labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }} /> : null}
                <InputGroup className="mt-3">
                  <InputGroupTextarea
                    value={input}
                    onChange={(event) => setInput(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" && !event.shiftKey) {
                        event.preventDefault();
                        void submit();
                      }
                    }}
                    placeholder={t("chat.placeholder")}
                    aria-label={t("chat.inputLabel")}
                    disabled={busy}
                    maxLength={8000}
                    className="min-h-20 text-start"
                    dir={inputDirection}
                  />
                  <InputGroupAddon align="block-end" className="justify-between border-t border-border">
                    <span className="text-xs text-muted-foreground">{t("chat.inputHint")}</span>
                    {busy ? (
                      <Button type="button" variant="outline" size="sm" onClick={stop}>
                        <CircleStop data-icon="inline-start" aria-hidden="true" />{t("chat.stop")}
                      </Button>
                    ) : (
                      <Button type="button" size="sm" onClick={() => void submit()} disabled={!input.trim()}>
                        <Send data-icon="inline-start" aria-hidden="true" />{t("chat.send")}
                      </Button>
                    )}
                  </InputGroupAddon>
                </InputGroup>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      <Card className="h-fit min-w-0">
        <CardHeader>
          <CardTitle>{t("chat.openWebUI.title")}</CardTitle>
          <CardDescription>{t("chat.openWebUI.description")}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <CopyableValue value={openAIBaseURL} labels={{ valueAriaLabel: t("chat.openWebUI.baseURL"), copy: t("copy.copy"), copying: t("copy.copying"), copied: t("copy.copied") }} />
          <dl className="flex flex-col gap-3 text-sm">
            <div><dt className="text-muted-foreground">{t("chat.openWebUI.model")}</dt><dd className="font-mono text-xs" dir="ltr">{chatbot.public_id}</dd></div>
            <div><dt className="text-muted-foreground">{t("fields.access")}</dt><dd>{t(`chatbots.access.${chatbot.access_mode}`)}</dd></div>
          </dl>
          <Alert>
            <AlertDescription>{chatbot.access_mode === "private" ? t("chat.openWebUI.privateHint") : t("chat.openWebUI.publicHint")}</AlertDescription>
          </Alert>
          {tokenError ? <OperationNotice notice={{ kind: "error", message: tokenError }} labels={{ success: t("notice.success"), error: t("notice.error"), warning: t("notice.warning"), info: t("notice.info") }} /> : null}
          {chatbot.access_mode === "private" ? (
            <Button type="button" variant="outline" onClick={() => void issueSessionToken()} disabled={issuingToken || !chatbot.published_version_id}>
              {issuingToken ? <Spinner /> : <KeyRound data-icon="inline-start" aria-hidden="true" />}
              {issuingToken ? t("chat.openWebUI.issuingToken") : t("chat.openWebUI.issueToken")}
            </Button>
          ) : null}
        </CardContent>
      </Card>
      </div>
      <OneTimeSecretDialog
        open={Boolean(sessionToken)}
        onOpenChange={(open) => { if (!open) setSessionToken(null); }}
        value={sessionToken}
        title={t("chat.openWebUI.tokenTitle")}
        description={t("chat.openWebUI.tokenDescription", {
          expiresAt: tokenExpiresAt ? new Date(tokenExpiresAt).toLocaleString() : "",
        })}
        warning={t("chat.openWebUI.tokenWarning")}
        closeLabel={t("chat.openWebUI.closeToken")}
        copyLabels={{
          valueAriaLabel: t("chat.openWebUI.tokenValue"),
          copy: t("copy.copy"),
          copying: t("copy.copying"),
          copied: t("copy.copied"),
        }}
        masked
        allowReveal
        onClear={() => {
          setSessionToken(null);
          setTokenExpiresAt(null);
        }}
      />
    </>
  );
}

function ChatMessage({ message }: { message: ManagedRAGUIMessage }) {
  const t = useTranslations("managedRag");
  const assistant = message.role === "assistant";
  const result = message.parts.find((part) => part.type === "data-result");
  const citations = result?.type === "data-result" ? result.data.citations : [];

  return (
    <Message align={assistant ? "start" : "end"}>
      <MessageAvatar>
        <Avatar className="size-8"><AvatarFallback>{assistant ? <Bot aria-hidden="true" /> : <UserRound aria-hidden="true" />}</AvatarFallback></Avatar>
      </MessageAvatar>
      <MessageContent className="max-w-3xl">
        <MessageHeader>{assistant ? t("chat.assistant") : t("chat.you")}</MessageHeader>
        <div className={cn(
          "rounded-lg px-4 py-3",
          assistant ? "border border-border bg-muted/40" : "bg-primary text-primary-foreground",
        )}>
          {message.parts.filter((part) => part.type === "text").map((part, index) => assistant ? (
            <MarkdownRenderer key={`${message.id}-${index}`} content={part.text} />
          ) : <p key={`${message.id}-${index}`} className="whitespace-pre-wrap leading-7" dir="auto">{part.text}</p>)}
        </div>
        {assistant && citations.length > 0 ? <Citations citations={citations} /> : null}
      </MessageContent>
    </Message>
  );
}

function ChatStatus({ stage }: { stage: "retrieving" | "reranking" }) {
  const t = useTranslations("managedRag");
  return (
    <div className="flex items-center gap-2 px-10 text-sm text-muted-foreground" role="status">
      <Spinner />{t(`chat.stage.${stage}`)}
    </div>
  );
}

function Citations({ citations }: { citations: Citation[] }) {
  const t = useTranslations("managedRag");
  return (
    <Collapsible>
      <CollapsibleTrigger asChild>
        <Button type="button" variant="ghost" size="sm">
          <FileText data-icon="inline-start" aria-hidden="true" />
          {t("chat.citations", { count: citations.length })}
          <ChevronDown data-icon="inline-end" aria-hidden="true" />
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="flex flex-col gap-2 pt-2">
        {citations.map((citation, index) => (
          <div key={citation.id} className="rounded-lg border border-border bg-card p-3 text-sm">
            <p className="font-medium" dir="auto">[{index + 1}] {citation.title}{citation.page != null ? ` — ${t("chat.page", { page: citation.page })}` : ""}</p>
            <p className="mt-1 text-muted-foreground" dir="auto">{citation.quote}</p>
          </div>
        ))}
      </CollapsibleContent>
    </Collapsible>
  );
}

function gatewayBaseURL(publicID: string) {
  if (!publicID || typeof window === "undefined") return "";
  const configured = process.env.NEXT_PUBLIC_GODARAI_API_BASE_URL?.replace(/\/$/, "");
  const fallback = `${window.location.protocol}//${window.location.hostname}:4000`;
  return `${configured || fallback}/v1/chatbots/${publicID}/openai`;
}
