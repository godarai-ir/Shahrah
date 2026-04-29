"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

type InviteAcceptPanelProps = {
  token: string;
};

type InviteAcceptState = "idle" | "loading" | "success" | "error";

export default function InviteAcceptPanel({ token }: InviteAcceptPanelProps) {
  const t = useTranslations("invite");
  const [state, setState] = useState<InviteAcceptState>("idle");
  const [message, setMessage] = useState<string>("");

  useEffect(() => {
    let cancelled = false;

    async function run() {
      setState("loading");
      try {
        const response = await fetch("/api/users/invitations/accept", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ token }),
          cache: "no-store",
        });
        const payload = (await response.json().catch(() => ({}))) as {
          error?: { message?: string };
        };
        if (!response.ok) {
          throw new Error(payload.error?.message || t("messages.acceptFailed"));
        }

        if (cancelled) return;
        setState("success");
        setMessage(t("messages.accepted"));
      } catch (error) {
        if (cancelled) return;
        setState("error");
        setMessage(error instanceof Error ? error.message : t("messages.acceptFailed"));
      }
    }

    void run();
    return () => {
      cancelled = true;
    };
  }, [token, t]);

  const statusText = useMemo(() => {
    if (state === "loading") return t("status.loading");
    if (state === "success") return t("status.success");
    if (state === "error") return t("status.error");
    return t("status.idle");
  }, [state, t]);

  return (
    <div className="mx-auto max-w-xl space-y-4 rounded-xl border bg-card p-6">
      <h1 className="text-2xl font-bold">{t("title")}</h1>
      <p className="text-sm text-muted-foreground">{statusText}</p>

      {state === "loading" ? null : (
        <Alert variant={state === "error" ? "destructive" : "default"}>
          <AlertTitle>{state === "success" ? t("alerts.successTitle") : t("alerts.errorTitle")}</AlertTitle>
          <AlertDescription>{message}</AlertDescription>
        </Alert>
      )}

      <div className="flex gap-2">
        <Button asChild variant="default">
          <Link href="/login">{t("actions.goToLogin")}</Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/">{t("actions.goToDashboard")}</Link>
        </Button>
      </div>
    </div>
  );
}

