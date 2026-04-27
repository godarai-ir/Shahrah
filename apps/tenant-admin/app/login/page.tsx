import { redirect } from "next/navigation";

import LoginForm from "@/app/components/login-form";
import { getAuthContext } from "@/lib/auth/session";
import { createCaptchaChallenge } from "@/lib/auth/store";

export const dynamic = "force-dynamic";

type LoginPageProps = {
  searchParams?: Promise<{
    next?: string;
  }>;
};

function sanitizeNextPath(nextPath: string | undefined) {
  if (!nextPath || !nextPath.startsWith("/") || nextPath.startsWith("//")) {
    return "/";
  }
  return nextPath;
}

export default async function LoginPage({ searchParams }: LoginPageProps) {
  const auth = await getAuthContext();
  const resolvedSearchParams = (await searchParams) ?? {};
  const redirectTo = sanitizeNextPath(resolvedSearchParams.next);

  if (auth) {
    redirect(redirectTo);
  }

  const initialCaptcha = await createCaptchaChallenge();
  return <LoginForm initialCaptcha={initialCaptcha} redirectTo={redirectTo} />;
}
