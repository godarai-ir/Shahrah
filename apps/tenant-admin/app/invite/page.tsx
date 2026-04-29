import { getTranslations } from "next-intl/server";

import InviteAcceptPanel from "@/app/components/invite-accept-panel";

type InvitePageProps = {
  searchParams?: Promise<{
    token?: string;
  }>;
};

export const dynamic = "force-dynamic";

export default async function InvitePage({ searchParams }: InvitePageProps) {
  const t = await getTranslations("invite");
  const params = (await searchParams) ?? {};
  const token = typeof params.token === "string" ? params.token.trim() : "";

  if (!token) {
    return (
      <div className="mx-auto max-w-xl rounded-xl border bg-card p-6">
        <h1 className="text-2xl font-bold">{t("title")}</h1>
        <p className="mt-2 text-sm text-destructive">{t("messages.missingToken")}</p>
      </div>
    );
  }

  return <InviteAcceptPanel token={token} />;
}
