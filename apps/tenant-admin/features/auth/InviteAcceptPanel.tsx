"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import {
  ArrowLeft,
  CheckCircle2,
  LockKeyhole,
  MailCheck,
  ShieldCheck,
  UserRoundCheck,
} from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";
import { GodarAILogo } from "@/components/app-ui/GodarAILogo";
import { acceptInvitation } from "@/features/auth/api";
import { cn } from "@/lib/utils";

type InviteAcceptPanelProps = {
  token?: string;
};

type InviteAcceptState = "idle" | "loading" | "success" | "error" | "invalid";

export default function InviteAcceptPanel({ token }: InviteAcceptPanelProps) {
  const t = useTranslations("invite");
  const [state, setState] = useState<InviteAcceptState>(token ? "idle" : "invalid");
  const [message, setMessage] = useState(token ? "" : t("messages.missingToken"));

  const content = useMemo(() => {
    if (state === "loading") {
      return { title: t("states.loadingTitle"), description: t("states.loadingDescription") };
    }
    if (state === "success") {
      return { title: t("states.successTitle"), description: t("states.successDescription") };
    }
    if (state === "error" || state === "invalid") {
      return { title: t("states.errorTitle"), description: t("states.errorDescription") };
    }
    return { title: t("states.idleTitle"), description: t("states.idleDescription") };
  }, [state, t]);

  async function acceptInvite() {
    if (!token || state === "loading") return;

    setState("loading");
    setMessage("");

    try {
      await acceptInvitation(token);

      setState("success");
      setMessage(t("messages.accepted"));
    } catch (error) {
      setState("error");
      const errorMessage = error instanceof Error ? error.message.toLowerCase() : "";
      setMessage(errorMessage.includes("already used") ? t("messages.alreadyUsed") : t("messages.acceptFailed"));
    }
  }

  const isLoading = state === "loading";
  const isComplete = state === "success";
  const hasError = state === "error" || state === "invalid";

  return (
    <main className="relative isolate flex min-h-screen items-center justify-center overflow-hidden bg-background px-4 py-8 sm:px-6 lg:px-8">
      <div aria-hidden="true" className="absolute inset-0 bg-gradient-to-br from-primary/10 via-background to-accent/40" />
      <div aria-hidden="true" className="absolute -start-24 top-16 size-72 rounded-full bg-primary/10 blur-3xl" />
      <div aria-hidden="true" className="absolute -end-24 bottom-10 size-80 rounded-full bg-accent/60 blur-3xl" />

      <div className="relative grid w-full max-w-5xl overflow-hidden rounded-3xl border border-card-border bg-card shadow-2xl lg:grid-cols-5">
        <aside className="relative hidden flex-col justify-between overflow-hidden bg-primary p-8 text-primary-foreground lg:col-span-2 lg:flex">
          <div aria-hidden="true" className="absolute inset-0 bg-gradient-to-b from-primary-foreground/10 to-transparent" />
          <div className="relative flex flex-col gap-8">
            <div className="flex items-center gap-3">
              <span className="flex size-12 items-center justify-center rounded-xl bg-primary-foreground shadow-sm">
                <GodarAILogo
                  decorative
                  variant="light"
                  priority
                  className="h-9 w-auto max-w-full object-contain"
                  sizes="2.25rem"
                />
              </span>
              <div>
                <p className="text-lg font-bold">{t("brand.name")}</p>
                <p className="text-xs text-primary-foreground/75">{t("brand.tagline")}</p>
              </div>
            </div>

            <div className="flex flex-col gap-4">
              <Badge className="w-fit border-primary-foreground/25 bg-primary-foreground/10 text-primary-foreground">
                {t("eyebrow")}
              </Badge>
              <h2 className="text-2xl font-bold leading-relaxed">{t("brand.heading")}</h2>
              <p className="text-sm leading-7 text-primary-foreground/80">{t("brand.description")}</p>
            </div>
          </div>

          <div className="relative flex items-center gap-3 text-xs text-primary-foreground/75">
            <ShieldCheck aria-hidden="true" />
            <span>{t("brand.security")}</span>
          </div>
        </aside>

        <section className="flex flex-col p-4 sm:p-8 lg:col-span-3 lg:p-10">
          <div className="mb-8 flex items-center gap-3 lg:hidden">
            <span className="flex size-11 items-center justify-center rounded-xl border border-primary/15 bg-primary/5">
              <GodarAILogo
                decorative
                priority
                className="h-8 w-auto max-w-full object-contain"
                sizes="2rem"
              />
            </span>
            <div>
              <p className="font-bold">{t("brand.name")}</p>
              <p className="text-xs text-muted-foreground">{t("brand.tagline")}</p>
            </div>
          </div>

          <Card className="flex flex-1 flex-col border-0 shadow-none hover:transform-none hover:shadow-none">
            <CardHeader className="items-center px-0 pb-6 text-center md:px-0">
              <div
                className={cn(
                  "flex size-16 items-center justify-center rounded-2xl border",
                  hasError ? "tone-danger" : isComplete ? "tone-success" : "surface-primary-soft",
                )}
              >
                {isLoading ? <Spinner size="lg" /> : isComplete ? <CheckCircle2 aria-hidden="true" /> : hasError ? <LockKeyhole aria-hidden="true" /> : <MailCheck aria-hidden="true" />}
              </div>
              <div className="flex flex-col gap-2">
                <CardTitle className="text-2xl md:text-3xl">{content.title}</CardTitle>
                <CardDescription className="mx-auto max-w-md leading-7">{content.description}</CardDescription>
              </div>
            </CardHeader>

            <CardContent className="flex flex-1 flex-col gap-6 px-0 md:px-0">
              {state === "idle" ? (
                <div className="rounded-xl border border-border bg-muted/40 p-4">
                  <div className="flex items-start gap-3">
                    <UserRoundCheck className="mt-0.5 shrink-0 text-primary" aria-hidden="true" />
                    <div className="flex flex-col gap-1">
                      <p className="text-sm font-semibold">{t("summary.title")}</p>
                      <p className="text-sm leading-6 text-muted-foreground">{t("summary.description")}</p>
                    </div>
                  </div>
                </div>
              ) : null}

              {isLoading ? (
                <div className="flex items-center justify-center gap-3 rounded-xl border border-border bg-muted/40 p-4 text-sm text-muted-foreground" role="status" aria-live="polite">
                  <Spinner className="text-primary" />
                  <span>{t("status.loading")}</span>
                </div>
              ) : null}

              {message ? (
                <Alert variant={hasError ? "destructive" : "success"} aria-live="polite">
                  <AlertTitle>{hasError ? t("alerts.errorTitle") : t("alerts.successTitle")}</AlertTitle>
                  <AlertDescription>{message}</AlertDescription>
                </Alert>
              ) : null}

              <Separator />
              <div className="flex items-start gap-3 text-xs leading-6 text-muted-foreground">
                <ShieldCheck className="mt-0.5 shrink-0 text-primary" aria-hidden="true" />
                <p>{t("privacyNote")}</p>
              </div>
            </CardContent>

            <CardFooter className="flex-col px-0 pb-0 pt-6 md:px-0 md:pb-0">
              {state === "idle" ? (
                <Button size="lg" className="w-full" onClick={acceptInvite}>
                  <UserRoundCheck data-icon="inline-start" />
                  {t("actions.accept")}
                </Button>
              ) : null}

              {isLoading ? (
                <Button size="lg" className="w-full" disabled>
                  <Spinner data-icon="inline-start" />
                  {t("actions.accepting")}
                </Button>
              ) : null}

              {isComplete ? (
                <Button asChild size="lg" className="w-full">
                  <Link href="/login">
                    {t("actions.continue")}
                    <ArrowLeft data-icon="inline-end" />
                  </Link>
                </Button>
              ) : null}

              {hasError ? (
                <div className="flex w-full flex-col gap-3 sm:flex-row">
                  {token ? <Button size="lg" className="flex-1" onClick={acceptInvite}>{t("actions.retry")}</Button> : null}
                  <Button asChild size="lg" variant={token ? "outline" : "default"} className="flex-1">
                    <Link href="/login">{t("actions.goToLogin")}</Link>
                  </Button>
                </div>
              ) : null}
            </CardFooter>
          </Card>
        </section>
      </div>
    </main>
  );
}
